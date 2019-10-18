// Copyright © 2019 Michael Gruener
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vault

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/vault/api"
	vaultapi "github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gopkg.in/fsnotify.v1"
)

const (
	initialTokenTimeout = 10 * time.Second
)

type clientOptions struct {
	role     string
	authPath string
}

// ClientOption configures a Vault client using the functional options paradigm popularized by Rob Pike and Dave Cheney.
// If you're unfamiliar with this style,
// see https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html and
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis.
type ClientOption interface {
	apply(o *clientOptions)
}

// ClientRole is the vault role which the client would like to receive
type ClientRole string

func (co ClientRole) apply(o *clientOptions) {
	o.role = string(co)
}

// ClientAuthPath is the mount path where the auth method is enabled.
type ClientAuthPath string

func (co ClientAuthPath) apply(o *clientOptions) {
	o.authPath = string(co)
}

// Client manages the connection to the vault, especially refreshing of the
// auth token
type Client struct {
	client       *vaultapi.Client
	logical      *vaultapi.Logical
	tokenRenewer *vaultapi.Renewer
	closed       bool
	watch        *fsnotify.Watcher
	mu           sync.Mutex
	approle      *Approle
}

// NewClient creates a new Vault client.
func NewClient() (*Client, error) {
	return NewClientWithOptions()
}

// NewClientWithOptions creates a new Vault client with custom options.
func NewClientWithOptions(opts ...ClientOption) (*Client, error) {
	return NewClientFromConfig(vaultapi.DefaultConfig(), opts...)
}

// NewClientFromConfig creates a new Vault client from custom configuration.
func NewClientFromConfig(config *vaultapi.Config, opts ...ClientOption) (*Client, error) {
	rawClient, err := vaultapi.NewClient(config)
	if err != nil {
		return nil, err
	}

	client, err := NewClientFromRawClient(rawClient, opts...)
	if err != nil {
		return nil, err
	}

	caCertPath := os.Getenv(vaultapi.EnvVaultCACert)
	caCertReload := os.Getenv("VAULT_CACERT_RELOAD") != "false"

	if caCertPath != "" && caCertReload {
		watch, err := fsnotify.NewWatcher()
		if err != nil {
			return nil, err
		}

		caCertFile := filepath.Clean(caCertPath)
		configDir, _ := filepath.Split(caCertFile)

		_ = watch.Add(configDir)

		go func() {
			for {
				client.mu.Lock()
				if client.closed {
					client.mu.Unlock()
					break
				}
				client.mu.Unlock()

				select {
				case event := <-watch.Events:
					// we only care about the CA cert file or the Secret mount directory (if in Kubernetes)
					if filepath.Clean(event.Name) == caCertFile || filepath.Base(event.Name) == "..data" {
						if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
							err := config.ReadEnvironment()
							if err != nil {
								log.Error("failed to reload Vault config:", err)
							} else {
								log.Info("CA certificate reloaded")
							}
						}
					}
				case err := <-watch.Errors:
					log.Error("watcher error:", err)
				}
			}
		}()

		client.watch = watch
	}

	return client, nil
}

// NewClientFromRawClient creates a new Vault client from custom raw client.
func NewClientFromRawClient(rawClient *vaultapi.Client, opts ...ClientOption) (*Client, error) {
	logical := rawClient.Logical()
	var tokenRenewer *vaultapi.Renewer

	o := &clientOptions{}

	for _, opt := range opts {
		opt.apply(o)
	}

	// Default role
	if o.role == "" {
		o.role = viper.GetString("vault-role")
	}

	// Default auth path
	if o.authPath == "" {
		o.authPath = viper.GetString("vault-auth-path")
	}

	approle, err := NewApprole()
	if err != nil {
		return nil, fmt.Errorf("failed to create approle: %s", err)
	}

	client := &Client{client: rawClient, logical: logical, approle: approle}

	initialTokenArrived := make(chan string, 1)
	initialTokenSent := false

	go func() {
		for {
			client.mu.Lock()
			if client.closed {
				client.mu.Unlock()
				break
			}
			// Login must be done while locked as the Login method
			// changes the internal state of the approle instance
			// by creating and storing a new secret id
			secret, err := approle.Login(rawClient, o.authPath, o.role)
			client.mu.Unlock()

			if err != nil {
				log.Info("Failed to request new Vault token", err.Error())
				time.Sleep(1 * time.Second)
				continue
			}

			if secret == nil {
				log.Info("Received empty answer from Vault, retrying")
				time.Sleep(1 * time.Second)
				continue
			}

			log.Println("Received new Vault token")

			if !initialTokenSent {
				initialTokenArrived <- secret.LeaseID
				initialTokenSent = true
			}

			// Start the renewing process
			tokenRenewer, err = rawClient.NewRenewer(&vaultapi.RenewerInput{Secret: secret})
			if err != nil {
				log.Info("Failed to renew Vault token", err.Error())
				continue
			}

			client.mu.Lock()
			client.tokenRenewer = tokenRenewer
			client.mu.Unlock()

			go tokenRenewer.Renew()

			runRenewChecker(tokenRenewer)
		}
		log.Info("Vault token renewal closed")
	}()

	select {
	case <-initialTokenArrived:
		log.Info("Initial Vault token arrived")

	case <-time.After(initialTokenTimeout):
		client.Close()
		return nil, fmt.Errorf("timeout [%s] during waiting for Vault token", initialTokenTimeout)
	}

	return client, nil
}

func runRenewChecker(tokenRenewer *vaultapi.Renewer) {
	for {
		select {
		case err := <-tokenRenewer.DoneCh():
			if err != nil {
				log.Error("Vault token renewal error:", err.Error())
			}
			return
		case <-tokenRenewer.RenewCh():
			log.Info("Renewed Vault Token")
		}
	}
}

// RawClient returns the underlying raw Vault client.
func (client *Client) RawClient() *vaultapi.Client {
	return client.client
}

// Close stops the token renewing process of this client
func (client *Client) Close() {
	client.mu.Lock()
	defer client.mu.Unlock()

	if client.tokenRenewer != nil {
		client.closed = true
		client.tokenRenewer.Stop()
	}

	if client.watch != nil {
		_ = client.watch.Close()
	}
}

// NewRawClient creates a new raw Vault client.
func NewRawClient() (*api.Client, error) {
	config := vaultapi.DefaultConfig()
	if config.Error != nil {
		return nil, config.Error
	}

	config.HttpClient.Transport.(*http.Transport).TLSHandshakeTimeout = 5 * time.Second

	return vaultapi.NewClient(config)
}

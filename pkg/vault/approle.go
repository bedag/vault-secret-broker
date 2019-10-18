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
	"io/ioutil"

	vaultapi "github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Approle implements an opinionated Vault AppRole authentication
// based on single use SecredIDs
// See https://www.vaultproject.io/api/auth/approle/index.html
type Approle struct {
	roleID   string
	secretID string
	// the auth token retrieved by authenticating with the roleID/secretID
	token string
	// the persistent storage path for the SecretID
	// new SecretIDs generated during the auth refresh process will be
	// stored here
	secretIDStorePath string

	// true if the current SecredID has been persisted to disk
	// false if not
	persisted bool
}

// options to initialize a new approle authenticator
type approleOptions struct {
	roleID              string
	roleIDPath          string
	initialSecretID     string
	initialSecretIDPath string
	secretIDStorePath   string
}

// ApproleOption configures a Vault client using the functional options paradigm popularized by Rob Pike and Dave Cheney.
// If you're unfamiliar with this style,
// see https://commandcenter.blogspot.com/2014/01/self-referential-functions-and-design.html and
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis.
type ApproleOption interface {
	apply(o *approleOptions)
}

// ApproleRoleID is the RoleID to be used for the AppRole authentication
type ApproleRoleID string

func (co ApproleRoleID) apply(o *approleOptions) {
	o.roleID = string(co)
}

// ApproleRoleIDPath is the path to initially read the AppRole RoleID from (optional)
type ApproleRoleIDPath string

func (co ApproleRoleIDPath) apply(o *approleOptions) {
	o.roleIDPath = string(co)
}

// ApproleInitialSecretID is the initial AppRole SecretID
type ApproleInitialSecretID string

func (co ApproleInitialSecretID) apply(o *approleOptions) {
	o.initialSecretID = string(co)
}

// ApproleInitialSecretIDPath is the path to initially read the Approle SecretID from
type ApproleInitialSecretIDPath string

func (co ApproleInitialSecretIDPath) apply(o *approleOptions) {
	o.initialSecretIDPath = string(co)
}

// ApproleSecretIDStorePath is the path where the AppRole SecretID will be stored after initialization
type ApproleSecretIDStorePath string

func (co ApproleSecretIDStorePath) apply(o *approleOptions) {
	o.initialSecretIDPath = string(co)
}

// NewApprole creates a new AppRole.
func NewApprole() (*Approle, error) {
	return NewApproleWithOptions()
}

// NewApproleWithOptions creates a new AppRole with custom options.
func NewApproleWithOptions(opts ...ApproleOption) (*Approle, error) {
	o := &approleOptions{}

	for _, opt := range opts {
		opt.apply(o)
	}

	// To proceed with the authentication we need a RoleID and an SecretID which
	// we can either get explicitely as method parameter, as a cli option, from the
	// environment or stored at different file path
	// Precedence:
	//   RoleID:   method parameter > cli option > environment variable > file
	//   SecretID: method parameter > store file > cli option > environment variable > initial file

	// get the paths for retrieving the RoleID and the initial SecretID from files
	if o.roleIDPath == "" {
		o.roleIDPath = viper.GetString("vault-approle-roleid-path")
	}

	if o.initialSecretIDPath == "" {
		o.initialSecretIDPath = viper.GetString("vault-approle-initial-secretid-path")
	}

	if o.secretIDStorePath == "" {
		o.secretIDStorePath = viper.GetString("vault-approle-secretid-store-path")
	}

	// Get the RoleID ...
	// ... from cli parameters or the environment
	if o.roleID == "" {
		o.roleID = viper.GetString("vault-approle-roleid")
	}

	// ... from the RoleID path
	if o.roleID == "" {
		roleID, err := ioutil.ReadFile(o.roleIDPath)
		if err != nil {
			return nil, err
		}
		o.roleID = string(roleID)
	}

	// still no valid RoleID? Fail!
	if o.roleID == "" {
		return nil, fmt.Errorf("failed to retrieve the AppRole RoleID")
	}

	// Get the initial SecretID
	// ... from the persistent SecretID store
	if o.initialSecretID == "" {
		secretID, err := ioutil.ReadFile(o.secretIDStorePath)
		if err == nil {
			o.initialSecretID = string(secretID)
		}
	}

	// ... from the cli or the environment
	if o.initialSecretID == "" {
		o.initialSecretID = viper.GetString("vault-approle-initial-secretid")
	}

	// ... from the initial SecretID path
	if o.initialSecretID == "" {
		secretID, err := ioutil.ReadFile(o.initialSecretIDPath)
		if err != nil {
			return nil, err
		}
		o.initialSecretID = string(secretID)
	}

	// Still no valid initial SecretID? Fail!
	if o.initialSecretID == "" {
		return nil, fmt.Errorf("failed to retrieve the initial AppRole SecretID")
	}

	approle := &Approle{roleID: o.roleID, secretIDStorePath: o.secretIDStorePath}
	approle.SetSecretID(o.initialSecretID)

	return approle, nil
}

// Login with AppRole authentication at the given authentication path (/auth/<authPath>/login) and the given client
// As this Approle type is build around the idea that the SecretID is single use only,
// the Login method also tries to retrieve and store a new SecretID and destroys the old one
func (approle *Approle) Login(rawClient *vaultapi.Client, authPath string, role string) (*vaultapi.Secret, error) {
	payload := map[string]interface{}{"role_id": approle.roleID, "secret_id": approle.secretID}
	logical := rawClient.Logical()

	// perform the login
	tokenSecret, err := logical.Write(fmt.Sprintf("auth/%s/login", authPath), payload)
	if err != nil {
		return nil, err
	}

	// store the token and make it the active auth token for the client as
	// we need to create a new secret id in the next step which requires
	// authenticated requests
	approle.token = string(tokenSecret.Auth.ClientToken)
	rawClient.SetToken(approle.token)

	payload = map[string]interface{}{}
	secretIDSecret, err := logical.Write(fmt.Sprintf("auth/%s/role/%s/secret-id", authPath, role), payload)
	// failing to generate a new the secret id is bad but not immediately fatal
	// so we do return the token secret with the error and hope for the best
	if err != nil {
		return tokenSecret, err
	}
	oldSecretID := approle.secretID
	approle.SetSecretID(secretIDSecret.Data["secret_id"].(string))

	// Enforce single use SecretIDs by destroying the old SecretID
	payload = map[string]interface{}{"secret_id": oldSecretID}
	_, err = logical.Write(fmt.Sprintf("auth/%s/role/%s/secret-id/destroy", authPath, role), payload)
	if err != nil {
		log.Warn(fmt.Sprintf("Failed to destroy old SecretID: %s", err.Error()))
	}

	return tokenSecret, err
}

// SetSecretID sets the secret id for the approle and tries to persist it to disk
func (approle *Approle) SetSecretID(secretID string) {
	approle.persisted = false
	approle.secretID = secretID
	approle.Persist()
}

// Persist the currently active SecretID to the SecretID store
// If persisting was successfull can be checked with the "persisted"
// field. Failing to persist the SecretID is not fatal as the in
// memory one can still be used
func (approle *Approle) Persist() {
	err := ioutil.WriteFile(approle.secretIDStorePath, []byte(approle.secretID), 0600)
	if err != nil {
		log.Warn(fmt.Sprintf("Failed to persist SecretID: %s", err.Error()))
		return
	}
	approle.persisted = true
}

// Persisted returns if the current secret ID has been persisted
// to disk successfully
func (approle *Approle) Persisted() bool {
	return approle.persisted
}

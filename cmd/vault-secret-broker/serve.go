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

package main

import (
	"fmt"
	"net/http"

	"github.com/bedag/vault-secret-broker/pkg/vault"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var client *vault.Client

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start serving the broker api",
	Run: func(cmd *cobra.Command, args []string) {
		http.HandleFunc("/", APIRoot)
		var listenAddress string
		if viper.GetBool("tls") {
			log.Info("Creating TLS listener")
			listenAddress = fmt.Sprint(viper.GetString("listen-ip"), ":", viper.GetInt("listen-tls-port"))
			err := http.ListenAndServeTLS(listenAddress, viper.GetString("tls-cert"), viper.GetString("tls-key"), nil)
			if err != nil {
				log.Fatal(err.Error())
			}
		} else {
			log.Info("Creating plain http listener")
			listenAddress = fmt.Sprint(viper.GetString("listen-ip"), ":", viper.GetInt("listen-port"))
			err := http.ListenAndServe(listenAddress, nil)
			if err != nil {
				log.Fatal(err.Error())
			}
		}

		log.Info(fmt.Sprint("Listening on ", listenAddress))

	},
}

func init() {
	serveCmd.Flags().StringP("listen-ip", "", "", "API server listen ip")
	serveCmd.Flags().IntP("listen-port", "", 8080, "API server listen port")
	serveCmd.Flags().IntP("listen-tls-port", "", 8443, "API server tls listen port")
	serveCmd.Flags().StringP("tls-cert", "", "server.crt", "TLS certificate file")
	serveCmd.Flags().StringP("tls-key", "", "server.key", "TLS private key")
	serveCmd.Flags().BoolP("tls", "", false, "Enable TLS")
	viper.BindPFlag("listen-ip", serveCmd.Flags().Lookup("listen-ip"))
	viper.BindPFlag("listen-port", serveCmd.Flags().Lookup("listen-port"))
	viper.BindPFlag("listen-tls-port", serveCmd.Flags().Lookup("listen-tls-port"))
	viper.BindPFlag("tls-cert", serveCmd.Flags().Lookup("tls-cert"))
	viper.BindPFlag("tls-key", serveCmd.Flags().Lookup("tls-key"))
	viper.BindPFlag("tls", serveCmd.Flags().Lookup("tls"))

	var err error
	client, err = vault.NewClient()
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to initialize Vault client: %s", err.Error()))
	}

	rootCmd.AddCommand(serveCmd)
}

// APIRoot is the request handler for requests to "/"
// Currently only returns the app name and the version
func APIRoot(w http.ResponseWriter, r *http.Request) {
	versionString := fmt.Sprint(appName, " ", version)
	fmt.Fprintf(w, versionString)
	_ = client.RawClient()
}

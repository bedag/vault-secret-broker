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

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start serving the broker api",
	Run: func(cmd *cobra.Command, args []string) {
		http.HandleFunc("/", ApiRoot)
		listenAddress := fmt.Sprint(viper.GetString("listen-ip"), ":", viper.GetInt("listen-port"))
		err := http.ListenAndServe(listenAddress, nil)
		if err != nil {
			log.Fatal(err.Error())
		}
	},
}

func init() {
	serveCmd.Flags().StringP("listen-ip", "", "", "API server listen ip")
	serveCmd.Flags().IntP("listen-port", "", 8080, "API server listen port ")
	viper.BindPFlag("listen-ip", serveCmd.Flags().Lookup("listen-ip"))
	viper.BindPFlag("listen-port", serveCmd.Flags().Lookup("listen-port"))

	rootCmd.AddCommand(serveCmd)
}

func ApiRoot(w http.ResponseWriter, r *http.Request) {
	versionString := fmt.Sprint(appName, " ", version)
	fmt.Fprintf(w, versionString)
}

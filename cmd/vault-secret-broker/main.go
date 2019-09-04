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
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "vault-secret-broker",
	Short: "CI/CD interface to access vault",
	Long: `Vault-secret-broker is an interface between Hashicorp Vault and a CI/CD process
that requires access to secrets stored in Vault. Instead of directly handing
out vault credentials to CI/CD servers, the secret broker adds another layer of
protection by enforcing that secrets will only be handed out to actually
running jobs.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if viper.GetBool("json-log") {
			log.SetFormatter(&log.JSONFormatter{})
		}

		logLevel, err := log.ParseLevel(viper.GetString("log-level"))

		if err != nil {
			log.Fatal(err.Error())
		}

		log.SetLevel(logLevel)
	},
}

func init() {
	viper.SetEnvPrefix(appName)
	dashReplacer := strings.NewReplacer("-", "_", ".", "_")
	viper.SetEnvKeyReplacer(dashReplacer)
	viper.AutomaticEnv()

	rootCmd.PersistentFlags().StringP("log-level", "", log.WarnLevel.String(), "log level (trace,debug,info,warn/warning,error,fatal,panic)")
	rootCmd.PersistentFlags().BoolP("json-log", "", false, "log as json")

	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("json-log", rootCmd.PersistentFlags().Lookup("json-log"))
}

func execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func main() {
	execute()
}

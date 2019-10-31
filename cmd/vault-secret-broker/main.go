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
	// use a short EnvPrefix as appName would be unpractically long
	viper.SetEnvPrefix("vsb")
	dashReplacer := strings.NewReplacer("-", "_", ".", "_")
	viper.SetEnvKeyReplacer(dashReplacer)
	viper.AutomaticEnv()

	userHome := os.Getenv("HOME")
	defaultVaultApproleSecretIDStorePath := userHome + "/.vault-secret-broker-secret-id"
	defaultVaultApproleRoleIDPath := userHome + "/.vault-secret-broker-role-id"

	rootCmd.PersistentFlags().StringP("log-level", "", log.WarnLevel.String(), "log level (trace,debug,info,warn/warning,error,fatal,panic)")
	rootCmd.PersistentFlags().BoolP("json-log", "", false, "log as json")

	rootCmd.PersistentFlags().StringP("vault-role", "", "default", "Vault role")
	rootCmd.PersistentFlags().StringP("vault-auth-path", "", "approle", "Vault auth-path, e.g. /v1/auth/<vault-auth-path>/")
	rootCmd.PersistentFlags().StringP("vault-approle-role-id", "", "", "Vault AppRole RoleID")
	rootCmd.PersistentFlags().StringP("vault-approle-role-id-path", "", defaultVaultApproleRoleIDPath, "Vault AppRole RoleID path")
	rootCmd.PersistentFlags().StringP("vault-approle-initial-secret-id", "", "", "Initial Vault AppRole SecretID")
	rootCmd.PersistentFlags().StringP("vault-approle-initial-secret-id-path", "", defaultVaultApproleSecretIDStorePath, "Initial Vault AppRole SecretID path")
	rootCmd.PersistentFlags().StringP("vault-approle-secret-id-store-path", "", defaultVaultApproleSecretIDStorePath, "Vault AppRole SecretID storage path")

	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("json-log", rootCmd.PersistentFlags().Lookup("json-log"))
	viper.BindPFlag("vault-role", rootCmd.PersistentFlags().Lookup("vault-role"))
	viper.BindPFlag("vault-auth-path", rootCmd.PersistentFlags().Lookup("vault-auth-path"))
	viper.BindPFlag("vault-approle-role-id", rootCmd.PersistentFlags().Lookup("vault-approle-role-id"))
	viper.BindPFlag("vault-approle-role-id-path", rootCmd.PersistentFlags().Lookup("vault-approle-role-id-path"))
	viper.BindPFlag("vault-approle-initial-secret-id", rootCmd.PersistentFlags().Lookup("vault-approle-intitial-secret-id"))
	viper.BindPFlag("vault-approle-initial-secret-id-path", rootCmd.PersistentFlags().Lookup("vault-approle-initial-secret-id-path"))
	viper.BindPFlag("vault-approle-secret-id-store-path", rootCmd.PersistentFlags().Lookup("vault-approle-secret-id-store-path"))
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

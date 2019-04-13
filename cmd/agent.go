// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
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

package cmd

import (
	"log"

	"github.com/dirty49374/coreddns/pkg/coreddns"
	"github.com/spf13/cobra"
)

// agentCmd represents the server command
var agentCmd = &cobra.Command{
	Use:   "agent",
	Short: "start DDNS agent",
	Run: func(cmd *cobra.Command, args []string) {
		domain, err := cmd.Flags().GetString("domain")
		if err != nil {
			panic(err)
		}

		servers, err := cmd.Flags().GetStringArray("etcd")
		if err != nil {
			panic(err)
		}

		log.Printf("* domain  = %s\n", domain)
		log.Printf("  servers = %+v\n", servers)

		lease, err := cmd.Flags().GetInt64("lease")
		if err != nil {
			panic(err)
		}
		coreddns.StartServer(servers, domain, lease)
	},
}

func init() {
	rootCmd.AddCommand(agentCmd)

	agentCmd.Flags().StringP("domain", "d", "lan.", "domain")
	agentCmd.Flags().Int64P("lease", "l", 600, "lease time in seconds")
	agentCmd.Flags().StringArrayP("etcd", "e", []string{"http://127.0.0.1:2379"}, "etcd server list")
}

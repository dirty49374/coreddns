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

// unsetCmd represents the agent command
var unsetCmd = &cobra.Command{
	Use:   "unset",
	Short: "delete dns record",
	Args:  cobra.ExactArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		servers, err := cmd.Flags().GetStringArray("server")
		if err != nil {
			log.Println(err)
			return
		}

		name, _ := cmd.Flags().GetString("name")
		ip, _ := cmd.Flags().GetString("ip")

		coreddns.Unset(servers, name, ip)

	},
}

func init() {
	rootCmd.AddCommand(unsetCmd)

	unsetCmd.Flags().String("name", "", "dns name (ex, --name server1)")
	unsetCmd.Flags().String("ip", "", "ip address (ex, --ip 10.0.1.1)")

	unsetCmd.Flags().StringArrayP("server", "s", []string{"127.0.0.1"}, "server ip addresses")
}

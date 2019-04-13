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
	"os"
	"strings"
	"time"

	"github.com/dirty49374/coreddns/pkg/coreddns"
	"github.com/spf13/cobra"
)

// leaseCmd represents the agent command
var leaseCmd = &cobra.Command{
	Use:   "lease",
	Short: "lease DDNS record",
	Run: func(cmd *cobra.Command, args []string) {
		servers, err := cmd.Flags().GetStringArray("server")
		if err != nil {
			log.Println(err)
			return
		}

		if len(args) == 0 {
			longName, err := os.Hostname()
			if err != nil {
				panic(err)
			}

			name := strings.Split(longName, ".")[0]

			args = []string{name}
		}

		interval, _ := cmd.Flags().GetInt64("interval")
		if interval == 0 {
			coreddns.Lease(servers, args)
			return
		}

		for {
			err := coreddns.Lease(servers, args)
			if err != nil {
				time.Sleep(10 * time.Second)
			} else {
				time.Sleep(time.Duration(interval) * time.Second)
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(leaseCmd)

	leaseCmd.Flags().StringArrayP("server", "s", []string{"127.0.0.1"}, "server ip addresses")
	leaseCmd.Flags().Int64P("interval", "i", 180, "update interval in seconds")
}

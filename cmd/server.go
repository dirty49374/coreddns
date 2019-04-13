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
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/spf13/cobra"
)

func runETCD(ctx context.Context, myName string, myAddr string, servers map[string]string) {
	clusters := ""
	for name, addr := range servers {
		if len(clusters) == 0 {
			clusters = name + "=http://" + addr + ":2380"
		} else {
			clusters = clusters + "," + name + "=http://" + addr + ":2380"
		}
	}

	args := []string{
		"--name", myName,
		"--data-dir", "data",
		"--listen-client-urls", "http://0.0.0.0:2379",
		"--advertise-client-urls", "http://" + myAddr + ":2379",
		"--listen-peer-urls", "http://0.0.0.0:2380",
		"--initial-advertise-peer-urls", "http://" + myAddr + ":2380",
		"--initial-cluster", clusters,
		"--initial-cluster-token", "etcd-token",
		"--initial-cluster-state", "new",
	}
	fmt.Println("=================================================")
	fmt.Printf("etcd: ./bin/etcd %v\n", args)
	cmd := exec.Command("./bin/etcd", args...)
	fmt.Println("=================================================")

	go func() {
		for {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()

			fmt.Printf("etcd: %s\n", err)
			time.Sleep(10 * time.Second)
		}
	}()
}

func runCoreDNS(ctx context.Context, name string, servers []string) {
	args := []string{
		"-conf", "conf/Corefile",
	}
	fmt.Println("=================================================")
	fmt.Printf("coredns: ./bin/coredns %v\n", args)
	fmt.Println("=================================================")

	go func() {
		for {
			cmd := exec.Command("./bin/coredns", args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()

			fmt.Printf("coredns: %s\n", err)
			time.Sleep(10 * time.Second)
		}
	}()
}

func runAgent(ctx context.Context, domain string, lease int64) {
	args := []string{
		"agent",
		"--domain", domain,
		"--etcd", "http://127.0.0.1:2379",
		"--lease", strconv.FormatInt(lease, 10),
	}
	fmt.Println("=================================================")
	fmt.Printf("agent: ./coreddns %v\n", args)
	fmt.Println("=================================================")

	go func() {
		for {
			cmd := exec.Command("./coreddns", args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()

			fmt.Printf("agent: %s\n", err)
			time.Sleep(10 * time.Second)
		}
	}()
}

type DnsConf struct {
	Name   string
	Addr   string
	Domain string
}

func buildCorefile(conf DnsConf) {

	buf, err := ioutil.ReadFile("Corefile.tpl")
	if err != nil {
		panic(err)
	}

	tpl, err := template.New("Corefile.tpl").Parse(string(buf))
	if err != nil {
		panic(err)
	}

	var out bytes.Buffer
	err = tpl.Execute(&out, conf)
	if err != nil {
		panic(err)
	}

	text := out.String()
	fmt.Println(text)

	ioutil.WriteFile("conf/Corefile", out.Bytes(), 0644)
}

// serverCmd represents the start command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start dns + agent server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain, err := cmd.Flags().GetString("domain")
		if err != nil || domain == "" {
			panic(err)
		}

		servers, err := cmd.Flags().GetStringArray("server")
		if err != nil {
			panic(err)
		}

		serverMap := make(map[string]string)
		for _, server := range servers {
			splits := strings.Split(server, "=")
			serverMap[splits[0]] = splits[1]
		}

		myName := args[0]
		myAddr := serverMap[myName]
		if myAddr == "" {
			panic("cannot find server entry for " + myName)
		}

		lease, err := cmd.Flags().GetInt64("lease")
		if err != nil {
			panic(err)
		}

		buildCorefile(DnsConf{
			Name:   myName,
			Addr:   myAddr,
			Domain: domain,
		})

		ctx := context.Background()

		runETCD(ctx, myName, myAddr, serverMap)
		time.Sleep(1 * time.Second)

		runCoreDNS(ctx, myName, servers)
		time.Sleep(1 * time.Second)

		runAgent(ctx, domain, lease)

		<-ctx.Done()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	serverCmd.Flags().StringP("domain", "d", "", "domain (ex, --domain local.)")
	serverCmd.Flags().Int64P("lease", "l", 600, "default lease time in seconds")
	serverCmd.Flags().StringArrayP("server", "e", []string{}, "dns server list (ex, --server ns1=10.0.0.1 --server ns2=10.0.0.2)")
}

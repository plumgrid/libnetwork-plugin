// Copyright 2016 PLUMgrid
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.package main

package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	Log "github.com/Sirupsen/logrus"
	driver "github.com/plumgrid/libnetwork-plugin/plugin/driver"
)

var version = "v1.0"

func main() {
	var (
		justVersion bool
		address     string
	)

	flag.BoolVar(&justVersion, "version", false, "print version and exit")
	flag.StringVar(&address, "socket", "/run/docker/plugins/plumgrid.sock", "socket on which to listen")
	flag.Parse()

	if justVersion {
		fmt.Printf("PLUMgrid plugin %s\n", version)
		os.Exit(0)
	}

	driver.ReadConfig()

	var d driver.Driver
	d, err := driver.New(version)
	if err != nil {
		Log.Fatalf("Unable to create driver: %s", err)
	}

	var listener net.Listener

	listener, err = net.Listen("unix", address)
	if err != nil {
		Log.Fatal(err)
	}
	defer listener.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	endChan := make(chan error, 1)
	go func() {
		endChan <- d.Listen(listener)
	}()

	select {
	case sig := <-sigChan:
		Log.Debugf("Caught signal %s; shutting down", sig)
	case err := <-endChan:
		if err != nil {
			Log.Errorf("Error from listener: ", err)
			listener.Close()
			os.Exit(1)
		}
	}
}

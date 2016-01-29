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
		//nameserver  string
	)

	flag.BoolVar(&justVersion, "version", false, "print version and exit")
	flag.StringVar(&address, "socket", "/run/docker/plugins/plumgrid.sock", "socket on which to listen")
	//flag.StringVar(&nameserver, "nameserver", "", "nameserver to provide to containers")

	flag.Parse()

	if justVersion {
		fmt.Printf("PLUMgrid plugin %s\n", version)
		os.Exit(0)
	}

	var d driver.Driver
	d, err := driver.New(version)
	if err != nil {
		Log.Fatalf("unable to create driver: %s", err)
	}

	/*if nameserver != "" {
		if err := d.SetNameserver(nameserver); err != nil {
			Log.Fatalf("could not set nameserver: %s", err)
		}
	}*/

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

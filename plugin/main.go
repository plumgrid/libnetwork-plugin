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
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/docker/go-plugins-helpers/network"
	"github.com/plumgrid/libnetwork-plugin/plugin/driver"
	"github.com/urfave/cli"
)

const (
	version = "1.0.0"
)

func main() {

	var flagDebug = cli.BoolFlag{
		Name:  "debug, d",
		Usage: "enable debugging",
	}
	app := cli.NewApp()
	app.Name = "plumgrid"
	app.Usage = "PLUMgrid Libnetwork Plugin"
	app.Version = version
	app.Flags = []cli.Flag{
		flagDebug,
	}
	app.Action = Run
	app.Run(os.Args)
}

// Run initializes the driver
func Run(ctx *cli.Context) {
	if ctx.Bool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	driver.ReadConfig()

	d, err := driver.New(version)
	if err != nil {
		panic(err)
	}
	h := network.NewHandler(d)
	h.ServeUnix("root", "plumgrid")
}

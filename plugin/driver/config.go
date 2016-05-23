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

package driver

import (
	"github.com/alyu/configparser"
	"log"
)

var vip string

// Read configuration file
func ReadConfig() {
	config, err := configparser.Read("/etc/config.ini")
	if err != nil {
		log.Fatal(err)
	}
	// get a PLUMgrid virtual IP
	section, err := config.Section("PLUMgrid")
	if err != nil {
		log.Fatal(err)
	} else {
		vip = section.ValueOf("virtual_ip")
	}

}

const (
	username       = "plumgrid"
	password       = "plumgrid"
	default_domain = "admin"
	scope          = "local"
)

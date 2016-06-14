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

var (
	vip            string
	username       string
	password       string
	default_vd     string
	scope          string
)

// Read configuration file
func ReadConfig() {
	config, err := configparser.Read("/opt/pg/libnetwork/config.ini")
	if err != nil {
		log.Fatal(err)
	}

	// get PLUMgrid config
	pg_section, pg_err := config.Section("PLUMgrid")
	if pg_err != nil {
		log.Fatal(pg_err)
	} else {
		vip = pg_section.ValueOf("virtual_ip")
		username = pg_section.ValueOf("pg_username")
		password = pg_section.ValueOf("pg_password")
		default_vd = pg_section.ValueOf("default_domain")
	}

	// get Libnetwork config
	lib_section, lib_err := config.Section("Libnetwork")
	if lib_err != nil {
		log.Fatal(lib_err)
	} else {
		scope = lib_section.ValueOf("scope")
	}
}

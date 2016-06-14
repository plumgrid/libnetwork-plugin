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
	"encoding/json"

	Log "github.com/Sirupsen/logrus"
)

func pgBridgeCreate(ID string, domainid string, gatewayip string) error {

	// PUT Domain Data

	url := "/0/connectivity/domain/" + domainid + "/ne/bri" + ID

	data := []byte(`{"action": {
				"action1": {
					"action_text": "create_and_link_ifc(DYN_1)"}
				},
				"ifc": {},
				"mobility": "true",
				"ne_description": "PLUMgrid Bridge",
				"ne_dname": "net-` + ID[:10] + `",
				"ne_group": "Bridge",
				"ne_type": "bridge"
	}`)

	if _, err := rest_Call("PUT", url, data); err != nil {
		return err
	}

	// PUT Rule group Data

	url = "/0/connectivity/domain/" + domainid + "/rule_group/cnf" + ID

	data = []byte(`{"mark_disabled": false,
			"ne_dest": "/ne/bri` + ID + `/action/action1",
			"ne_dname": "cont-` + ID[10:] + `",
			"ne_type": "cnf-vmgroup",
					"rule": {
						"rules` + ID + `": {
							"add_context": "",
							"criteria": "pgtag2",
							"match": "bridge-` + ID[:10] + `"
						}
					}
	}`)

	if _, err := rest_Call("PUT", url, data); err != nil {
		return err
	}

	// PUT Domain prop

	url = "/0/connectivity/domain_prop/" + domainid + "/ne/brii" + ID

	data = []byte(`{"ne_metadata": "` + gatewayip + `"}`)

	if _, err := rest_Call("PUT", url, data); err != nil {
		return err
	}
	return nil
}

func pgBridgeDestroy(ID string, domainid string) error {

	// Delete domain prop

	url := "/0/connectivity/domain_prop/" + domainid + "/ne/brii" + ID

	if _, err := rest_Call("DELETE", url, nil); err != nil {
		return err
	}

	// Delete Domain Data

	url = "/0/connectivity/domain/" + domainid + "/ne/bri" + ID

	if _, err := rest_Call("DELETE", url, nil); err != nil {
		return err
	}

	// DELETE Rule group Data

	url = "/0/connectivity/domain/" + domainid + "/rule_group/cnf" + ID

	if _, err := rest_Call("DELETE", url, nil); err != nil {
		return err
	}
	return nil
}

func FindDomainFromNetwork(ID string) (string, error) {

	url := "/0/connectivity/domain?configonly=true&level=3"

	body, err := rest_Call("GET", url, nil)
	if err != nil {
		return "", err
	}

	var domain_data map[string]interface{}
	err = json.Unmarshal([]byte(body), &domain_data)
	if err != nil {
		Log.Error(err)
		return "", err
	}
	for domains, domain_val := range domain_data {
		if nes, ok := domain_val.(map[string]interface{})["ne"]; ok {
			if _, ok := nes.(map[string]interface{})["bri"+ID]; ok {
				return domains, nil
			}
		}
	}
	return "", nil
}

func FindNetworkGateway(domainID string, networkID string) (string, error) {

	url := "/0/connectivity/domain_prop/" + domainID + "/ne/brii" + networkID

	body, err := rest_Call("GET", url, nil)
	if err != nil {
		return "", err
	}

	var domain_prop map[string]interface{}
	err = json.Unmarshal([]byte(body), &domain_prop)
	if err != nil {
		Log.Error(err)
		return "", err
	}
	gatewayIP := domain_prop["ne_metadata"].(string)
	return gatewayIP, nil
}

func pgVDCreate(domainID string) error {

	url := "/0/connectivity/domain?configonly=true"

	body, err := rest_Call("GET", url, nil)
	if err != nil {
		return err
	}

	var domain_data map[string]interface{}
	err = json.Unmarshal([]byte(body), &domain_data)
	if err != nil {
		Log.Error(err)
		return err
	}
	for domains := range domain_data {
		if domains == domainID {
			return nil
		}
	}

	// PLUMgrid Tenant Manager Configuration

	url = "/0/tenant_manager/tenants" + "/" + domainID

	data := []byte(`{"containers": {
				"` + domainID + `": {
					"enable": "True",
					"qos_marking": "9",
					"context": "` + domainID + `",
					"type": "Gold",
					"property": "Container ` + domainID + ` Property",
					"domains": {}, "rules": {}}}}`)

	if _, err = rest_Call("PUT", url, data); err != nil {
		return err
	}

	// PLUMgrid Tunnel Configuration

	url = "/0/tunnel_service/vnd_config/" + domainID

	data = []byte(`{"profile_name": "VXLAN",
			"add_vlan": "False"}`)

	if _, err = rest_Call("PUT", url, data); err != nil {
		return err
	}

	// Create VD in Connectivity Manager

	url = "/0/connectivity/domain/" + domainID

	data = []byte(`{"container_group": "` + domainID + `",
			"topology_name": "docker"}`)

	if _, err = rest_Call("PUT", url, data); err != nil {
		return err
	}

	// Create VD in Domain Prop

	url = "/0/connectivity/domain_prop/" + domainID

	data = []byte(`{}`)

	if _, err = rest_Call("PUT", url, data); err != nil {
		return err
	}

	// Create PEM Master Rules

	url = "/0/pem_master/log_rule/" + domainID

	data = []byte(`{"rule": {
				"rule_` + domainID + `": {
					"pgtag1": "` + domainID + `",
					"log_ifc_type": "ACCESS_VM"}}}`)

	if _, err = rest_Call("PUT", url, data); err != nil {
		return err
	}
	return nil
}

func pgVDDelete(domainID string) error {

	url := "/0/connectivity/domain?configonly=true"

	body, err := rest_Call("GET", url, nil)
	if err != nil {
		return err
	}

	var domain_data map[string]interface{}
	err = json.Unmarshal([]byte(body), &domain_data)
	if err != nil {
		Log.Error(err)
		return err
	}
	for domains, domain_val := range domain_data {
		if domains == domainID {
			res := domain_val.(map[string]interface{})["ne"]
			if len(res.(map[string]interface{})) == 0 {

				// Delete PEM Master Rules

				url = "/0/pem_master/log_rule/" + domainID

				if _, err = rest_Call("DELETE", url, nil); err != nil {
					return err
				}

				// Delete VD from Domain Prop

				url = "/0/connectivity/domain_prop/" + domainID

				if _, err = rest_Call("DELETE", url, nil); err != nil {
					return err
				}

				// Delete VD from Connectivity Manager

				url = "/0/connectivity/domain/" + domainID

				if _, err = rest_Call("DELETE", url, nil); err != nil {
					return err
				}

				// Delete PLUMgrid Tunnel Configuration

				url = "/0/tunnel_service/vnd_config/" + domainID

				if _, err = rest_Call("DELETE", url, nil); err != nil {
					return err
				}

				// Delete PLUMgrid Tenant Manager Configuration

				url = "/0/tenant_manager/tenants" + "/" + domainID

				if _, err = rest_Call("DELETE", url, nil); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

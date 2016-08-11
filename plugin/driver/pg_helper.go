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
	"strings"
)

func BridgeCreate(ID string, domainid string, gatewayip string) {

	url := "/0/connectivity/domain/" + domainid + "/ne/bri" + ID
	data := []byte(`{"action": {
				"action1": {
					"action_text": "create_and_link_ifc(DYN_1)"}
				},
				"ifc": {},
				"metadata": "` + ID + `",
				"mobility": "true",
				"ne_description": "PLUMgrid Bridge",
				"ne_dname": "net-` + ID[:10] + `",
				"ne_group": "Bridge",
				"ne_type": "bridge"
	}`)

	RestCall("PUT", url, data)

	url = "/0/connectivity/domain/" + domainid + "/rule_group/cnf" + ID
	data = []byte(`{"mark_disabled": false,
			"ne_dest": "/ne/bri` + ID + `/action/action1",
			"ne_dname": "cont-` + ID[10:] + `",
			"ne_type": "cnf-contgroup",
					"rule": {
						"rules` + ID + `": {
							"add_context": "",
							"criteria": "pgtag2",
							"match": "bri` + ID + `"
						}
					}
	}`)

	RestCall("PUT", url, data)

	url = "/0/connectivity/domain_prop/" + domainid + "/ne/brii" + ID
	data = []byte(`{"ne_metadata": "` + gatewayip + `"}`)
	RestCall("PUT", url, data)
}

func AddGatewayInfo(ID string, domainid string, gatewayip string) {

	url := "/0/connectivity/domain_prop/" + domainid + "/ne/brii" + ID
	data := []byte(`{"ne_metadata": "` + gatewayip + `"}`)
	RestCall("PUT", url, data)
}

func BridgeDelete(ID string, domainid string) {

	url := "/0/connectivity/domain_prop/" + domainid + "/ne/brii" + ID
	RestCall("DELETE", url, nil)

	url = "/0/connectivity/domain/" + domainid + "/ne/bri" + ID
	RestCall("DELETE", url, nil)

	url = "/0/connectivity/domain/" + domainid + "/rule_group/cnf" + ID
	RestCall("DELETE", url, nil)
}

func AddNetworkInfo(dNetwork string, neName string, domain string) {
        id := GetNeId(neName, domain)
        var neId string
        if id == "" {
            neId = neName
        } else {
            neId = id
        }
        url := "/0/connectivity/domain/" + domain + "/ne/" + neId
        body, _ := RestCall("GET", url+"?configonly=true", nil)
        var netData map[string]interface{}
        err := json.Unmarshal([]byte(body), &netData)
        if err != nil {
                panic(err)
        }
        netData["metadata"] = dNetwork
        finalData, _ := json.Marshal(netData)
        RestCall("PUT", url, finalData)
}

func FindDomainFromNetwork(ID string) (domainid string, netid string) {

	url := "/0/connectivity/domain?configonly=true&level=3"
	body, _ := RestCall("GET", url, nil)
	var domain_data map[string]interface{}
	err := json.Unmarshal([]byte(body), &domain_data)
	if err != nil {
		panic(err)
	}
	for domains, domain_val := range domain_data {
		if nes, ok := domain_val.(map[string]interface{})["ne"]; ok {
			for ne, data := range nes.(map[string]interface{}) {
				if data.(map[string]interface{})["metadata"] == ID {
					domainid = domains
					netid = ne
					break
				}
			}
		}
	}
	return
}

func FindNetworkGateway(domainID string, networkID string) (gatewayIP string) {

	url := "/0/connectivity/domain_prop/" + domainID + "/ne/brii" + networkID
	body, _ := RestCall("GET", url, nil)
	var domain_prop map[string]interface{}
	err := json.Unmarshal([]byte(body), &domain_prop)
	if err != nil {
		panic(err)
	}
	gatewayIP = domain_prop["ne_metadata"].(string)
	return
}

func DomainCreate(domainID string) {

	url := "/0/connectivity/domain?configonly=true"
	body, _ := RestCall("GET", url, nil)
	var domain_data map[string]interface{}
	err := json.Unmarshal([]byte(body), &domain_data)
	if err != nil {
		panic(err)
	}
	for domains := range domain_data {
		if domains == domainID {
			return
		}
	}

	url = "/0/tenant_manager/tenants" + "/" + domainID
	data := []byte(`{"containers": {
				"` + domainID + `": {
					"enable": "True",
					"context": "` + domainID + `",
                                        "vd_display_name": "` + domainID + `",
					"domains": {}, "rules": {}}}}`)
	RestCall("PUT", url, data)

	url = "/0/tenant_manager/metaconfig/" + domainID
	data = []byte(`{}`)
	RestCall("PUT", url, data)

	url = "/0/tunnel_service/vnd_config/" + domainID
	data = []byte(`{"profile_name": "VXLAN",
			"add_vlan": "False"}`)
	RestCall("PUT", url, data)

	url = "/0/connectivity/domain/" + domainID
	data = []byte(`{"container_group": "` + domainID + `",
			"topology_name": "docker"}`)
	RestCall("PUT", url, data)

	url = "/0/connectivity/domain_prop/" + domainID
	data = []byte(`{}`)
	RestCall("PUT", url, data)

	url = "/0/pem_master/log_rule/" + domainID
	data = []byte(`{"rule": {
				"rule_` + domainID + `": {
					"pgtag1": "` + domainID + `",
					"log_ifc_type": "ACCESS_CONTAINER"}}}`)

	RestCall("PUT", url, data)
}

func DomainDelete(domainID string) {

	url := "/0/connectivity/domain?configonly=true"
	body, _ := RestCall("GET", url, nil)
	var domain_data map[string]interface{}
	err := json.Unmarshal([]byte(body), &domain_data)
	if err != nil {
		panic(err)
	}
	for domains, domain_val := range domain_data {
		if domains == domainID {
			res := domain_val.(map[string]interface{})["ne"]
			if len(res.(map[string]interface{})) == 0 {

				url = "/0/pem_master/log_rule/" + domainID
				RestCall("DELETE", url, nil)

				url = "/0/connectivity/domain_prop/" + domainID
				RestCall("DELETE", url, nil)

				url = "/0/connectivity/domain/" + domainID
				RestCall("DELETE", url, nil)

				url = "/0/tunnel_service/vnd_config/" + domainID
				RestCall("DELETE", url, nil)

				url = "/0/tenant_manager/metaconfig/" + domainID
				RestCall("DELETE", url, nil)

				url = "/0/tenant_manager/tenants/" + domainID
				RestCall("DELETE", url, nil)
			}
		}
	}
}

func GetNeId(NeName string, DomainID string) (NeID string) {

        url := "/0/connectivity/domain/" + DomainID + "/ne?configonly=true&level=1"
        body, _ := RestCall("GET", url, nil)
        var ne_data map[string]interface{}
        err := json.Unmarshal([]byte(body), &ne_data)
        if err != nil {
                panic(err)
        }
        for nes, ne_val := range ne_data {
                if ne_val.(map[string]interface{})["ne_dname"] == NeName {
                        NeID = nes
                        return NeID
                }
        }
        return ""
}

func CreateNetworkLink(NeName string, DomainID string, NetworkID string, IP string, Netmask string) {

	ne_ID := GetNeId(NeName, DomainID)
	CheckNeChildList(ne_ID, DomainID, "ifc")
	CheckNeChildList(NetworkName(NetworkID), DomainID, "ifc")
	ne_ifc := NetworkID
	url := "/0/connectivity/domain/" + DomainID + "/ne/" + ne_ID + "/ifc/" + ne_ifc
	data := []byte(`{"attachable": "true",
                         "list": "true",
                         "attach_type": "static,dynamic",
                         "mobility": "true",
                         "ifc_name": "` + ne_ifc + `",
                         "ifc_type": "static",
                         "ip_address": "` + IP + `",
                         "ip_address_mask": "` + Netmask + `"}`)
	RestCall("PUT", url, data)

	net_ifc := ne_ID
	url = "/0/connectivity/domain/" + DomainID + "/ne/" + NetworkName(NetworkID) + "/ifc/" + net_ifc
	data = []byte(`{"attachable": "true",
                         "list": "true",
                         "attach_type": "static,dynamic",
                         "mobility": "true",
                         "ifc_type": "static"}`)
	RestCall("PUT", url, data)

	link_name := ne_ID + NetworkID
	url = "/0/connectivity/domain/" + DomainID + "/link/" + link_name
	data = []byte(`{"link_type": "static",
                         "link_name": "` + link_name + `",
                         "attachment1": "/ne/` + ne_ID + `/ifc/` + ne_ifc + `",
                         "attachment2": "/ne/` + NetworkName(NetworkID) + `/ifc/` + net_ifc + `"}`)
	RestCall("PUT", url, data)
}

func DeleteNetworkLinks(DomainID string, NetworkID string) {

	var ne_ID string
	var ne_ifc string
	url := "/0/connectivity/domain/" + DomainID
	body, _ := RestCall("GET", url+"?configonly=true", nil)
	var domain_data map[string]interface{}
	err := json.Unmarshal([]byte(body), &domain_data)
	if err != nil {
		panic(err)
	}
	for domains, domain_val := range domain_data {
		if domains == "link" {
			links := domain_val.(map[string]interface{})
			for att, att_info := range links {
				att_1 := att_info.(map[string]interface{})["attachment1"].(string)
				att_2 := att_info.(map[string]interface{})["attachment2"].(string)
				if strings.Split(att_1, "/")[2][3:] == NetworkID {
					ne_ID = strings.Split(att_2, "/")[2]
					ne_ifc = strings.Split(att_2, "/")[4]
					RestCall("DELETE", url+"/link/"+att, nil)
					ifc_url := url + "/ne/" + ne_ID + "/ifc/" + ne_ifc
					RestCall("DELETE", ifc_url, nil)
				} else if strings.Split(att_2, "/")[2][3:] == NetworkID {
					ne_ID = strings.Split(att_1, "/")[2]
					ne_ifc = strings.Split(att_1, "/")[4]
					RestCall("DELETE", url+"/link/"+att, nil)
					ifc_url := url + "/ne/" + ne_ID + "/ifc/" + ne_ifc
					RestCall("DELETE", ifc_url, nil)
				}
			}
		}
	}
}

func NetworkName(ID string) (name string) {
	name = "bri" + ID
	return
}

func CheckNeChildList(NeID string, DomainID string, childList string) {

	url := "/0/connectivity/domain/" + DomainID + "/ne/" + NeID + "?configonly=true"
	body, _ := RestCall("GET", url, nil)
	var ne_data map[string]interface{}
	err := json.Unmarshal([]byte(body), &ne_data)
	if err != nil {
		panic(err)
	}
	if _, ok := ne_data[childList]; ok {
		return
	} else {
		data := []byte(`{}`)
		RestCall("PUT", url+"/"+childList, data)
		return
	}
}

func AddMetaconfig(domainID string, netID string, deviceID string, endpointID string, macaddr string) {

	url := "/0/tenant_manager/metaconfig/" + domainID + "?configonly=true"
	body, _ := RestCall("GET", url, nil)
	var tm_data map[string]interface{}
	err := json.Unmarshal([]byte(body), &tm_data)
	if err != nil {
		panic(err)
	}

	for device, _ := range tm_data["workloads"].(map[string]interface{}) {
		if device == deviceID {
			url = "/0/tenant_manager/metaconfig/" + domainID + "/workloads/" + deviceID + "/prop/" + endpointID
			data := []byte(`{"phy_address": "` + macaddr + `",
					 "hint": "` + netID + `"}`)
			RestCall("PUT", url, data)
			return
		}
	}

	url = "/0/tenant_manager/metaconfig/" + domainID + "/workloads/" + deviceID
	data := []byte(`{"ne_type": "container",
			 "prop": {
				"` + endpointID + `": {
						"phy_address": "` + macaddr + `",
						"hint": "` + netID + `"}}}`)
	RestCall("PUT", url, data)
}

func RemoveMetaconfig(domainID string, netID string, endpointID string) {

	url := "/0/tenant_manager/metaconfig/" + domainID + "?configonly=true"
	body, _ := RestCall("GET", url, nil)
	var tm_data map[string]interface{}
	err := json.Unmarshal([]byte(body), &tm_data)
	if err != nil {
		panic(err)
	}

	for device, prop := range tm_data["workloads"].(map[string]interface{}) {
		for endpoint, _ := range prop.(map[string]interface{})["prop"].(map[string]interface{}) {
			if endpoint == endpointID {
				if len(prop.(map[string]interface{})["prop"].(map[string]interface{})) < 2 {
					url = "/0/tenant_manager/metaconfig/" + domainID + "/workloads/" + device
					RestCall("DELETE", url, nil)
					return
				}
				url = "/0/tenant_manager/metaconfig/" + domainID + "/workloads/" + device + "/prop/" + endpoint
				RestCall("DELETE", url, nil)
			}
		}
	}
}

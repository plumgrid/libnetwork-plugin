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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"

	Log "github.com/Sirupsen/logrus"
)

func pgBridgeCreate(ID string, domainid string, gatewayip string) {
	cookieJar, _ := cookiejar.New(nil)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Jar:       cookieJar,
		Transport: tr,
	}

	url := "https://" + vip + "/0/login"
	Log.Infof("URL:> %s", url)

	var jsonStr = []byte(`{"userName":"` + username + `", "password":"` + password + `"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	fmt.Println("response Cookies:", resp.Cookies())
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	//== PUT call

	url2 := "https://" + vip + "/0/connectivity/domain/" + domainid + "/ne/bri" + ID
	fmt.Println("URL:>", url2)

	var jsonStr1 = []byte(`{
				"action": {
					"action1": {
						"action_text": "create_and_link_ifc(DYN_1)"
					}
				},
				"ifc": {},
				"mobility": "true",
				"ne_description": "PLUMgrid Bridge",
				"ne_dname": "net-` + ID[:10] + `",
				"ne_group": "Bridge",
				"ne_type": "bridge"
	}`)

	req2, err := http.NewRequest("PUT", url2, bytes.NewBuffer(jsonStr1))
	req2.Header.Set("Accept", "application/json")
	req2.Header.Set("Content-Type", "application/json")

	resp2, err2 := client.Do(req2)
	if err2 != nil {
		fmt.Println(err2)
	}
	defer resp2.Body.Close()

	fmt.Println("response Status:", resp2.Status)
	//fmt.Println("response Headers:", resp2.Header)
	body2, _ := ioutil.ReadAll(resp2.Body)
	fmt.Println("response Body:", string(body2))

	//== PUT call

	url3 := "https://" + vip + "/0/connectivity/domain/" + domainid + "/rule_group/cnf" + ID
	fmt.Println("URL:>", url3)

	var jsonStr3 = []byte(`{
					"mark_disabled": false,
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

	req3, err := http.NewRequest("PUT", url3, bytes.NewBuffer(jsonStr3))
	req3.Header.Set("Accept", "application/json")
	req3.Header.Set("Content-Type", "application/json")

	resp3, err3 := client.Do(req3)
	if err3 != nil {
		fmt.Println(err3)
	}
	defer resp3.Body.Close()

	fmt.Println("response Status:", resp3.Status)
	//fmt.Println("response Headers:", resp3.Header)
	body3, _ := ioutil.ReadAll(resp3.Body)
	fmt.Println("response Body:", string(body3))

	//== PUT domain prop

	url4 := "https://" + vip + "/0/connectivity/domain_prop/" + domainid + "/ne/brii" + ID
	fmt.Println("URL:>", url4)

	var jsonStr4 = []byte(`{"ne_metadata": "` + gatewayip + `"
        }`)

	req4, err := http.NewRequest("PUT", url4, bytes.NewBuffer(jsonStr4))
	req4.Header.Set("Accept", "application/json")
	req4.Header.Set("Content-Type", "application/json")

	resp4, err4 := client.Do(req4)
	if err4 != nil {
		fmt.Println(err4)
	}
	defer resp4.Body.Close()

	fmt.Println("response Status:", resp4.Status)
	body4, _ := ioutil.ReadAll(resp4.Body)
	fmt.Println("response Body:", string(body4))

}

func pgBridgeDestroy(ID string, domainid string) {
	cookieJar, _ := cookiejar.New(nil)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Jar:       cookieJar,
		Transport: tr,
	}

	url := "https://" + vip + "/0/login"
	Log.Infof("URL:> %s", url)

	var jsonStr = []byte(`{"userName":"` + username + `", "password":"` + password + `"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	fmt.Println("response Cookies:", resp.Cookies())
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	//== Delete domain prop

	url1 := "https://" + vip + "/0/connectivity/domain_prop/" + domainid + "/ne/brii" + ID
	fmt.Println("URL:>", url1)

	req1, err := http.NewRequest("DELETE", url1, nil)
	req1.Header.Set("Accept", "application/json")
	req1.Header.Set("Content-Type", "application/json")

	resp1, err1 := client.Do(req1)
	if err1 != nil {
		fmt.Println(err1)
	}
	defer resp1.Body.Close()

	fmt.Println("response Status:", resp1.Status)
	body1, _ := ioutil.ReadAll(resp1.Body)
	fmt.Println("response Body:", string(body1))

	//== Delete call

	url2 := "https://" + vip + "/0/connectivity/domain/" + domainid + "/ne/bri" + ID
	fmt.Println("URL:>", url2)

	req2, err := http.NewRequest("DELETE", url2, nil)
	req2.Header.Set("Accept", "application/json")
	req2.Header.Set("Content-Type", "application/json")

	resp2, err2 := client.Do(req2)
	if err2 != nil {
		fmt.Println(err2)
	}
	defer resp2.Body.Close()

	fmt.Println("response Status:", resp2.Status)
	//fmt.Println("response Headers:", resp2.Header)
	body2, _ := ioutil.ReadAll(resp2.Body)
	fmt.Println("response Body:", string(body2))

	//== DELETE call

	url3 := "https://" + vip + "/0/connectivity/domain/" + domainid + "/rule_group/cnf" + ID
	fmt.Println("URL:>", url3)

	req3, err := http.NewRequest("DELETE", url3, nil)
	req3.Header.Set("Accept", "application/json")
	req3.Header.Set("Content-Type", "application/json")

	resp3, err3 := client.Do(req3)
	if err3 != nil {
		fmt.Println(err3)
	}
	defer resp3.Body.Close()

	fmt.Println("response Status:", resp3.Status)
	//fmt.Println("response Headers:", resp3.Header)
	body3, _ := ioutil.ReadAll(resp3.Body)
	fmt.Println("response Body:", string(body3))

}

func FindDomainFromNetwork(ID string) (domainid string) {
	cookieJar, _ := cookiejar.New(nil)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Jar:       cookieJar,
		Transport: tr,
	}

	url := "https://" + vip + "/0/login"
	Log.Infof("URL:> %s", url)

	var jsonStr = []byte(`{"userName":"` + username + `", "password":"` + password + `"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	fmt.Println("response Cookies:", resp.Cookies())
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	url2 := "https://" + vip + "/0/connectivity/domain?configonly=true&level=3"
	fmt.Println("URL:>", url2)

	req2, err := http.NewRequest("GET", url2, nil)
	req2.Header.Set("Accept", "application/json")

	resp2, err2 := client.Do(req2)
	if err2 != nil {
		fmt.Println(err2)
	}
	defer resp2.Body.Close()
	body2, _ := ioutil.ReadAll(resp2.Body)
	var domain_data map[string]interface{}
	err3 := json.Unmarshal([]byte(body2), &domain_data)
	if err3 != nil {
		panic(err3)
	}
	for domains, domain_val := range domain_data {
		if nes, ok := domain_val.(map[string]interface{})["ne"]; ok {
			if _, ok := nes.(map[string]interface{})["bri"+ID]; ok {
				domainid = domains
				break
			}
		}
	}
	return
}

func FindNetworkGateway(domainID string, networkID string) (gatewayIP string) {
	cookieJar, _ := cookiejar.New(nil)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Jar:       cookieJar,
		Transport: tr,
	}

	url := "https://" + vip + "/0/login"
	Log.Infof("URL:> %s", url)

	var jsonStr = []byte(`{"userName":"` + username + `", "password":"` + password + `"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	fmt.Println("response Cookies:", resp.Cookies())
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	url2 := "https://" + vip + "/0/connectivity/domain_prop/" + domainID + "/ne/brii" + networkID
	fmt.Println("URL:>", url2)

	req2, err := http.NewRequest("GET", url2, nil)
	req2.Header.Set("Accept", "application/json")

	resp2, err2 := client.Do(req2)
	if err2 != nil {
		fmt.Println(err2)
	}
	defer resp2.Body.Close()
	body2, _ := ioutil.ReadAll(resp2.Body)
	var domain_prop map[string]interface{}
	err3 := json.Unmarshal([]byte(body2), &domain_prop)
	if err3 != nil {
		panic(err3)
	}
	gatewayIP = domain_prop["ne_metadata"].(string)
	return
}

func pgVDCreate(domainID string) {
	cookieJar, _ := cookiejar.New(nil)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Jar:       cookieJar,
		Transport: tr,
	}

	url := "https://" + vip + "/0/login"
	Log.Infof("URL:> %s", url)

	var jsonStr = []byte(`{"userName":"` + username + `", "password":"` + password + `"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	fmt.Println("response Cookies:", resp.Cookies())
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	url0 := "https://" + vip + "/0/connectivity/domain?configonly=true"
	fmt.Println("URL:>", url0)

	req0, err := http.NewRequest("GET", url0, nil)
	req0.Header.Set("Accept", "application/json")

	resp0, err0 := client.Do(req0)
	if err0 != nil {
		fmt.Println(err0)
	}
	defer resp0.Body.Close()
	body0, _ := ioutil.ReadAll(resp0.Body)
	var domain_data map[string]interface{}
	err0 = json.Unmarshal([]byte(body0), &domain_data)
	if err0 != nil {
		panic(err0)
	}
	for domains := range domain_data {
		if domains == domainID {
			return
		}
	}

	// PLUMgrid Tenant Manager Configuration

	url2 := "https://" + vip + "/0/tenant_manager/tenants" + "/" + domainID
	Log.Infof("URL:>", url2)

	var jsonStr2 = []byte(`{"containers": {
					"` + domainID + `": {
						"enable": "True",
						"qos_marking": "9",
						"context": "` + domainID + `",
						"type": "Gold",
						"property": "Container ` + domainID + ` Property",
						"domains": {}, "rules": {}}}}`)
	req2, err := http.NewRequest("PUT", url2, bytes.NewBuffer(jsonStr2))
	req2.Header.Set("Accept", "application/json")
	req2.Header.Set("Content-Type", "application/json")

	resp2, err2 := client.Do(req2)
	if err2 != nil {
		fmt.Println(err2)
	}
	defer resp2.Body.Close()

	fmt.Println("response Status:", resp2.Status)
	//fmt.Println("response Headers:", resp2.Header)
	body2, _ := ioutil.ReadAll(resp2.Body)
	fmt.Println("response Body:", string(body2))

	// PLUMgrid Tunnel Configuration

	url3 := "https://" + vip + "/0/tunnel_service/vnd_config/" + domainID
	fmt.Println("URL:>", url3)

	var jsonStr3 = []byte(`{"profile_name": "VXLAN",
				"add_vlan": "False"}`)

	req3, err := http.NewRequest("PUT", url3, bytes.NewBuffer(jsonStr3))
	req3.Header.Set("Accept", "application/json")
	req3.Header.Set("Content-Type", "application/json")

	resp3, err3 := client.Do(req3)
	if err3 != nil {
		fmt.Println(err3)
	}
	defer resp3.Body.Close()

	fmt.Println("response Status:", resp3.Status)
	//fmt.Println("response Headers:", resp3.Header)
	body3, _ := ioutil.ReadAll(resp3.Body)
	fmt.Println("response Body:", string(body3))

	// Create VD in Connectivity Manager

	url4 := "https://" + vip + "/0/connectivity/domain/" + domainID
	fmt.Println("URL:>", url4)

	var jsonStr4 = []byte(`{"container_group": "` + domainID + `",
				"topology_name": "docker"}`)

	req4, err := http.NewRequest("PUT", url4, bytes.NewBuffer(jsonStr4))
	req4.Header.Set("Accept", "application/json")
	req4.Header.Set("Content-Type", "application/json")

	resp4, err4 := client.Do(req4)
	if err4 != nil {
		fmt.Println(err4)
	}
	defer resp4.Body.Close()

	fmt.Println("response Status:", resp4.Status)
	body4, _ := ioutil.ReadAll(resp4.Body)
	fmt.Println("response Body:", string(body4))

	// Create VD in Domain Prop

	url5 := "https://" + vip + "/0/connectivity/domain_prop/" + domainID
	fmt.Println("URL:>", url5)

	var jsonStr5 = []byte(`{}`)

	req5, err := http.NewRequest("PUT", url5, bytes.NewBuffer(jsonStr5))
	req5.Header.Set("Accept", "application/json")
	req5.Header.Set("Content-Type", "application/json")

	resp5, err5 := client.Do(req5)
	if err5 != nil {
		fmt.Println(err5)
	}
	defer resp5.Body.Close()

	fmt.Println("response Status:", resp5.Status)
	body5, _ := ioutil.ReadAll(resp5.Body)
	fmt.Println("response Body:", string(body5))

	// Create PEM Master Rules

	url6 := "https://" + vip + "/0/pem_master/log_rule/" + domainID
	fmt.Println("URL:>", url6)

	var jsonStr6 = []byte(`{"rule": {
					"rule_` + domainID + `": {
						"pgtag1": "` + domainID + `",
						"log_ifc_type": "ACCESS_VM"}}}`)
	req6, err := http.NewRequest("PUT", url6, bytes.NewBuffer(jsonStr6))
	req6.Header.Set("Accept", "application/json")
	req6.Header.Set("Content-Type", "application/json")

	resp6, err6 := client.Do(req6)
	if err6 != nil {
		fmt.Println(err6)
	}
	defer resp6.Body.Close()

	fmt.Println("response Status:", resp6.Status)
	body6, _ := ioutil.ReadAll(resp6.Body)
	fmt.Println("response Body:", string(body6))
}

func pgVDDelete(domainID string) {
	cookieJar, _ := cookiejar.New(nil)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Jar:       cookieJar,
		Transport: tr,
	}

	url := "https://" + vip + "/0/login"
	Log.Infof("URL:> %s", url)

	var jsonStr = []byte(`{"userName":"` + username + `", "password":"` + password + `"}`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	fmt.Println("response Cookies:", resp.Cookies())
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	url0 := "https://" + vip + "/0/connectivity/domain?configonly=true"
	fmt.Println("URL:>", url0)

	req0, err := http.NewRequest("GET", url0, nil)
	req0.Header.Set("Accept", "application/json")

	resp0, err0 := client.Do(req0)
	if err0 != nil {
		fmt.Println(err0)
	}
	defer resp0.Body.Close()
	body0, _ := ioutil.ReadAll(resp0.Body)
	var domain_data map[string]interface{}
	err0 = json.Unmarshal([]byte(body0), &domain_data)
	if err0 != nil {
		panic(err0)
	}
	for domains, domain_val := range domain_data {
		if domains == domainID {
			res := domain_val.(map[string]interface{})["ne"]
			if len(res.(map[string]interface{})) == 0 {
				fmt.Println("-----------deleted--------------")

				// PLUMgrid Tenant Manager Configuration

				url2 := "https://" + vip + "/0/tenant_manager/tenants" + "/" + domainID
				Log.Infof("URL:>", url2)

				req2, _ := http.NewRequest("DELETE", url2, nil)
				req2.Header.Set("Accept", "application/json")
				req2.Header.Set("Content-Type", "application/json")

				resp2, err2 := client.Do(req2)
				if err2 != nil {
					fmt.Println(err2)
				}
				defer resp2.Body.Close()

				fmt.Println("response Status:", resp2.Status)
				//fmt.Println("response Headers:", resp2.Header)
				body2, _ := ioutil.ReadAll(resp2.Body)
				fmt.Println("response Body:", string(body2))

				// PLUMgrid Tunnel Configuration

				url3 := "https://" + vip + "/0/tunnel_service/vnd_config/" + domainID
				fmt.Println("URL:>", url3)

				req3, _ := http.NewRequest("DELETE", url3, nil)
				req3.Header.Set("Accept", "application/json")
				req3.Header.Set("Content-Type", "application/json")

				resp3, err3 := client.Do(req3)
				if err3 != nil {
					fmt.Println(err3)
				}
				defer resp3.Body.Close()

				fmt.Println("response Status:", resp3.Status)
				//fmt.Println("response Headers:", resp3.Header)
				body3, _ := ioutil.ReadAll(resp3.Body)
				fmt.Println("response Body:", string(body3))

				// Create VD in Connectivity Manager

				url4 := "https://" + vip + "/0/connectivity/domain/" + domainID
				fmt.Println("URL:>", url4)

				req4, _ := http.NewRequest("DELETE", url4, nil)
				req4.Header.Set("Accept", "application/json")
				req4.Header.Set("Content-Type", "application/json")

				resp4, err4 := client.Do(req4)
				if err4 != nil {
					fmt.Println(err4)
				}
				defer resp4.Body.Close()

				fmt.Println("response Status:", resp4.Status)
				body4, _ := ioutil.ReadAll(resp4.Body)
				fmt.Println("response Body:", string(body4))

				// Create VD in Domain Prop

				url5 := "https://" + vip + "/0/connectivity/domain_prop/" + domainID
				fmt.Println("URL:>", url5)

				req5, _ := http.NewRequest("DELETE", url5, nil)
				req5.Header.Set("Accept", "application/json")
				req5.Header.Set("Content-Type", "application/json")

				resp5, err5 := client.Do(req5)
				if err5 != nil {
					fmt.Println(err5)
				}
				defer resp5.Body.Close()

				fmt.Println("response Status:", resp5.Status)
				body5, _ := ioutil.ReadAll(resp5.Body)
				fmt.Println("response Body:", string(body5))

				// Create PEM Master Rules

				url6 := "https://" + vip + "/0/pem_master/log_rule/" + domainID
				fmt.Println("URL:>", url6)

				req6, _ := http.NewRequest("DELETE", url6, nil)
				req6.Header.Set("Accept", "application/json")
				req6.Header.Set("Content-Type", "application/json")

				resp6, err6 := client.Do(req6)
				if err6 != nil {
					fmt.Println(err6)
				}
				defer resp6.Body.Close()

				fmt.Println("response Status:", resp6.Status)
				body6, _ := ioutil.ReadAll(resp6.Body)
				fmt.Println("response Body:", string(body6))
			}
		}
	}
}

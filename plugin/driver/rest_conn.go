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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"

	Log "github.com/Sirupsen/logrus"
)

func rest_Call(method string, uri string, data []byte) ([]byte, error) {

	url := "https://" + vip

	cookieJar, _ := cookiejar.New(nil)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Jar:       cookieJar,
		Transport: tr,
	}

	login_url := url + "/0/login"

	var cred_data = []byte(`{"userName":"` + username + `", "password":"` + password + `"}`)
	req, err := http.NewRequest("POST", login_url, bytes.NewBuffer(cred_data))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		Log.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	Log.Info("response Status:", resp.Status)
	resp_body, _ := ioutil.ReadAll(resp.Body)
	Log.Info("response Body:", string(resp_body))

	if resp.Status != "200 OK" {
		Log.Error(resp_body)
		return nil, fmt.Errorf("PLUMgrid service not available.")
	}

	// REST Call

	rest_url := url + uri
	Log.Info("URL:>", rest_url)

	req, err = http.NewRequest(method, rest_url, bytes.NewBuffer(data))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	if err != nil {
		Log.Error(err)
		return nil, err
	}
	defer resp.Body.Close()

	Log.Info("response Status:", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	Log.Info("response Body:", string(body))

	if resp.Status != "200 OK" {
		Log.Error(body)
		return nil, fmt.Errorf("PLUMgrid service not available.")
	}

	return body, nil
}

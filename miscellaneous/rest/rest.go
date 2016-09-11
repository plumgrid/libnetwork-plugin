package rest

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
)

type PgRestHandle struct {
	Ip_or_host string
	Port       int
	User       string
	Password   string

	client_cookies *cookiejar.Jar
	httpClient     *http.Client
}

func CreatePGRestClient(rest_ip string, rest_Port int, Username string, Password string) *PgRestHandle {

	cookies, _ := cookiejar.New(nil)

	return &PgRestHandle{
		Ip_or_host: rest_ip,
		Port:       rest_Port,
		User:       Username,
		Password:   Password,

		client_cookies: cookies,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Jar: cookies,
		},
	}
}

func RestGet(handle *PgRestHandle, path string) (error, int, []byte) {

	req, _ := http.NewRequest("GET", path, nil)

	req.Header.Set("Accept", "application/json")

	res, err := handle.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to login on PG plat: %v", err), 0, nil
	}

	// Read body data
	body_data, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		return fmt.Errorf("Error reading body data %v", err), 0, nil
	}

	return nil, res.StatusCode, body_data
}

func RestPost(handle *PgRestHandle, path string, data string) (error, int, []byte) {

	req, _ := http.NewRequest("POST", path, bytes.NewBufferString(data))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := handle.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Faild to POST: %v", err), 0, nil
	}

	// Read body data
	body_data, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		return fmt.Errorf("Error reading body data %v", err), 0, nil
	}

	return nil, res.StatusCode, body_data
}

func RestPut(handle *PgRestHandle, path string, data string) (error, int, []byte) {

	req, _ := http.NewRequest("PUT", path, bytes.NewBufferString(data))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := handle.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Faild to PUT: %v", err), 0, nil
	}

	// Read body data
	body_data, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		return fmt.Errorf("Error reading body data %v", err), 0, nil
	}

	return nil, res.StatusCode, body_data
}

func RestDelete(handle *PgRestHandle, path string) (error, int, []byte) {

	req, _ := http.NewRequest("DELETE", path, nil)

	req.Header.Set("Accept", "application/json")

	res, err := handle.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Failed to DELETE: %v", err), 0, nil
	}

	// Read body data
	body_data, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		return fmt.Errorf("Error reading body data %v", err), 0, nil
	}

	return nil, res.StatusCode, body_data
}

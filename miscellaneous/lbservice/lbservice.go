package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"

	pg_rest "github.com/msagheer/miscellaneous/rest"
)

func main() {

	var net, lb string
	var del bool
	flag.StringVar(&net, "network", "", "Network ID")
	flag.StringVar(&lb, "lb", "", "loab balancer name")
	flag.BoolVar(&del, "delete", false, "delete the loab balancer")

	flag.Parse()

	pg_ip := os.Getenv("PG_API_IP")
	pg_port := os.Getenv("PG_API_PORT")
	pg_username := os.Getenv("PG_USERNAME")
	pg_password := os.Getenv("PG_PASSWORD")

	i, err := strconv.Atoi(pg_port)
	rest_handle := pg_rest.CreatePGRestClient(pg_ip, i, pg_username, pg_password)
	if err := pg_rest.AttemptLogin(rest_handle); err != nil {
		fmt.Printf("Login Failed: %v", err)
	}

	domain, netID := FindDomainFromNetwork(rest_handle, net)

	url := pg_rest.GetRestPath(rest_handle) + "/v0/vd/" + domain + "/lbaas/" + lb

	if !del {
		fmt.Printf("Creating Load Balancer (%v) in domain (%v) against bridge (%v)\n", lb, domain, netID)
		interface1 := make([]interface{}, 0)
		interface1 = append(interface1, map[string]interface{}{
			"name":       "vip",
			"vip":        "0.0.0.0",
			"attachment": netID,
			"port_id":    "123-456-7890"})

		data := map[string]interface{}{
			"origin":       "openstack",
			"display_name": lb,
			"interfaces":   interface1}

		a, _ := json.Marshal(data)
		err, status_code, _ := pg_rest.RestPut(rest_handle, url, string(a))
		if err != nil {
			fmt.Printf("Error while posting VD : %v", err)
		} else if status_code != 200 {
			fmt.Printf("Failed to put VND domain: %s", status_code)
		}
	} else if del {
		fmt.Printf("Deleting Load Balancer (%v) in domain (%v) against bridge (%v)\n", lb, domain, netID)
		err, status_code, _ := pg_rest.RestDelete(rest_handle, url)
		if err != nil {
			fmt.Printf("Error while deleting VD : %v", err)
		} else if status_code != 200 {
			fmt.Printf("Failed to delete VND domain: %s", status_code)
		}

	}

	if err = pg_rest.AttemptLogout(rest_handle); err != nil {
		fmt.Println("Logout Failed: %v", err)
	}
}

func FindDomainFromNetwork(rest_handle *pg_rest.PgRestHandle, ID string) (domainid string, netid string) {
	url := pg_rest.GetRestPath(rest_handle) + "/0/connectivity/domain?configonly=true&level=3"

	err, status_code, body := pg_rest.RestGet(rest_handle, url)
	if err != nil {
		fmt.Printf("Error while getting VD : %v", err)
	} else if status_code != 200 {
		fmt.Printf("Failed to get VND domain: %s", status_code)
	}

	var domain_data map[string]interface{}
	err = json.Unmarshal([]byte(body), &domain_data)
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

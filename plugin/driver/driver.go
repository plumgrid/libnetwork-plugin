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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"

	Log "github.com/Sirupsen/logrus"
	"github.com/docker/libnetwork/drivers/remote/api"
	"github.com/docker/libnetwork/netlabel"
	docker "github.com/fsouza/go-dockerclient"

	"github.com/gorilla/mux"
	"github.com/vishvananda/netlink"
)

const (
	MethodReceiver = "NetworkDriver"
)

type Driver interface {
	Listen(net.Listener) error
}

type driver struct {
	client  *docker.Client
	version string
}

func New(version string) (Driver, error) {
	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		return nil, fmt.Errorf("could not connect to docker: %s", err)
	}

	return &driver{
		client:  client,
		version: version,
	}, nil
}

func (driver *driver) Listen(socket net.Listener) error {
	router := mux.NewRouter()
	router.NotFoundHandler = http.HandlerFunc(notFound)

	router.Methods("GET").Path("/status").HandlerFunc(driver.status)
	router.Methods("POST").Path("/Plugin.Activate").HandlerFunc(driver.handshake)

	handleMethod := func(method string, h http.HandlerFunc) {
		router.Methods("POST").Path(fmt.Sprintf("/%s.%s", MethodReceiver, method)).HandlerFunc(h)
	}

	handleMethod("GetCapabilities", driver.getCapabilities)
	handleMethod("CreateNetwork", driver.createNetwork)
	handleMethod("DeleteNetwork", driver.deleteNetwork)
	handleMethod("CreateEndpoint", driver.createEndpoint)
	handleMethod("DeleteEndpoint", driver.deleteEndpoint)
	handleMethod("EndpointOperInfo", driver.infoEndpoint)
	handleMethod("Join", driver.joinEndpoint)
	handleMethod("Leave", driver.leaveEndpoint)

	return http.Serve(socket, router)
}

// Handle Functions

type handshakeResp struct {
	Implements []string
}

func (driver *driver) handshake(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(&handshakeResp{
		[]string{"NetworkDriver"},
	})
	if err != nil {
		sendError(w, "encode error", http.StatusInternalServerError)
		Log.Error("handshake encode:", err)
		return
	}
	Log.Infof("Handshake completed")
}

func (driver *driver) status(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, fmt.Sprintln("PLUMgrid Libnetwork-plugin", driver.version))
}

type getcapabilitiesResp struct {
	Scope string
}

func (driver *driver) getCapabilities(w http.ResponseWriter, r *http.Request) {
	err := json.NewEncoder(w).Encode(&getcapabilitiesResp{
		scope,
	})
	if err != nil {
		Log.Fatal("get capability encode:", err)
		sendError(w, "encode error", http.StatusInternalServerError)
		return
	}
	Log.Infof("Get Capabilities completed")
}

// create network call
func (driver *driver) createNetwork(w http.ResponseWriter, r *http.Request) {
	var create api.CreateNetworkRequest
	err := json.NewDecoder(r.Body).Decode(&create)
	if err != nil {
		sendError(w, "Unable to decode JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	Log.Infof("Create network request %+v", &create)

	domainid := create.Options[netlabel.GenericData].(map[string]interface{})["domain"]
	if domainid == nil {
		domainid = default_vd
	}
	DomainCreate(domainid.(string))

	gatewayip := create.IPv4Data[0].Gateway.IP.String()
	neName := create.Options[netlabel.GenericData].(map[string]interface{})["bridge"]

	if neName != nil {
		if err := AddNetworkInfo(create.NetworkID, neName.(string), domainid.(string)); err != nil {
			Log.Error(err)
			DomainDelete(domainid.(string))
			errorResponse(w, fmt.Sprintf("Bridge (%s) doesnot exist in doamin (%s).", neName.(string), domainid.(string)))
			return
		}
		AddGatewayInfo(create.NetworkID, domainid.(string), gatewayip)
		emptyResponse(w)

		return
	}

	router := create.Options[netlabel.GenericData].(map[string]interface{})["router"]
	BridgeCreate(create.NetworkID, domainid.(string), gatewayip)

	if router != nil {
		cidr := create.IPv4Data[0].Pool.String()
		_, ipnet, _ := net.ParseCIDR(cidr)
		tm, _ := hex.DecodeString(ipnet.Mask.String())
		netmask := fmt.Sprintf("%v.%v.%v.%v", tm[0], tm[1], tm[2], tm[3])
		Log.Infof("Adding router interface for : ", router, gatewayip, netmask)
		if err := CreateNetworkLink(router.(string), domainid.(string), create.NetworkID, gatewayip, netmask); err != nil {
			Log.Error(err)
			BridgeDelete(create.NetworkID, domainid.(string))
			DomainDelete(domainid.(string))
			errorResponse(w, fmt.Sprintf("Router (%s) doesnot exist in doamin (%s).", router.(string), domainid.(string)))
			return
		}
	}

	emptyResponse(w)

	Log.Infof("Create network %s", create.NetworkID)
}

// delete network call
func (driver *driver) deleteNetwork(w http.ResponseWriter, r *http.Request) {
	var delete api.DeleteNetworkRequest
	if err := json.NewDecoder(r.Body).Decode(&delete); err != nil {
		sendError(w, "Unable to decode JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	Log.Infof("Delete network request: %+v", &delete)

	domainid, _ := FindDomainFromNetwork(delete.NetworkID)
	if domainid == "" {
		domainid = default_vd
	}
	DeleteNetworkLinks(domainid, delete.NetworkID)
	DeleteAttachedLB(domainid, delete.NetworkID)
	BridgeDelete(delete.NetworkID, domainid)
	DomainDelete(domainid)

	emptyResponse(w)
	Log.Infof("Destroy network %s", delete.NetworkID)
}

// create endpoint call
func (driver *driver) createEndpoint(w http.ResponseWriter, r *http.Request) {
	var create api.CreateEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&create); err != nil {
		sendError(w, "Unable to decode JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}
	Log.Infof("Create endpoint request %+v", &create)

	endID := create.EndpointID

	ip := create.Interface.Address
	Log.Infof("Got IP from IPAM %s", ip)

	local := vethPair(endID[:5])
	if err := netlink.LinkAdd(local); err != nil {
		Log.Error(err)
		errorResponsef(w, "could not create veth pair")
		return
	}
	link, _ := netlink.LinkByName(local.PeerName)
	mac := link.Attrs().HardwareAddr.String()

	resp := &api.CreateEndpointResponse{
		Interface: &api.EndpointInterface{
			MacAddress: mac,
		},
	}

	objectResponse(w, resp)
	Log.Infof("Create endpoint %s %+v", endID, resp)
}

// delete endpoint call
func (driver *driver) deleteEndpoint(w http.ResponseWriter, r *http.Request) {
	var delete api.DeleteEndpointRequest
	if err := json.NewDecoder(r.Body).Decode(&delete); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	Log.Debugf("Delete endpoint request: %+v", &delete)

	local := vethPair(delete.EndpointID[:5])
	if err := netlink.LinkDel(local); err != nil {
		Log.Warningf("unable to delete veth on leave: %s", err)
	}

	emptyResponse(w)
	Log.Infof("Delete endpoint %s", delete.EndpointID)
}

// endpoint info request
func (driver *driver) infoEndpoint(w http.ResponseWriter, r *http.Request) {
	var info api.EndpointInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&info); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	Log.Infof("Endpoint info request: %+v", &info)
	objectResponse(w, &api.EndpointInfoResponse{Value: map[string]interface{}{}})
	Log.Infof("Endpoint info %s", info.EndpointID)
}

// join call
func (driver *driver) joinEndpoint(w http.ResponseWriter, r *http.Request) {
	var j api.JoinRequest
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	Log.Infof("Join request: %+v", &j)

	netID := j.NetworkID
	endID := j.EndpointID
	domainid, bridgeID := FindDomainFromNetwork(netID)
	if domainid == "" {
		domainid = default_vd
	}
	gatewayIP := FindNetworkGateway(domainid, netID)

	local := vethPair(endID[:5])

	if_local_name := "tap" + endID[:5]

	link, _ := netlink.LinkByName(local.PeerName)
	mac := link.Attrs().HardwareAddr.String()
	Log.Infof("mac address: %s\n", mac)

	cmdStr := "sudo /opt/pg/bin/ifc_ctl gateway add_port " + if_local_name
	Log.Infof("addport cmd: %s", cmdStr)
	cmd := exec.Command("/bin/sh", "-c", cmdStr)
	var addport bytes.Buffer
	cmd.Stdout = &addport
	if err := cmd.Run(); err != nil {
		Log.Error("Error thrown: ", err)
	}
	if addport.String() != "" {
		Log.Error(fmt.Errorf(addport.String()))
		errorResponse(w, "Unable to on-board container onto PLUMgrid")
		return
	}

	cmdStr = "sudo /opt/pg/bin/ifc_ctl gateway ifup " + if_local_name + " access_container cont_" + endID[:8] + " " + mac + " pgtag2=" + bridgeID + " pgtag1=" + domainid
	Log.Infof("ifup cmd: %s", cmdStr)
	cmd = exec.Command("/bin/sh", "-c", cmdStr)
	var ifup bytes.Buffer
	cmd.Stdout = &ifup
	if err := cmd.Run(); err != nil {
		Log.Error("Error thrown: ", err)
	}
	if ifup.String() != "" {
		Log.Error(fmt.Errorf(ifup.String()))
		errorResponse(w, "Unable to on-board container onto PLUMgrid")
		return
	}

	if netlink.LinkSetUp(local) != nil {
		errorResponsef(w, `unable to bring veth up`)
		return
	}

	ifname := &api.InterfaceName{
		SrcName:   local.PeerName,
		DstPrefix: "ethpg",
	}

	res := &api.JoinResponse{
		InterfaceName: ifname,
		Gateway:       gatewayIP,
	}

	AddMetaconfig(domainid, bridgeID, j.SandboxKey[22:], endID, mac)

	objectResponse(w, res)
	Log.Infof("Join endpoint %s:%s to %s", j.NetworkID, j.EndpointID, j.SandboxKey)
}

// leave call
func (driver *driver) leaveEndpoint(w http.ResponseWriter, r *http.Request) {
	var l api.LeaveRequest
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		sendError(w, "Could not decode JSON encode payload", http.StatusBadRequest)
		return
	}
	Log.Infof("Leave request: %+v", &l)

	domainid, bridgeID := FindDomainFromNetwork(l.NetworkID)

	if_local_name := "tap" + l.EndpointID[:5]

	cmdStr := "sudo /opt/pg/bin/ifc_ctl gateway ifdown " + if_local_name
	Log.Infof("ifdown cmd: %s", cmdStr)
	cmd := exec.Command("/bin/sh", "-c", cmdStr)
	var ifdown bytes.Buffer
	cmd.Stdout = &ifdown
	if err := cmd.Run(); err != nil {
		Log.Error("Error thrown: ", err)
	}
	if ifdown.String() != "" {
		Log.Error(fmt.Errorf(ifdown.String()))
		errorResponse(w, "Unable to off-board container from PLUMgrid")
	}

	cmdStr = "sudo /opt/pg/bin/ifc_ctl gateway del_port " + if_local_name
	Log.Infof("delport cmd: %s", cmdStr)
	cmd = exec.Command("/bin/sh", "-c", cmdStr)
	var delport bytes.Buffer
	cmd.Stdout = &delport
	if err := cmd.Run(); err != nil {
		Log.Error("Error thrown: ", err)
	}
	Log.Infof("output: %+v\n", delport.String())
	if delport.String() != "" {
		Log.Error(fmt.Errorf(delport.String()))
		errorResponse(w, "Unable to off-board container from PLUMgrid")
	}

	RemoveMetaconfig(domainid, bridgeID, l.EndpointID)

	emptyResponse(w)
	Log.Infof("Leave %s:%s", l.NetworkID, l.EndpointID)
}

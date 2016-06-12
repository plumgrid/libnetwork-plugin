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
	"fmt"
	"net"
	"os/exec"

	Log "github.com/Sirupsen/logrus"
	api "github.com/docker/go-plugins-helpers/network"
	docker "github.com/fsouza/go-dockerclient"

	"github.com/vishvananda/netlink"
)

type Driver struct {
	client  *docker.Client
	version string
}

func New(version string) (*Driver, error) {
	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		return nil, fmt.Errorf("could not connect to docker: %s", err)
	}

	return &Driver{
		client:  client,
		version: version,
	}, nil
}

// GetCapabilities tells libnetwork scope of driver
func (driver *Driver) GetCapabilities() (*api.CapabilitiesResponse, error) {
	scope := &api.CapabilitiesResponse{Scope: scope}

	Log.Infof("Libnetwork Plugin scope is %+v", scope)
	return scope, nil
}

// CreateNetwork creates a new network in PLUMgrid
func (driver *Driver) CreateNetwork(create *api.CreateNetworkRequest) error {
	Log.Infof("Create network request %+v", create)

	gatewayip, _, err := parseGatewayIP(create.IPv4Data[0].Gateway)
	if err != nil {
		return err
	}

	domainid := create.Options["com.docker.network.generic"].(map[string]interface{})["domain"]
	if domainid == nil {
		domainid = default_domain
	}

	if err = pgVDCreate(domainid.(string)); err != nil {
		return err
	}

	if err = pgBridgeCreate(create.NetworkID, domainid.(string), gatewayip); err != nil {
		return err
	}
	return nil
}

// DeleteNetwork deletes a network from PLUMgrid
func (driver *Driver) DeleteNetwork(delete *api.DeleteNetworkRequest) error {
	Log.Infof("Delete network request: %+v", delete)

	domainid, err := FindDomainFromNetwork(delete.NetworkID)
	if err != nil {
		return err
	}
	if domainid == "" {
		domainid = default_domain
	}

	if err = pgBridgeDestroy(delete.NetworkID, domainid); err != nil {
		return err
	}

	if err = pgVDDelete(domainid); err != nil {
		return err
	}
	return nil
}

// CreateEndpoint creates a new Endpoint
func (driver *Driver) CreateEndpoint(create *api.CreateEndpointRequest) (*api.CreateEndpointResponse, error) {
	Log.Infof("Create endpoint request %+v", create)

	ip := create.Interface.Address
	Log.Infof("Got IP from IPAM %s", ip)

	ip_mac, ipnet_mac, err_mac := net.ParseCIDR(ip)
	if err_mac == nil {
		ipnet_mac.IP = ip_mac
	}

	mac := makeMac(ipnet_mac.IP)

	res := &api.CreateEndpointResponse{
		Interface: &api.EndpointInterface{
			MacAddress: mac,
		},
	}

	Log.Info("Create endpoint response: %+v", res)
	return res, nil
}

// DeleteEndpoint deletes an endpoint
func (driver *Driver) DeleteEndpoint(delete *api.DeleteEndpointRequest) error {
	Log.Infof("Delete endpoint request: %+v", delete)
	return nil
}

// EndpointInfo returns information about an endpoint
func (driver *Driver) EndpointInfo(info *api.InfoRequest) (*api.InfoResponse, error) {
	Log.Infof("Endpoint info request: %+v", info)

	res := &api.InfoResponse{
		Value: make(map[string]string),
	}

	Log.Infof("Endpoint info response: %+v", res)
	return res, nil
}

// Join call
func (driver *Driver) Join(join *api.JoinRequest) (*api.JoinResponse, error) {
	Log.Infof("Join request: %+v", join)

	netID := join.NetworkID
	endID := join.EndpointID

	domainid, err := FindDomainFromNetwork(netID)
	if err != nil {
		return nil, err
	}
	if domainid == "" {
		domainid = default_domain
	}

	gatewayIP, err_gw := FindNetworkGateway(domainid, netID)
	if err_gw != nil {
		return nil, err_gw
	}

	// create and attach local name to the bridge
	local := vethPair(endID[:5])
	if err = netlink.LinkAdd(local); err != nil {
		Log.Error(err)
		return nil, err
	}

	if_local_name := "tap" + endID[:5]

	//getting mac address of tap...
	cmdStr0 := "ifconfig " + if_local_name + " | awk '/HWaddr/ {print $NF}'"
	Log.Infof("mac address cmd: %s", cmdStr0)
	cmd0 := exec.Command("/bin/sh", "-c", cmdStr0)
	var out0 bytes.Buffer
	cmd0.Stdout = &out0
	err0 := cmd0.Run()
	if err0 != nil {
		Log.Error("Error thrown: ", err0)
	}
	mac := out0.String()
	Log.Infof("output of cmd: %s\n", mac)

	//first command {adding port on plumgrid}
	cmdStr1 := "sudo /opt/pg/bin/ifc_ctl gateway add_port " + if_local_name
	Log.Infof("second cmd: %s", cmdStr1)
	cmd1 := exec.Command("/bin/sh", "-c", cmdStr1)
	var out1 bytes.Buffer
	cmd1.Stdout = &out1
	err1 := cmd1.Run()
	if err1 != nil {
		Log.Error("Error thrown: ", err1)
	}
	Log.Infof("output of cmd: %+v\n", out1.String())

	//second command {up the port on plumgrid}
	cmdStr2 := "sudo /opt/pg/bin/ifc_ctl gateway ifup " + if_local_name + " access_vm cont_" + endID[:2] + " " + mac[:17] + " pgtag2=bridge-" + netID[:10] + " pgtag1=" + domainid
	Log.Infof("third cmd: %s", cmdStr2)
	cmd2 := exec.Command("/bin/sh", "-c", cmdStr2)
	var out2 bytes.Buffer
	cmd2.Stdout = &out2
	err2 := cmd2.Run()
	if err2 != nil {
		Log.Error("Error thrown: ", err2)
	}
	Log.Infof("output of cmd: %+v\n", out2.String())

	if netlink.LinkSetUp(local) != nil {
		return nil, fmt.Errorf("Unable to bring veth up")
	}

	ifname := &api.InterfaceName{
		SrcName:   local.PeerName,
		DstPrefix: "ethpg",
	}

	res := &api.JoinResponse{
		InterfaceName: *ifname,
		Gateway:       gatewayIP,
	}

	Log.Infof("Join response: %+v", res)
	return res, nil
}

// Leave call
func (driver *Driver) Leave(leave *api.LeaveRequest) error {
	Log.Infof("Leave request: %+v", leave)

	if_local_name := "tap" + leave.EndpointID[:5]

	//getting mac address of tap...
	cmdStr0 := "ifconfig " + if_local_name + " | awk '/HWaddr/ {print $NF}'"
	Log.Infof("mac address cmd: %s", cmdStr0)
	cmd0 := exec.Command("/bin/sh", "-c", cmdStr0)
	var out0 bytes.Buffer
	cmd0.Stdout = &out0
	err0 := cmd0.Run()
	if err0 != nil {
		Log.Error("Error thrown: ", err0)
	}
	mac := out0.String()
	Log.Infof("output of cmd: %s\n", mac)

	//first command {adding port on plumgrid}
	cmdStr1 := "sudo /opt/pg/bin/ifc_ctl gateway ifdown " + if_local_name + " access_vm cont_" + leave.EndpointID[:5] + " " + mac[:17]
	Log.Infof("second cmd: %s", cmdStr1)
	cmd1 := exec.Command("/bin/sh", "-c", cmdStr1)
	var out1 bytes.Buffer
	cmd1.Stdout = &out1
	err1 := cmd1.Run()
	if err1 != nil {
		Log.Error("Error thrown: ", err1)
	}
	Log.Infof("output of cmd: %+v\n", out1.String())

	//second command {up the port on plumgrid}
	cmdStr2 := "sudo /opt/pg/bin/ifc_ctl gateway del_port " + if_local_name
	Log.Infof("third cmd: %s", cmdStr2)
	cmd2 := exec.Command("/bin/sh", "-c", cmdStr2)
	var out2 bytes.Buffer
	cmd2.Stdout = &out2
	err2 := cmd2.Run()
	if err2 != nil {
		Log.Error("Error thrown: ", err2)
	}
	Log.Infof("output of cmd: %+v\n", out2.String())

	local := vethPair(leave.EndpointID[:5])
	if err := netlink.LinkDel(local); err != nil {
		Log.Warningf("unable to delete veth on leave: %s", err)
	}
	return nil
}

// DiscoverNew
func (driver *Driver) DiscoverNew(n *api.DiscoveryNotification) error {
	return nil
}

// DiscoverDelete
func (driver *Driver) DiscoverDelete(d *api.DiscoveryNotification) error {
	return nil
}

// ProgramExternalConnectivity
func (driver *Driver) ProgramExternalConnectivity(r *api.ProgramExternalConnectivityRequest) error {
	return nil
}

// RevokeExternalConnectivity
func (driver *Driver) RevokeExternalConnectivity(r *api.RevokeExternalConnectivityRequest) error {
	return nil
}

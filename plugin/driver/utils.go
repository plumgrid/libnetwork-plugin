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
	"fmt"
	"net"
	"strings"

	"github.com/vishvananda/netlink"
)

// Helper Functions

func vethPair(suffix string) *netlink.Veth {
	return &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{Name: "tap" + suffix},
		PeerName:  "ns" + suffix,
	}
}

func makeMac(ip net.IP) string {
	hw := make(net.HardwareAddr, 6)
	hw[0] = 0x7a
	hw[1] = 0x42
	copy(hw[2:], ip.To4())
	return hw.String()
}

func parseGatewayIP(gatewayIP string) (string, string, error) {
	if gatewayIP == "" {
		return "", "", fmt.Errorf("No Gateway IP Found")
	}

	parts := strings.Split(gatewayIP, "/")
	if parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("Cannot Split Gateway IP Address")
	}

	return parts[0], parts[1], nil
}

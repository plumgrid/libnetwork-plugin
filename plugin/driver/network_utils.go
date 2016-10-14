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

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	Log "github.com/Sirupsen/logrus"
	"github.com/eduser25/cni/pkg/ns"
)

// Plumgrid just specifies the prefix, so we will arping on all ifcs
// that match with the prefix given to libnetwork
func GetIPandGarp(net_ns_path string, mac string, auto_arp bool) (string, error) {

	var netns ns.NetNS
	var err error

	netns, err = ns.GetNS(net_ns_path)
	if err != nil {
		return "", fmt.Errorf("Failed to open netns %s: %v", net_ns_path, err)
	}
	defer netns.Close()

	var ifc_ip string

	// Let's go in
	err = netns.Do(func(_ ns.NetNS) error {

		// Get Ifcs on container
		ifc_list, err := net.Interfaces()
		if err != nil {
			return fmt.Errorf("Failed to get interfaces in Netns: %v", err)
		}

		// Iterate Ifcs
		for _, ifc := range ifc_list {
			// Valid PG-created ifc

			hw_addr := ifc.HardwareAddr
			if hw_addr == nil {
				// Supressing this output, as for loopback is quite spammy
				// Log.Errorf("Failed to get HW Addr for %s: %v", ifc.Name, err)
				continue
			}

			hw_addr_s := hw_addr.String()

			if hw_addr_s != mac {
				// not the ifc we are lookin for
				continue
			}

			address_slice, err := ifc.Addrs()
			if err != nil {
				Log.Errorf("Failed to get Address slice for ifc %s: %s", ifc.Name, err)
				continue
			}

			if len(address_slice) == 0 {
				Log.Printf("Interface %s had no IP address at the moment we checked", ifc.Name)
				continue
			}

			for _, ifc_addr := range address_slice {
				ip, _, err := net.ParseCIDR(ifc_addr.String())
				if err != nil {
					Log.Printf("Failed to parse network %s for ifc %s : %s", ifc_addr.String(), ifc.Name, err)
					continue
				}

				if ip.To4() == nil {
					// ignoring any IPV6 addr
					continue
				}

				// Ip for this IFC found, this will be returned already
				ifc_ip = ip.To4().String()

				// See if we have to garp or just return the IP
				if auto_arp {
					handle, err := pcap.OpenLive(ifc.Name, 65536, true, pcap.BlockForever)
					if err != nil {
						return fmt.Errorf("Failed to open raw socket %s : %s", ifc.Name, err)
					}
					defer handle.Close()

					// Get ARP packet
					arp_pkt := getGratuitousArp(ifc.HardwareAddr, ip)

					Log.Printf("Sending %s GARP on %s : %s",
						ifc.Name, ifc.HardwareAddr.String(), ip.String())

					if err := handle.WritePacketData(arp_pkt); err != nil {
						return fmt.Errorf("Failed to Write GARP in raw socket %s : %s", ifc.Name, err)
					}
				}
			}
		}
		return nil
	})

	return ifc_ip, err
}

func getGratuitousArp(mac net.HardwareAddr, ip net.IP) []byte {
	// Set up all the layers' fields we can.
	eth := layers.Ethernet{
		SrcMAC:       mac,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPReply,
		SourceHwAddress:   mac,
		SourceProtAddress: ip.To4(),
		DstHwAddress:      []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		DstProtAddress:    ip.To4(),
	}
	// Set up buffer and options for serialization.
	buf := gopacket.NewSerializeBuffer()

	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	// Send one packet for every address.
	gopacket.SerializeLayers(buf, opts, &eth, &arp)

	return buf.Bytes()
}

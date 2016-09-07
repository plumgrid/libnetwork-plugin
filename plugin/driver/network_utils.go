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
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"

	Log "github.com/Sirupsen/logrus"
	"github.com/eduser25/cni/pkg/ns"
)

// Plumgrid just specifies the prefix, so we will arping on all ifcs
// that match with the prefix given to libnetwork
func RunContainerArping(net_ns_path string, pg_ifc_prefix string) error {

	Log.Printf("Attempting to acces the container %s NetNS", net_ns_path)
	var netns ns.NetNS
	var err error
	success := false

	for attempts := 0; attempts <= 5; attempts++ {

		netns, err = ns.GetNS(net_ns_path)

		if err != nil {
			Log.Printf("Attempt %d: failed to open netns %s: %v", attempts, net_ns_path, err)
			time.Sleep(1 * time.Second)
		} else {
			success = true
			break
		}
		attempts += 1
	}

	if !success {
		return fmt.Errorf("Attempts exhaused, could not access %s, exiting", net_ns_path)
	}
	defer netns.Close()

	// Let's go in
	err = netns.Do(func(_ ns.NetNS) error {

		// Get Ifcs on container
		ifc_list, err := net.Interfaces()
		if err != nil {
			return fmt.Errorf("Failed to get interfaces in Netns: %v", err)
		}

		// Iterate Ifcs
		for _, ifc := range ifc_list {
			if strings.HasPrefix(ifc.Name, pg_ifc_prefix) {
				// Valid PG-created ifc

				// TODO we might want to add a retry here to make sure,
				// in case the IFC is up but still has no IP, we retry
				address_slice, err := ifc.Addrs()
				if err != nil {
					Log.Errorf("Failed to get Address slice for ifc %s: %v", ifc.Name, err)
					continue
				}

				if len(address_slice) == 0 {
					Log.Errorf("Interface %s had no IP address at the moment we checked", ifc.Name)
					continue
				}

				Log.Printf("Attempting raw socket open on %s", ifc.Name)
				handle, err := pcap.OpenLive(ifc.Name, 65536, true, pcap.BlockForever)
				if err != nil {
					return fmt.Errorf("Failed to open raw socket %s : %v", ifc.Name, err)
				}

				// We will just arping for all IPv4 addresses on this ifc, just in case
				for _, ifc_addr := range address_slice {
					ip, _, err := net.ParseCIDR(ifc_addr.String())
					if err != nil {
						Log.Printf("Failed to parse network %s : %v", ifc_addr.String(), err)
						continue
					}

					if ip.To4() == nil {
						// ignoring IPV6 addr
						continue
					}

					// Get ARP packet
					arp_pkt := getGratuitousArp(ifc.HardwareAddr, ip)

					for i := 0; i <= 3; i++ {
						Log.Printf("Sending %d GARP on %s : %s - %s",
							i, ifc.Name, ifc.HardwareAddr.String(), ip.String())

						if err := handle.WritePacketData(arp_pkt); err != nil {
							handle.Close()
							return fmt.Errorf("Failed to write ARP packet data on %s : %v", ifc.Name, err)
						}
						time.Sleep(750 * time.Millisecond)
					}
				}

				// Close the socket
				handle.Close()
			}
		}
		return nil
	})

	return err
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

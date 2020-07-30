// Copyright 2012 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Author: jsing@google.com (Joel Sing)

/*
Package client implements the client interface to the Seesaw v2
Network Control Centre component, which provides an interface for the
Seesaw engine to manipulate and control network related configuration.
*/
package client

import (
	"net"

	"github.com/google/seesaw/common/seesaw"
	"github.com/google/seesaw/ipvs"
	ncctypes "github.com/google/seesaw/ncc/types"
	"github.com/google/seesaw/quagga"
)

// NCC provides a client interface to the network control component.
type NCC interface {
	// NewLBInterface returns an initialised NCC network LB interface.
	NewLBInterface(name string, cfg *ncctypes.LBConfig) LBInterface

	// Close closes the connection to the Seesaw NCC.
	Close() error

	// ARPSendGratuitous sends a gratuitious ARP message.
	ARPSendGratuitous(iface string, ip net.IP) error

	// BGPConfig returns the configuration for the Quagga BGP daemon.
	BGPConfig() ([]string, error)

	// BGPNeighbors returns the neighbors that the Quagga BGP daemon
	// is peering with.
	BGPNeighbors() ([]*quagga.Neighbor, error)

	// BGPWithdrawAll requests the Quagga BGP daemon to withdraw all
	// configured network advertisements.
	BGPWithdrawAll() error

	// BGPAdvertiseVIP requests the Quagga BGP daemon to advertise the
	// specified VIP.
	BGPAdvertiseVIP(vip net.IP) error

	// BGPWithdrawVIP requests the Quagga BGP daemon to withdraw the
	// specified VIP.
	BGPWithdrawVIP(vip net.IP) error

	// IPVSFlush flushes all services and destinations from the IPVS table.
	IPVSFlush() error

	// IPVSGetServices returns all services configured in the IPVS table.
	IPVSGetServices() ([]*ipvs.Service, error)

	// IPVSGetService returns the service entry currently configured in
	// the kernel IPVS table, which matches the specified service.
	IPVSGetService(svc *ipvs.Service) (*ipvs.Service, error)

	// IPVSAddService adds the specified service to the IPVS table.
	IPVSAddService(svc *ipvs.Service) error

	// IPVSUpdateService updates the specified service in the IPVS table.
	IPVSUpdateService(svc *ipvs.Service) error

	// IPVSDeleteService deletes the specified service from the IPVS table.
	IPVSDeleteService(svc *ipvs.Service) error

	// IPVSAddDestination adds the specified destination to the IPVS table.
	IPVSAddDestination(svc *ipvs.Service, dst *ipvs.Destination) error

	// IPVSUpdateDestination updates the specified destination in
	// the IPVS table.
	IPVSUpdateDestination(svc *ipvs.Service, dst *ipvs.Destination) error

	// IPVSDeleteDestination deletes the specified destination from
	// the IPVS table.
	IPVSDeleteDestination(svc *ipvs.Service, dst *ipvs.Destination) error

	// RouteDefaultIPv4 returns the default route for IPv4 traffic.
	RouteDefaultIPv4() (net.IP, error)
}

// LBInterface provides an interface for manipulating a load balancing
// network interface.
type LBInterface interface {
	// Init initialises the load balancing network interface.
	Init() error

	// Up attempts to bring the network interface up.
	Up() error

	// Down attempts to bring the network interface down.
	Down() error

	// AddVserver adds a Vserver to the load balancing interface.
	AddVserver(*seesaw.Vserver, seesaw.AF) error

	// Delete Vserver removes a Vserver from the load balancing interface.
	DeleteVserver(*seesaw.Vserver, seesaw.AF) error

	// AddVIP adds a VIP to the load balancing interface.
	AddVIP(*seesaw.VIP) error

	// DeleteVIP removes a VIP from the load balancing interface.
	DeleteVIP(*seesaw.VIP) error

	// AddVLAN adds a VLAN interface to the load balancing interface.
	AddVLAN(vlan *seesaw.VLAN) error

	// DeleteVLAN removes a VLAN interface from the load balancing interface.
	DeleteVLAN(vlan *seesaw.VLAN) error
}

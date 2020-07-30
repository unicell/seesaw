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
	"fmt"
	"net"
	"net/rpc"
	"time"

	"github.com/google/seesaw/common/seesaw"
	"github.com/google/seesaw/ipvs"
	ncctypes "github.com/google/seesaw/ncc/types"
	"github.com/google/seesaw/quagga"
)

const nccMaxRetries = 3
const nccRetryDelay = 500 * time.Millisecond
const nccRPCTimeout = 20 * time.Second

// nccClient implements NCC interface.
type nccClient struct {
	client    *rpc.Client
	nccSocket string
}

// NewNCC returns an initialised NCC client.
// Only unix socket is supported.
// It keeps using the same single connection for RPC assuming underlying unix socket connection is stable.
// But caller should be prepared for failures caused by connection disruptions. The client itself doesn't handle reconnecting.
func NewNCC(socket string) (*nccClient, error) {
	ncc := &nccClient{nccSocket: socket}
	if err := ncc.dial(); err != nil {
		return nil, err
	}
	return ncc, nil
}

func (nc *nccClient) NewLBInterface(name string, cfg *ncctypes.LBConfig) LBInterface {
	iface := &ncctypes.LBInterface{LBConfig: *cfg}
	iface.Name = name
	return &nccLBIface{nc: nc, nccLBInterface: iface}
}

// call performs an RPC call to the Seesaw v2 nccClient.
func (nc *nccClient) call(name string, in interface{}, out interface{}) error {
	client := nc.client
	if client == nil {
		return fmt.Errorf("Not connected")
	}

	replyCall := client.Go(name, in, out, nil)

	select {
	case <-replyCall.Done:
		return replyCall.Error
	case <-time.After(nccRPCTimeout):
		return fmt.Errorf("RPC call timed out after %v", nccRPCTimeout)
	}
}

func (nc *nccClient) dial() error {
	var err error
	for i := 0; i < nccMaxRetries; i++ {
		nc.client, err = rpc.Dial("unix", nc.nccSocket)
		if err == nil {
			return nil
		}
		time.Sleep(time.Duration(i+1) * nccRetryDelay)
	}
	return fmt.Errorf("Failed to establish connection: %v", err)
}

func (nc *nccClient) Close() error {
	if nc.client == nil {
		return nil
	}
	if err := nc.client.Close(); err != nil {
		return fmt.Errorf("Failed to close connection: %v", err)
	}
	nc.client = nil
	return nil
}

func (nc *nccClient) ARPSendGratuitous(iface string, ip net.IP) error {
	arp := ncctypes.ARPGratuitous{
		IfaceName: iface,
		IP:        ip,
	}
	return nc.call("SeesawNCC.ARPSendGratuitous", &arp, nil)
}

func (nc *nccClient) BGPConfig() ([]string, error) {
	bc := &ncctypes.BGPConfig{}
	if err := nc.call("SeesawNCC.BGPConfig", 0, bc); err != nil {
		return nil, err
	}
	return bc.Config, nil
}

func (nc *nccClient) BGPNeighbors() ([]*quagga.Neighbor, error) {
	bn := &ncctypes.BGPNeighbors{}
	if err := nc.call("SeesawNCC.BGPNeighbors", 0, bn); err != nil {
		return nil, err
	}
	return bn.Neighbors, nil
}

func (nc *nccClient) BGPWithdrawAll() error {
	return nc.call("SeesawNCC.BGPWithdrawAll", 0, nil)
}

// TODO(ncope): Use seesaw.VIP here for consistency
func (nc *nccClient) BGPAdvertiseVIP(vip net.IP) error {
	return nc.call("SeesawNCC.BGPAdvertiseVIP", vip, nil)
}

// TODO(ncope): Use seesaw.VIP here for consistency
func (nc *nccClient) BGPWithdrawVIP(vip net.IP) error {
	return nc.call("SeesawNCC.BGPWithdrawVIP", vip, nil)
}

func (nc *nccClient) IPVSFlush() error {
	return nc.call("SeesawNCC.IPVSFlush", 0, nil)
}

func (nc *nccClient) IPVSGetServices() ([]*ipvs.Service, error) {
	s := &ncctypes.IPVSServices{}
	if err := nc.call("SeesawNCC.IPVSGetServices", 0, s); err != nil {
		return nil, err
	}
	return s.Services, nil
}

func (nc *nccClient) IPVSGetService(svc *ipvs.Service) (*ipvs.Service, error) {
	s := &ncctypes.IPVSServices{}
	if err := nc.call("SeesawNCC.IPVSGetService", svc, s); err != nil {
		return nil, err
	}
	return s.Services[0], nil
}

func (nc *nccClient) IPVSAddService(svc *ipvs.Service) error {
	return nc.call("SeesawNCC.IPVSAddService", svc, nil)
}

func (nc *nccClient) IPVSUpdateService(svc *ipvs.Service) error {
	return nc.call("SeesawNCC.IPVSUpdateService", svc, nil)
}

func (nc *nccClient) IPVSDeleteService(svc *ipvs.Service) error {
	return nc.call("SeesawNCC.IPVSDeleteService", svc, nil)
}

func (nc *nccClient) IPVSAddDestination(svc *ipvs.Service, dst *ipvs.Destination) error {
	ipvsDst := ncctypes.IPVSDestination{Service: svc, Destination: dst}
	return nc.call("SeesawNCC.IPVSAddDestination", ipvsDst, nil)
}

func (nc *nccClient) IPVSUpdateDestination(svc *ipvs.Service, dst *ipvs.Destination) error {
	ipvsDst := ncctypes.IPVSDestination{Service: svc, Destination: dst}
	return nc.call("SeesawNCC.IPVSUpdateDestination", ipvsDst, nil)
}

func (nc *nccClient) IPVSDeleteDestination(svc *ipvs.Service, dst *ipvs.Destination) error {
	ipvsDst := ncctypes.IPVSDestination{Service: svc, Destination: dst}
	return nc.call("SeesawNCC.IPVSDeleteDestination", ipvsDst, nil)
}

func (nc *nccClient) RouteDefaultIPv4() (net.IP, error) {
	var ip net.IP
	err := nc.call("SeesawNCC.RouteDefaultIPv4", 0, &ip)
	return ip, err
}

// nccLBIface contains the data needed to manipulate a LB network interface.
type nccLBIface struct {
	nc             *nccClient
	nccLBInterface *ncctypes.LBInterface
}

func (iface *nccLBIface) Init() error {
	return iface.nc.call("SeesawNCC.LBInterfaceInit", iface.nccLBInterface, nil)
}

func (iface *nccLBIface) Up() error {
	return iface.nc.call("SeesawNCC.LBInterfaceUp", iface.nccLBInterface, nil)
}

func (iface *nccLBIface) Down() error {
	return iface.nc.call("SeesawNCC.LBInterfaceDown", iface.nccLBInterface, nil)
}

func (iface *nccLBIface) AddVserver(v *seesaw.Vserver, af seesaw.AF) error {
	lbVserver := ncctypes.LBInterfaceVserver{
		Iface:   iface.nccLBInterface,
		Vserver: v,
		AF:      af,
	}
	return iface.nc.call("SeesawNCC.LBInterfaceAddVserver", &lbVserver, nil)
}

func (iface *nccLBIface) DeleteVserver(v *seesaw.Vserver, af seesaw.AF) error {
	lbVserver := ncctypes.LBInterfaceVserver{
		Iface:   iface.nccLBInterface,
		Vserver: v,
		AF:      af,
	}
	return iface.nc.call("SeesawNCC.LBInterfaceDeleteVserver", &lbVserver, nil)
}

func (iface *nccLBIface) AddVIP(vip *seesaw.VIP) error {
	lbVIP := ncctypes.LBInterfaceVIP{
		Iface: iface.nccLBInterface,
		VIP:   vip,
	}
	return iface.nc.call("SeesawNCC.LBInterfaceAddVIP", &lbVIP, nil)
}

func (iface *nccLBIface) DeleteVIP(vip *seesaw.VIP) error {
	lbVIP := ncctypes.LBInterfaceVIP{
		Iface: iface.nccLBInterface,
		VIP:   vip,
	}
	return iface.nc.call("SeesawNCC.LBInterfaceDeleteVIP", &lbVIP, nil)
}

func (iface *nccLBIface) AddVLAN(vlan *seesaw.VLAN) error {
	lbVLAN := ncctypes.LBInterfaceVLAN{Iface: iface.nccLBInterface, VLAN: vlan}
	return iface.nc.call("SeesawNCC.LBInterfaceAddVLAN", &lbVLAN, nil)
}

func (iface *nccLBIface) DeleteVLAN(vlan *seesaw.VLAN) error {
	lbVLAN := ncctypes.LBInterfaceVLAN{Iface: iface.nccLBInterface, VLAN: vlan}
	return iface.nc.call("SeesawNCC.LBInterfaceDeleteVLAN", &lbVLAN, nil)
}

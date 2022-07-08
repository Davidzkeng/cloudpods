// Copyright 2019 Yunion
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

package winstack

import (
	"fmt"
	"strings"
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/multicloud"
	"yunion.io/x/onecloud/pkg/util/rbacutils"
	"yunion.io/x/pkg/errors"
	"yunion.io/x/pkg/util/netutils"
)

const (
	NETWORKS_LIST_URL  = "/api/network/vpcs/%s/networks"
	NETWORK_CREATE_URL = "/api/network/vpcs/%s/networks"
	NETWORK_DELETE_URL = "/api/network/vpcs/%s/networks/%s/delete"
)

type SNetwork struct {
	multicloud.SResourceBase
	multicloud.STagBase
	wire *SWire

	Id       string
	SubNetId string
	Name     string
	Cidr     string
	IpRange  string
	Gateway  string
}

func (s *SNetwork) GetId() string {
	return s.Id
}

func (s *SNetwork) GetName() string {
	if len(s.Name) > 0 {
		return s.Name
	}
	return s.Id
}

func (s *SNetwork) GetGlobalId() string {
	return s.Id
}

func (s *SNetwork) GetStatus() string {
	return api.NETWORK_STATUS_AVAILABLE
}

func (s *SNetwork) GetProjectId() string {
	return ""
}

func (s *SNetwork) GetIWire() cloudprovider.ICloudWire {
	return s.wire
}

func (s *SNetwork) GetIpStart() string {
	var startIp string
	ips := strings.Split(s.IpRange, "-")
	if len(ips) > 0 {
		startIp = ips[0]
	}
	return startIp
}

func (s *SNetwork) GetIpEnd() string {
	var endIp string
	ips := strings.Split(s.IpRange, "-")
	if len(ips) > 1 {
		endIp = ips[1]
	}
	return endIp
}

func (s *SNetwork) GetIpMask() int8 {
	pref, _ := netutils.NewIPV4Prefix(s.Cidr)
	return pref.MaskLen
}

func (s *SNetwork) GetGateway() string {
	return s.Gateway
}

func (s *SNetwork) GetServerType() string {
	return api.NETWORK_TYPE_GUEST
}

func (s *SNetwork) GetPublicScope() rbacutils.TRbacScope {
	return rbacutils.ScopeDomain
}

func (s *SNetwork) Delete() error {
	URL := fmt.Sprintf(NETWORK_DELETE_URL, s.wire.vpc.Id, s.Id)
	_, err := s.wire.cluster.region.client.invokePOST(URL, nil, nil, nil)
	return err
}

func (s *SNetwork) GetAllocTimeoutSeconds() int {
	return 300
}

func (s *SNetwork) Contains(ip string) bool {
	start, err := netutils.NewIPV4Addr(s.GetIpStart())
	if err != nil {
		return false
	}
	end, err := netutils.NewIPV4Addr(s.GetIpEnd())
	if err != nil {
		return false
	}
	addr, err := netutils.NewIPV4Addr(ip)
	if err != nil {
		return false
	}
	return netutils.NewIPV4AddrRange(start, end).Contains(addr)
}

func (s *SWire) GetINetworkById(id string) (cloudprovider.ICloudNetwork, error) {
	return s.getNetworkById(id)
}

func (s *SWire) getNetworkById(id string) (*SNetwork, error) {
	networks, err := s.vpc.region.GetNetworks(s.vpc.Id)
	if err != nil {
		return nil, err
	}
	for i := range networks {
		if networks[i].GetGlobalId() == id {
			networks[i].wire = s
			return &networks[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, id)
}

func (s *SWire) GetINetworks() ([]cloudprovider.ICloudNetwork, error) {
	networks, err := s.vpc.region.GetNetworks(s.vpc.Id)
	if err != nil {
		return nil, err
	}
	var ret []cloudprovider.ICloudNetwork
	for i := range networks {
		networks[i].wire = s
		ret = append(ret, &networks[i])
	}
	return ret, err
}

func (s *SRegion) GetNetworks(vpcId string) ([]SNetwork, error) {
	URL := fmt.Sprintf(NETWORKS_LIST_URL, vpcId)
	resp, err := s.client.invokeGET(URL, nil, nil)
	if err != nil {
		return nil, err
	}
	var networks []SNetwork
	return networks, resp.Unmarshal(&networks)
}

func (s *SRegion) GetNetworkById(vpcId string, networkId string) (*SNetwork, error) {
	networks, err := s.GetNetworks(vpcId)
	if err != nil {
		return nil, err
	}
	for i := range networks {
		if networks[i].Id == networkId {
			return &networks[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, networkId)
}

func (s *SRegion) CreateNetwork(vpcId, name string, cidr string, desc string) (*SNetwork, error) {
	URL := fmt.Sprintf(NETWORK_CREATE_URL, vpcId)
	gateway, err := getDefaultGateWay(cidr)
	if err != nil {
		return nil, err
	}
	body := make(map[string]string)
	body["cidr"] = cidr
	body["gateway"] = gateway
	body["ipVersion"] = "4"
	body["name"] = name
	var ret SNetwork
	resp, err := s.client.invokePOST(URL, nil, nil, body)
	if err != nil {
		return nil, err
	}
	return &ret, resp.Unmarshal(&ret)
}

func getDefaultGateWay(cidr string) (string, error) {
	pref, err := netutils.NewIPV4Prefix(cidr)
	if err != nil {
		return "", errors.Wrap(err, "getDefaultGateWay.NewIPV4Prefix")
	}
	startIp := pref.Address.NetAddr(pref.MaskLen) // 0
	startIp = startIp.StepUp()                    // 1
	return startIp.String(), nil
}

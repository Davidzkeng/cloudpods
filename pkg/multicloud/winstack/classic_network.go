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
	CLASSIC_NETWORK_LIST_URL   = "/api/network/networks/external_nets"
	CLASSIC_NETWORK_CREATE_URL = "/api/network/networks/external_nets"
	CLASSIC_NETWORK_DELETE_URL = "/api/network/networks/external_nets/%s/delete"
)

type SClassicNetwork struct {
	multicloud.SResourceBase
	multicloud.STagBase
	wire *SClassicWire

	Id              string `json:"id"`
	Name            string `json:"name"`
	Cidr            string `json:"cidr"`
	Gateway         string `json:"gateway"`
	Type            int    `json:"type"`
	NetType         string `json:"netType"`
	PhysicalNetwork string `json:"physicalNetwork"`
	IpRange         string `json:"ipRange"`
	IpVersion       string `json:"ipVersion"`
	CreateTime      string `json:"createTime"`
}

func (s *SClassicNetwork) GetId() string {
	return s.Id
}

func (s *SClassicNetwork) GetName() string {
	if len(s.Name) > 0 {
		return s.Name
	}
	return s.Id
}

func (s *SClassicNetwork) GetGlobalId() string {
	return s.Id
}

func (s *SClassicNetwork) GetStatus() string {
	return api.NETWORK_STATUS_AVAILABLE
}

func (s *SClassicNetwork) GetProjectId() string {
	return ""
}

func (s *SClassicNetwork) GetIWire() cloudprovider.ICloudWire {
	return s.wire
}

func (s *SClassicNetwork) GetIpStart() string {
	var startIp string
	ips := strings.Split(s.IpRange, "-")
	if len(ips) > 0 {
		startIp = ips[0]
	}
	return startIp
}

func (s *SClassicNetwork) GetIpEnd() string {
	var endIp string
	ips := strings.Split(s.IpRange, "-")
	if len(ips) > 1 {
		endIp = ips[1]
	}
	return endIp
}

func (s *SClassicNetwork) GetIpMask() int8 {
	pref, _ := netutils.NewIPV4Prefix(s.Cidr)
	return pref.MaskLen
}

func (s *SClassicNetwork) GetGateway() string {
	return s.Gateway
}

func (s *SClassicNetwork) GetServerType() string {
	return api.NETWORK_TYPE_EIP
}

func (s *SClassicNetwork) GetPublicScope() rbacutils.TRbacScope {
	return rbacutils.ScopeDomain
}

func (s *SClassicNetwork) Delete() error {
	URL := fmt.Sprintf(CLASSIC_NETWORK_DELETE_URL, s.Id)
	_, err := s.wire.cluster.region.client.invokePOST(URL, nil, nil, nil)
	return err
}

func (s *SClassicNetwork) GetAllocTimeoutSeconds() int {
	return 300
}

func (s *SClassicNetwork) Contains(ip string) bool {
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

func (s *SClassicWire) GetINetworkById(id string) (cloudprovider.ICloudNetwork, error) {
	return s.getNetworkById(id)
}

func (s *SClassicWire) getNetworkById(id string) (*SClassicNetwork, error) {
	networks, err := s.vpc.region.GetClassicNetworks()
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

func (s *SClassicWire) GetINetworks() ([]cloudprovider.ICloudNetwork, error) {
	networks, err := s.vpc.region.GetClassicNetworks()
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

func (s *SRegion) GetClassicNetworks() ([]SClassicNetwork, error) {
	resp, err := s.client.invokeGET(CLASSIC_NETWORK_LIST_URL, nil, nil)
	if err != nil {
		return nil, err
	}
	var ret []SClassicNetwork
	return ret, resp.Unmarshal(&ret)
}

func (s *SRegion) GetClassicNetworkById(id string) (*SClassicNetwork, error) {
	networks, err := s.GetClassicNetworks()
	if err != nil {
		return nil, err
	}
	for i := range networks {
		if networks[i].Id == id {
			return &networks[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, id)
}

func (s *SRegion) CreateClassicNetwork(vpcId, wireName, name string, cidr string, startIp, endIp, gateway string, vlanId int, desc string) (*SClassicNetwork, error) {
	var err error
	if len(gateway) == 0 {
		gateway, err = getDefaultGateWay(cidr)
		if err != nil {
			return nil, err
		}
	}
	body := make(map[string]interface{})
	body["cidr"] = cidr
	body["gateway"] = gateway
	body["ipVersion"] = "4"
	body["name"] = name
	body["netType"] = "vlan"
	body["physicalNetwork"] = wireName
	body["type"] = "1"
	body["segmentationId"] = vlanId
	body["ipRange"] = startIp + "-" + endIp
	var ret SClassicNetwork
	resp, err := s.client.invokePOST(CLASSIC_NETWORK_CREATE_URL, nil, nil, body)
	if err != nil {
		return nil, err
	}
	return &ret, resp.Unmarshal(&ret)
}

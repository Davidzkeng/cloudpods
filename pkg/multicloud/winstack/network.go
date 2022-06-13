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
	NETWORKS_LIST_URL = "/api/network/vpcs/%s/networks"
)

type SNetwork struct {
	multicloud.SResourceBase
	multicloud.STagBase
	wire *SWire

	SubNetId string
	Name     string
	Cidr     string
	IpRange  string
	Gateway  string
}

func (s *SNetwork) GetId() string {
	return s.SubNetId
}

func (s *SNetwork) GetName() string {
	if len(s.Name) > 0 {
		return s.Name
	}
	return s.SubNetId
}

func (s *SNetwork) GetGlobalId() string {
	return s.SubNetId
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
	return cloudprovider.ErrNotImplemented
}

func (s *SNetwork) GetAllocTimeoutSeconds() int {
	return 300
}

func (s *SWire) GetINetworkById(id string) (cloudprovider.ICloudNetwork, error) {
	networks, err := s.vpc.region.GetNetworks(s.vpc.Id)
	if err != nil {
		return nil, err
	}
	for i := range networks {
		if networks[i].GetGlobalId() == id {
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
	return networks, resp.Unmarshal(&networks, "data")
}

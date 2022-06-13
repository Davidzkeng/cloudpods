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
	"yunion.io/x/onecloud/pkg/cloudprovider"
)

const (
	INSTANCE_NIC_LIST = "/api/compute/domains/%s/domainInterfaceInfo"
)

type SInstanceNic struct {
	Bridge        string `json:"bridge"`
	IsEnabled     bool   `json:"isEnabled"`
	BootOrder     int    `json:"bootOrder"`
	Model         int    `json:"model"`
	IsVhostDriver bool   `json:"isVhostDriver"`
	Mac           string `json:"mac"`
	Ip            string `json:"ip"`
	NetworkType   int    `json:"networkType"`
	InterfaceId   string `json:"interfaceId"`
	Vpc           struct {
		VpcId              string   `json:"vpcId"`
		VpcName            string   `json:"vpcName"`
		PrivateNetId       string   `json:"privateNetId"`
		PrivateNetName     string   `json:"privateNetName"`
		SecurityGroupNames []string `json:"securityGroupNames"`
	} `json:"vpc"`
	NetWorkList []struct {
		Id         string `json:"id"`
		NetIpType  int    `json:"netIpType"`
		IsBindIp   bool   `json:"isBindIp"`
		NetGateway string `json:"netGateway"`
		NetMask    string `json:"netMask"`
		IpAddress  string `json:"ipAddress"`
	} `json:"netWorkList"`

	NoIpMacSpoofing bool `json:"noIpMacSpoofing"`
}

func (s *SInstanceNic) GetId() string {
	return s.InterfaceId
}

func (s *SInstanceNic) GetIP() string {
	return s.Ip
}

func (s *SInstanceNic) GetMAC() string {
	return s.Mac
}

func (s *SInstanceNic) InClassicNetwork() bool {
	return false
}

func (s *SInstanceNic) GetDriver() string {
	//1.virtio 2.rtl8139 3.e1000
	switch s.Model {
	case 1:
		return "virtio"
	case 2:
		return "rtl8139"
	case 3:
		return "e1000"
	}
	return "virtio"
}

func (s *SInstanceNic) GetINetworkId() string {
	return s.Vpc.PrivateNetId
}

func (s *SInstanceNic) GetSubAddress() ([]string, error) {
	var ret []string
	for i := range s.NetWorkList {
		ret = append(ret, s.NetWorkList[i].IpAddress)
	}
	return ret, nil
}

func (s *SInstanceNic) AssignNAddress(count int) ([]string, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (s *SInstanceNic) AssignAddress(ipAddrs []string) error {
	return cloudprovider.ErrNotImplemented
}

func (s *SInstanceNic) UnassignAddress(ipAddrs []string) error {
	return cloudprovider.ErrNotImplemented
}

func (s *SRegion) GetInstanceNics(instanceId string) ([]SInstanceNic, error) {
	URL := fmt.Sprintf(INSTANCE_NIC_LIST, instanceId)
	resp, err := s.client.invokeGET(URL, nil, nil)
	if err != nil {
		return nil, err
	}
	var ret []SInstanceNic
	return ret, resp.Unmarshal(&ret, "bridgeInterfaces")
}

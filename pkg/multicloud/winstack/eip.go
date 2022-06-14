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
	"strconv"
	"yunion.io/x/log"
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/multicloud"
	"yunion.io/x/pkg/errors"
)

const (
	FLOAT_IP_LIST = "/api/network/floatIps"
)

type SEip struct {
	multicloud.SEipBase
	multicloud.WinStackTags

	region *SRegion

	Id             string `json:"id"`
	Ip             string `json:"ip"`
	VpcId          string `json:"vpcId"`
	VpcName        string `json:"vpcName"`
	MappingIp      string `json:"mappingIp"`
	BindDevType    string `json:"bindDevType"`
	BindDevName    string `json:"bindDevName"`
	BindDevId      string `json:"bindDevId"`
	ExtNetworkId   string `json:"extNetworkId"`
	ExtNetworkName string `json:"extNetworkName"`
}

func (s *SEip) GetId() string {
	return s.Ip
}

func (s *SEip) GetName() string {
	return s.Ip
}

func (s *SEip) GetGlobalId() string {
	return s.Ip
}

func (s *SEip) GetStatus() string {
	return api.EIP_STATUS_READY
}

func (s *SEip) GetProjectId() string {
	return ""
}

func (s *SEip) GetIpAddr() string {
	return s.Ip
}

func (s *SEip) GetMode() string {
	return api.EIP_MODE_STANDALONE_EIP
}

func (s *SEip) GetINetworkId() string {
	networks, err := s.region.GetNetworks(s.VpcId)
	if err != nil {
		log.Errorf("failed to find vpc id for eip %s(%s), error: %v", s.Ip, s.VpcId, err)
		return ""
	}
	for i := range networks {
		if networks[i].Contains(s.Ip) {
			return networks[i].GetGateway()
		}
	}
	log.Errorf("failed to find eip %s(%s) networkId", s.Ip, s.VpcId)

	return ""
}

func (s *SEip) GetAssociationType() string {
	switch s.BindDevType {
	case "VM":
		return api.EIP_ASSOCIATE_TYPE_SERVER
	}
	return ""
}

func (s *SEip) GetAssociationExternalId() string {
	if s.BindDevType == "VM" {
		return s.BindDevId
	}
	return ""
}

func (s *SEip) GetBandwidth() int {
	return 0
}

func (s *SEip) GetInternetChargeType() string {
	return api.EIP_CHARGE_TYPE_BY_TRAFFIC
}

func (s *SEip) Delete() error {
	return cloudprovider.ErrNotImplemented
}

func (s *SEip) Associate(conf *cloudprovider.AssociateConfig) error {
	return cloudprovider.ErrNotImplemented
}

func (s *SEip) Dissociate() error {
	return cloudprovider.ErrNotImplemented
}

func (s *SEip) ChangeBandwidth(bw int) error {
	return cloudprovider.ErrNotSupported
}

func (s *SRegion) GetIEips() ([]cloudprovider.ICloudEIP, error) {
	ret := make([]cloudprovider.ICloudEIP, 0)
	eips, err := s.getEips()
	if err != nil {
		return nil, err
	}
	for i := range eips {
		eips[i].region = s
		ret = append(ret, &eips[i])
	}
	return ret, nil
}

func (s *SRegion) GetIEipById(eipId string) (cloudprovider.ICloudEIP, error) {
	eip, err := s.getEip(eipId)
	if err != nil {
		return nil, err
	}
	eip.region = s
	return eip, nil
}

func (s *SRegion) getEips() ([]SEip, error) {
	var eips []SEip
	start, size := 0, 10
	for {
		ret, err := s.GetEips("", start, size)
		if err != nil {
			return nil, err
		}
		for i := range ret {
			eips = append(eips, ret[i])
		}
		if len(ret) < size {
			break
		}
		start += size
	}
	return eips, nil
}

func (s *SRegion) getEip(id string) (*SEip, error) {
	eips, err := s.getEips()
	if err != nil {
		return nil, err
	}
	for i := range eips {
		if eips[i].GetGlobalId() == id {
			return &eips[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, id)
}
func (s *SRegion) GetEips(id string, start, size int) ([]SEip, error) {
	query := make(map[string]string)
	if size <= 0 {
		size = 10
	}
	if start < 0 {
		start = 0
	}
	if len(id) > 0 {
		query["id"] = id
		start = 0
	}
	query["start"] = strconv.Itoa(start)
	query["size"] = strconv.Itoa(size)
	resp, err := s.client.invokeGET(FLOAT_IP_LIST, nil, query)
	if err != nil {
		return nil, err
	}
	var ret []SEip

	return ret, resp.Unmarshal(&ret, "data")
}

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
	"strconv"

	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/multicloud"
	"yunion.io/x/pkg/errors"
)

const (
	FLOAT_IP_LIST_URL    = "/api/network/floatIps"
	FLOAT_IP_ASSIGN_URL  = "/api/network/floatIps/assgin"
	FLOAT_IP_BIND_URL    = "/api/network/floatIps/%s/bind"
	FLOAT_IP_UMBIND_URL  = "/api/network/floatIps/%s/unbind"
	FLOAT_IP_RELEASE_URL = "/api/network/floatIps/%s/release"
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
	return s.Id
}

func (s *SEip) GetName() string {
	return s.Ip
}

func (s *SEip) GetGlobalId() string {
	return s.Id
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
	network, err := s.region.GetClassicNetworkById(s.ExtNetworkId)
	if err != nil {
		log.Errorf("failed to find extNetwork %s for eip (%s), error: %v", s.ExtNetworkId, s.Ip, err)
		return ""
	}
	if network.Contains(s.Ip) {
		return network.GetGlobalId()
	}

	log.Errorf("failed to find eip %s(%s) extNetworkId", s.Ip, s.ExtNetworkId)
	return ""
}

func (s *SEip) GetIVpcId() string {
	vpc, err := s.region.GetIVpcById(s.VpcId)
	if err != nil {
		return ""
	}
	return vpc.GetGlobalId()
}

func (s *SEip) GetAssociationType() string {
	switch s.BindDevType {
	case "VM":
		return api.EIP_ASSOCIATE_TYPE_SERVER
	case "ROUTER":
		return api.EIP_ASSOCIATE_TYPE_ROUTETABLE

	}
	return ""
}

func (s *SEip) GetAssociationExternalId() string {
	switch s.BindDevType {
	case "VM", "ROUTER":
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
	URL := fmt.Sprintf(FLOAT_IP_RELEASE_URL, s.Id)
	_, err := s.region.client.invokePOST(URL, nil, nil, nil)
	return err
}

func (s *SEip) Associate(conf *cloudprovider.AssociateConfig) error {
	URL := fmt.Sprintf(FLOAT_IP_BIND_URL, s.Id)
	body := make(map[string]string)

	switch conf.AssociateType {
	case api.EIP_ASSOCIATE_TYPE_SERVER:
		body["deviceType"] = "VM"
	default:
		return cloudprovider.ErrNotImplemented
	}
	//通过instance_id 查询端口，使用端口分配eip
	nics, err := s.region.GetInstanceNics(conf.InstanceId)
	if err != nil {
		return err
	}
	if len(nics) == 0 {
		return errors.Errorf("GetInstanceNics is empty")
	}
	//查询instanceId的名词
	instance, err := s.region.GetInstanceById(conf.InstanceId)
	if err != nil {
		return err
	}
	for i := range nics {
		body["vpcId"] = nics[i].Vpc.VpcId
		portName := nics[i].InterfaceId
		body["portName"] = portName

	}
	body["deviceId"] = conf.InstanceId
	body["deviceName"] = instance.Name
	_, err = s.region.client.invokePOST(URL, nil, nil, body)
	if err != nil {
		return err
	}
	return nil
}

func (s *SEip) Dissociate() error {
	URL := fmt.Sprintf(FLOAT_IP_UMBIND_URL, s.Id)
	_, err := s.region.client.invokePOST(URL, nil, nil, nil)
	return err
}

func (s *SEip) ChangeBandwidth(bw int) error {
	return cloudprovider.ErrNotSupported
}

func (s *SEip) Refresh() error {
	newEip, err := s.region.getEipByIp(s.Ip)
	if err != nil {
		return err
	}
	return jsonutils.Update(s, newEip)
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
	eips, err := s.getEips()
	if err != nil {
		return nil, err
	}
	for i := range eips {
		if eips[i].GetGlobalId() == eipId {
			eips[i].region = s
			return &eips[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, eipId)

}

func (s *SRegion) getEips() ([]SEip, error) {
	var eips []SEip
	start, size := 1, 10
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
		start += 1
	}
	return eips, nil
}

func (s *SRegion) getEipByIp(ip string) (*SEip, error) {
	eips, err := s.GetEips(ip, 0, 1)
	if err != nil {
		return nil, err
	}
	for i := range eips {
		if eips[i].Ip == ip {
			return &eips[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, ip)
}

func (s *SRegion) GetEipByBindId(bindId string) (*SEip, error) {
	eips, err := s.getEips()
	if err != nil {
		return nil, err
	}
	for i := range eips {
		if eips[i].BindDevId == bindId {
			return &eips[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, bindId)
}

func (s *SRegion) GetEips(ip string, start, size int) ([]SEip, error) {
	query := make(map[string]string)
	if size <= 0 {
		size = 10
	}
	if start < 0 {
		start = 0
	}
	if len(ip) > 0 {
		query["ip"] = ip
		start = 0
	}

	query["start"] = strconv.Itoa(start)
	query["size"] = strconv.Itoa(size)
	resp, err := s.client.invokeGET(FLOAT_IP_LIST_URL, nil, query)
	if err != nil {
		return nil, err
	}
	var ret []SEip

	return ret, resp.Unmarshal(&ret, "data")
}

func (s *SRegion) CreateEIP(opts *cloudprovider.SEip) (cloudprovider.ICloudEIP, error) {
	classicNetwork, err := s.GetClassicNetworkById(opts.NetworkExternalId)
	if err != nil {
		return nil, err
	}
	if classicNetwork == nil {
		return nil, errors.Errorf("external_net is empty")
	}
	body := make(map[string]string)
	body["extNetId"] = classicNetwork.Id
	body["ip"] = opts.IP
	body["vpcId"] = opts.VpcExternalId
	resp, err := s.client.invokePOST(FLOAT_IP_ASSIGN_URL, nil, nil, body)

	if err != nil {
		return nil, err
	}
	var ret SEip
	ret.region = s
	return &ret, resp.Unmarshal(&ret)
}

type SExternalNet struct {
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

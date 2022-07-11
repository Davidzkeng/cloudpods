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

	"github.com/pkg/errors"
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/multicloud"
)

const (
	VPC_LIST_URL   = "/api/network/vpcs"
	VPC_CREATE_URL = "/api/network/vpcs"
	VPC_DELETE_URL = "/api/network/vpcs/%s/delete"
)

type SVpc struct {
	multicloud.SVpc
	multicloud.WinStackTags

	region *SRegion

	iwires []cloudprovider.ICloudWire

	Id   string
	Name string
}

func (s *SVpc) GetIsDefault() bool {
	return true
}

func (s *SVpc) GetCidrBlock() string {
	return ""
}

func (s *SVpc) GetISecurityGroups() ([]cloudprovider.ICloudSecurityGroup, error) {
	groups, err := s.region.getSecurityGroups()
	if err != nil {
		return nil, err
	}
	var ret []cloudprovider.ICloudSecurityGroup
	for i := range groups {
		if groups[i].ProjectId != s.Id {
			continue
		}
		groups[i].region = s.region
		ret = append(ret, &groups[i])
	}
	return ret, nil
}

func (s *SVpc) Refresh() error {
	newVpc, err := s.region.GetIVpcById(s.Id)
	if err != nil {
		return err
	}
	return jsonutils.Update(s, newVpc)

}

func (s *SVpc) Delete() error {
	URL := fmt.Sprintf(VPC_DELETE_URL, s.Id)
	_, err := s.region.client.invokePOST(URL, nil, nil, nil)
	return err
}

func (s *SVpc) GetId() string {
	return s.Id
}

func (s *SVpc) GetName() string {
	return s.Name
}

func (s *SVpc) GetGlobalId() string {
	return s.Id
}

func (s *SVpc) GetRegion() cloudprovider.ICloudRegion {
	return s.region
}

func (s *SVpc) GetStatus() string {
	return api.VPC_STATUS_AVAILABLE
}

func (s *SRegion) GetVpcs(id string, start, size int) ([]SVpc, error) {
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
	resp, err := s.client.invokeGET(VPC_LIST_URL, nil, query)
	if err != nil {
		return nil, err
	}
	var vpcs []SVpc
	return vpcs, resp.Unmarshal(&vpcs, "data")
}

func (s *SRegion) GetIVpcs() ([]cloudprovider.ICloudVpc, error) {
	vpcs, err := s.getVpcs()
	if err != nil {
		return nil, err
	}
	var ret []cloudprovider.ICloudVpc
	for i := range vpcs {
		vpcs[i].region = s
		ret = append(ret, &vpcs[i])
	}
	//增加一个id为default vpc用来标识云宏的网络管理
	classicVpcs, err := s.getClassicVpcs()
	if err != nil {
		return nil, errors.Wrapf(err, "ListClassicVpcs")
	}
	for i := range classicVpcs {
		classicVpcs[i].region = s
		ret = append(ret, &classicVpcs[i])
	}
	return ret, nil
}

func (s *SRegion) GetIVpcById(id string) (cloudprovider.ICloudVpc, error) {
	vpcs, err := s.GetIVpcs()
	if err != nil {
		return nil, err
	}
	for i := range vpcs {
		if vpcs[i].GetGlobalId() == id {
			return vpcs[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, id)
}

func (s *SRegion) getVpcs() ([]SVpc, error) {
	var vpcs []SVpc
	start, size := 1, 10
	for {
		ret, err := s.GetVpcs("", start, size)
		if err != nil {
			return nil, err
		}
		for i := range ret {
			vpcs = append(vpcs, ret[i])
		}
		if len(ret) < size {
			break
		}
		start += 1
	}
	return vpcs, nil
}

func (s *SRegion) CreateIVpc(opts *cloudprovider.VpcCreateOptions) (cloudprovider.ICloudVpc, error) {
	body := make(map[string]string)
	body["name"] = opts.NAME
	body["remark"] = opts.Desc
	resp, err := s.client.invokePOST(VPC_CREATE_URL, nil, nil, body)
	if err != nil {
		return nil, err
	}
	var ret SVpc
	return &ret, resp.Unmarshal(&ret)
}

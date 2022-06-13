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
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/multicloud"
)

const (
	VPC_LIST_URL = "/api/network/vpcs"
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
		groups[i].region = s.region
		ret = append(ret, &groups[i])
	}
	return ret, nil
}

func (s *SVpc) GetIRouteTables() ([]cloudprovider.ICloudRouteTable, error) {
	panic("implement me")
}

func (s *SVpc) GetIRouteTableById(routeTableId string) (cloudprovider.ICloudRouteTable, error) {
	panic("implement me")
}

func (s *SVpc) Delete() error {
	panic("implement me")
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

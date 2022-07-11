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
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/multicloud"
	"yunion.io/x/pkg/errors"
)

type SClassicVpc struct {
	multicloud.SVpc
	multicloud.WinStackTags

	region *SRegion

	iwires []cloudprovider.ICloudWire

	Id   string
	Name string
}

func (s *SClassicVpc) GetIsDefault() bool {
	return false
}

func (s *SClassicVpc) GetCidrBlock() string {
	return ""
}

func (s *SClassicVpc) GetISecurityGroups() ([]cloudprovider.ICloudSecurityGroup, error) {
	return nil, nil
}

func (s *SClassicVpc) GetIRouteTables() ([]cloudprovider.ICloudRouteTable, error) {
	rts := []cloudprovider.ICloudRouteTable{}
	return rts, nil
}

func (s *SClassicVpc) GetIRouteTableById(routeTableId string) (cloudprovider.ICloudRouteTable, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (s *SClassicVpc) Delete() error {
	return nil
}

func (s *SClassicVpc) GetId() string {
	return s.Id
}

func (s *SClassicVpc) GetName() string {
	return s.Name
}

func (s *SClassicVpc) GetGlobalId() string {
	return s.Id
}

func (s *SClassicVpc) GetRegion() cloudprovider.ICloudRegion {
	return s.region
}

func (s *SClassicVpc) GetStatus() string {
	return api.VPC_STATUS_AVAILABLE
}

func (s *SClassicVpc) GetExternalAccessMode() string {
	return api.VPC_EXTERNAL_ACCESS_MODE_DISTGW
}

func (s *SClassicVpc) GetIsExternalNet() bool {
	return true
}

func (s *SRegion) getClassicVpcs() ([]SClassicVpc, error) {
	query := make(map[string]string)
	query["type"] = "1"
	resp, err := s.client.invokeGET(CLASSIC_NETWORK_LIST_URL, nil, query)
	if err != nil {
		return nil, err
	}
	var ret []SClassicVpc
	return ret, resp.Unmarshal(&ret)
}

func (s *SRegion) GetClassicVpcById(id string) (*SClassicVpc, error) {
	vpcs, err := s.getClassicVpcs()
	if err != nil {
		return nil, err
	}
	for i := range vpcs {
		if vpcs[i].Id == id {
			return &vpcs[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, id)
}

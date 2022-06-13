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
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/multicloud"
	"yunion.io/x/pkg/errors"
)

const (
	REGION_LIST_URL = "/api/compute/pools"
)

type SRegion struct {
	multicloud.SRegion
	multicloud.SRegionEipBase
	multicloud.SRegionLbBase
	multicloud.SRegionOssBase
	multicloud.SRegionSecurityGroupBase
	multicloud.SRegionVpcBase
	multicloud.SRegionZoneBase

	client *SWinStackClient

	Id   string
	Name string
}

func (s *SRegion) GetId() string {
	return s.Id
}

func (s *SRegion) GetName() string {
	return s.Name
}

func (s *SRegion) GetGlobalId() string {
	return fmt.Sprintf("%s/%s", CLOUD_PROVIDER_WINSTACK, s.Id)
}

func (s *SRegion) GetStatus() string {
	return api.CLOUD_REGION_STATUS_INSERVER
}

func (s *SRegion) GetI18n() cloudprovider.SModelI18nTable {
	return cloudprovider.SModelI18nTable{}
}

func (s *SRegion) GetGeographicInfo() cloudprovider.SGeographicInfo {
	return cloudprovider.SGeographicInfo{}
}

func (s *SRegion) GetIHosts() ([]cloudprovider.ICloudHost, error) {
	zones, err := s.GetIZones()
	if err != nil {
		return nil, err
	}
	ret := []cloudprovider.ICloudHost{}
	for i := range zones {
		hosts, err := zones[i].GetIHosts()
		if err != nil {
			return nil, err
		}
		ret = append(ret, hosts...)
	}
	return ret, nil
}

func (s *SRegion) GetIHostById(id string) (cloudprovider.ICloudHost, error) {
	zones, err := s.GetIZones()
	if err != nil {
		return nil, err
	}
	for i := range zones {
		hosts, err := zones[i].GetIHosts()
		if err != nil {
			return nil, err
		}
		for i := range hosts {
			if hosts[i].GetGlobalId() == id {
				return hosts[i], nil
			}
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, id)
}

func (s *SRegion) GetCloudEnv() string {
	return ""
}

func (s *SRegion) GetProvider() string {
	return CLOUD_PROVIDER_WINSTACK
}

func (s *SRegion) GetCapabilities() []string {
	return s.client.GetCapabilities()
}

func (client *SWinStackClient) GetRegions() ([]SRegion, error) {
	var ret []SRegion
	resp, err := client.invokeGET(REGION_LIST_URL, nil, nil)
	if err != nil {
		return nil, err
	}
	return ret, resp.Unmarshal(&ret, "data")
}

func (s *SRegion) GetClient() *SWinStackClient {
	return s.client
}

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

const (
	CLUSTER_LIST_URL = "/api/compute/clusters"
)

type SCluster struct {
	multicloud.SResourceBase
	multicloud.STagBase

	region *SRegion

	Id   string
	Name string
}

func (s *SCluster) GetIHosts() ([]cloudprovider.ICloudHost, error) {
	hosts, err := s.region.getHosts(s.Id, "")
	if err != nil {
		return nil, err
	}
	var ret []cloudprovider.ICloudHost
	for i := range hosts {
		hosts[i].cluster = s
		ret = append(ret, &hosts[i])
	}
	return ret, nil
}

func (s *SCluster) GetIHostById(id string) (cloudprovider.ICloudHost, error) {
	hosts, err := s.region.getHosts(s.Id, id)
	if err != nil {
		return nil, err
	}
	for i := range hosts {
		if hosts[i].Id == id {
			return &hosts[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, id)
}

func (s *SCluster) GetId() string {
	return s.Id
}

func (s *SCluster) GetName() string {
	return s.Name
}

func (s *SCluster) GetGlobalId() string {
	return s.Id
}

func (s *SCluster) GetStatus() string {
	return api.ZONE_ENABLE
}

func (s *SCluster) GetI18n() cloudprovider.SModelI18nTable {
	return cloudprovider.SModelI18nTable{}
}

func (s *SCluster) GetIRegion() cloudprovider.ICloudRegion {
	return s.region
}

func (s *SRegion) GetIZones() ([]cloudprovider.ICloudZone, error) {
	clusters, err := s.GetClusters()
	if err != nil {
		return nil, errors.Wrapf(err, "GetClusters")
	}
	ret := []cloudprovider.ICloudZone{}
	for i := range clusters {
		clusters[i].region = s
		ret = append(ret, &clusters[i])
	}
	return ret, nil
}

func (s *SRegion) GetIZoneById(id string) (cloudprovider.ICloudZone, error) {
	zones, err := s.GetIZones()
	if err != nil {
		return nil, err
	}
	for i := range zones {
		if zones[i].GetGlobalId() == id {
			return zones[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, id)
}

func (s *SRegion) GetClusters() ([]SCluster, error) {
	resp, err := s.client.invokeGET(CLUSTER_LIST_URL, nil, nil)
	if err != nil {
		return nil, err
	}
	var ret []SCluster
	return ret, resp.Unmarshal(&ret, "data")
}

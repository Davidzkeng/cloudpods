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
	"yunion.io/x/jsonutils"
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/multicloud"
)

const (
	HOST_LIST_URL = "/api/compute/hosts"
)

type SHost struct {
	multicloud.SHostBase
	multicloud.STagBase

	cluster *SCluster

	Id          string
	Name        string
	ClusterId   string
	CpuCores    int
	IsConnected bool
	IsMaintain  bool
}

func (s *SHost) GetIVMs() ([]cloudprovider.ICloudVM, error) {
	//vms := []SInstance{}
	return nil, nil
}

func (s *SHost) GetIVMById(id string) (cloudprovider.ICloudVM, error) {
	panic("implement me")
}

func (s *SHost) GetIWires() ([]cloudprovider.ICloudWire, error) {
	panic("implement me")
}

func (s *SHost) GetIStorages() ([]cloudprovider.ICloudStorage, error) {
	panic("implement me")
}

func (s *SHost) GetIStorageById(id string) (cloudprovider.ICloudStorage, error) {
	panic("implement me")
}

func (s *SHost) GetEnabled() bool {
	panic("implement me")
}

func (s *SHost) GetHostStatus() string {
	if !s.IsConnected {
		return api.HOST_OFFLINE
	}
	return api.HOST_ONLINE
}

func (s *SHost) GetAccessIp() string {
	panic("implement me")
}

func (s *SHost) GetAccessMac() string {
	panic("implement me")
}

func (s *SHost) GetSysInfo() jsonutils.JSONObject {
	panic("implement me")
}

func (s *SHost) GetSN() string {
	panic("implement me")
}

func (s *SHost) GetCpuCount() int {
	panic("implement me")
}

func (s *SHost) GetNodeCount() int8 {
	panic("implement me")
}

func (s *SHost) GetCpuDesc() string {
	panic("implement me")
}

func (s *SHost) GetCpuMhz() int {
	return 0
}

func (s *SHost) GetMemSizeMB() int {
	panic("implement me")
}

func (s *SHost) GetStorageSizeMB() int {
	panic("implement me")
}

func (s *SHost) GetStorageType() string {
	panic("implement me")
}

func (s *SHost) GetHostType() string {
	return api.HOST_TYPE_WINSTACK
}

func (s *SHost) GetIsMaintenance() bool {
	return s.IsMaintain
}

func (s *SHost) GetVersion() string {
	return ""
}

func (s *SHost) CreateVM(desc *cloudprovider.SManagedVMCreateConfig) (cloudprovider.ICloudVM, error) {
	panic("implement me")
}

func (s *SHost) GetIHostNics() ([]cloudprovider.ICloudHostNetInterface, error) {
	panic("implement me")
}

func (s *SHost) GetId() string {
	return s.Id
}

func (s *SHost) GetName() string {
	return s.Name
}

func (s *SHost) GetGlobalId() string {
	return s.Id
}

func (s *SHost) GetStatus() string {
	return api.HOST_STATUS_RUNNING
}

func (s *SRegion) GetHosts(clusterId, hostId string, start, size int) ([]SHost, error) {
	query := make(map[string]string)
	if len(clusterId) > 0 {
		query["clusterId"] = clusterId
	}
	if len(hostId) > 0 {
		query["hostIds"] = hostId
	}
	if size <= 0 {
		size = 10
	}
	if start < 0 {
		start = 0
	}
	query["start"] = strconv.Itoa(start)
	query["size"] = strconv.Itoa(size)
	resp, err := s.client.invokeGET(HOST_LIST_URL, nil, query)
	if err != nil {
		return nil, err
	}
	var ret []SHost
	return ret, resp.Unmarshal(&ret, "data")
}

func (s *SRegion) getHosts(clusterId, hostId string) ([]SHost, error) {
	var ret []SHost
	start, size := 0, 10
	for {
		hosts, err := s.GetHosts(clusterId, hostId, start, size)
		if err != nil {
			return nil, err
		}
		for i := range hosts {
			ret = append(ret, hosts[i])
		}
		if len(hosts) < size {
			break
		}
		start += size
	}
	return ret, nil
}

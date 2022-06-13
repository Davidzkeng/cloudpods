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
	"context"
	"strconv"
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/multicloud"
)

const (
	INSTANCE_LIST_URL = "/api/compute/domains"
)

type SInstance struct {
	multicloud.SInstanceBase
	multicloud.WinStackTags

	host *SHost

	Id                  string
	Name                string
	Status              int
	OsType              int
	VCpu                int
	Memory              int
	DomainDiskDbRspList []struct {
		Id              string
		VolId           string
		StoragePoolId   string
		StoragePoolType int
		FileName        string
		Type            string
		Device          string
		Dev             string
		Bus             string
		QemuType        string
		Path            string
		BootOrder       int
		VirtualSize     int64
		DiskSize        int64
		IsPersistence   bool
		Shareable       bool
		ByPathId        string
	}
}

func (s *SInstance) GetProjectId() string {
	return ""
}

func (s *SInstance) GetHostname() string {
	return s.Name
}

func (s *SInstance) GetIHost() cloudprovider.ICloudHost {
	return s.host
}

func (s *SInstance) GetIDisks() ([]cloudprovider.ICloudDisk, error) {
	var ret []cloudprovider.ICloudDisk
	for i := range s.DomainDiskDbRspList {
		storage, err := s.host.cluster.region.getStorage(s.DomainDiskDbRspList[i].StoragePoolId)
		if err != nil {
			return nil, err
		}
		disk, err := storage.GetIDiskById(s.DomainDiskDbRspList[i].VolId)
		if err != nil {
			return nil, err
		}
		ret = append(ret, disk)
	}
	return ret, nil
}

func (s *SInstance) GetINics() ([]cloudprovider.ICloudNic, error) {
	nics, err := s.host.cluster.region.GetInstanceNics(s.Id)
	if err != nil {
		return nil, err
	}
	var ret []cloudprovider.ICloudNic
	for i := range nics {
		ret = append(ret, &nics[i])
	}
	return ret, nil
}

func (s *SInstance) GetIEIP() (cloudprovider.ICloudEIP, error) {
	return nil, nil
}

func (s *SInstance) GetVcpuCount() int {
	return s.VCpu
}

func (s *SInstance) GetVmemSizeMB() int {
	return s.Memory / 1024 / 1024
}

func (s *SInstance) GetBootOrder() string {
	return "dcn"
}

func (s *SInstance) GetVga() string {
	return ""
}

func (s *SInstance) GetVdi() string {
	return ""
}

func (s *SInstance) GetOsType() cloudprovider.TOsType {
	if s.OsType == 1 {
		return cloudprovider.OsTypeLinux
	}
	return cloudprovider.OsTypeWindows
}

func (s *SInstance) GetOSName() string {
	return ""
}

func (s *SInstance) GetBios() string {
	return "BIOS"
}

func (s *SInstance) GetMachine() string {
	return ""
}

func (s *SInstance) GetInstanceType() string {
	panic("implement me")
}

func (s *SInstance) GetSecurityGroupIds() ([]string, error) {
	panic("implement me")
}

func (s *SInstance) AssignSecurityGroup(secgroupId string) error {
	panic("implement me")
}

func (s *SInstance) SetSecurityGroups(secgroupIds []string) error {
	panic("implement me")
}

func (s *SInstance) GetHypervisor() string {
	panic("implement me")
}

func (s *SInstance) StartVM(ctx context.Context) error {
	panic("implement me")
}

func (s *SInstance) StopVM(ctx context.Context, opts *cloudprovider.ServerStopOptions) error {
	panic("implement me")
}

func (s *SInstance) DeleteVM(ctx context.Context) error {
	panic("implement me")
}

func (s *SInstance) UpdateVM(ctx context.Context, name string) error {
	panic("implement me")
}

func (s *SInstance) UpdateUserData(userData string) error {
	panic("implement me")
}

func (s *SInstance) RebuildRoot(ctx context.Context, config *cloudprovider.SManagedVMRebuildRootConfig) (string, error) {
	panic("implement me")
}

func (s *SInstance) DeployVM(ctx context.Context, name string, username string, password string, publicKey string, deleteKeypair bool, description string) error {
	panic("implement me")
}

func (s *SInstance) ChangeConfig(ctx context.Context, config *cloudprovider.SManagedVMChangeConfig) error {
	panic("implement me")
}

func (s *SInstance) GetVNCInfo(input *cloudprovider.ServerVncInput) (*cloudprovider.ServerVncOutput, error) {
	panic("implement me")
}

func (s *SInstance) AttachDisk(ctx context.Context, diskId string) error {
	panic("implement me")
}

func (s *SInstance) DetachDisk(ctx context.Context, diskId string) error {
	panic("implement me")
}

func (s *SInstance) GetError() error {
	panic("implement me")
}

func (s *SInstance) GetId() string {
	return s.Id
}

func (s *SInstance) GetName() string {
	if len(s.Name) > 0 {
		return s.Name
	}
	return s.Id
}

func (s *SInstance) GetGlobalId() string {
	return s.Id
}

func (s *SInstance) GetStatus() string {
	switch s.Status {
	case 1: //运行
		return api.VM_RUNNING
	case 2: //关闭
		return api.VM_STOPPING
	case 3: //暂停
		return api.VM_STARTING
	default:
		return api.VM_UNKNOWN
	}
}

func (s *SRegion) GetInstances(id, hostId, clusterId string, start, size int) ([]SInstance, error) {
	query := make(map[string]string)
	if size <= 0 {
		size = 10
	}
	if start < 0 {
		start = 0
	}
	if len(hostId) > 0 {
		query["hostId"] = hostId
	}
	if len(id) > 0 {
		start = 0
		query["id"] = id
	}
	if len(clusterId) > 0 {
		query["clusterId"] = clusterId
	}
	query["start"] = strconv.Itoa(start)
	query["size"] = strconv.Itoa(size)
	resp, err := s.client.invokeGET(INSTANCE_LIST_URL, nil, query)
	if err != nil {
		return nil, err
	}
	var ret []SInstance
	return ret, resp.Unmarshal(&ret, "data")
}

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
	"fmt"
	"strconv"
	"time"
	"yunion.io/x/jsonutils"
	"yunion.io/x/onecloud/pkg/apis"
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/multicloud"
	"yunion.io/x/pkg/errors"
)

const (
	INSTANCE_LIST_URL     = "/api/compute/domains"
	VNC_URL               = "/api/compute/domains/%s/noVNC/vnc"
	INSTANCE_SHUTDOWN_URL = "/api/compute/domains/%s/act/shutdown"
	INSTANCE_START_URL    = "/api/compute/domains/%s/act/start"
	INSTANCE_DELETE_URL   = "/api/compute/domains/%s/2"
	INSTANCE_DETAIL_URL   = "/api/compute/domains/%s"
	VPC_INSTANCE_LIST_URL = "/api/compute/domainVpc/%s"
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

func (s *SInstance) GetOSArch() string {
	detail, err := s.host.cluster.region.GetInstanceDetailById(s.Id)
	if err != nil {
		return ""
	}
	switch detail.Cpu.Arch {
	case 1:
		return apis.OS_ARCH_X86
	case 2:
		return apis.OS_ARCH_X86_64
	case 3:
		return apis.OS_ARCH_AARCH64
	case 4:
		return "mips64el"
	default:
		return ""
	}
}

func (s *SInstance) GetIDisks() ([]cloudprovider.ICloudDisk, error) {
	var ret []cloudprovider.ICloudDisk
	for i := range s.DomainDiskDbRspList {
		storage, err := s.host.cluster.region.getStorage(s.DomainDiskDbRspList[i].StoragePoolId)
		if err != nil {
			return nil, err
		}
		storage.cluster = s.host.cluster
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
	eips, err := s.host.cluster.region.getEips()
	if err != nil {
		return nil, err
	}
	for i := range eips {
		if eips[i].BindDevId == s.Id {
			eips[i].region = s.host.cluster.region
			return &eips[i], nil
		}
	}
	return nil, cloudprovider.ErrNotFound
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
	detail, err := s.host.cluster.region.GetInstanceDetailById(s.Id)
	if err != nil {
		return "UEFI"
	}
	if detail.BootType == 0 {
		return "BIOS"
	}
	return "UEFI"
}

func (s *SInstance) GetMachine() string {
	return ""
}

func (s *SInstance) GetInstanceType() string {
	return fmt.Sprintf("ecs.g1.c%dm%d", s.GetVcpuCount(), s.GetVmemSizeMB()/1024)
}

func (s *SInstance) GetSecurityGroupIds() ([]string, error) {
	securitys, err := s.host.cluster.region.GetSecurityByVmId(s.Id)
	if err != nil {
		return nil, err
	}
	var securityIds []string
	for i := range securitys {
		securityIds = append(securityIds, securitys[i].Id)
	}
	return securityIds, nil
}

func (s *SInstance) AssignSecurityGroup(secgroupId string) error {
	return cloudprovider.ErrNotImplemented
}

func (s *SInstance) SetSecurityGroups(secgroupIds []string) error {
	return nil
	//return cloudprovider.ErrNotImplemented
}

func (s *SInstance) GetHypervisor() string {
	return api.HYPERVISOR_WINSTACK
}

func (s *SInstance) Refresh() error {
	newInstance, err := s.host.cluster.region.GetInstanceById(s.Id)
	if err != nil {
		return err
	}
	return jsonutils.Update(s, newInstance)
}

func (s *SInstance) StartVM(ctx context.Context) error {
	err := s.host.cluster.region.StartVM(s.GetId())
	if err != nil {
		return errors.Wrap(err, "Instance.StartVM")
	}
	err = cloudprovider.WaitStatus(s, api.VM_RUNNING, 5*time.Second, 300*time.Second)
	if err != nil {
		return errors.Wrap(err, "Instance.StartVM.WaitStatus")
	}
	return nil
}

func (s *SInstance) StopVM(ctx context.Context, opts *cloudprovider.ServerStopOptions) error {
	err := s.host.cluster.region.StopVM(s.GetId())
	if err != nil {
		return errors.Wrap(err, "Instance.StopVM")
	}
	err = cloudprovider.WaitStatus(s, api.VM_READY, 5*time.Second, 300*time.Second)
	if err != nil {
		return errors.Wrap(err, "Instance.StopVM.WaitStatus")
	}
	return nil
}

func (s *SInstance) DeleteVM(ctx context.Context) error {
	err := s.host.cluster.region.DeleteVM(s.GetId())
	if err != nil {
		return errors.Wrap(err, "Instance.DeleteVM")
	}
	return nil
}

func (s *SInstance) UpdateVM(ctx context.Context, name string) error {
	return cloudprovider.ErrNotImplemented
}

func (s *SInstance) UpdateUserData(userData string) error {
	return cloudprovider.ErrNotImplemented
}

func (s *SInstance) RebuildRoot(ctx context.Context, config *cloudprovider.SManagedVMRebuildRootConfig) (string, error) {
	return "", cloudprovider.ErrNotImplemented
}

func (s *SInstance) DeployVM(ctx context.Context, name string, username string, password string, publicKey string, deleteKeypair bool, description string) error {
	return cloudprovider.ErrNotImplemented
}

func (s *SInstance) ChangeConfig(ctx context.Context, config *cloudprovider.SManagedVMChangeConfig) error {
	return cloudprovider.ErrNotImplemented
}

func (s *SInstance) GetVNCInfo(input *cloudprovider.ServerVncInput) (*cloudprovider.ServerVncOutput, error) {
	return s.host.cluster.region.GetVNCURL(s.Id)

}

func (s *SInstance) AttachDisk(ctx context.Context, diskId string) error {
	return cloudprovider.ErrNotImplemented
}

func (s *SInstance) DetachDisk(ctx context.Context, diskId string) error {
	return cloudprovider.ErrNotImplemented
}

func (s *SInstance) GetError() error {
	return nil
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
		return api.VM_READY
	case 3: //暂停
		return api.VM_STARTING
	default:
		return api.VM_UNKNOWN
	}
}

func (s *SRegion) GetInstancesByHostId(hostId string) ([]SInstance, error) {
	var instances []SInstance
	start, size := 1, 10
	for {
		ret, err := s.GetInstances("", hostId, "", "", start, size)
		if err != nil {
			return nil, err
		}
		for i := range ret {
			instances = append(instances, ret[i])
		}
		if len(ret) < size {
			break
		}
		start += 1
	}
	return instances, nil
}

func (s *SRegion) GetInstancesByClusterId(clusterId string) ([]SInstance, error) {
	var instances []SInstance
	start, size := 1, 10
	for {
		ret, err := s.GetInstances("", "", clusterId, "", start, size)
		if err != nil {
			return nil, err
		}
		for i := range ret {
			instances = append(instances, ret[i])
		}
		if len(ret) < size {
			break
		}
		start += 1
	}
	return instances, nil
}

func (s *SRegion) GetInstancesByName(name string) (*SInstance, error) {
	instance, err := s.GetInstances("", "", "", name, 0, 1)
	if err != nil {
		return nil, err
	}
	for i := range instance {
		if instance[i].GetName() == name {
			return &instance[i], nil
		}
	}
	return nil, cloudprovider.ErrNotFound
}

func (s *SRegion) GetInstanceById(id string) (*SInstance, error) {
	instances, err := s.GetInstances(id, "", "", "", 0, 1)
	if err != nil {
		return nil, err
	}
	for i := range instances {
		if instances[i].GetId() == id {
			return &instances[i], nil
		}
	}
	return nil, cloudprovider.ErrNotFound
}

type SInstanceDetail struct {
	BootType int
	Cpu      struct {
		Arch int
	}
}

func (s *SRegion) GetInstanceDetailById(id string) (*SInstanceDetail, error) {
	URL := fmt.Sprintf(INSTANCE_DETAIL_URL, id)
	resp, err := s.client.invokeGET(URL, nil, nil)
	if err != nil {
		return nil, err
	}
	var ret SInstanceDetail
	return &ret, resp.Unmarshal(&ret)
}

func (s *SRegion) GetInstances(id, hostId, clusterId string, name string, start, size int) ([]SInstance, error) {
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
	if len(name) > 0 {
		query["name"] = name
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

func (s *SRegion) GetVNCURL(id string) (*cloudprovider.ServerVncOutput, error) {
	client := s.GetClient()
	URL := client.endpoint + fmt.Sprintf(VNC_URL, id)
	ret := &cloudprovider.ServerVncOutput{
		Url:        URL,
		Protocol:   "winstack",
		InstanceId: id,
		Hypervisor: api.HYPERVISOR_WINSTACK,
	}
	return ret, nil
}

func (s *SRegion) StartVM(vmId string) error {
	URL := fmt.Sprintf(INSTANCE_START_URL, vmId)
	_, err := s.client.invokePATCH(URL, nil, nil)
	if err != nil {
		return errors.Wrap(err, "SRegion.StartVm")
	}

	return nil
}

func (s *SRegion) StopVM(vmId string) error {
	URL := fmt.Sprintf(INSTANCE_SHUTDOWN_URL, vmId)
	_, err := s.client.invokePATCH(URL, nil, nil)
	if err != nil {
		return errors.Wrap(err, "SRegion.StopVm")
	}

	return nil
}

func (s *SRegion) DeleteVM(vmId string) error {
	URL := fmt.Sprintf(INSTANCE_DELETE_URL, vmId)
	_, err := s.client.invokeDELETE(URL, nil, nil)
	if err != nil {
		return errors.Wrap(err, "SRegion.DeleteVm")
	}
	return nil
}

type SVpcInstance struct {
	VpcId   string
	VpcName string
	VpcList []struct {
		DomainId       string
		DomainShowName string
		InterfaceList  []struct {
			PrivateNetId   string
			PrivateNetName string
			InterfaceId    string
			Mac            string
			Ip             string
		}
	}
}

func (s *SRegion) GetInstancesByVpcId(vpcId string) (*SVpcInstance, error) {
	URL := fmt.Sprintf(VPC_INSTANCE_LIST_URL, vpcId)
	resp, err := s.client.invokeGET(URL, nil, nil)
	if err != nil {
		return nil, err
	}
	var ret SVpcInstance
	return &ret, resp.Unmarshal(&ret)
}

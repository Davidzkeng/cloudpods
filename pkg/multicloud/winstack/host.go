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
	"time"
	"yunion.io/x/jsonutils"
	"yunion.io/x/pkg/errors"

	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/multicloud"
	"yunion.io/x/onecloud/pkg/util/billing"
)

const (
	HOST_LIST_URL     = "/api/compute/hosts"
	VM_DEPLOY_URL     = "/api/compute/domain_templates/%s/deploy/hosts/%s"
	VM_MODIFY_PWD_URL = "/api/compute/domains/%s/pwd"
)

type SHost struct {
	multicloud.SHostBase
	multicloud.STagBase

	cluster *SCluster

	Id           string
	Name         string
	ClusterId    string
	CpuCores     int
	IsConnected  bool
	IsMaintain   bool
	Ip           string
	CpuSockets   int8
	CpuModelName string
	Memory       int64
	HMemory      string
	Storage      int64
}

func (s *SHost) GetIVMs() ([]cloudprovider.ICloudVM, error) {
	var ret []cloudprovider.ICloudVM
	instances, err := s.cluster.region.GetInstancesByHostId(s.Id)
	if err != nil {
		return nil, err
	}
	for i := range instances {
		instances[i].host = s
		ret = append(ret, &instances[i])

	}
	return ret, nil
}

func (s *SHost) GetIVMById(id string) (cloudprovider.ICloudVM, error) {
	instance, err := s.cluster.region.GetInstancesByHostId(s.Id)
	if err != nil {
		return nil, err
	}
	for i := range instance {
		if instance[i].GetGlobalId() == id {
			instance[i].host = s
			return &instance[i], nil
		}
	}
	return nil, cloudprovider.ErrNotFound

}

func (s *SHost) GetIWires() ([]cloudprovider.ICloudWire, error) {
	return s.cluster.GetIWires()
}

func (s *SHost) GetIStorages() ([]cloudprovider.ICloudStorage, error) {
	return s.cluster.GetIStorages()
}

func (s *SHost) GetIStorageById(id string) (cloudprovider.ICloudStorage, error) {
	return s.cluster.GetIStorageById(id)
}

func (s *SHost) GetEnabled() bool {
	return true
}

func (s *SHost) GetHostStatus() string {
	if !s.IsConnected {
		return api.HOST_OFFLINE
	}
	return api.HOST_ONLINE
}

func (s *SHost) GetAccessIp() string {
	return s.Ip
}

func (s *SHost) GetAccessMac() string {
	return ""
}

func (s *SHost) GetSysInfo() jsonutils.JSONObject {
	return jsonutils.NewDict()
}

func (s *SHost) GetSN() string {
	return ""
}

func (s *SHost) GetCpuCount() int {
	return s.CpuCores
}

func (s *SHost) GetNodeCount() int8 {
	return s.CpuSockets
}

func (s *SHost) GetCpuDesc() string {
	return s.CpuModelName
}

func (s *SHost) GetCpuMhz() int {
	return 0
}

func (s *SHost) GetMemSizeMB() int {
	return int(s.Memory / 1024 / 1024)
}

func (s *SHost) GetStorageSizeMB() int {
	return int(s.Storage / 1024 / 1024)
}

func (s *SHost) GetStorageType() string {
	return api.STORAGE_LOCAL_SSD
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
	instanceId, err := s._createVM(desc.Name, desc.Hostname, desc.ExternalImageId, desc.SysDisk, desc.Cpu, desc.MemoryMB,
		desc.InstanceType, desc.ExternalNetworkId, desc.IpAddr, desc.Description, desc.Account, desc.Password,
		desc.DataDisks, desc.PublicKey, desc.ExternalSecgroupId, desc.UserData, desc.BillingCycle,
		desc.ProjectId, desc.OsType, desc.Tags, desc.SPublicIpInfo)
	if err != nil {
		return nil, err
	}
	//查询虚拟机详情
	instance, err := s.GetIVMById(instanceId)
	if err != nil {
		return nil, errors.Wrapf(err, "GetIVMById %s", instanceId)
	}
	//绑定安全组
	err = s.cluster.region.AssignSecurityGroup(instanceId, desc.ExternalVpcId, desc.ExternalSecgroupId)
	if err != nil {
		return nil, errors.Wrapf(err, "AssignSecurityGroup instance:%s,vpcId:%s,secGroupId:%s", instanceId, desc.ExternalVpcId, desc.ExternalSecgroupId)
	}
	if len(desc.Password) > 0 {
		//修改密码
		err = s.cluster.region.ChangeVMPassword(instanceId, desc.Account, desc.Password)
		if err != nil {
			return nil, errors.Wrapf(err, "ChangeVMPassword %s,account:%s,pwd:%s", instanceId, desc.Account, desc.Password)
		}
	}
	return instance, nil
}

func (s *SHost) GetIHostNics() ([]cloudprovider.ICloudHostNetInterface, error) {
	return nil, cloudprovider.ErrNotSupported
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
	start, size := 1, 10
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
		start += 1
	}
	return ret, nil
}

func (s *SHost) _createVM(name, hostname string, imgId string,
	sysDisk cloudprovider.SDiskInfo, cpu int, memMB int, instanceType string,
	netId string, ipAddr string, desc string, account, passwd string,
	dataDisks []cloudprovider.SDiskInfo, publicKey string, secgroupId string,
	userData string, bc *billing.SBillingCycle, projectId, osType string,
	tags map[string]string, publicIp cloudprovider.SPublicIpInfo) (string, error) {
	net := s.cluster.getNetworkById(netId)
	if net == nil {
		return "", errors.Errorf("invalid net ID %v", netId)
	}
	if net.wire == nil {
		return "", errors.Errorf("net's wire is empty")
	}
	if net.wire.vpc == nil {
		return "", fmt.Errorf("net's wire's vpc is empty")
	}
	var err error

	img, err := s.cluster.region.getImage(imgId)
	if err != nil {
		return "", errors.Wrapf(err, "GetImage fail")
	}
	if img.GetStatus() != api.CACHED_IMAGE_STATUS_ACTIVE {
		return "", fmt.Errorf("image %s not ready", imgId)
	}
	disks := make([]SDisk, len(dataDisks)+1)
	disks[0].Capacity = img.GetSizeByte()
	if sysDisk.SizeGB > 0 && sysDisk.SizeGB > img.GetMinOsDiskSizeGb() {
		disks[0].Capacity = int64(sysDisk.SizeGB) * 1024 * 1024 * 1024
	}
	disks[0].storage, err = s.cluster.getStorageById(sysDisk.StorageExternalId)
	if err != nil {
		return "", err
	}

	for i, dataDisk := range dataDisks {
		disks[i+1].Capacity = int64(dataDisk.SizeGB) * 1024 * 1024 * 1024
		disks[i+1].storage, err = s.cluster.getStorageById(sysDisk.StorageExternalId)
		if err != nil {
			return "", err
		}
	}
	_, err = s.cluster.region.CreateInstance(name, s.Id, imgId, cpu, memMB, netId, net.wire.vpc.Id, disks, ipAddr, account, passwd, publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to create specification (c%d.m%d).%s", cpu, memMB, err.Error())
	}
	var vmId string
	cloudprovider.WaitCreated(15*time.Second, 5*time.Minute, func() bool {
		instance, err := s.cluster.region.GetInstancesByName(name)
		if err != nil {
			return false
		}
		if len(instance.Id) > 0 {
			vmId = instance.Id
			return true
		}
		return false
	})
	return vmId, nil
}

// CreateInstance
//{
//    "deployNum":1,
//    "deployStartNo":1,
//    "deployType":1,
//    "isStartNewDomain":false,
//    "isHa":false,
//    "baseInfo":{
//        "name":"kylin_tem",
//        "cpu":{
//            "sockets":1,
//            "cores":96,
//            "threads":1,
//            "vcpus":96,
//            "current":2,
//            "arch":3,
//            "shares":2,
//            "mode":3,
//            "gurantees":0,
//            "quota":100,
//            "bindCpu":null
//        },
//        "memory":{
//            "currentMemory":4294967296,
//            "memory":4294967296,
//            "maxMemory":null,
//            "minMemory":1073741824,
//            "locked":0,
//            "priority":null,
//            "limit":null,
//            "autoMem":false,
//            "mode":null,
//            "nodeSet":null,
//            "isHugePages":false,
//            "isOpenVnuma":false,
//            "size":4294967296
//        }
//    },
//    "diskDevices":[
//        {
//            "bus":1,
//            "dev":"vda",
//            "source":4,
//            "shareable":false,
//            "isEncrypt":null,
//            "type":null,
//            "capacity":85899345920,
//            "poolName":"sata",
//            "cache":2
//        }
//    ],
//    "bridgeInterfaces":[
//        {
//            "networkType":"1",
//            "mac":null,
//            "oldMac":"52:54:00:50:93:b9",
//            "portGroupId":null,
//            "isVhostDriver":true,
//            "model":1,
//            "vpc":{
//                "vpcId":"8b6b813f-a56d-4c45-bef8-6f9ef83bed43",
//                "privateNetId":"ddaf0052-4094-43e3-a593-0f6a897e22ff"
//            },
//            "ip":"192.168.1.2"
//        }
//    ]
//}
func (s *SRegion) CreateInstance(name string, nodeId string, imageId string, cpu int, memMB int, SubnetId string,
	vpcId string, disks []SDisk, ipAddr string, account string, passwd string, keypair string) (string, error) {
	//查找imageId对应的镜像信息
	imageInfo, err := s.getImage(imageId)
	if err != nil {
		return "", err
	}

	//cloudInitInfo := jsonutils.NewDict()
	//cloudInitInfo.Set("osType", jsonutils.NewString(string(imageInfo.GetOsType())))
	//if len(passwd) > 0 {
	//	changePasswdListInfo := jsonutils.NewDict()
	//	changePasswdListInfo.Set("password", jsonutils.NewString(passwd))
	//	changePasswdListInfo.Set("username", jsonutils.NewString(account))
	//	changePasswdListInfoArr := jsonutils.NewArray(changePasswdListInfo)
	//	cloudInitInfo.Set("changePassWordList", changePasswdListInfoArr)
	//}
	if len(keypair) > 0 { //todo:镜像ssh密钥登录？
	}

	cpuInfo := jsonutils.NewDict()
	cpuInfo.Set("arch", jsonutils.NewInt(int64(imageInfo.CpuArch)))
	cpuInfo.Set("mode", jsonutils.NewInt(3))
	cpuInfo.Set("current", jsonutils.NewInt(int64(cpu)))
	cpuInfo.Set("sockets", jsonutils.NewInt(1))
	cpuInfo.Set("cores", jsonutils.NewInt(96))
	cpuInfo.Set("threads", jsonutils.NewInt(1))

	memoryInfo := jsonutils.NewDict()
	memByte := int64(memMB) * 1024 * 1024
	memoryInfo.Set("currentMemory", jsonutils.NewInt(memByte))
	memoryInfo.Set("memory", jsonutils.NewInt(memByte))
	memoryInfo.Set("size", jsonutils.NewInt(memByte))

	baseInfo := jsonutils.NewDict()
	baseInfo.Set("name", jsonutils.NewString(name))
	baseInfo.Set("memory", memoryInfo)
	baseInfo.Set("cpu", cpuInfo)

	diskInfo := jsonutils.NewDict()
	for i := range disks {
		if i == 0 { //系统盘
			//查询镜像详情接口获取bus
			diskInfo.Set("bus", jsonutils.NewInt(int64(imageInfo.DiskDevices[i].Bus)))
			//查询镜像详情接口获取dev
			diskInfo.Set("dev", jsonutils.NewString(imageInfo.DiskDevices[i].Dev))
			diskInfo.Set("source", jsonutils.NewInt(4))
			diskInfo.Set("capacity", jsonutils.NewInt(disks[i].Capacity))
			diskInfo.Set("poolName", jsonutils.NewString(disks[i].storage.Name))

		}
	}

	vpcInfo := jsonutils.NewDict()
	vpcInfo.Set("privateNetId", jsonutils.NewString(SubnetId))
	vpcInfo.Set("vpcId", jsonutils.NewString(vpcId))

	bridgeInterfacesInfo := jsonutils.NewDict()
	bridgeInterfacesInfo.Set("networkType", jsonutils.NewInt(1))
	bridgeInterfacesInfo.Set("model", jsonutils.NewInt(1))
	bridgeInterfacesInfo.Set("vpc", vpcInfo)
	bridgeInterfacesInfo.Set("ip", jsonutils.NewString(ipAddr))

	diskInfoArr := jsonutils.NewArray(diskInfo)
	bridgeInterfacesInfoArr := jsonutils.NewArray(bridgeInterfacesInfo)
	serverParams := jsonutils.NewDict()
	serverParams.Set("deployNum", jsonutils.NewInt(1))
	serverParams.Set("deployType", jsonutils.NewInt(1))
	serverParams.Set("isStartNewDomain", jsonutils.NewBool(true))
	serverParams.Set("baseInfo", baseInfo)
	serverParams.Set("diskDevices", diskInfoArr)
	serverParams.Set("bridgeInterfaces", bridgeInterfacesInfoArr)

	URL := fmt.Sprintf(VM_DEPLOY_URL, imageId, nodeId)
	resp, err := s.client.invokePOST(URL, nil, nil, serverParams.Interface())
	if err != nil {
		return "", err
	}
	return resp.GetString("taskId")
}

func (s *SRegion) ChangeVMPassword(instanceId string, user string, passwd string) error {
	URL := fmt.Sprintf(VM_MODIFY_PWD_URL, instanceId)
	body := make(map[string]string)
	body["user"] = user
	body["passwd"] = passwd
	_, err := s.client.invokePOST(URL, nil, nil, body)
	return err
}

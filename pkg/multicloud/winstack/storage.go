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
	"yunion.io/x/pkg/errors"
)

type TStorageType string

//储的类型,1 FC-SAN， 2 IP-SAN，3 NAS，4 分布式存储,5 本地存储,6.NVME存储
const (
	StorageTypeFCSAN   = TStorageType("FC-SAN")
	StorageTypeIPSAN   = TStorageType("IP-SAN")
	StorageTypeNAS     = TStorageType("NAS")
	StorageTypeCeph    = TStorageType("CEPH")
	StorageTypeLocal   = TStorageType("Local")
	StorageTypeNVME    = TStorageType("NVME")
	StorageTypeUnknown = TStorageType("Unknown")
)

const (
	STORAGE_LIST_URL = "/api/storage/storagePools"
)

type SStorage struct {
	multicloud.STagBase
	multicloud.SStorageBase

	cluster *SCluster

	Id           string
	Name         string
	Status       int
	StorageType  int
	Capacity     int64
	Allocation   int64
	UsedCapacity int64
	Type         TStorageType
}

func (s *SStorage) GetIStoragecache() cloudprovider.ICloudStoragecache {
	return &SStoragecache{storageId: s.Id, storageName: s.Name, region: s.cluster.region}
}

func (s *SStorage) GetIZone() cloudprovider.ICloudZone {
	return s.cluster
}

func (s *SStorage) GetStorageType() string {
	switch s.StorageType {
	case 1:
		return string(StorageTypeFCSAN)
	case 2:
		return string(StorageTypeIPSAN)
	case 3:
		return string(StorageTypeNAS)
	case 4:
		return string(StorageTypeCeph)
	case 5:
		return string(StorageTypeLocal)
	case 6:
		return string(StorageTypeNVME)
	default:
		return string(StorageTypeUnknown)
	}
}

func (s *SStorage) GetMediumType() string {
	return api.DISK_TYPE_SSD
}

func (s *SStorage) GetCapacityMB() int64 {
	return s.Capacity / 1024
}

func (s *SStorage) GetCapacityUsedMB() int64 {
	return s.UsedCapacity / 1024
}

func (s *SStorage) GetStorageConf() jsonutils.JSONObject {
	conf := jsonutils.NewDict()
	return conf
}

func (s *SStorage) GetEnabled() bool {
	return true
}

func (s *SStorage) CreateIDisk(conf *cloudprovider.DiskCreateConfig) (cloudprovider.ICloudDisk, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (s *SStorage) GetMountPoint() string {
	return ""
}

func (s *SStorage) IsSysDiskStore() bool {
	return true
}

func (s *SStorage) GetId() string {
	return s.Id
}

func (s *SStorage) GetName() string {
	return s.Name
}

func (s *SStorage) GetGlobalId() string {
	return s.Id
}

func (s *SStorage) GetStatus() string {
	switch s.Status {
	case 2:
		return api.STORAGE_ONLINE
	default:
		return api.STORAGE_OFFLINE
	}
}

func (s *SCluster) GetIStorageById(id string) (cloudprovider.ICloudStorage, error) {
	storages, err := s.region.getStorages()
	if err != nil {
		return nil, err
	}

	for i := range storages {
		if storages[i].GetGlobalId() == id {
			storages[i].cluster = s
			return &storages[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, id)
}

func (s *SCluster) getStorageById(id string) (*SStorage, error) {
	storages, err := s.region.getStorages()
	if err != nil {
		return nil, err
	}

	for i := range storages {
		if storages[i].GetGlobalId() == id {
			storages[i].cluster = s
			return &storages[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, id)
}

func (s *SCluster) GetIStorages() ([]cloudprovider.ICloudStorage, error) {
	var ret []cloudprovider.ICloudStorage
	storages, err := s.region.getStorages()
	if err != nil {
		return nil, err
	}
	for i := range storages {
		storages[i].cluster = s
		ret = append(ret, &storages[i])
	}
	return ret, nil
}

func (s *SRegion) GetStorages(id string, start, size int) ([]SStorage, error) {
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
	resp, err := s.client.invokeGET(STORAGE_LIST_URL, nil, query)
	if err != nil {
		return nil, err
	}
	var ret []SStorage
	return ret, resp.Unmarshal(&ret, "data")
}

func (s *SRegion) getStorages() ([]SStorage, error) {
	var ret []SStorage
	start, size := 1, 10
	for {
		storages, err := s.GetStorages("", start, size)
		if err != nil {
			return nil, err
		}
		for i := range storages {
			ret = append(ret, storages[i])
		}
		if len(storages) < size {
			break
		}
		start += 1
	}
	return ret, nil
}

func (s *SRegion) getStorage(id string) (*SStorage, error) {
	storages, err := s.GetStorages(id, 0, 0)
	if err != nil {
		return nil, err
	}
	for i := range storages {
		if storages[i].GetId() == id {
			return &storages[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, id)
}

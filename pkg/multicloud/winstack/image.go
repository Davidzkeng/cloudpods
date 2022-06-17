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
	"yunion.io/x/onecloud/pkg/apis"
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/multicloud"
	"yunion.io/x/onecloud/pkg/util/rbacutils"
	"yunion.io/x/pkg/errors"
)

const (
	IMAGE_LIST_URL = "/api/compute/domain_templates"
)

type SImage struct {
	multicloud.SResourceBase
	multicloud.WinStackTags

	cache *SStoragecache

	Id                      string `json:"id"`
	Name                    string `json:"name"`
	StorageId               string `json:"storageId"`
	StorageType             int    `json:"storageType"`
	StorageMountPath        string `json:"storageMountPath"`
	DeviceCdromCount        int    `json:"deviceCdromCount"`
	DeviceInterfaceCount    int    `json:"deviceInterfaceCount"`
	DeviceDiskCount         int    `json:"deviceDiskCount"`
	DeviceDiskTotalCapacity int64  `json:"deviceDiskTotalCapacity"`
	OsType                  int    `json:"osType"`
	OsVersion               string `json:"osVersion"`
	DumpxmlFileMd5          string `json:"dumpxmlFileMd5"`
	DumpxmlFileFullPath     string `json:"dumpxmlFileFullPath"`
	HasVnc                  bool   `json:"hasVnc"`
	HasSpice                bool   `json:"hasSpice"`
	Share                   int    `json:"share"`
	UserId                  string `json:"userId"`
	CpuCurrent              int    `json:"cpuCurrent"`
	CpuArch                 int    `json:"cpuArch"`
	Memory                  int64  `json:"memory"`
	CreateByMyself          int    `json:"createByMyself"`
	CreateUserName          string `json:"createUserName"`
	ManagerOwned            int    `json:"managerOwned"`
	IsEncrypt               bool   `json:"isEncrypt"`
}

func (s *SImage) GetProjectId() string {
	return ""
}

func (s *SImage) Delete(ctx context.Context) error {
	return cloudprovider.ErrNotImplemented
}

func (s *SImage) GetIStoragecache() cloudprovider.ICloudStoragecache {
	return s.cache
}

func (s *SImage) GetSizeByte() int64 {
	return 0
}

func (s *SImage) GetImageType() cloudprovider.TImageType {
	return cloudprovider.ImageTypeCustomized
}

func (s *SImage) GetImageStatus() string {
	return ""
}

func (s *SImage) GetOsType() cloudprovider.TOsType {
	if s.OsType == 1 {
		return cloudprovider.OsTypeLinux
	}
	return cloudprovider.OsTypeWindows
}

func (s *SImage) GetOsDist() string {
	return ""
}

func (s *SImage) GetOsVersion() string {
	return ""
}

func (s *SImage) GetOsArch() string {
	//cpu架构 1.32位 2.64位 3.aarch64 4.mips64el
	switch s.CpuArch {
	case 1:
		return apis.OS_ARCH_X86
	case 2:
		return apis.OS_ARCH_X86_64
	case 3:
		return apis.OS_ARCH_AARCH64
	default:
		return "mips64el"
	}
}

func (s *SImage) GetMinOsDiskSizeGb() int {
	return 40
}

func (s *SImage) GetMinRamSizeMb() int {
	return 0
}

func (s *SImage) GetImageFormat() string {
	return "raw"
}

func (s *SImage) UEFI() bool {
	return true
}

func (s *SImage) GetPublicScope() rbacutils.TRbacScope {
	if s.Share == 2 {
		return rbacutils.ScopeSystem
	}
	return rbacutils.ScopeDomain
}

func (s *SImage) GetSubImages() []cloudprovider.SSubImage {
	return []cloudprovider.SSubImage{}
}

func (s *SImage) GetId() string {
	return s.Id
}

func (s *SImage) GetName() string {
	return s.Name
}

func (s *SImage) GetGlobalId() string {
	return s.Id
}

func (s *SImage) GetStatus() string {
	return api.CACHED_IMAGE_STATUS_ACTIVE
}

func (s *SRegion) GetImages(id string, start, size int) ([]SImage, error) {
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
	resp, err := s.client.invokeGET(IMAGE_LIST_URL, nil, nil)
	if err != nil {
		return nil, err
	}
	var ret []SImage
	return ret, resp.Unmarshal(&ret, "data")
}

func (s *SRegion) getImages() ([]SImage, error) {
	var images []SImage
	start, size := 1, 10
	for {
		ret, err := s.GetImages("", start, size)
		if err != nil {
			return nil, err
		}
		for i := range ret {
			images = append(images, ret[i])
		}
		if len(ret) < size {
			break
		}
		start += 1
	}
	return images, nil
}

func (s *SStoragecache) GetICloudImages() ([]cloudprovider.ICloudImage, error) {
	images, err := s.region.getImages()
	if err != nil {
		return nil, err
	}
	var ret []cloudprovider.ICloudImage
	for i := range images {
		if images[i].StorageId == s.storageId {
			images[i].cache = s
			ret = append(ret, &images[i])
		}
	}
	return ret, nil
}

func (s *SStoragecache) GetIImageById(id string) (cloudprovider.ICloudImage, error) {
	images, err := s.region.getImages()
	if err != nil {
		return nil, err
	}
	for i := range images {
		if images[i].GetGlobalId() == id {
			images[i].cache = s
			return &images[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, id)
}

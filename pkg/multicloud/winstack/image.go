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
	"yunion.io/x/onecloud/pkg/apis"
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/multicloud"
	"yunion.io/x/onecloud/pkg/util/rbacutils"
	"yunion.io/x/pkg/errors"
)

const (
	IMAGE_LIST_URL   = "/api/compute/domain_templates"
	IMAGE_DETAIL_URL = "/api/compute/domain_templates/%s"
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
	DiskDevices             []struct {
		Id               string `json:"id"`
		DomainTemplateId string `json:"domainTemplateId"`
		VolFullPath      string `json:"volFullPath"`
		VolName          string `json:"volName"`
		VolMd5           string `json:"volMd5"`
		VolSm3           string `json:"volSm3"`
		Capacity         int64  `json:"capacity"`
		Allocation       int64  `json:"allocation"`
		Bus              int    `json:"bus"`
		Dev              string `json:"dev"`
		Cache            int    `json:"cache"`
	} `json:"diskDevices"`
	DomainMemoryResp struct {
		CurrentMemory int64 `json:"currentMemory"`
		Memory        int64 `json:"memory"`
		MaxMemory     int   `json:"maxMemory"`
		MinMemory     int   `json:"minMemory"`
	} `json:"domainMemoryResp"`
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
	return s.DeviceDiskTotalCapacity
}

func (s *SImage) GetImageType() cloudprovider.TImageType {
	return cloudprovider.ImageTypeCustomized
}

func (s *SImage) GetImageStatus() string {
	return cloudprovider.IMAGE_STATUS_ACTIVE
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
	return s.OsVersion
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
	case 4:
		return "mips64el"
	default:
		return ""
	}
}

func (s *SImage) GetMinOsDiskSizeGb() int {
	return int(s.GetSizeByte() / 1024 / 1024 / 1024)
}

func (s *SImage) GetMinRamSizeMb() int {
	return int(s.DomainMemoryResp.CurrentMemory / 1024 / 1024)
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
	if err := resp.Unmarshal(&ret, "data"); err != nil {
		return nil, err
	}
	for i := range ret {
		imgDetail, err := s.client.invokeGET(fmt.Sprintf(IMAGE_DETAIL_URL, ret[i].Id), nil, nil)
		if err != nil {
			continue
		}
		imgDetail.Unmarshal(&ret[i])
	}
	return ret, nil
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

func (s *SRegion) getImage(id string) (*SImage, error) {
	images, err := s.getImages()
	if err != nil {
		return nil, err
	}
	for i := range images {
		if images[i].GetGlobalId() == id {
			return &images[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, id)
}

func (s *SStoragecache) GetICloudImages() ([]cloudprovider.ICloudImage, error) {
	images, err := s.region.getImages()
	if err != nil {
		return nil, err
	}
	var ret []cloudprovider.ICloudImage
	for i := range images {
		images[i].cache = s
		ret = append(ret, &images[i])
	}
	return ret, nil
}

func (s *SStoragecache) GetIImageById(id string) (cloudprovider.ICloudImage, error) {
	image, err := s.region.getImage(id)
	if err != nil {
		return nil, err
	}
	image.cache = s
	return image, nil
}

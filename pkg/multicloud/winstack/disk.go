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
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/multicloud"
	"yunion.io/x/pkg/errors"
)

const (
	DISK_LIST_URL = "/api/storage/storagePools/%s/storageVolumes"
)

type SDisk struct {
	multicloud.SDisk
	multicloud.WinStackTags

	storage *SStorage

	Id         string
	Name       string
	Status     int
	Type       int
	Capacity   int
	Allocation int
	Path       string
}

func (s *SDisk) GetIStorage() (cloudprovider.ICloudStorage, error) {
	return s.storage, nil
}

func (s *SDisk) GetDiskFormat() string {
	switch s.Type {
	case 1:
		return "qcow2"
	case 2:
		return "raw"
	case 3:
		return "iso"
	default:
		return ""
	}
}

func (s *SDisk) GetDiskSizeMB() int {
	return int(s.Capacity) / 1024 / 1024
}

func (s *SDisk) GetIsAutoDelete() bool {
	return false
}

func (s *SDisk) GetTemplateId() string {
	return ""
}

func (s *SDisk) GetDiskType() string {
	return api.DISK_TYPE_SYS
}

func (s *SDisk) GetFsFormat() string {
	return ""
}

func (s *SDisk) GetIsNonPersistent() bool {
	return false
}

func (s *SDisk) GetDriver() string {
	return "scsi"
}

func (s *SDisk) GetCacheMode() string {
	return "none"
}

func (s *SDisk) GetMountpoint() string {
	return ""
}

func (s *SDisk) GetAccessPath() string {
	return s.Path
}

func (s *SDisk) Delete(ctx context.Context) error {
	return cloudprovider.ErrNotImplemented
}

func (s *SDisk) CreateISnapshot(ctx context.Context, name string, desc string) (cloudprovider.ICloudSnapshot, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (s *SDisk) GetISnapshots() ([]cloudprovider.ICloudSnapshot, error) {
	return nil, cloudprovider.ErrNotImplemented
}

func (s *SDisk) Resize(ctx context.Context, newSizeMB int64) error {
	return cloudprovider.ErrNotImplemented
}

func (s *SDisk) Reset(ctx context.Context, snapshotId string) (string, error) {
	return "", cloudprovider.ErrNotImplemented
}

func (s *SDisk) Rebuild(ctx context.Context) error {
	return cloudprovider.ErrNotImplemented
}

func (s *SDisk) GetId() string {
	return s.Id
}

func (s *SDisk) GetName() string {
	return s.Name
}

func (s *SDisk) GetGlobalId() string {
	return s.Id
}

func (s *SDisk) GetStatus() string {
	switch s.Status {
	case 1:
		return api.DISK_READY
	default:
		return api.DISK_UNKNOWN
	}
}

func (s *SStorage) GetIDisks() ([]cloudprovider.ICloudDisk, error) {
	ret := make([]cloudprovider.ICloudDisk, 0)
	disks, err := s.getDisks()
	if err != nil {
		return nil, err
	}
	for i := range disks {
		disks[i].storage = s
		ret = append(ret, &disks[i])
	}
	return ret, nil
}

func (s *SStorage) GetIDiskById(id string) (cloudprovider.ICloudDisk, error) {
	disk, err := s.getDisk(id)
	if err != nil {
		return nil, err
	}
	disk.storage = s
	return disk, nil
}

func (s *SStorage) getDisk(id string) (*SDisk, error) {
	disks, err := s.getDisks()
	if err != nil {
		return nil, err
	}
	for i := range disks {
		if disks[i].GetGlobalId() == id {
			return &disks[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, id)
}

func (s *SStorage) getDisks() ([]SDisk, error) {
	var disks []SDisk
	start, size := 1, 10
	for {
		ret, err := s.GetDisks("", start, size)
		if err != nil {
			return nil, err
		}
		for i := range ret {
			disks = append(disks, ret[i])
		}
		if len(ret) < size {
			break
		}
		start += 1
	}
	return disks, nil
}

func (s *SStorage) GetDisks(id string, start, size int) ([]SDisk, error) {
	query := make(map[string]string)
	URL := fmt.Sprintf(DISK_LIST_URL, s.Id)
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
	resp, err := s.cluster.region.client.invokeGET(URL, nil, query)
	if err != nil {
		return nil, err
	}
	var ret []SDisk

	return ret, resp.Unmarshal(&ret, "data")

}

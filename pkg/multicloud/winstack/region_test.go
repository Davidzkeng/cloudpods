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
	"testing"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/multicloud"
)

var region *SRegion

func init() {
	InitRegion()
}

func InitRegion() {
	cfg := WinStackConfig{
		cpcfg:    cloudprovider.ProviderConfig{},
		endpoint: "https://10.252.226.12/",
		user:     "admin",
		password: "passw0rd",
		debug:    false,
	}
	client, _ := NewWinStackClient(&cfg)

	region = &SRegion{
		SRegion:                  multicloud.SRegion{},
		SRegionEipBase:           multicloud.SRegionEipBase{},
		SRegionLbBase:            multicloud.SRegionLbBase{},
		SRegionOssBase:           multicloud.SRegionOssBase{},
		SRegionSecurityGroupBase: multicloud.SRegionSecurityGroupBase{},
		SRegionVpcBase:           multicloud.SRegionVpcBase{},
		SRegionZoneBase:          multicloud.SRegionZoneBase{},
		client:                   client,
	}
}

func TestSRegion_GetInstances(t *testing.T) {
	instance, err := region.GetInstances("", "", "", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(instance)
}

func TestSRegion_GetInstanceNics(t *testing.T) {
	nics, err := region.GetInstanceNics("f86d0a33-067d-4640-b74d-5ebb17b9b429")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(nics)
}

func TestSRegion_GetSecurityByVmId(t *testing.T) {
	security, err := region.GetSecurityByVmId("a0804ee6-1921-426e-b770-b723e33ffa06")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(security)
}

func TestSRegion_GetInstancesByHostId(t *testing.T) {
	instances, err := region.GetInstancesByHostId("c90dfc02-32a0-48d4-84a6-2c261434167a")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(instances)
}

func TestSRegion_GetStorages(t *testing.T) {
	storages, err := region.getStorages()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(storages))
}

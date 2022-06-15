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
	nics, err := region.GetInstanceNics("7593234e-2065-40eb-9544-b887c9eb7645")
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

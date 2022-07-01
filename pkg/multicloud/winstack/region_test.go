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
	"log"
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
	instance, err := region.GetInstancesByName("winstack-create-56")
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

func TestSRegion_GetImages(t *testing.T) {
	images, err := region.getImages()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", images)
}

func TestSRegion_CreateSecurityGroup(t *testing.T) {
	resp, err := region.CreateSecurityGroup("392dd5e8-0848-4725-bc10-b719692a5f00", "hello", "t")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(resp)
}

func TestSRegion_GetSecurityGroups(t *testing.T) {
	ret, err := region.getSecurityGroups()
	if err != nil {
		t.Fatal(err)
	}
	log.Println(ret)
}

func TestSRegion_AssignSecurityGroup(t *testing.T) {
	err := region.AssignSecurityGroup("0d040bb3-4e65-4b26-af68-8e135c2e1939", "392dd5e8-0848-4725-bc10-b719692a5f00", "d5d8123a-b531-4da5-bfb5-1976c2ff27aa")
	if err != nil {
		t.Fatalf("%+v", err)
	}
}

func TestSRegion_getEip(t *testing.T) {
	eip, err := region.getEipByIp("10.252.226.52")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(eip)
}

func TestSRegion_CreateEIP(t *testing.T) {
	opts := cloudprovider.SEip{
		Name:              "",
		BandwidthMbps:     0,
		ChargeType:        "",
		BGPType:           "",
		NetworkExternalId: "",
		IP:                "",
		ProjectId:         "",
		VpcExternalId:     "",
	}
	region.CreateEIP(&opts)
}

func TestSRegion_GetEips(t *testing.T) {
	eips, err := region.getEips()
	if err != nil {
		t.Fatal(err)
	}
	for i := range eips {
		log.Printf("%+v", eips[i].Ip)
	}
}

func TestSRegion_GetRouterNetworkId(t *testing.T) {
	ids, err := region.GetRouterNetworkId("419c5191-8329-447a-8663-9df791733d16")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ids)
}

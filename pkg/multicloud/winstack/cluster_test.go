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
	"yunion.io/x/onecloud/pkg/multicloud"
)

var cluster *SCluster

func init() {
	InitCluster()
}

func InitCluster() {
	InitRegion()
	cluster = &SCluster{
		SResourceBase: multicloud.SResourceBase{},
		STagBase:      multicloud.STagBase{},
		region:        region,
	}
}

func TestSRegion_GetClusters(t *testing.T) {
	cluster, err := region.GetClusters()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cluster)
}

func TestSRegion_GetIZoneById(t *testing.T) {
	zone, err := region.GetIZoneById("06f13108-422f-46f6-a592-6ad39844ce6b")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(zone)
}

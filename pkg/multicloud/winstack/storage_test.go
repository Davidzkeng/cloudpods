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

var storage *SStorage

func init() {
	InitStorage()
}

func InitStorage() {
	InitCluster()
	storage = &SStorage{
		STagBase:     multicloud.STagBase{},
		SStorageBase: multicloud.SStorageBase{},
		cluster:      cluster,
		Id:           "1ec3f1ac-8d4f-4e30-b332-eef092eb27df",
	}
}
func TestSStorage_GetIDisks(t *testing.T) {
	disks, err := storage.GetIDisks()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(disks[0].GetId())
}

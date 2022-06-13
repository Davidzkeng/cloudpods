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

package shell

import (
	"yunion.io/x/onecloud/pkg/multicloud/winstack"
	"yunion.io/x/onecloud/pkg/util/shellutils"
)

func init() {
	type DiskListOptions struct {
		Id        string
		MaxResult int
		NextToken string
	}
	shellutils.R(&DiskListOptions{}, "disk-list", "list disks", func(cli *winstack.SRegion, args *DiskListOptions) error {
		cluster, err := cli.GetIZones()
		if err != nil {
			return err
		}
		var disks []*winstack.SDisk

		for cl1 := range cluster {
			storages, err := cluster[cl1].GetIStorages()
			if err != nil {
				return err
			}
			for i := range storages {
				d, err := storages[i].GetIDisks()
				if err != nil {
					return err
				}
				for i2 := range d {
					t, ok := d[i2].(*winstack.SDisk)
					if ok {
						disks = append(disks, t)
					}
				}
			}
		}

		printList(disks, 0, 0, 0, []string{})
		return nil
	})
}

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
	"fmt"
	"strconv"
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/multicloud"
)

const (
	SECURITY_GROUP_LIST    = "api/network/securityGroups"
	VM_SECURITY_GROUP_LIST = "/api/network/vpc/vms/%s/security-groups"
)

type SSecurityGroup struct {
	multicloud.SResourceBase
	multicloud.WinStackTags
	region *SRegion

	Id     string
	Name   string
	Remark string
}

func (s *SSecurityGroup) GetProjectId() string {
	return ""
}

func (s *SSecurityGroup) GetDescription() string {
	return s.Remark
}

func (s *SSecurityGroup) GetRules() ([]cloudprovider.SecurityRule, error) {
	panic("implement me")
}

func (s *SSecurityGroup) GetVpcId() string {
	return api.NORMAL_VPC_ID
}

func (s *SSecurityGroup) SyncRules(common, inAdds, outAdds, inDels, outDels []cloudprovider.SecurityRule) error {
	return cloudprovider.ErrNotImplemented
}

func (s *SSecurityGroup) GetReferences() ([]cloudprovider.SecurityGroupReference, error) {
	return []cloudprovider.SecurityGroupReference{}, nil
}

func (s *SSecurityGroup) Delete() error {
	return cloudprovider.ErrNotImplemented
}

func (s *SSecurityGroup) GetId() string {
	return s.Id
}

func (s *SSecurityGroup) GetName() string {
	return s.Name
}

func (s *SSecurityGroup) GetGlobalId() string {
	return s.Id
}

func (s *SSecurityGroup) GetStatus() string {
	return api.SECGROUP_STATUS_READY
}

func (s *SRegion) GetISecurityGroupById(id string) (cloudprovider.ICloudSecurityGroup, error) {
	group, err := s.getSecurityGroupsById(id)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (s *SRegion) GetISecurityGroupByName(opts *cloudprovider.SecurityGroupFilterOptions) (cloudprovider.ICloudSecurityGroup, error) {
	group, err := s.getSecurityGroupsByName(opts.Name)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (s *SRegion) getSecurityGroupsById(id string) (*SSecurityGroup, error) {
	groups, err := s.getSecurityGroups()
	if err != nil {
		return nil, err
	}
	for i := range groups {
		if groups[i].GetGlobalId() == id {
			return &groups[i], nil
		}
	}
	return nil, cloudprovider.ErrNotFound
}

func (s *SRegion) getSecurityGroupsByName(name string) (*SSecurityGroup, error) {
	groups, err := s.GetSecurityGroups("", name, 0, 0)
	if err != nil {
		return nil, err
	}
	for i := range groups {
		if groups[i].Name == name {
			return &groups[i], nil
		}
	}
	return nil, cloudprovider.ErrNotFound
}

func (s *SRegion) getSecurityGroups() ([]SSecurityGroup, error) {
	var securityGroups []SSecurityGroup
	start, size := 0, 10
	for {
		ret, err := s.GetSecurityGroups("", "", start, size)
		if err != nil {
			return nil, err
		}
		for i := range ret {
			securityGroups = append(securityGroups, ret[i])
		}
		if len(ret) < size {
			break
		}
		start += size
	}
	return securityGroups, nil
}

func (s *SRegion) GetSecurityGroups(id, name string, start, size int) ([]SSecurityGroup, error) {
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
	if len(name) > 0 {
		query["name"] = name
		start = 0
	}
	query["start"] = strconv.Itoa(start)
	query["size"] = strconv.Itoa(size)
	resp, err := s.client.invokeGET(SECURITY_GROUP_LIST, nil, query)
	if err != nil {
		return nil, err
	}
	var ret []SSecurityGroup

	return ret, resp.Unmarshal(&ret, "data")
}

func (s *SRegion) GetSecurityByVmId(id string) ([]SSecurityGroup, error) {
	URL := fmt.Sprintf(VM_SECURITY_GROUP_LIST, id)
	resp, err := s.client.invokeGET(URL, nil, nil)
	if err != nil {
		return nil, err
	}
	var ret []SSecurityGroup
	return ret, resp.Unmarshal(&ret, "security_groups")
}

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
	"regexp"
	"strconv"
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/multicloud"
	"yunion.io/x/pkg/util/secrules"
)

const (
	SECURITY_GROUP_LIST    = "/api/sdn/v2.0/security-groups"
	VM_SECURITY_GROUP_LIST = "/api/network/vpc/vms/%s/security-groups"
)

type SSecurityGroup struct {
	multicloud.SResourceBase
	multicloud.WinStackTags
	region *SRegion

	Id                 string
	Name               string
	Description        string
	SecurityGroupRules []struct {
		Direction       string `json:"direction"`
		Ethertype       string `json:"ethertype"`
		Id              string `json:"id"`
		RemoteGroupId   string `json:"remote_group_id"`
		SecurityGroupId string `json:"security_group_id"`
		Priority        int    `json:"priority"`
		Protocol        string `json:"protocol"`
		PortRangeMax    string `json:"port_range_max"`
		PortRangeMin    string `json:"port_range_min"`
		RemoteIpPrefix  string `json:"remote_ip_prefix"`
	}
}

func (s *SSecurityGroup) GetProjectId() string {
	return ""
}

func (s *SSecurityGroup) GetDescription() string {
	return s.Description
}

func (s *SSecurityGroup) GetRules() ([]cloudprovider.SecurityRule, error) {
	var ret []cloudprovider.SecurityRule
	for _, _rule := range s.SecurityGroupRules {
		rule := cloudprovider.SecurityRule{}
		rule.Direction = secrules.DIR_IN
		rule.Priority = 1
		rule.Action = secrules.SecurityRuleAllow
		rule.Protocol = secrules.PROTO_ANY
		rule.Description = s.Description
		if _rule.Direction == "egress" {
			rule.Direction = secrules.DIR_OUT
		}

		if _rule.Protocol != "" {
			rule.Protocol = _rule.Protocol
		}
		if rule.Protocol == secrules.PROTO_TCP || rule.Protocol == secrules.PROTO_UDP {
			re := regexp.MustCompile("[0-9]+")
			var portMax, portMin int
			if res := re.FindAllString(_rule.PortRangeMax, -1); len(res) > 0 {
				portMax, _ = strconv.Atoi(res[0])
			}
			if res := re.FindAllString(_rule.PortRangeMin, -1); len(res) > 0 {
				portMin, _ = strconv.Atoi(res[0])
			}
			rule.PortStart, rule.PortEnd = portMin, portMax
		}
		if _rule.RemoteIpPrefix != "" {
			if _rule.RemoteIpPrefix == "::/0" {
				_rule.RemoteIpPrefix = "0.0.0.0/0"
			}
			rule.ParseCIDR(_rule.RemoteIpPrefix)
			err := rule.ValidateRule()
			if err != nil {
				return nil, err
			}
			ret = append(ret, rule)
		}
	}
	return ret, nil
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
	start, size := 1, 10
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
		start += 1
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

	return ret, resp.Unmarshal(&ret, "security_groups")
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

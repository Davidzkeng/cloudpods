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
	"yunion.io/x/pkg/errors"
	"yunion.io/x/pkg/util/secrules"
)

const (
	SECURITY_GROUP_LIST             = "/api/sdn/v2.0/security-groups"
	VM_SECURITY_GROUP_LIST          = "/api/network/vpc/vms/%s/security-groups"
	SECURITY_GROUP_CREATE_URL       = "/api/sdn/v2.0/security-groups"
	SECURITY_GROUP_ASSIGIN_URL      = "/api/sdn/v2.0/security-groups/%s/ports"
	SECURITY_GROUP_PORT_LIST_URL    = "/api/sdn/v2.0/security-groups/%s/ports"
	SECURITY_GROUP_RULES_DELETE_URL = "/api/sdn/v2.0/security-group-rules/%s"
	SECURITY_GROUP_RULES_CREATE_URL = "/api/sdn/v2.0/security-group-rules"
)

type SSecurityGroup struct {
	multicloud.SResourceBase
	multicloud.WinStackTags
	region *SRegion

	Id                 string
	Name               string
	Description        string
	ProjectId          string
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
	if s.ProjectId != "" {
		return s.ProjectId
	}
	return api.NORMAL_VPC_ID
}

func (s *SSecurityGroup) SyncRules(common, inAdds, outAdds, inDels, outDels []cloudprovider.SecurityRule) error {
	for _, r := range append(inDels, outDels...) {
		if len(r.ExternalId) == 0 {
			continue
		}
		err := s.region.DeleteSecurityGroupRules(r.ExternalId)
		if err != nil {
			return errors.Wrapf(err, "delSecurityGroupRule(%s)", r.ExternalId)
		}
	}
	for _, r := range append(inAdds, outAdds...) {
		err := s.region.AddSecurityGroupRule(s.Id, r)
		if err != nil {
			//if jsonError, ok := err.(*httputils.JSONClientError); ok {
			//	if jsonError.Class == "SecurityGroupRuleExists" {
			//		continue
			//	}
			//}
			return errors.Wrapf(err, "addSecgroupRules(%s)", r.String())
		}
	}
	return nil
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
	groups, err := s.getSecurityGroups()
	if err != nil {
		return nil, err
	}
	for i := range groups {
		if groups[i].GetGlobalId() == id {
			groups[i].region = s
			return &groups[i], nil
		}
	}
	return nil, cloudprovider.ErrNotFound
}

func (s *SRegion) GetISecurityGroupByName(opts *cloudprovider.SecurityGroupFilterOptions) (cloudprovider.ICloudSecurityGroup, error) {
	groups, err := s.GetSecurityGroups("", opts.Name)
	if err != nil {
		return nil, err
	}
	for i := range groups {
		if groups[i].GetName() == opts.Name {
			return &groups[i], nil
		}
	}
	return nil, cloudprovider.ErrNotFound
}

func (s *SRegion) getSecurityGroupsById(id string) (*SSecurityGroup, error) {
	groups, err := s.GetSecurityGroups(id, "")
	if err != nil {
		return nil, err
	}
	for i := range groups {
		if groups[i].Id == id {
			return &groups[i], nil
		}
	}
	return nil, cloudprovider.ErrNotFound
}

func (s *SRegion) getSecurityGroups() ([]SSecurityGroup, error) {
	var securityGroups []SSecurityGroup
	ret, err := s.GetSecurityGroups("", "")
	if err != nil {
		return nil, err
	}
	for i := range ret {
		securityGroups = append(securityGroups, ret[i])
	}

	return securityGroups, nil
}

func (s *SRegion) GetSecurityGroups(id, name string) ([]SSecurityGroup, error) {
	query := make(map[string]string)
	if len(id) > 0 {
		query["id"] = id
	}
	if len(name) > 0 {
		query["name"] = name
	}

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

func (s *SRegion) CreateISecurityGroup(conf *cloudprovider.SecurityGroupCreateInput) (cloudprovider.ICloudSecurityGroup, error) {
	return s.CreateSecurityGroup(conf.VpcId, conf.Name, conf.Desc)
}

func (s *SRegion) CreateSecurityGroup(vpcId string, name string, desc string) (cloudprovider.ICloudSecurityGroup, error) {
	securityParam := make(map[string]string)
	securityParam["name"] = name
	securityParam["description"] = desc
	securityParam["project_Id"] = vpcId
	body := make(map[string]map[string]string)
	body["security_group"] = securityParam
	resp, err := s.client.invokePOST(SECURITY_GROUP_CREATE_URL, nil, nil, body)
	if err != nil {
		return nil, err
	}
	var ret SSecurityGroup
	return &ret, resp.Unmarshal(&ret, "security_group")
}

func (s *SRegion) AssignSecurityGroup(instanceId, vpcId, secgroupId string) error {
	if _, err := s.GetInstanceById(instanceId); err != nil {
		return err
	}
	if _, err := s.getSecurityGroupsById(secgroupId); err != nil {
		return err
	}
	var portIds []string
	//查找已加入安全组的虚拟机
	secGroupInstances, err := s.GetInstancesBySecGroupId(secgroupId)
	if err != nil {
		return err
	}
	for i := range secGroupInstances {
		portIds = append(portIds, secGroupInstances[i].Name)
	}
	//查询vpc下的port列表
	vpcInstances, err := s.GetInstancesByVpcId(vpcId)
	if err != nil {
		return err
	}
	for i := range vpcInstances.VpcList {
		if vpcInstances.VpcList[i].DomainId != instanceId || len(vpcInstances.VpcList[i].InterfaceList) <= 0 {
			continue
		}
		portIds = append(portIds, vpcInstances.VpcList[i].InterfaceList[0].InterfaceId)
	}

	//todo：找出虚拟机对应的端口名字，调用加入安全组接口
	URL := fmt.Sprintf(SECURITY_GROUP_ASSIGIN_URL, secgroupId)
	body := make(map[string][]string)
	body["portIds"] = portIds
	_, err = s.client.invokePOST(URL, nil, nil, body)
	return err
}

type SSecGroupInstance struct {
	MacAddress string `json:"mac_address"`
	Name       string `json:"name"`
}

func (s *SRegion) GetInstancesBySecGroupId(secgroupId string) ([]SSecGroupInstance, error) {
	URL := fmt.Sprintf(SECURITY_GROUP_PORT_LIST_URL, secgroupId)
	resp, err := s.client.invokeGET(URL, nil, nil)
	if err != nil {
		return nil, err
	}
	var ret []SSecGroupInstance
	return ret, resp.Unmarshal(&ret, "ports")
}

func (s *SRegion) DeleteSecurityGroupRules(secgroupRuleId string) error {
	URL := fmt.Sprintf(SECURITY_GROUP_RULES_DELETE_URL, secgroupRuleId)
	_, err := s.client.invokeDELETE(URL, nil, nil)
	return err
}

func (s *SRegion) AddSecurityGroupRule(secgroupId string, rule cloudprovider.SecurityRule) error {
	direction := "ingress"
	if rule.Direction == secrules.SecurityRuleEgress {
		direction = "egress"
	}

	if rule.Protocol == secrules.PROTO_ANY {
		rule.Protocol = ""
	}

	ruleInfo := map[string]interface{}{
		"direction":         direction,
		"security_group_id": secgroupId,
		"remote_ip_prefix":  rule.IPNet.String(),
		"ethertype":         "Ipv4",
	}
	if len(rule.Protocol) > 0 {
		ruleInfo["protocol"] = rule.Protocol
	}

	params := map[string]map[string]interface{}{
		"security_group_rule": ruleInfo,
	}
	if len(rule.Ports) > 0 {
		for _, port := range rule.Ports {
			params["security_group_rule"]["port_range_max"] = port
			params["security_group_rule"]["port_range_min"] = port
			_, err := s.client.invokePOST(SECURITY_GROUP_RULES_CREATE_URL, nil, nil, params)
			if err != nil {
				return errors.Wrap(err, "SecurityGroup create")
			}
		}
		return nil
	}
	if rule.PortEnd > 0 && rule.PortStart > 0 {
		params["security_group_rule"]["port_range_min"] = rule.PortStart
		params["security_group_rule"]["port_range_max"] = rule.PortEnd
	}
	_, err := s.client.invokePOST(SECURITY_GROUP_RULES_CREATE_URL, nil, nil, params)
	return err

}

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

import "fmt"

const (
	ROUTER_LIST_URL        = "/api/sdn/v2.0/routers"
	ROUTER_CREATE_URL      = "/api/network/vpcs/%s/routers"
	ROUTER_SUBNET_LIST_URL = "/api/sdn/v2.0/routers/%s/interfaces"
)

type SRouter struct {
	Status              string        `json:"status"`
	Name                string        `json:"name"`
	Routes              []SRouteEntry `json:"routes"`
	Id                  string        `json:"id"`
	ProjectId           string        `json:"project_Id"`
	TenantId            string        `json:"tenant_id"`
	ExternalGatewayInfo struct {
		NetworkId        string `json:"network_id"`
		PortName         string `json:"port_name"`
		ExternalFixedIps []struct {
			IpAddress string `json:"ip_address"`
			SubnetId  string `json:"subnet_id"`
		} `json:"external_fixed_ips"`
	} `json:"external_gateway_info"`
	AdminStateUp bool        `json:"admin_state_up"`
	Remark       interface{} `json:"remark"`
	CreateTime   string      `json:"createTime"`
}

func (s *SRegion) GetRouter() ([]SRouter, error) {
	resp, err := s.client.invokeGET(ROUTER_LIST_URL, nil, nil)
	if err != nil {
		return nil, err
	}
	var ret []SRouter
	return ret, resp.Unmarshal(&ret, "routers")
}

type SSubnet struct {
	NetworkId string
}

func (s *SRegion) GetRouterSubnet(routerId string) ([]SSubnet, error) {
	URL := fmt.Sprintf(ROUTER_SUBNET_LIST_URL, routerId)
	resp, err := s.client.invokeGET(URL, nil, nil)
	if err != nil {
		return nil, err
	}
	var subnet []SSubnet
	return subnet, resp.Unmarshal(&subnet, "subnets")
}

func (s *SRegion) CreateRouter(vpcId string, name string, networkId string) (*SRouter, error) {
	URL := fmt.Sprintf(ROUTER_CREATE_URL, vpcId)
	body := make(map[string]string)
	body["name"] = name
	body["network_id"] = networkId
	resp, err := s.client.invokePOST(URL, nil, nil, body)
	if err != nil {
		return nil, err
	}
	var ret SRouter
	return &ret, resp.Unmarshal(&ret)
}

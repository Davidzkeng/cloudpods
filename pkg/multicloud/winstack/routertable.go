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

	"github.com/pkg/errors"
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/multicloud"
)

const (
	ROUTE_INTERFACE_CREATE_URL = "/api/sdn/v2.0/routers/%s/add_router_interface"
)

type SRouteEntry struct {
	multicloud.SResourceBase
	multicloud.WinStackTags
	Destination string `json:"destination"`
	Nexthop     string `json:"nexthop"`
}

func (s *SRouteEntry) GetId() string {
	return s.Destination + ":" + s.Nexthop
}
func (s *SRouteEntry) GetName() string {
	return ""
}
func (s *SRouteEntry) GetGlobalId() string {
	return s.GetId()
}

func (s *SRouteEntry) GetStatus() string {
	return ""
}

func (s *SRouteEntry) Refresh() error {
	return nil
}

func (s *SRouteEntry) IsEmulated() bool {
	return false
}

func (s *SRouteEntry) GetType() string {
	return api.ROUTE_ENTRY_TYPE_CUSTOM
}

func (s *SRouteEntry) GetCidr() string {
	return s.Destination
}

func (s *SRouteEntry) GetNextHopType() string {
	return s.Nexthop
}

func (s *SRouteEntry) GetNextHop() string {
	return s.Nexthop
}

type SRouteTable struct {
	multicloud.SResourceBase
	multicloud.WinStackTags
	vpc     *SVpc
	entries []SRouteEntry
	router  *SRouter
}

func (s *SRouteTable) GetDescription() string {
	return ""
}

func (s *SRouteTable) GetId() string {
	return s.GetGlobalId()
}

func (s *SRouteTable) GetGlobalId() string {
	return s.router.Id
}

func (s *SRouteTable) GetName() string {
	return s.router.Name
}

func (s *SRouteTable) GetRegionId() string {
	return s.vpc.region.GetId()
}

func (s *SRouteTable) GetType() cloudprovider.RouteTableType {
	return cloudprovider.RouteTableTypeSystem
}

func (s *SRouteTable) GetVpcId() string {
	return s.vpc.GetId()
}

func (s *SRouteTable) GetIRoutes() ([]cloudprovider.ICloudRoute, error) {
	ret := []cloudprovider.ICloudRoute{}
	for index := range s.entries {
		ret = append(ret, &s.entries[index])
	}
	return ret, nil
}

func (s *SRouteTable) GetStatus() string {
	if s.router.Status == "ACTIVE" {
		return api.ROUTE_TABLE_AVAILABLE
	}
	return s.router.Status
}

func (s *SRouteTable) IsEmulated() bool {
	return false
}

func (s *SRouteTable) Refresh() error {
	return nil
}

func (s *SRouteTable) GetAssociations() []cloudprovider.RouteTableAssociation {
	var ret []cloudprovider.RouteTableAssociation
	networks, err := s.vpc.region.GetRouterSubnet(s.router.Id)
	if err != nil {
		return ret
	}
	for i := range networks {
		association := cloudprovider.RouteTableAssociation{
			AssociationId:        s.GetId() + ":" + networks[i].NetworkId,
			AssociationType:      cloudprovider.RouteTableAssociaToSubnet,
			AssociatedResourceId: networks[i].NetworkId,
		}
		ret = append(ret, association)
	}
	return ret
}

func (s *SRouteTable) CreateRoute(route cloudprovider.RouteSet) error {
	return cloudprovider.ErrNotSupported
}

func (s *SRouteTable) UpdateRoute(route cloudprovider.RouteSet) error {
	return cloudprovider.ErrNotSupported
}

func (s *SRouteTable) RemoveRoute(route cloudprovider.RouteSet) error {
	return cloudprovider.ErrNotSupported
}

func (s *SRouteTable) AddRouteInterface(routeInterface cloudprovider.RouteInterface) error {
	URL := fmt.Sprintf(ROUTE_INTERFACE_CREATE_URL, s.router.Id)
	body := make(map[string]string)
	body["project_Id"] = routeInterface.NetworkId
	_, err := s.vpc.region.client.invokePUT(URL, nil, nil, body)
	return err
}

func (s *SVpc) GetIRouteTables() ([]cloudprovider.ICloudRouteTable, error) {
	routers, err := s.region.GetRouter()
	if err != nil {
		return nil, errors.Wrapf(err, "vpc.region.GetRouter")
	}

	var routeTables []SRouteTable
	for i := range routers {
		if len(routers[i].Routes) < 1 {
			continue
		}
		if routers[i].ProjectId == s.GetId() {
			var routerTable SRouteTable
			routerTable.entries = routers[i].Routes
			routerTable.router = &routers[i]
			routerTable.vpc = s
			routeTables = append(routeTables, routerTable)
			break
		}
	}
	ret := []cloudprovider.ICloudRouteTable{}
	for i := range routeTables {
		routeTables[i].vpc = s
		ret = append(ret, &routeTables[i])
	}
	return ret, nil
}

func (s *SVpc) GetIRouteTableById(id string) (cloudprovider.ICloudRouteTable, error) {
	tables, err := s.GetIRouteTables()
	if err != nil {
		return nil, err
	}
	for i := range tables {
		if tables[i].GetGlobalId() == id {
			return tables[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, id)
}

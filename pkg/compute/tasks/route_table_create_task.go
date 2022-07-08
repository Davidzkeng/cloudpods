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

package tasks

import (
	"context"
	"time"

	"yunion.io/x/onecloud/pkg/cloudprovider"

	"yunion.io/x/onecloud/pkg/cloudcommon/notifyclient"

	"yunion.io/x/jsonutils"
	"yunion.io/x/pkg/errors"

	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/cloudcommon/db/taskman"
	"yunion.io/x/onecloud/pkg/compute/models"
	"yunion.io/x/onecloud/pkg/util/logclient"
)

type RouteTableCreateTask struct {
	taskman.STask
}

func (self *RouteTableCreateTask) taskFailed(ctx context.Context, routeTable *models.SRouteTable, err error) {
	routeTable.SetStatus(self.GetUserCred(), api.ROUTE_TABLE_CREATEFAIL, err.Error())
	db.OpsLog.LogEvent(routeTable, db.ACT_CREATE, err, self.GetUserCred())
	logclient.AddActionLogWithContext(ctx, routeTable, logclient.ACT_CREATE, err, self.UserCred, false)
	self.SetStageFailed(ctx, jsonutils.NewString(err.Error()))
}

func init() {
	taskman.RegisterTask(RouteTableCreateTask{})
}

func (self *RouteTableCreateTask) OnInit(ctx context.Context, obj db.IStandaloneModel, data jsonutils.JSONObject) {
	routeTable := obj.(*models.SRouteTable)
	routeTable.SetStatus(self.UserCred, api.ROUTE_TABLE_PENDING, "")

	_vpc, err := routeTable.GetVpc()
	if err != nil {
		self.taskFailed(ctx, routeTable, errors.Wrapf(err, "GetVpc"))
		return
	}
	ivpc, err := _vpc.GetIVpc(ctx)
	if err != nil {
		self.taskFailed(ctx, routeTable, errors.Wrapf(err, "GetIVpc"))
		return
	}
	self.SetStage("OnCreateRouteTableComplete", nil)
	taskman.LocalTaskRun(self, func() (jsonutils.JSONObject, error) {
		_network, err := db.FetchById(models.NetworkManager, routeTable.NetworkId)
		if err != nil {
			self.taskFailed(ctx, routeTable, errors.Wrapf(err, "GetNetworkManager"))
			return nil, errors.Wrapf(err, "GetNetworkManager")
		}
		network := _network.(*models.SNetwork)
		opts := &cloudprovider.RouteTableCreateOptions{Name: routeTable.Name, NetworkId: network.ExternalId}
		iRouteTable, err := ivpc.CreateIRouteTable(opts)
		if err != nil {
			self.taskFailed(ctx, routeTable, errors.Wrapf(err, "CreateIRouteTable"))
			return nil, errors.Wrapf(err, "CreateIRouteTable")
		}
		err = cloudprovider.WaitStatus(iRouteTable, api.ROUTE_TABLE_AVAILABLE, time.Second*5, time.Minute*10)
		if err != nil {
			self.taskFailed(ctx, routeTable, errors.Wrapf(err, "cloudprovider.WaitStatus"))
			return nil, errors.Wrapf(err, "cloudprovider.WaitStatus")
		}
		routes, _ := iRouteTable.GetIRoutes()
		//删除静态路由，再绑定自定义静态路由
		for i := range routes {
			iRouteTable.RemoveRoute(cloudprovider.RouteSet{
				Destination: routes[i].GetCidr(),
				NextHopType: routes[i].GetNextHopType(),
				NextHop:     routes[i].GetNextHop(),
			})
		}
		if routeTable.Routes != nil {
			for _, v := range *routeTable.Routes {
				iRouteTable.CreateRoute(cloudprovider.RouteSet{
					Destination: v.Cidr,
					NextHopType: v.NextHopType,
					NextHop:     v.NextHopId,
				})
			}
		}
		iNewRouteTable, err := ivpc.GetIRouteTableById(iRouteTable.GetId())
		if err != nil {
			self.taskFailed(ctx, routeTable, errors.Wrapf(err, "GetRouteTable"))
			return nil, errors.Wrapf(err, "GetNetwork")
		}
		routeTable.SyncWithCloudRouteTable(ctx, self.UserCred, _vpc, iNewRouteTable, routeTable.GetCloudprovider())
		routeTable.SyncRouteTableRouteSets(ctx, self.UserCred, iNewRouteTable, routeTable.GetCloudprovider())
		return nil, nil
	})

}

func (self *RouteTableCreateTask) OnCreateRouteTableComplete(ctx context.Context, routeTable *models.SRouteTable, data jsonutils.JSONObject) {
	logclient.AddActionLogWithStartable(self, routeTable, logclient.ACT_CREATE, nil, self.UserCred, true)
	notifyclient.EventNotify(ctx, self.UserCred, notifyclient.SEventNotifyParam{
		Obj:    routeTable,
		Action: notifyclient.ActionCreate,
	})
	self.SetStageComplete(ctx, nil)
}

func (self *RouteTableCreateTask) OnCreateRouteTableCompleteFailed(ctx context.Context, routeTable *models.SRouteTable, data jsonutils.JSONObject) {
	self.taskFailed(ctx, routeTable, errors.Errorf(data.String()))
}

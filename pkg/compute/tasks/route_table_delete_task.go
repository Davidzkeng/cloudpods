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

	"yunion.io/x/onecloud/pkg/cloudcommon/notifyclient"

	"yunion.io/x/jsonutils"
	"yunion.io/x/pkg/errors"

	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/cloudcommon/db/taskman"
	"yunion.io/x/onecloud/pkg/compute/models"
	"yunion.io/x/onecloud/pkg/util/logclient"
)

type RouteTableDeleteTask struct {
	taskman.STask
}

func (self *RouteTableDeleteTask) taskFailed(ctx context.Context, routeTable *models.SRouteTable, err error) {
	routeTable.SetStatus(self.GetUserCred(), api.ROUTE_TABLE_DELETEFAIL, err.Error())
	db.OpsLog.LogEvent(routeTable, db.ACT_DELETE, err, self.GetUserCred())
	logclient.AddActionLogWithContext(ctx, routeTable, logclient.ACT_DELETE, err, self.UserCred, false)
	self.SetStageFailed(ctx, jsonutils.NewString(err.Error()))
}

func init() {
	taskman.RegisterTask(RouteTableDeleteTask{})
}

func (self *RouteTableDeleteTask) OnInit(ctx context.Context, obj db.IStandaloneModel, data jsonutils.JSONObject) {
	routeTable := obj.(*models.SRouteTable)
	routeTable.SetStatus(self.UserCred, api.ROUTE_TABLE_DELETING, "")

	iRouteTable, err := routeTable.GetICloudRouteTable(ctx)
	if err != nil {
		self.taskFailed(ctx, routeTable, errors.Wrapf(err, "GetICloudRouteTable"))
		return
	}
	self.SetStage("OnDeleteRouteTableComplete", nil)
	err = iRouteTable.Delete()
	if err != nil {
		self.taskFailed(ctx, routeTable, err)
		return
	}
}

func (self *RouteTableDeleteTask) OnDeleteRouteTableComplete(ctx context.Context, routeTable *models.SRouteTable, data jsonutils.JSONObject) {
	logclient.AddActionLogWithStartable(self, routeTable, logclient.ACT_DELETE, nil, self.UserCred, true)
	notifyclient.EventNotify(ctx, self.UserCred, notifyclient.SEventNotifyParam{
		Obj:    routeTable,
		Action: notifyclient.ActionDelete,
	})
	self.SetStageComplete(ctx, nil)
}

func (self *RouteTableDeleteTask) OnDeleteTableCompleteFailed(ctx context.Context, routeTable *models.SRouteTable, data jsonutils.JSONObject) {
	self.taskFailed(ctx, routeTable, errors.Errorf(data.String()))
}

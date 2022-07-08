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
	"yunion.io/x/log"

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

type RouteTableAssociationCreateTask struct {
	taskman.STask
}

func (self *RouteTableAssociationCreateTask) taskFailed(ctx context.Context, routeTableAssociation *models.SRouteTableAssociation, err error) {
	routeTableAssociation.SetStatus(self.GetUserCred(), api.ROUTE_TABLE_CREATEFAIL, err.Error())
	db.OpsLog.LogEvent(routeTableAssociation, db.ACT_CREATE, err, self.GetUserCred())
	logclient.AddActionLogWithContext(ctx, routeTableAssociation, logclient.ACT_CREATE, err, self.UserCred, false)
	self.SetStageFailed(ctx, jsonutils.NewString(err.Error()))
}

func init() {
	taskman.RegisterTask(RouteTableAssociationCreateTask{})
}

func (self *RouteTableAssociationCreateTask) OnInit(ctx context.Context, obj db.IStandaloneModel, data jsonutils.JSONObject) {
	association := obj.(*models.SRouteTableAssociation)
	association.SetStatus(self.UserCred, api.ROUTE_TABLE_ASSOCIATION_PENDING, "")
	_routeTable, err := association.GetRouteTable()
	if err != nil {
		self.taskFailed(ctx, association, errors.Wrapf(err, "GetRouteTable"))
		return
	}
	log.Errorf("getrouteTable:%v", _routeTable)
	iRouteTable, err := _routeTable.GetICloudRouteTable(ctx)
	if err != nil {
		self.taskFailed(ctx, association, errors.Wrapf(err, "GetIRouteTable"))
		return
	}

	self.SetStage("OnCreateRouteTableAssociationComplete", nil)
	taskman.LocalTaskRun(self, func() (jsonutils.JSONObject, error) {
		opts := cloudprovider.RouteTableAssociation{
			AssociationType:      cloudprovider.RouteTableAssociationType(association.AssociationType),
			AssociatedResourceId: association.ExtAssociatedResourceId,
		}
		err := iRouteTable.CreateAssociations(opts)
		if err != nil {
			self.taskFailed(ctx, association, errors.Wrapf(err, "CreateAssociations"))
			return nil, errors.Wrapf(err, "CreateAssociations")
		}
		err = cloudprovider.WaitCreated(time.Second*5, time.Minute*5, func() bool {
			associations := iRouteTable.GetAssociations()
			for i := range associations {
				if associations[i].AssociatedResourceId == association.ExtAssociatedResourceId {
					db.Update(association, func() error {
						association.ExternalId = associations[i].GetGlobalId()
						return nil
					})
					return true
				}
			}
			return false
		})
		if err != nil {
			self.taskFailed(ctx, association, errors.Wrapf(err, "CreateAssociations"))
			return nil, errors.Wrapf(err, "CreateAssociations")
		}
		association.SetStatus(self.UserCred, api.ROUTE_TABLE_ASSOCIATION_AVALIABLE, "")
		//查询绑定的数据
		return nil, nil
	})

}

func (self *RouteTableAssociationCreateTask) OnCreateRouteTableAssociationComplete(ctx context.Context, routeTableAssociation *models.SRouteTableAssociation, data jsonutils.JSONObject) {
	logclient.AddActionLogWithStartable(self, routeTableAssociation, logclient.ACT_CREATE, nil, self.UserCred, true)
	notifyclient.EventNotify(ctx, self.UserCred, notifyclient.SEventNotifyParam{
		Obj:    routeTableAssociation,
		Action: notifyclient.ActionCreate,
	})
	self.SetStageComplete(ctx, nil)
}

func (self *RouteTableAssociationCreateTask) OnCreateRouteTableAssociationCompleteFailed(ctx context.Context, routeTableAssociation *models.SRouteTableAssociation, data jsonutils.JSONObject) {
	self.taskFailed(ctx, routeTableAssociation, errors.Errorf(data.String()))
}

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

package models

import (
	"context"
	"yunion.io/x/jsonutils"
	"yunion.io/x/onecloud/pkg/cloudcommon/db/taskman"
	"yunion.io/x/onecloud/pkg/httperrors"

	"yunion.io/x/pkg/errors"
	"yunion.io/x/sqlchemy"

	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudcommon/db"
	"yunion.io/x/onecloud/pkg/cloudcommon/db/lockman"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/util/stringutils2"
)

type SRouteTableAssociationManager struct {
	db.SStatusStandaloneResourceBaseManager
	db.SExternalizedResourceBaseManager
	SRouteTableResourceBaseManager
}

var RouteTableAssociationManager *SRouteTableAssociationManager

func init() {
	RouteTableAssociationManager = &SRouteTableAssociationManager{
		SStatusStandaloneResourceBaseManager: db.NewStatusStandaloneResourceBaseManager(
			SRouteTableAssociation{},
			"route_table_associations_tbl",
			"route_table_association",
			"route_table_associations",
		),
	}
	RouteTableAssociationManager.SetVirtualObject(RouteTableAssociationManager)
}

type SRouteTableAssociation struct {
	db.SStatusStandaloneResourceBase
	db.SExternalizedResourceBase
	SRouteTableResourceBase
	AssociationType         string `width:"36" charset:"ascii" nullable:"false" list:"user" update:"domain" create:"domain_required"`
	AssociatedResourceId    string `width:"36" charset:"ascii" nullable:"true" list:"user" update:"domain" create:"domain_required"`
	ExtAssociatedResourceId string `width:"36" charset:"ascii" nullable:"false" list:"user" update:"domain" create:"domain_required"`
}

func (manager *SRouteTableAssociationManager) GetContextManagers() [][]db.IModelManager {
	return [][]db.IModelManager{
		{RouteTableManager},
	}
}

func (manager *SRouteTableAssociationManager) ListItemFilter(
	ctx context.Context,
	q *sqlchemy.SQuery,
	userCred mcclient.TokenCredential,
	query api.RouteTableAssociationListInput,
) (*sqlchemy.SQuery, error) {
	var err error
	q, err = manager.SStatusStandaloneResourceBaseManager.ListItemFilter(ctx, q, userCred, query.StatusStandaloneResourceListInput)
	if err != nil {
		return nil, errors.Wrap(err, "SStatusStandaloneResourceBaseManager.ListItemFilter")
	}

	q, err = manager.SExternalizedResourceBaseManager.ListItemFilter(ctx, q, userCred, query.ExternalizedResourceBaseListInput)
	if err != nil {
		return nil, errors.Wrap(err, "SExternalizedResourceBaseManager.ListItemFilter")
	}

	q, err = manager.SRouteTableResourceBaseManager.ListItemFilter(ctx, q, userCred, query.RouteTableFilterList)
	if err != nil {
		return nil, errors.Wrap(err, "SRouteTableResourceBaseManager.ListItemFilter")
	}
	return q, nil
}

func (self *SRouteTableAssociation) syncRemoveAssociation(ctx context.Context, userCred mcclient.TokenCredential) error {
	lockman.LockObject(ctx, self)
	defer lockman.ReleaseObject(ctx, self)

	err := self.ValidateDeleteCondition(ctx, nil)
	if err != nil {
		return err
	}
	err = self.RealDelete(ctx, userCred)
	return err
}

func (self *SRouteTableAssociation) syncWithCloudAssociation(ctx context.Context,
	userCred mcclient.TokenCredential,
	provider *SCloudprovider,
	cloudAssociation cloudprovider.RouteTableAssociation,
) error {
	AssociatedResourceId := ""
	if cloudAssociation.AssociationType == cloudprovider.RouteTableAssociaToSubnet {
		routeTable, err := self.GetRouteTable()
		if err != nil {
			return errors.Wrap(err, "self.GetRouteTable()")
		}
		vpc, _ := routeTable.GetVpc()
		subnet, err := vpc.GetNetworkByExtId(cloudAssociation.AssociatedResourceId)
		if err == nil {
			AssociatedResourceId = subnet.GetId()
		}
	}

	diff, err := db.UpdateWithLock(ctx, self, func() error {
		self.AssociationType = string(cloudAssociation.AssociationType)
		self.ExtAssociatedResourceId = cloudAssociation.AssociatedResourceId
		self.AssociatedResourceId = AssociatedResourceId
		return nil
	})
	if err != nil {
		return err
	}

	db.OpsLog.LogSyncUpdate(self, diff, userCred)
	return nil
}

func (manager *SRouteTableAssociationManager) newAssociationFromCloud(
	ctx context.Context,
	userCred mcclient.TokenCredential,
	routeTable *SRouteTable,
	provider *SCloudprovider,
	cloudAssociation cloudprovider.RouteTableAssociation,
) (*SRouteTableAssociation, error) {
	association := &SRouteTableAssociation{
		AssociationType:         string(cloudAssociation.AssociationType),
		ExtAssociatedResourceId: cloudAssociation.AssociatedResourceId,
	}
	association.RouteTableId = routeTable.GetId()
	association.ExternalId = cloudAssociation.GetGlobalId()
	if association.AssociationType == string(cloudprovider.RouteTableAssociaToSubnet) {
		vpc, _ := routeTable.GetVpc()
		subnet, err := vpc.GetNetworkByExtId(association.ExtAssociatedResourceId)
		if err == nil {
			association.AssociatedResourceId = subnet.GetId()
		}
	}

	association.SetModelManager(manager, association)
	if err := manager.TableSpec().Insert(ctx, association); err != nil {
		return nil, err
	}

	db.OpsLog.LogEvent(association, db.ACT_CREATE, association.GetShortDesc(ctx), userCred)
	return association, nil
}

func (self *SRouteTableAssociation) RealDelete(ctx context.Context, userCred mcclient.TokenCredential) error {
	return self.SStatusStandaloneResourceBase.Delete(ctx, userCred)
}

func (self *SRouteTableAssociation) GetRouteTable() (*SRouteTable, error) {
	routeTable, err := RouteTableManager.FetchById(self.RouteTableId)
	if err != nil {
		return nil, errors.Wrapf(err, "RouteTableManager.FetchById(%s)", self.RouteTableId)
	}
	return routeTable.(*SRouteTable), nil
}

func (manager *SRouteTableAssociationManager) ListItemExportKeys(ctx context.Context, q *sqlchemy.SQuery, userCred mcclient.TokenCredential, keys stringutils2.SSortedStrings) (*sqlchemy.SQuery, error) {
	var err error
	q, err = manager.SStatusStandaloneResourceBaseManager.ListItemExportKeys(ctx, q, userCred, keys)
	if err != nil {
		return nil, errors.Wrap(err, "SStatusStandaloneResourceBaseManager.ListItemExportKeys")
	}

	q, err = manager.SRouteTableResourceBaseManager.ListItemExportKeys(ctx, q, userCred, keys)
	if err != nil {
		return nil, errors.Wrap(err, "SRouteTableResourceBaseManager.ListItemExportKeys")
	}

	return q, nil
}

func (manager *SRouteTableAssociationManager) CreateAssociation(ctx context.Context, association *SRouteTableAssociation) error {
	err := manager.TableSpec().Insert(ctx, association)
	if err != nil {
		return errors.Wrap(err, "Insert classic network")
	}
	return nil
}

func (manager *SRouteTableAssociationManager) ValidateCreateData(
	ctx context.Context,
	userCred mcclient.TokenCredential,
	ownerId mcclient.IIdentityProvider,
	query jsonutils.JSONObject,
	input api.RouteTableAssociationCreateInput,
) (api.RouteTableAssociationCreateInput, error) {
	var err error
	input, err = manager.validateCreateData(input)
	if err != nil {
		return input, errors.Wrap(err, "SStatusStandaloneResourceBaseManager.ValidateCreateData")
	}
	input.StatusStandaloneResourceCreateInput, err = manager.SStatusStandaloneResourceBaseManager.ValidateCreateData(ctx, userCred, ownerId, query, input.StatusStandaloneResourceCreateInput)
	if err != nil {
		return input, errors.Wrap(err, "SStatusStandaloneResourceBaseManager.ValidateCreateData")
	}
	return input, nil
}

func (self *SRouteTableAssociation) PostCreate(ctx context.Context, userCred mcclient.TokenCredential, ownerId mcclient.IIdentityProvider, query jsonutils.JSONObject, data jsonutils.JSONObject) {
	defer func() {
		self.SStatusStandaloneResourceBase.PostCreate(ctx, userCred, ownerId, query, data)
	}()
	task, err := taskman.TaskManager.NewTask(ctx, "RouteTableAssociationCreateTask", self, userCred, nil, "", "", nil)
	if err != nil {
		self.SetStatus(userCred, api.ROUTE_TABLE_ASSOCIATION_CREATEFAIL, errors.Wrapf(err, "NewTask").Error())
		return
	}
	task.ScheduleRun(nil)
	return
}

func (manager *SRouteTableAssociationManager) PerformBatchcreate(ctx context.Context, userCred mcclient.TokenCredential, query jsonutils.JSONObject, data jsonutils.JSONObject) (jsonutils.JSONObject, error) {
	routeTableId, err := data.GetString("route_table_id")
	if err != nil {
		return nil, err
	}
	associations, err := data.GetArray("association")
	if err != nil {
		return nil, err
	}
	for i := range associations {
		var input api.RouteTableAssociationCreateInput
		input.RouteTableId = routeTableId
		err := associations[i].Unmarshal(&input)
		if err != nil {
			return nil, err
		}
		input, err = manager.validateCreateData(input)
		if err != nil {
			return nil, err
		}
		add := SRouteTableAssociation{}
		add.AssociationType = string(input.AssociationType)
		add.AssociatedResourceId = input.AssociatedResourceId
		add.ExtAssociatedResourceId = input.ExtAssociatedResourceId
		add.RouteTableId = routeTableId
		add.SetModelManager(manager, &add)
		err = manager.TableSpec().Insert(ctx, &add)
		if err != nil {
			return nil, err
		}
		task, err := taskman.TaskManager.NewTask(ctx, "RouteTableAssociationCreateTask", &add, userCred, nil, "", "", nil)
		if err != nil {
			return nil, errors.Wrapf(err, "RouteTableAssociationCreateTask")
		}

		task.ScheduleRun(nil)
	}
	return nil, nil
}

func (manager *SRouteTableAssociationManager) validateCreateData(input api.RouteTableAssociationCreateInput) (api.RouteTableAssociationCreateInput, error) {
	if input.AssociationType != api.RouteTableAssociaToSubnet && input.AssociationType != api.RouteTableAssociaToRouter {
		return input, httperrors.NewInputParameterError("invalid association_type %s ", input.AssociationType)
	}
	n, _ := RouteTableAssociationManager.Query().
		Equals("route_table_id", input.RouteTableId).
		Equals("associated_resource_id", input.AssociatedResourceId).
		Equals("association_type", input.AssociationType).CountWithError()
	if n > 0 {
		return input, httperrors.NewInputParameterError("associated_resource_id[%s] already exist ", input.AssociatedResourceId)
	}

	if input.AssociationType == api.RouteTableAssociaToSubnet {
		_network, err := NetworkManager.FetchById(input.AssociatedResourceId)
		if err != nil {
			return input, httperrors.NewInputParameterError("invalid associated_resource_id %s ", input.AssociatedResourceId)
		}
		network := _network.(*SNetwork)
		input.ExtAssociatedResourceId = network.ExternalId
	}
	return input, nil
}

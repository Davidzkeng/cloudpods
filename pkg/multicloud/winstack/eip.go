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
	"time"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/util/billing"
)

type SEip struct {
}

func (s *SEip) GetBillingType() string {
	panic("implement me")
}

func (s *SEip) GetExpiredAt() time.Time {
	panic("implement me")
}

func (s *SEip) SetAutoRenew(bc billing.SBillingCycle) error {
	panic("implement me")
}

func (s *SEip) Renew(bc billing.SBillingCycle) error {
	panic("implement me")
}

func (s *SEip) IsAutoRenew() bool {
	panic("implement me")
}

func (s *SEip) GetId() string {
	panic("implement me")
}

func (s *SEip) GetName() string {
	panic("implement me")
}

func (s *SEip) GetGlobalId() string {
	panic("implement me")
}

func (s *SEip) GetCreatedAt() time.Time {
	panic("implement me")
}

func (s *SEip) GetStatus() string {
	panic("implement me")
}

func (s *SEip) Refresh() error {
	panic("implement me")
}

func (s *SEip) IsEmulated() bool {
	panic("implement me")
}

func (s *SEip) GetSysTags() map[string]string {
	panic("implement me")
}

func (s *SEip) GetTags() (map[string]string, error) {
	panic("implement me")
}

func (s *SEip) SetTags(tags map[string]string, replace bool) error {
	panic("implement me")
}

func (s *SEip) GetProjectId() string {
	panic("implement me")
}

func (s *SEip) GetIpAddr() string {
	panic("implement me")
}

func (s *SEip) GetMode() string {
	panic("implement me")
}

func (s *SEip) GetINetworkId() string {
	panic("implement me")
}

func (s *SEip) GetAssociationType() string {
	panic("implement me")
}

func (s *SEip) GetAssociationExternalId() string {
	panic("implement me")
}

func (s *SEip) GetBandwidth() int {
	panic("implement me")
}

func (s *SEip) GetInternetChargeType() string {
	panic("implement me")
}

func (s *SEip) Delete() error {
	panic("implement me")
}

func (s *SEip) Associate(conf *cloudprovider.AssociateConfig) error {
	panic("implement me")
}

func (s *SEip) Dissociate() error {
	panic("implement me")
}

func (s *SEip) ChangeBandwidth(bw int) error {
	panic("implement me")
}

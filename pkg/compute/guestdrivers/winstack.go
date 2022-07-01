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

package guestdrivers

import (
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudcommon/db/quotas"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/compute/models"
	"yunion.io/x/onecloud/pkg/compute/options"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/util/billing"
	"yunion.io/x/onecloud/pkg/util/rbacutils"
)

type SWinStackGuestDriver struct {
	SManagedVirtualizedGuestDriver
}

func init() {
	driver := SWinStackGuestDriver{}
	models.RegisterGuestDriver(&driver)
}

func (self *SWinStackGuestDriver) GetHypervisor() string {
	return api.HYPERVISOR_WINSTACK
}

func (self *SWinStackGuestDriver) GetProvider() string {
	return api.CLOUD_PROVIDER_WINSTACK
}

func (self *SWinStackGuestDriver) GetMinimalSysDiskSizeGb() int {
	return options.Options.DefaultDiskSizeMB / 1024
}

func (self *SWinStackGuestDriver) GetInstanceCapability() cloudprovider.SInstanceCapability {
	return cloudprovider.SInstanceCapability{
		Hypervisor: self.GetHypervisor(),
		Provider:   self.GetProvider(),
		DefaultAccount: cloudprovider.SDefaultAccount{
			Linux: cloudprovider.SOsDefaultAccount{
				DefaultAccount: api.VM_DEFAULT_LINUX_LOGIN_USER,
				Changeable:     true,
			},
			Windows: cloudprovider.SOsDefaultAccount{
				DefaultAccount: api.VM_DEFAULT_WINDOWS_LOGIN_USER,
				Changeable:     false,
			},
		},
	}
}

func (self *SWinStackGuestDriver) GetComputeQuotaKeys(scope rbacutils.TRbacScope, ownerId mcclient.IIdentityProvider, brand string) models.SComputeResourceKeys {
	keys := models.SComputeResourceKeys{}
	keys.SBaseProjectQuotaKeys = quotas.OwnerIdProjectQuotaKeys(scope, ownerId)
	keys.CloudEnv = api.CLOUD_ENV_PRIVATE_CLOUD
	keys.Provider = api.CLOUD_PROVIDER_WINSTACK
	keys.Brand = api.CLOUD_PROVIDER_WINSTACK
	keys.Hypervisor = api.HYPERVISOR_WINSTACK
	return keys
}

func (self *SWinStackGuestDriver) GetDefaultSysDiskBackend() string {
	return ""
}

func (self *SWinStackGuestDriver) GetStorageTypes() []string {
	return []string{
		api.STORAGE_WINSTACK_LOCAL,
		api.STORAGE_WINSTACK_IPSAN,
		api.STORAGE_WINSTACK_FCSAN,
		api.STORAGE_WINSTACK_CEPH,
		api.STORAGE_WINSTACK_NAS,
		api.STORAGE_WINSTACK_NVME,
	}
}

func (self *SWinStackGuestDriver) IsNeedInjectPasswordByCloudInit() bool {
	return true
}

func (self *SWinStackGuestDriver) IsWindowsUserDataTypeNeedEncode() bool {
	return true
}

func (self *SWinStackGuestDriver) GetGuestInitialStateAfterCreate() string {
	return api.VM_RUNNING
}

func (self *SWinStackGuestDriver) GetGuestInitialStateAfterRebuild() string {
	return api.VM_READY
}

func (self *SWinStackGuestDriver) GetUserDataType() string {
	return cloudprovider.CLOUD_SHELL
}

func (self *SWinStackGuestDriver) AllowReconfigGuest() bool {
	return true
}

func (self *SWinStackGuestDriver) IsSupportedBillingCycle(bc billing.SBillingCycle) bool {
	return false
}

func (self *SWinStackGuestDriver) DoScheduleCPUFilter() bool { return false }

func (self *SWinStackGuestDriver) DoScheduleMemoryFilter() bool { return true }

func (self *SWinStackGuestDriver) DoScheduleSKUFilter() bool { return false }

func (self *SWinStackGuestDriver) DoScheduleStorageFilter() bool { return true }

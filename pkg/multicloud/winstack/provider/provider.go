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

package provider

import (
	"context"
	"fmt"
	"yunion.io/x/jsonutils"
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/httperrors"
	"yunion.io/x/onecloud/pkg/mcclient"
	"yunion.io/x/onecloud/pkg/multicloud/winstack"
	"yunion.io/x/pkg/errors"
)

type SWinStackProviderFactory struct {
	cloudprovider.SPrivateCloudBaseProviderFactory
}

func (s *SWinStackProviderFactory) GetProvider(cfg cloudprovider.ProviderConfig) (cloudprovider.ICloudProvider, error) {
	client, err := winstack.NewWinStackClient(
		winstack.NewWinStackConfig(
			cfg.URL, cfg.Account, cfg.Secret,
		).CloudproviderConfig(cfg),
	)
	if err != nil {
		return nil, err
	}
	return &SWinStackProvider{
		SBaseProvider: cloudprovider.NewBaseProvider(s),
		client:        client,
	}, nil
}

func (s *SWinStackProviderFactory) GetClientRC(info cloudprovider.SProviderInfo) (map[string]string, error) {
	return map[string]string{
		"WINSTACK_ENDPOINT": info.Url,
		"WINSTACK_USER":     info.Account,
		"WINSTACK_PASSWORD": info.Secret,
	}, nil
}

func (s *SWinStackProviderFactory) GetId() string {
	return winstack.CLOUD_PROVIDER_WINSTACK
}

func (s *SWinStackProviderFactory) GetName() string {
	return winstack.CLOUD_PROVIDER_WINSTACK
}

func (s *SWinStackProviderFactory) ValidateChangeBandwidth(instanceId string, bandwidth int64) error {
	return fmt.Errorf("changing %s bandwidth is not supported", winstack.CLOUD_PROVIDER_WINSTACK)

}

func (s *SWinStackProviderFactory) ValidateCreateCloudaccountData(ctx context.Context, userCred mcclient.TokenCredential, input cloudprovider.SCloudaccountCredential) (cloudprovider.SCloudaccount, error) {
	output := cloudprovider.SCloudaccount{}
	if len(input.Username) == 0 {
		return output, errors.Wrap(httperrors.ErrMissingParameter, "username")
	}
	if len(input.Password) == 0 {
		return output, errors.Wrap(httperrors.ErrMissingParameter, "password")
	}
	if len(input.AuthUrl) == 0 {
		return output, errors.Wrap(httperrors.ErrMissingParameter, "auth_url")
	}

	output.AccessUrl = input.AuthUrl
	output.Account = input.Username
	output.Secret = input.Password
	return output, nil
}

func (s *SWinStackProviderFactory) ValidateUpdateCloudaccountCredential(ctx context.Context, userCred mcclient.TokenCredential, input cloudprovider.SCloudaccountCredential, cloudaccount string) (cloudprovider.SCloudaccount, error) {
	output := cloudprovider.SCloudaccount{}
	if len(input.Username) == 0 {
		return output, errors.Wrap(httperrors.ErrMissingParameter, "username")
	}
	if len(input.Password) == 0 {
		return output, errors.Wrap(httperrors.ErrMissingParameter, "password")
	}
	output = cloudprovider.SCloudaccount{
		Account: input.Username,
		Secret:  input.Password,
	}
	return output, nil
}

func init() {
	factory := SWinStackProviderFactory{}
	cloudprovider.RegisterFactory(&factory)
}

type SWinStackProvider struct {
	cloudprovider.SBaseProvider
	client *winstack.SWinStackClient
}

func (s *SWinStackProvider) GetSysInfo() (jsonutils.JSONObject, error) {
	return jsonutils.NewDict(), nil
}

func (s *SWinStackProvider) GetVersion() string {
	return "2009-08-15"
}

func (s *SWinStackProvider) GetIRegions() []cloudprovider.ICloudRegion {
	return s.client.GetIRegions()
}

func (s *SWinStackProvider) GetIProjects() ([]cloudprovider.ICloudProject, error) {
	return []cloudprovider.ICloudProject{}, nil
}

func (s *SWinStackProvider) GetIRegionById(id string) (cloudprovider.ICloudRegion, error) {
	return s.client.GetIRegionById(id)
}

func (s *SWinStackProvider) GetBalance() (float64, string, error) {
	return 0.0, api.CLOUD_PROVIDER_HEALTH_NORMAL, cloudprovider.ErrNotSupported
}

func (s *SWinStackProvider) GetSubAccounts() ([]cloudprovider.SSubAccount, error) {
	return s.client.GetSubAccounts()
}

func (s *SWinStackProvider) GetAccountId() string {
	return s.client.GetAccountId()

}

func (s *SWinStackProvider) GetStorageClasses(regionId string) []string {
	return nil
}

func (s *SWinStackProvider) GetBucketCannedAcls(regionId string) []string {
	return nil
}

func (s *SWinStackProvider) GetObjectCannedAcls(regionId string) []string {
	return nil
}

func (s *SWinStackProvider) GetCapabilities() []string {
	return s.client.GetCapabilities()
}

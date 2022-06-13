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
	"testing"
	"yunion.io/x/onecloud/pkg/cloudprovider"
)

func Test_invoke(t *testing.T) {
	cfg := WinStackConfig{
		cpcfg:    cloudprovider.ProviderConfig{},
		endpoint: "https://10.252.226.12/",
		user:     "admin",
		password: "passw0rd",
		debug:    false,
	}
	client, err := NewWinStackClient(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.invokeGET("api/compute/domains", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	//t.Log(resp)
}

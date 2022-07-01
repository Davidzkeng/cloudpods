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
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"
	"yunion.io/x/jsonutils"
	"yunion.io/x/log"
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/util/httputils"
	"yunion.io/x/pkg/errors"
)

const (
	CLOUD_PROVIDER_WINSTACK = api.CLOUD_PROVIDER_WINSTACK

	CHECK_SESSION_URL = "/api/check/session"
	LOGIN_URL         = "/api/login"
)

var (
	UnauthorizedError = errors.Errorf("WinStack Unauthorized")
)

var client *SWinStackClient
var clientOnce sync.Once

type ListOps struct {
	Total int
	Start int
	Size  int
	Sort  string
	Order string
}

type WinStackConfig struct {
	cpcfg cloudprovider.ProviderConfig

	endpoint string
	user     string
	password string

	debug bool
}

func NewWinStackConfig(endpoint, user, password string) *WinStackConfig {
	return &WinStackConfig{
		endpoint: endpoint,
		user:     user,
		password: password,
	}
}

func (cfg *WinStackConfig) CloudproviderConfig(cpcfg cloudprovider.ProviderConfig) *WinStackConfig {
	cfg.cpcfg = cpcfg
	return cfg
}

func (cfg *WinStackConfig) Debug(debug bool) *WinStackConfig {
	cfg.debug = debug
	return cfg
}

type SWinStackClient struct {
	*WinStackConfig

	cli   *http.Client
	h     http.Header
	hLock sync.RWMutex

	regions []SRegion
}

func (client *SWinStackClient) SetHeader(key, value string) {
	client.hLock.Lock()
	client.h.Set(key, value)
	client.hLock.Unlock()
}

func (client *SWinStackClient) GetHeader() http.Header {
	client.hLock.RLock()
	h := client.h
	client.hLock.RUnlock()
	return h
}

func NewWinStackClient(cfg *WinStackConfig) (*SWinStackClient, error) {
	clientOnce.Do(func() {
		client = &SWinStackClient{WinStackConfig: cfg, h: http.Header{}}
		client.cli = client.getDefaultClient(0)
		var err error
		client.regions, err = client.GetRegions()
		if err != nil {
			return
		}
		for i := range client.regions {
			client.regions[i].client = client
		}
	})
	return client, nil
}

func (client *SWinStackClient) GetAccountId() string {
	return client.endpoint
}

func (client *SWinStackClient) GetRegion(id string) (*SRegion, error) {
	for i := range client.regions {
		if client.regions[i].Id == id {
			return &client.regions[i], nil
		}
	}
	if len(id) == 0 {
		return &client.regions[0], nil
	}
	return nil, cloudprovider.ErrNotFound
}

func (client *SWinStackClient) GetIRegions() []cloudprovider.ICloudRegion {
	ret := []cloudprovider.ICloudRegion{}
	for i := range client.regions {
		client.regions[i].client = client
		ret = append(ret, &client.regions[i])
	}
	return ret
}

func (client *SWinStackClient) GetIRegionById(id string) (cloudprovider.ICloudRegion, error) {
	iregions := client.GetIRegions()
	for i := range iregions {
		if iregions[i].GetGlobalId() == id {
			return iregions[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, id)
}

func (client *SWinStackClient) GetCapabilities() []string {
	return []string{
		cloudprovider.CLOUD_CAPABILITY_PROJECT,
		cloudprovider.CLOUD_CAPABILITY_COMPUTE,
		cloudprovider.CLOUD_CAPABILITY_NETWORK,
		cloudprovider.CLOUD_CAPABILITY_EIP,
		cloudprovider.CLOUD_CAPABILITY_LOADBALANCER,
		cloudprovider.CLOUD_CAPABILITY_OBJECTSTORE,
		cloudprovider.CLOUD_CAPABILITY_RDS,
		cloudprovider.CLOUD_CAPABILITY_CACHE,
		cloudprovider.CLOUD_CAPABILITY_EVENT,
		cloudprovider.CLOUD_CAPABILITY_CLOUDID,
		cloudprovider.CLOUD_CAPABILITY_DNSZONE,
		cloudprovider.CLOUD_CAPABILITY_INTERVPCNETWORK,
		cloudprovider.CLOUD_CAPABILITY_SAML_AUTH,
		cloudprovider.CLOUD_CAPABILITY_NAT,
		cloudprovider.CLOUD_CAPABILITY_NAS,
		cloudprovider.CLOUD_CAPABILITY_WAF,
	}
}

func (client *SWinStackClient) GetSubAccounts() ([]cloudprovider.SSubAccount, error) {
	subAccount := cloudprovider.SSubAccount{
		Account: client.user,
		Name:    client.cpcfg.Name,

		HealthStatus: api.CLOUD_PROVIDER_HEALTH_NORMAL,
	}
	return []cloudprovider.SSubAccount{subAccount}, nil
}

func (client *SWinStackClient) getDefaultClient(timeout time.Duration) *http.Client {
	httpClient := httputils.GetDefaultClient()
	if timeout > 0 {
		httpClient = httputils.GetTimeoutClient(timeout)
	}
	if client.cpcfg.ProxyFunc != nil {
		httputils.SetClientProxyFunc(httpClient, client.cpcfg.ProxyFunc)
	}
	return httpClient
}

func (client *SWinStackClient) checkSession() bool {
	_, err := client.invokeGET(CHECK_SESSION_URL, nil, nil)
	if err != nil {
		return false
	}
	return true
}

func (client *SWinStackClient) refreshSession() error {
	type LoginReq struct {
		User string `json:"user"`
		Pwd  string `json:"pwd"`
	}
	body := LoginReq{
		User: client.user,
		Pwd:  client.password,
	}
	resp, err := client.invokePOST(LOGIN_URL, nil, nil, body)
	if err != nil {
		return err
	}
	if resp.Contains("sessionId") {
		sessionId, _ := resp.GetString("sessionId")
		client.SetHeader("Cookie", "SESSION="+sessionId)
	}
	return nil
}

type sWinStackError struct {
	ErrorCode int `json:"errorCode"`
	Message   string
	Exception string
}

func (e sWinStackError) Error() string {
	return jsonutils.Marshal(e).String()
}

func (client *SWinStackClient) skipRefreshSession(path string) bool {
	for _, v := range []string{LOGIN_URL, CHECK_SESSION_URL} {
		if path == v {
			return true
		}
	}
	return false
}

func (client *SWinStackClient) invokeGET(path string, header map[string]string, query map[string]string) (jsonutils.JSONObject, error) {
	return client.invoke(httputils.GET, path, header, query, nil)
}

func (client *SWinStackClient) invokePOST(path string, header map[string]string, query map[string]string, body interface{}) (jsonutils.JSONObject, error) {
	return client.invoke(httputils.POST, path, header, query, body)
}

func (client *SWinStackClient) invokePUT(path string, header map[string]string, query map[string]string, body interface{}) (jsonutils.JSONObject, error) {
	return client.invoke(httputils.POST, path, header, query, body)
}

func (client *SWinStackClient) invokePATCH(path string, header map[string]string, query map[string]string) (jsonutils.JSONObject, error) {
	return client.invoke(httputils.PATCH, path, header, query, nil)
}

func (client *SWinStackClient) invokeDELETE(path string, header map[string]string, query map[string]string) (jsonutils.JSONObject, error) {
	return client.invoke(httputils.DELETE, path, header, query, nil)
}

func (client *SWinStackClient) invoke(method httputils.THttpMethod, path string, header map[string]string, query map[string]string, body interface{}) (jsonutils.JSONObject, error) {
	if !client.skipRefreshSession(path) && !client.checkSession() {
		log.Printf("path:%s,checkSession:%v", path, client.checkSession())
		err := client.refreshSession()
		if err != nil {
			return nil, err
		}
	}

	var encode = func(k, v string) string {
		d := url.Values{}
		d.Set(k, v)
		return d.Encode()
	}
	var q string
	for k, v := range query {
		q += "&" + encode(k, v)
	}

	for k, v := range header {
		client.SetHeader(k, v)
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	client.SetHeader("Content-Type", "application/json")
	URL := client.endpoint + path + "?" + q
	resp, err := httputils.Request(client.cli, context.Background(), method, URL, client.GetHeader(), bytes.NewReader(b), client.debug)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return nil, UnauthorizedError
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if len(data) <= 0 {
		return nil, nil
	}
	obj, err := jsonutils.Parse(data)
	if err != nil {
		return nil, errors.Wrapf(err, "jsonutils.Parse")
	}

	if client.debug {
		log.Errorf("response: %s", obj.PrettyString())
	}
	wErr := &sWinStackError{}
	obj.Unmarshal(&wErr)
	if len(wErr.Message) > 0 {
		return nil, wErr
	}
	return obj, nil
}

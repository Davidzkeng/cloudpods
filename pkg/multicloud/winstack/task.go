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

import "fmt"

const (
	TASK_DETAIL_URL = "/api/notify/tasks/%s"
)

type STask struct {
	Id              string      `json:"id"`
	Code            string      `json:"code"`
	Name            string      `json:"name"`
	Description     string      `json:"description"`
	TimeCreate      string      `json:"timeCreate"`
	TimeStart       string      `json:"timeStart"`
	TimeEnd         interface{} `json:"timeEnd"`
	Status          int         `json:"status"`
	CreateLoginName string      `json:"createLoginName"`
	CreateLoginId   string      `json:"createLoginId"`
	CreateLoginIp   string      `json:"createLoginIp"`
	CancelLoginName interface{} `json:"cancelLoginName"`
	CancelLoginIp   interface{} `json:"cancelLoginIp"`
	TargetName      string      `json:"targetName"`
	PoolId          string      `json:"poolId"`
	ClusterId       string      `json:"clusterId"`
	HostId          string      `json:"hostId"`
	DomainId        string      `json:"domainId"`
	BusinessGroupId interface{} `json:"businessGroupId"`
	Module          string      `json:"module"`
	StepCount       int         `json:"stepCount"`
	StepIndex       int         `json:"stepIndex"`
	StepDesc        string      `json:"stepDesc"`
}

func (s *SRegion) GetTask(id string) (*STask, error) {
	URL := fmt.Sprintf(TASK_DETAIL_URL, id)
	resp, err := s.client.invokeGET(URL, nil, nil)
	if err != nil {
		return nil, err
	}
	var ret STask
	return &ret, resp.Unmarshal(&ret)
}

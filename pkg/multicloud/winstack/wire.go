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
	"fmt"
	api "yunion.io/x/onecloud/pkg/apis/compute"
	"yunion.io/x/onecloud/pkg/cloudprovider"
	"yunion.io/x/onecloud/pkg/multicloud"
	"yunion.io/x/pkg/errors"
)

type SWire struct {
	multicloud.STagBase
	multicloud.SResourceBase

	cluster *SCluster
	vpc     *SVpc
}

func (s *SWire) GetIVpc() cloudprovider.ICloudVpc {
	return s.vpc
}

func (s *SWire) GetIZone() cloudprovider.ICloudZone {
	return s.cluster
}

func (s *SWire) GetBandwidth() int {
	return 1000
}

func (s *SWire) CreateINetwork(opts *cloudprovider.SNetworkCreateOptions) (cloudprovider.ICloudNetwork, error) {
	panic("implement me")
}

func (s *SWire) GetId() string {
	return fmt.Sprintf("%s/%s", s.vpc.GetGlobalId(), s.cluster.GetGlobalId())
}

func (s *SWire) GetName() string {
	return fmt.Sprintf("%s-%s", s.vpc.GetName(), s.cluster.GetName())
}

func (s *SWire) GetGlobalId() string {
	return fmt.Sprintf("%s/%s", s.vpc.GetGlobalId(), s.cluster.GetGlobalId())
}

func (s *SWire) GetStatus() string {
	return api.WIRE_STATUS_AVAILABLE
}

func (s *SVpc) GetIWireById(id string) (cloudprovider.ICloudWire, error) {
	wires, err := s.GetIWires()
	if err != nil {
		return nil, err
	}
	for i := range wires {
		if wires[i].GetGlobalId() == id {
			return wires[i], nil
		}
	}
	return nil, errors.Wrapf(cloudprovider.ErrNotFound, id)
}

func (s *SVpc) GetIWires() ([]cloudprovider.ICloudWire, error) {
	clusters, err := s.region.GetClusters()
	if err != nil {
		return nil, err
	}
	var ret []cloudprovider.ICloudWire
	for i := range clusters {
		wire := &SWire{
			vpc:     s,
			cluster: &clusters[i],
		}
		ret = append(ret, wire)
	}
	return ret, nil
}

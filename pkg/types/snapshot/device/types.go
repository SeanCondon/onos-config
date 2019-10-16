// Copyright 2019-present Open Networking Foundation.
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

package device

import (
	"fmt"
	"github.com/onosproject/onos-config/pkg/types"
	"github.com/onosproject/onos-topo/pkg/northbound/device"
	"strings"
)

const separator = ":"

// ID is a snapshot identifier type
type ID types.ID

// GetSnapshotID returns the snapshot ID for the given network snapshot ID and device
func GetSnapshotID(networkID types.ID, deviceID device.ID) ID {
	return ID(fmt.Sprintf("%s%s%s", networkID, separator, deviceID))
}

// GetDeviceID returns the device ID for the snapshot ID
func (i ID) GetDeviceID() device.ID {
	return device.ID(string(i)[strings.Index(string(i), separator)+1:])
}

// Revision is a device snapshot revision number
type Revision types.Revision
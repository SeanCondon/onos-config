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

package admin

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/onosproject/onos-config/pkg/northbound"
	"google.golang.org/grpc"
	"gotest.tools/assert"
)

// TestMain initializes the test suite context.
func TestMain(m *testing.M) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	northbound.SetUpServer(10124, Service{}, &waitGroup)
	waitGroup.Wait()
	os.Exit(m.Run())
}

func getAdminClient() (*grpc.ClientConn, ConfigAdminServiceClient) {
	conn := northbound.Connect(northbound.Address, northbound.Opts...)
	return conn, CreateConfigAdminServiceClient(conn)
}

func Test_RollbackNetworkChange_BadName(t *testing.T) {
	t.Skip()
	conn, client := getAdminClient()
	defer conn.Close()
	_, err := client.RollbackNewNetworkChange(context.Background(), &RollbackRequest{Name: "BAD CHANGE"})
	assert.ErrorContains(t, err, "Rollback aborted. Network change BAD CHANGE not found")
}

func Test_RollbackNetworkChange_NoChange(t *testing.T) {
	t.Skip()
	conn, client := getAdminClient()
	defer conn.Close()
	_, err := client.RollbackNewNetworkChange(context.Background(), &RollbackRequest{Name: ""})
	assert.ErrorContains(t, err, "is not")
}

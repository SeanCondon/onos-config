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

package cli

import (
	"context"
	"github.com/onosproject/onos-config/pkg/northbound/diags"
	"github.com/spf13/cobra"
	"io"
	"text/template"
)

const changeTemplate = "CHANGE: {{.Id}} ({{.Desc}})\n" +
	"\t{{printf \"|%-50s|%-40s|%-7s|\" \"PATH\" \"VALUE\" \"REMOVED\"}}\n" +
	"{{range .ChangeValues}}" +
	"\t{{wrappath .Path 50 1| printf \"|%-50s|\"}}{{valuetostring .Value | printf \"(%s) %s\" .Value.Type | printf \"%-40s|\" }}{{printf \"%-7t|\" .Removed}}\n" +
	"{{end}}\n"

var funcMapChanges = template.FuncMap{
	"wrappath":      wrapPath,
	"valuetostring": valueToSstring,
}

// Deprecated: For old style changes
func getGetChangesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "changes [<changeId>]",
		Short: "Lists records of configuration changes (deprecated)",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runChangesCommand,
	}
	return cmd
}

// Deprecated: For old style changes
func runChangesCommand(cmd *cobra.Command, args []string) error {
	clientConnection, clientConnectionError := getConnection()

	if clientConnectionError != nil {
		return clientConnectionError
	}
	client := diags.CreateConfigDiagsClient(clientConnection)
	changesReq := &diags.ChangesRequest{ChangeIDs: make([]string, 0)}
	if len(args) == 1 {
		changesReq.ChangeIDs = append(changesReq.ChangeIDs, args[0])
	}

	tmplChanges, _ := template.New("change").Funcs(funcMapChanges).Parse(changeTemplate)

	stream, err := client.GetChanges(context.Background(), changesReq)
	if err != nil {
		return err
	}

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		_ = tmplChanges.Execute(GetOutput(), in)
	}
}

// Copyright 2025 Victor Palma <victor.palma@rackspace.com>
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

package cmd

import "testing"

func TestClusterDescribeRemovesEnvironmentExportFlags(t *testing.T) {
	cmd := newClusterDescribeCmd()

	if cmd.Flags().Lookup("export-only") != nil {
		t.Fatal("cluster describe must not expose --export-only")
	}
	if cmd.Flags().Lookup("shell") != nil {
		t.Fatal("cluster describe must not expose --shell")
	}
	if cmd.Flags().Lookup("json") != nil {
		t.Fatal("cluster describe must use global --output instead of local --json")
	}
	if cmd.Flags().Lookup("validate") == nil {
		t.Fatal("cluster describe should keep --validate")
	}
}

func TestClusterUseRemovesExportOnlyAndKeepsSelectionFlags(t *testing.T) {
	cmd := newClusterUseCmd()

	if cmd.Flags().Lookup("export-only") != nil {
		t.Fatal("cluster use must not expose --export-only")
	}
	for _, name := range []string{"clear", "clear-persistent", "persistent", "shell"} {
		if cmd.Flags().Lookup(name) == nil {
			t.Fatalf("cluster use should keep --%s", name)
		}
	}
}

func TestClusterEnvOwnsShellExportFlag(t *testing.T) {
	cmd := newClusterEnvCmd()

	if cmd.Flags().Lookup("shell") == nil {
		t.Fatal("cluster env should own --shell for environment export")
	}
}

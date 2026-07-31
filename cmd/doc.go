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

// Package cmd defines the Cobra command tree and the user-facing orchestration
// layer for opencenter.
//
// ExecuteWithContext in root.go is the primary wiring boundary. It builds the
// typed application graph, places it in the command context, registers the root
// commands, and discovers external command plugins. Individual command files
// should remain responsible for argument handling, prompts, terminal output,
// and sequencing calls into lower-level packages.
//
// Reusable configuration, cluster lifecycle, provider, GitOps, secrets, and
// validation behavior belongs in a specifically named package under internal.
// Do not add domain logic to a Cobra Run or RunE function when it can be tested
// independently behind a focused service API. Conversely, keep command-only
// presentation policy here instead of leaking it into reusable packages.
package cmd

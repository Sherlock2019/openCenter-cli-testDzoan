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

// Package v2 owns the authoritative typed cluster configuration schema and its
// load, normalization, reference resolution, defaulting, validation, and save
// behavior.
//
// Configuration processing is ordered: load serialized input, normalize legacy
// shapes, resolve environment, file, and configuration references, apply
// defaults, validate the effective configuration, and then treat the result as
// the caller's stable configuration value. Changing this order can change
// compatibility and error reporting.
//
// Add schema fields, normalization rules, defaults, and validation rules in this
// package with tests that cover both serialized and effective values. Command
// prompts, terminal output, and application wiring do not belong here. Dynamic
// reference syntax is a compatibility surface and should not be renamed or
// removed without migration coverage.
package v2

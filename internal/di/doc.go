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

// Package di constructs the application's runtime dependency graph.
//
// App and NewApp are the canonical wiring path. App exposes explicitly typed
// dependencies, while provider constructors keep construction details out of
// command handlers. NewAppContainer presents that graph through the older
// Container interface for call sites that have not yet migrated.
//
// SetupContainer, NewContainer, and reflection-based registration are legacy
// compatibility mechanisms. New dependencies should be added to App and NewApp,
// using focused provider constructors where useful; do not add new named
// registrations solely to extend the legacy service locator.
//
// This package may import lower-level domain and infrastructure packages to wire
// them together. It must not contain command presentation or business rules, and
// lower-level packages must not import di.
package di

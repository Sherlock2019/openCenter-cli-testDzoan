/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package sops integrates opencenter with SOPS and age for repository secret
// encryption, decryption, key management, configuration generation, and Git
// hooks.
//
// This package owns process and filesystem behavior that is reusable without a
// terminal. Prompts, command output, and command-specific policy remain in cmd.
// Overlay file selection is centralized in overlayFilesToEncrypt; its exact
// order is observable through failures and tests, so callers must not reorder or
// broaden the list without explicit behavior coverage.
//
// Similar-looking key loaders and file replacement paths currently preserve
// different empty-input, whitespace, mode, durability, and rollback semantics.
// Keep those paths local until characterization tests establish a shared policy.
package sops

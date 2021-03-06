// Copyright 2019 The Openstack-Service-Lifecyle Authors
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

package services

import (
	"errors"
)

var (
	// ErrNotFound indicates the release was not found.
	ErrNotFound = errors.New("Resource not found")

	// OwnershipMismatch indicates that one of the subresources does
	// not have the right ownership.
	OwnershipMismatch = errors.New("Ownership Mismatch")

	// Error detected during SyncResource
	SyncError = errors.New("Sync Error")

	// Error detected during InstallResource
	InstallError = errors.New("Install Error")

	// Error detected during UninstallResource
	UninstallError = errors.New("Uninstall Error")

	// Error detected during UpdateResource
	UpdateError = errors.New("Update Error")

	// Error detected during ReconcileResource
	ReconcileError = errors.New("Reconcile Error")
)

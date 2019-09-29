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
	"context"
	av1 "github.com/keleustes/armada-crd/pkg/apis/openstacklcm/v1alpha1"
)

// Manager manages a Openstack Service . It can deploy, upgrade, backup, restore
type OslcManager interface {
	ResourceName() string
	IsInstalled() bool
	IsUpdateRequired() bool
	SyncResource(context.Context) error
	InstallResource(context.Context) (*av1.LifecycleFlow, error)
	UpdateResource(context.Context) (*av1.LifecycleFlow, *av1.LifecycleFlow, error)
	ReconcileResource(context.Context) (*av1.LifecycleFlow, error)
	UninstallResource(context.Context) (*av1.LifecycleFlow, error)
}

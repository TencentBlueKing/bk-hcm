/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"context"

	types "hcm/cmd/woa-server/types/config"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/tools/metadata"
)

type model struct {
	requirement    Requirement
	idcZone        IdcZone
	deviceRestrict DeviceRestrict
	cvmDevice      CvmDevice
	dvmDevice      DvmDevice
	pmDevice       PmDevice
}

// Requirement get requirement operation interface
func (m *model) Requirement() Requirement {
	return m.requirement
}

// IdcZone get idc zone operation interface
func (m *model) IdcZone() IdcZone {
	return m.idcZone
}

// DeviceRestrict get device restrict operation interface
func (m *model) DeviceRestrict() DeviceRestrict {
	return m.deviceRestrict
}

// CvmDevice get cvm device operation interface
func (m *model) CvmDevice() CvmDevice {
	return m.cvmDevice
}

// DvmDevice get dvm device operation interface
func (m *model) DvmDevice() DvmDevice {
	return m.dvmDevice
}

// PmDevice get physical machine device operation interface
func (m *model) PmDevice() PmDevice {
	return m.pmDevice
}

var operation *model

func init() {
	operation = &model{
		requirement:    &requirement{},
		idcZone:        &idcZone{},
		deviceRestrict: &deviceRestrict{},
		cvmDevice:      &cvmDevice{},
		dvmDevice:      &dvmDevice{},
		pmDevice:       &pmDevice{},
	}
}

// Operation return all model operation interface
func Operation() *model {
	return operation
}

// Model provides storage interface for operations of models
type Model interface {
	Requirement() Requirement
	IdcZone() IdcZone
	DeviceRestrict() DeviceRestrict
	CvmDevice() CvmDevice
	DvmDevice() DvmDevice
	PmDevice() PmDevice
}

// Requirement requirement operation interface
type Requirement interface {
	// NextSequence returns next resource requirement type config sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateRequirement creates resource requirement type config in db
	CreateRequirement(ctx context.Context, inst *types.Requirement) error
	// GetRequirement gets resource requirement type config by filter from db
	GetRequirement(ctx context.Context, filter *mapstr.MapStr) (*types.Requirement, error)
	// FindManyRequirement gets resource requirement type config list by filter from db
	FindManyRequirement(ctx context.Context, filter *mapstr.MapStr, sortFields ...string) ([]*types.Requirement, error)
	// UpdateRequirement updates resource requirement type config by filter and doc in db
	UpdateRequirement(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
	// DeleteRequirement deletes resource requirement type config from db
	DeleteRequirement(ctx context.Context, filter *mapstr.MapStr) error
}

// IdcZone idc zone operation interface
type IdcZone interface {
	// NextSequence returns next zone config sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateZone creates zone config in db
	CreateZone(ctx context.Context, inst *types.IdcZone) error
	// GetZone gets resource zone config by filter from db
	GetZone(ctx context.Context, filter *mapstr.MapStr) (*types.IdcZone, error)
	// FindManyZone gets zone config list by filter from db
	FindManyZone(ctx context.Context, filter *mapstr.MapStr) ([]*types.IdcZone, error)
	// GetRegionList gets region list by filter from db
	GetRegionList(ctx context.Context, filter map[string]interface{}) ([]interface{}, error)
	// UpdateZone updates zone config by filter and doc in db
	UpdateZone(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
	// DeleteZone deletes zone config from db
	DeleteZone(ctx context.Context, filter *mapstr.MapStr) error
}

// DeviceRestrict device restrict operation interface
type DeviceRestrict interface {
	// NextSequence returns next device restrict config sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateDeviceRestrict creates device restrict config in db
	CreateDeviceRestrict(ctx context.Context, inst *types.DeviceRestrict) error
	// GetDeviceRestrict gets resource device restrict config by filter from db
	GetDeviceRestrict(ctx context.Context, filter *mapstr.MapStr) (*types.DeviceRestrict, error)
	// FindManyDeviceRestrict gets device restrict list by filter from db
	FindManyDeviceRestrict(ctx context.Context, filter *mapstr.MapStr) ([]*types.DeviceRestrict, error)
	// UpdateDeviceRestrict updates device restrict config by filter and doc in db
	UpdateDeviceRestrict(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
	// DeleteDeviceRestrict deletes device restrict config from db
	DeleteDeviceRestrict(ctx context.Context, filter *mapstr.MapStr) error
}

// CvmDevice cvm device operation interface
type CvmDevice interface {
	// NextSequence returns next device config sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// NextSequences returns next device config sequence ids from db
	NextSequences(ctx context.Context, num int) ([]uint64, error)
	// CreateDevice creates device config in db
	CreateDevice(ctx context.Context, inst *types.DeviceInfo) error
	// BatchCreateDevices creates multiple device configs in db
	BatchCreateDevices(ctx context.Context, insts []*types.DeviceInfo) error
	// GetDevice gets device config by filter from db
	GetDevice(ctx context.Context, filter *mapstr.MapStr) (*types.DeviceInfo, error)
	// CountDevice gets resource device count by filter from db
	CountDevice(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyDevice gets device list by filter from db
	FindManyDevice(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) ([]*types.DeviceInfo,
		error)
	// FindManyDeviceType gets resource device type config list by filter from db
	FindManyDeviceType(ctx context.Context, filter map[string]interface{}) ([]interface{}, error)
	// UpdateDevice updates device config by filter and doc in db
	UpdateDevice(ctx context.Context, filter, doc map[string]interface{}) error
	// DeleteDevice deletes device config from db
	DeleteDevice(ctx context.Context, filter *mapstr.MapStr) error
}

// DvmDevice dvm device operation interface
type DvmDevice interface {
	// NextSequence returns next device config sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateDevice creates device config in db
	CreateDevice(ctx context.Context, inst *types.DvmDeviceInfo) error
	// GetDevice gets device config by filter from db
	GetDevice(ctx context.Context, filter *mapstr.MapStr) (*types.DvmDeviceInfo, error)
	// CountDevice gets resource device count by filter from db
	CountDevice(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyDevice gets device list by filter from db
	FindManyDevice(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) ([]*types.DvmDeviceInfo,
		error)
	// FindManyDeviceType gets resource device type config list by filter from db
	FindManyDeviceType(ctx context.Context, filter map[string]interface{}) ([]interface{}, error)
	// UpdateDevice updates device config by filter and doc in db
	UpdateDevice(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
	// DeleteDevice deletes device config from db
	DeleteDevice(ctx context.Context, filter *mapstr.MapStr) error
}

// PmDevice physical machine device operation interface
type PmDevice interface {
	// NextSequence returns next device config sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateDevice creates device config in db
	CreateDevice(ctx context.Context, inst *types.PmDeviceInfo) error
	// GetDevice gets device config by filter from db
	GetDevice(ctx context.Context, filter *mapstr.MapStr) (*types.PmDeviceInfo, error)
	// CountDevice gets resource device count by filter from db
	CountDevice(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyDevice gets device list by filter from db
	FindManyDevice(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) ([]*types.PmDeviceInfo,
		error)
	// FindManyDeviceType gets resource device type config list by filter from db
	FindManyDeviceType(ctx context.Context, filter map[string]interface{}) ([]interface{}, error)
	// UpdateDevice updates device config by filter and doc in db
	UpdateDevice(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
	// DeleteDevice deletes device config from db
	DeleteDevice(ctx context.Context, filter *mapstr.MapStr) error
}

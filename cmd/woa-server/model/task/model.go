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

package model

import (
	"context"

	daltypes "hcm/cmd/woa-server/storage/dal/types"
	cfgtypes "hcm/cmd/woa-server/types/config"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/tools/metadata"
)

// model all model operation interface
type model struct {
	applyTicket     ApplyTicket
	applyOrder      ApplyOrder
	applyStep       ApplyStep
	generateRecord  GenerateRecord
	initRecord      InitRecord
	diskCheckRecord DiskCheckRecord
	deliverRecord   DeliverRecord
	deviceInfo      DeviceInfo
	zone            Zone
}

// ApplyTicket get apply ticket operation interface
func (m *model) ApplyTicket() ApplyTicket {
	return m.applyTicket
}

// ApplyOrder get apply order operation interface
func (m *model) ApplyOrder() ApplyOrder {
	return m.applyOrder
}

// ApplyStep get apply step operation interface
func (m *model) ApplyStep() ApplyStep {
	return m.applyStep
}

// GenerateRecord get apply generate record operation interface
func (m *model) GenerateRecord() GenerateRecord {
	return m.generateRecord
}

// InitRecord get apply init record operation interface
func (m *model) InitRecord() InitRecord {
	return m.initRecord
}

// DiskCheckRecord get apply disk check record operation interface
func (m *model) DiskCheckRecord() DiskCheckRecord {
	return m.diskCheckRecord
}

// DeliverRecord get apply deliver record operation interface
func (m *model) DeliverRecord() DeliverRecord {
	return m.deliverRecord
}

// DeviceInfo get device info operation interface
func (m *model) DeviceInfo() DeviceInfo {
	return m.deviceInfo
}

// Zone get zone operation interface
func (m *model) Zone() Zone {
	return m.zone
}

var operation *model

func init() {
	operation = &model{
		applyTicket:     &applyTicket{},
		applyOrder:      &applyOrder{},
		applyStep:       &applyStep{},
		generateRecord:  &generateRecord{},
		initRecord:      &initRecord{},
		diskCheckRecord: &diskCheckRecord{},
		deliverRecord:   &deliverRecord{},
		deviceInfo:      &deviceInfo{},
		zone:            &zone{},
	}
}

// Operation return all model operation interface
func Operation() *model {
	return operation
}

// Model provides storage interface for operations of models
type Model interface {
	ApplyTicket() ApplyTicket
	ApplyOrder() ApplyOrder
	ApplyStep() ApplyStep
	GenerateRecord() GenerateRecord
	InitRecord() InitRecord
	DiskCheckRecord() DiskCheckRecord
	DeliverRecord() DeliverRecord
	DeviceInfo() DeviceInfo
	Zone() Zone
}

// ApplyTicket apply ticket operation interface
type ApplyTicket interface {
	// NextSequence returns next apply ticket sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateApplyTicket creates apply ticket in db
	CreateApplyTicket(ctx context.Context, inst *types.ApplyTicket) error
	// GetApplyTicket gets apply ticket by filter from db
	GetApplyTicket(ctx context.Context, filter *mapstr.MapStr) (*types.ApplyTicket, error)
	// CountApplyTicket gets apply ticket count by filter from db
	CountApplyTicket(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyApplyTicket gets apply ticket list by filter from db
	FindManyApplyTicket(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) (
		[]*types.ApplyTicket, error)
	// UpdateApplyTicket updates apply ticket by filter and doc in db
	UpdateApplyTicket(ctx context.Context, filter *mapstr.MapStr, doc interface{}) error
	// DeleteApplyTicket deletes apply ticket from db
	DeleteApplyTicket()
	// AggregateAll apply ticket aggregate all operation
	AggregateAll(ctx context.Context, pipeline interface{}, result interface{}, opts ...*daltypes.AggregateOpts) error
}

// ApplyOrder apply order operation interface
type ApplyOrder interface {
	// NextSequence returns next apply order sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateApplyOrder creates apply order in db
	CreateApplyOrder(ctx context.Context, inst *types.ApplyOrder) error
	// GetApplyOrder gets apply order by filter from db
	GetApplyOrder(ctx context.Context, filter *mapstr.MapStr) (*types.ApplyOrder, error)
	// CountApplyOrder gets apply order count by filter from db
	CountApplyOrder(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyApplyOrder gets apply order list by filter from db
	FindManyApplyOrder(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) ([]*types.ApplyOrder,
		error)
	// UpdateApplyOrder updates apply order by filter and doc in db
	UpdateApplyOrder(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
	// DeleteApplyOrder deletes apply order from db
	DeleteApplyOrder()
	// AggregateAll apply order aggregate all operation
	AggregateAll(ctx context.Context, pipeline interface{}, result interface{}, opts ...*daltypes.AggregateOpts) error
}

// ApplyStep apply step operation interface
type ApplyStep interface {
	// NextSequence returns next apply order sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateApplyStep creates apply step in db
	CreateApplyStep(ctx context.Context, inst *types.ApplyStep) error
	// GetApplyStep gets apply order by filter from db
	GetApplyStep(ctx context.Context, filter *mapstr.MapStr) (*types.ApplyStep, error)
	// CountApplyStep gets apply step count by filter from db
	CountApplyStep(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyApplyStep gets apply order list by filter from db
	FindManyApplyStep(ctx context.Context, filter *mapstr.MapStr) ([]*types.ApplyStep, error)
	// UpdateApplyStep updates apply order by filter and doc in db
	UpdateApplyStep(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
	// DeleteApplyStep deletes apply step from db
	DeleteApplyStep()
}

// GenerateRecord apply generate record operation interface
type GenerateRecord interface {
	// NextSequence returns next apply order generate record sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateGenerateRecord creates apply order generate record in db
	CreateGenerateRecord(ctx context.Context, inst *types.GenerateRecord) error
	// GetGenerateRecord gets apply order generate record by filter from db
	GetGenerateRecord(ctx context.Context, filter *mapstr.MapStr) (*types.GenerateRecord, error)
	// CountGenerateRecord gets apply order generate record count by filter from db
	CountGenerateRecord(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyGenerateRecord gets generate record list by filter from db
	FindManyGenerateRecord(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) (
		[]*types.GenerateRecord, error)
	// UpdateGenerateRecord updates apply order generate record by filter and doc in db
	UpdateGenerateRecord(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
	// DeleteGenerateRecord deletes apply order generate record from db
	DeleteGenerateRecord()
	// AggregateAll generate record aggregate all operation
	AggregateAll(ctx context.Context, pipeline interface{}, result interface{}, opts ...*daltypes.AggregateOpts) error
}

// InitRecord apply init record operation interface
type InitRecord interface {
	// NextSequence returns next apply order init record sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateInitRecord creates apply order init record in db
	CreateInitRecord(ctx context.Context, inst *types.InitRecord) error
	// GetInitRecord gets apply order init record by filter from db
	GetInitRecord(ctx context.Context, filter *mapstr.MapStr) (*types.InitRecord, error)
	// CountInitRecord gets apply order init record count by filter from db
	CountInitRecord(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyInitRecord gets init record list by filter from db
	FindManyInitRecord(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) ([]*types.InitRecord,
		error)
	// UpdateInitRecord updates apply order init record by filter and doc in db
	UpdateInitRecord(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
	// DeleteInitRecord deletes apply order init record from db
	DeleteInitRecord()
}

// DiskCheckRecord apply disk check record operation interface
type DiskCheckRecord interface {
	// NextSequence returns next apply order disk check record sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateDiskCheckRecord creates apply order disk check record in db
	CreateDiskCheckRecord(ctx context.Context, inst *types.DiskCheckRecord) error
	// GetDiskCheckRecord gets apply order disk check record by filter from db
	GetDiskCheckRecord(ctx context.Context, filter *mapstr.MapStr) (*types.DiskCheckRecord, error)
	// CountDiskCheckRecord gets apply order disk check record count by filter from db
	CountDiskCheckRecord(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyDiskCheckRecord gets disk check record list by filter from db
	FindManyDiskCheckRecord(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) (
		[]*types.DiskCheckRecord, error)
	// UpdateDiskCheckRecord updates apply order disk check record by filter and doc in db
	UpdateDiskCheckRecord(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
	// DeleteDiskCheckRecord deletes apply order disk check record from db
	DeleteDiskCheckRecord()
}

// DeliverRecord apply deliver record operation interface
type DeliverRecord interface {
	// NextSequence returns next apply order deliver record sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateDeliverRecord creates apply order deliver record in db
	CreateDeliverRecord(ctx context.Context, inst *types.DeliverRecord) error
	// GetDeliverRecord gets apply order deliver record by filter from db
	GetDeliverRecord(ctx context.Context, filter *mapstr.MapStr) (*types.DeliverRecord, error)
	// CountDeliverRecord gets apply order deliver record count by filter from db
	CountDeliverRecord(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyDeliverRecord gets deliver record list by filter from db
	FindManyDeliverRecord(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) (
		[]*types.DeliverRecord, error)
	// UpdateDeliverRecord updates apply order deliver record by filter and doc in db
	UpdateDeliverRecord(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
	// DeleteDeliverRecord deletes apply order deliver record from db
	DeleteDeliverRecord()
}

// DeviceInfo device info operation interface
type DeviceInfo interface {
	// CreateDeviceInfo creates device info in db
	CreateDeviceInfo(ctx context.Context, inst *types.DeviceInfo) error
	// CreateDeviceInfos creates devices info in db
	CreateDeviceInfos(ctx context.Context, inst []*types.DeviceInfo) error
	// GetDeviceInfo gets device info by filter from db
	GetDeviceInfo(ctx context.Context, filter *mapstr.MapStr) ([]*types.DeviceInfo, error)
	// CountDeviceInfo gets device info count by filter from db
	CountDeviceInfo(ctx context.Context, filter map[string]interface{}) (uint64, error)
	// FindManyDeviceInfo gets device info list by filter from db
	FindManyDeviceInfo(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) ([]*types.DeviceInfo,
		error)
	// UpdateDeviceInfo updates device info by filter and doc in db
	UpdateDeviceInfo(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
	// DeleteDeviceInfo deletes device info from db
	DeleteDeviceInfo()
	// AggregateAll device info aggregate all operation
	AggregateAll(ctx context.Context, pipeline interface{}, result interface{}, opts ...*daltypes.AggregateOpts) error
	// Distinct gets device info distinct result from db
	Distinct(ctx context.Context, field string, filter map[string]interface{}) ([]interface{}, error)
}

// Zone zone operation interface
type Zone interface {
	// NextSequence returns next zone config sequence id from db
	NextSequence(ctx context.Context) (uint64, error)
	// CreateZone creates zone config in db
	CreateZone(ctx context.Context, inst *cfgtypes.Zone) error
	// GetZone gets resource zone config by filter from db
	GetZone(ctx context.Context, filter *mapstr.MapStr) (*cfgtypes.Zone, error)
	// FindManyZone gets zone config list by filter from db
	FindManyZone(ctx context.Context, filter *mapstr.MapStr) ([]*cfgtypes.Zone, error)
	// UpdateZone updates zone config by filter and doc in db
	UpdateZone(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error
	// DeleteZone deletes zone config from db
	DeleteZone(ctx context.Context, filter *mapstr.MapStr) error
}

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

// Package model implements model layer
package model

import (
	"context"

	daltypes "hcm/cmd/woa-server/storage/dal/types"
	"hcm/cmd/woa-server/storage/driver/mongodb"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/tools/metadata"
)

type deviceInfo struct {
}

// CreateDeviceInfo create device info in db
func (d *deviceInfo) CreateDeviceInfo(ctx context.Context, inst *types.DeviceInfo) error {
	return mongodb.Client().Table(pkg.BKTableNameDeviceInfo).Insert(ctx, inst)
}

// CreateDeviceInfos create device infos in db
func (d *deviceInfo) CreateDeviceInfos(ctx context.Context, inst []*types.DeviceInfo) error {
	return mongodb.Client().Table(pkg.BKTableNameDeviceInfo).Insert(ctx, inst)
}

// GetDeviceInfo gets device info by filter from db
func (d *deviceInfo) GetDeviceInfo(ctx context.Context, filter *mapstr.MapStr) ([]*types.DeviceInfo, error) {
	insts := make([]*types.DeviceInfo, 0)

	if err := mongodb.Client().Table(pkg.BKTableNameDeviceInfo).Find(filter).All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// CountDeviceInfo gets apply order device info count by filter from db
func (d *deviceInfo) CountDeviceInfo(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(pkg.BKTableNameDeviceInfo).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyDeviceInfo gets device info list by filter from db
func (d *deviceInfo) FindManyDeviceInfo(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) (
	[]*types.DeviceInfo, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(pkg.BKTableNameDeviceInfo).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("create_at:-1")
	}

	insts := make([]*types.DeviceInfo, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateDeviceInfo updates device info by filter and doc in db
func (d *deviceInfo) UpdateDeviceInfo(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	return mongodb.Client().Table(pkg.BKTableNameDeviceInfo).Update(ctx, filter, doc)
}

// DeleteDeviceInfo deletes device info from db
func (d *deviceInfo) DeleteDeviceInfo() {
	// TODO
}

// AggregateAll device info aggregate all operation
func (d *deviceInfo) AggregateAll(ctx context.Context, pipeline interface{}, result interface{},
	opts ...*daltypes.AggregateOpts) error {

	if err := mongodb.Client().Table(pkg.BKTableNameDeviceInfo).AggregateAll(ctx, pipeline, result,
		opts...); err != nil {
		return err
	}

	return nil
}

// Distinct gets device info distinct result from db
func (d *deviceInfo) Distinct(ctx context.Context, field string, filter map[string]interface{}) (
	[]interface{}, error) {
	insts, err := mongodb.Client().Table(pkg.BKTableNameDeviceInfo).Distinct(ctx, field, filter)
	if err != nil {
		return nil, err
	}

	return insts, nil
}

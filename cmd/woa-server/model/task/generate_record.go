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

// Package model implements all db operations of apply order generate record
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

type generateRecord struct {
}

// NextSequence returns next apply order generate record sequence id from db
func (g *generateRecord) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, pkg.BKTableNameGenerateRecord)
}

// CreateGenerateRecord creates apply order generate record in db
func (g *generateRecord) CreateGenerateRecord(ctx context.Context, inst *types.GenerateRecord) error {
	return mongodb.Client().Table(pkg.BKTableNameGenerateRecord).Insert(ctx, inst)
}

// GetGenerateRecord gets apply order generate record by filter from db
func (g *generateRecord) GetGenerateRecord(ctx context.Context, filter *mapstr.MapStr) (*types.GenerateRecord, error) {
	inst := new(types.GenerateRecord)

	if err := mongodb.Client().Table(pkg.BKTableNameGenerateRecord).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountGenerateRecord gets apply order generate record count by filter from db
func (g *generateRecord) CountGenerateRecord(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(pkg.BKTableNameGenerateRecord).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyGenerateRecord gets generate record list by filter from db
func (g *generateRecord) FindManyGenerateRecord(ctx context.Context, page metadata.BasePage,
	filter map[string]interface{}) ([]*types.GenerateRecord, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(pkg.BKTableNameGenerateRecord).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("generate_id")
	}

	insts := make([]*types.GenerateRecord, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateGenerateRecord updates apply order generate record by filter and doc in db
func (g *generateRecord) UpdateGenerateRecord(ctx context.Context, filter *mapstr.MapStr, doc *mapstr.MapStr) error {
	return mongodb.Client().Table(pkg.BKTableNameGenerateRecord).Update(ctx, filter, doc)
}

// DeleteGenerateRecord deletes apply order generate record from db
func (g *generateRecord) DeleteGenerateRecord() {

}

// AggregateAll generate record aggregate all operation
func (g *generateRecord) AggregateAll(ctx context.Context, pipeline interface{}, result interface{},
	opts ...*daltypes.AggregateOpts) error {

	if err := mongodb.Client().Table(pkg.BKTableNameGenerateRecord).AggregateAll(ctx, pipeline, result,
		opts...); err != nil {
		return err
	}

	return nil
}

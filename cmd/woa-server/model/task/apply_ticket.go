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

// Package model implements the object model for the service.
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

type applyTicket struct {
}

// NextSequence returns next apply ticket sequence id from db
func (a *applyTicket) NextSequence(ctx context.Context) (uint64, error) {
	return mongodb.Client().NextSequence(ctx, pkg.BKTableNameApplyTicket)
}

// CreateApplyTicket creates apply ticket in db
func (a *applyTicket) CreateApplyTicket(ctx context.Context, inst *types.ApplyTicket) error {
	return mongodb.Client().Table(pkg.BKTableNameApplyTicket).Insert(ctx, inst)
}

// GetApplyTicket gets apply ticket by filter from db
func (a *applyTicket) GetApplyTicket(ctx context.Context, filter *mapstr.MapStr) (*types.ApplyTicket, error) {
	inst := new(types.ApplyTicket)

	if err := mongodb.Client().Table(pkg.BKTableNameApplyTicket).Find(filter).One(ctx, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// CountApplyTicket gets apply ticket count by filter from db
func (a *applyTicket) CountApplyTicket(ctx context.Context, filter map[string]interface{}) (uint64, error) {
	total, err := mongodb.Client().Table(pkg.BKTableNameApplyTicket).Find(filter).Count(ctx)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// FindManyApplyTicket gets apply ticket list by filter from db
func (a *applyTicket) FindManyApplyTicket(ctx context.Context, page metadata.BasePage, filter map[string]interface{}) (
	[]*types.ApplyTicket, error) {

	limit := uint64(page.Limit)
	start := uint64(page.Start)
	query := mongodb.Client().Table(pkg.BKTableNameApplyTicket).Find(filter).Limit(limit).Start(start)
	if len(page.Sort) > 0 {
		query = query.Sort(page.Sort)
	} else {
		query = query.Sort("order_id")
	}

	insts := make([]*types.ApplyTicket, 0)
	if err := query.All(ctx, &insts); err != nil {
		return nil, err
	}

	return insts, nil
}

// UpdateApplyTicket updates apply ticket by filter and doc in db
func (a *applyTicket) UpdateApplyTicket(ctx context.Context, filter *mapstr.MapStr, doc interface{}) error {
	return mongodb.Client().Table(pkg.BKTableNameApplyTicket).Update(ctx, filter, doc)
}

// DeleteApplyTicket deletes apply ticket from db
func (a *applyTicket) DeleteApplyTicket() {
	// TODO
}

// AggregateAll apply ticket aggregate all operation
func (a *applyTicket) AggregateAll(ctx context.Context, pipeline interface{}, result interface{},
	opts ...*daltypes.AggregateOpts) error {

	if err := mongodb.Client().Table(pkg.BKTableNameApplyTicket).AggregateAll(ctx, pipeline, result, opts...); err != nil {
		return err
	}

	return nil
}

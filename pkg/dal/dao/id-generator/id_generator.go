/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package idgenerator

import (
	"database/sql"
	"fmt"
	"strconv"

	"hcm/pkg/dal/table"
	"hcm/pkg/kit"

	"github.com/jmoiron/sqlx"
)

// DefaultMaxRetryCount is default Max retry count when the generated unique id conflicts.
const DefaultMaxRetryCount = 3

// IDGenInterface supplies all the method to generate a resource's unique identity id.
type IDGenInterface interface {
	// Batch return a list of resource's unique id as required.
	Batch(kt *kit.Kit, resource table.Name, count int) ([]string, error)
	// One return one unique id for this resource.
	One(kt *kit.Kit, resource table.Name) (string, error)
}

var _ IDGenInterface = new(idGenerator)

// New create an id generator instance.
func New(db *sqlx.DB, retryCount int) IDGenInterface {
	return &idGenerator{db: db, maxRetryCount: retryCount}
}

type idGenerator struct {
	db *sqlx.DB
	// maxRetryCount is max retry count when the generated unique id conflicts.
	maxRetryCount int
}

// One generate one unique resource id.
func (ig idGenerator) One(kt *kit.Kit, resource table.Name) (string, error) {
	list, err := ig.Batch(kt, resource, 1)
	if err != nil {
		return "", err
	}

	if num := len(list); num != 1 {
		return "", fmt.Errorf("gen resource unique id, but %d returned", num)
	}

	return list[0], nil
}

// Batch is to generate distribute unique resource id list.
// returned with a number of unique ids as required.
func (ig idGenerator) Batch(kt *kit.Kit, resource table.Name, count int) ([]string, error) {
	if err := resource.Validate(); err != nil {
		return nil, err
	}

	txn, err := ig.db.BeginTx(kt.Ctx, new(sql.TxOptions))
	if err != nil {
		return nil, fmt.Errorf("gen %s unique id, but begin txn failed, err: %v", resource, err)
	}

	// get current max id
	queryExpr := fmt.Sprintf(`SELECT max_id from id_generator WHERE resource = "%s" FOR UPDATE`, resource)

	rows, err := txn.QueryContext(kt.Ctx, queryExpr)
	if err != nil {
		return nil, fmt.Errorf("gen %s unique id, but query max id failed, err: %v", resource, err)
	}

	var maxIDStr string
	for rows.Next() {
		if err := rows.Scan(&maxIDStr); err != nil {
			return nil, fmt.Errorf("gen %s unique id, but scan max id failed, err: %v", resource, err)
		}
		break
	}

	err = rows.Close()
	if err != nil {
		return nil, fmt.Errorf("gen %s unique id, but close rows failed, err: %v", resource, err)
	}

	// generate new max id and update it
	maxID, err := strconv.ParseUint(maxIDStr, 36, 64)
	if err != nil {
		return nil, fmt.Errorf("gen %s unique id, but parse max id failed, err: %v", resource, err)
	}

	newMaxID := strconv.FormatUint(maxID+uint64(count), 36)
	newMaxID = fmt.Sprintf("%08s", newMaxID)

	updateExpr := fmt.Sprintf(`UPDATE id_generator SET max_id = "%s" WHERE resource = "%s" AND max_id = "%s"`,
		newMaxID, resource, maxIDStr)

	result, err := txn.ExecContext(kt.Ctx, updateExpr)
	if err != nil {
		return nil, fmt.Errorf("gen %s unique id, but update max_id failed, err: %v", resource, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("gen %s unique id, but get rows affected failed, err: %v", resource, err)
	}

	if rowsAffected != 1 {
		return nil, fmt.Errorf("gen %s unique id, but rows affected %d is not 1", resource, rowsAffected)
	}

	if err := txn.Commit(); err != nil {
		return nil, fmt.Errorf("gen %s unique id, but commit failed, err: %v", resource, err)
	}

	// generate the id list that can be used.
	ids := make([]string, count)
	for idx := 0; idx < count; idx++ {
		id := maxID + uint64(idx+1)
		ids[idx] = fmt.Sprintf("%08s", strconv.FormatUint(id, 36))
	}

	return ids, nil
}

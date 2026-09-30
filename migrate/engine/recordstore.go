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

package engine

import (
	"fmt"
	"time"

	"hcm/migrate/register"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/migrate"
)

// Record is one row of the migration record table.
type Record struct {
	ID          uint64 `db:"id"`
	MigrationID string `db:"migration_id"`
	Version     string `db:"version"`
	// AppliedPkg is the Migration.Pkg of the last run of this ID.
	AppliedPkg string                 `db:"applied_pkg"`
	Status     enumor.MigrationStatus `db:"status"`
	Message    string                 `db:"message"`
	CreatedAt  time.Time              `db:"created_at"`
	UpdatedAt  time.Time              `db:"updated_at"`
}

// Records holds the migration records of one database, keyed by migration ID.
type Records map[string]Record

// RecordStore reads and writes the migration record table of one database.
// Writes go through the bare orm outside any migration transaction.
type RecordStore struct {
	o orm.Interface
}

// NewRecordStore returns the record store of the database behind o.
func NewRecordStore(o orm.Interface) *RecordStore {
	return &RecordStore{o: o}
}

// Load reads every record of the database. Returns ErrPrecondition for an
// unknown status, an unparsable version, or a success record without applied_pkg.
func (s *RecordStore) Load(kt *kit.Kit) (Records, error) {
	expr := fmt.Sprintf("SELECT `id`, `migration_id`, `version`, `applied_pkg`, `status`, `message`, "+
		"`created_at`, `updated_at` FROM `%s`", constant.MigrationRecordTable)
	rows := make([]Record, 0)
	if err := s.o.Do().Select(kt.Ctx, &rows, expr, map[string]interface{}{}); err != nil {
		logs.Errorf("list migration records failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list migration records failed, err: %v", err)
	}

	records := make(Records, len(rows))
	for _, r := range rows {
		if err := r.Status.Validate(); err != nil {
			return nil, fmt.Errorf("%w: migration record %s has unknown status %q",
				ErrPrecondition, r.MigrationID, r.Status)
		}
		if !register.IsPending(r.Version) {
			if _, err := register.Parse(r.Version); err != nil {
				return nil, fmt.Errorf("%w: migration record %s has unparsable version %q, err: %v",
					ErrPrecondition, r.MigrationID, r.Version, err)
			}
		}

		if r.Status == enumor.MigrationStatusSuccess && r.AppliedPkg == "" {
			return nil, fmt.Errorf("%w: success migration record %s has empty applied_pkg",
				ErrPrecondition, r.MigrationID)
		}
		records[r.MigrationID] = r
	}
	return records, nil
}

// MarkRunning records that m is about to run. Inserts a new row or sets an
// existing row to running with message cleared. Records m.Pkg as applied_pkg.
func (s *RecordStore) MarkRunning(kt *kit.Kit, m register.Migration) error {
	expr := fmt.Sprintf("INSERT INTO `%s` (`migration_id`, `version`, `applied_pkg`, `status`, `message`) "+
		"VALUES (:migration_id, :version, :applied_pkg, :status, '') "+
		"ON DUPLICATE KEY UPDATE `applied_pkg` = :applied_pkg, `status` = :status, `message` = '', "+
		"`updated_at` = NOW()", constant.MigrationRecordTable)
	arg := map[string]interface{}{
		"migration_id": m.ID,
		"version":      m.Version,
		"applied_pkg":  m.Pkg,
		"status":       enumor.MigrationStatusRunning,
	}
	if err := s.o.Do().Insert(kt.Ctx, expr, arg); err != nil {
		logs.Errorf("mark migration running failed, err: %v, id: %s, rid: %s", err, m.ID, kt.Rid)
		return fmt.Errorf("mark migration %s running failed, err: %v", m.ID, err)
	}
	return nil
}

// MarkSuccess records that m succeeded and refreshes the recorded version to
// the version m is registered with now.
func (s *RecordStore) MarkSuccess(kt *kit.Kit, m register.Migration) error {
	arg := map[string]interface{}{
		"migration_id": m.ID,
		"version":      m.Version,
		"status":       enumor.MigrationStatusSuccess,
	}
	expr := fmt.Sprintf("UPDATE `%s` SET `status` = :status, `message` = '', `version` = :version "+
		"WHERE `migration_id` = :migration_id", constant.MigrationRecordTable)
	return s.update(kt, m, expr, arg)
}

// MarkFailed records that m failed with runErr. The stored message is cut to
// constant.MigrationRecordMessageMaxBytes; the full error is logged here.
func (s *RecordStore) MarkFailed(kt *kit.Kit, m register.Migration, runErr error) error {
	msg := ""
	if runErr != nil {
		msg = runErr.Error()
	}
	logs.Errorf("migration failed, err: %s, id: %s, version: %s, rid: %s", msg, m.ID, m.Version, kt.Rid)

	arg := map[string]interface{}{
		"migration_id": m.ID,
		"status":       enumor.MigrationStatusFailed,
		"message":      migrate.TruncateUTF8(msg, constant.MigrationRecordMessageMaxBytes),
	}
	expr := fmt.Sprintf("UPDATE `%s` SET `status` = :status, `message` = :message "+
		"WHERE `migration_id` = :migration_id", constant.MigrationRecordTable)
	return s.update(kt, m, expr, arg)
}

// BackfillPendingVersion replaces the PENDING version of m's success record
// with the version m is registered with now, once m is released. Only the
// version changes; applied_pkg is kept. The condition is in the statement,
// so zero affected rows is normal and a repeated call changes nothing.
func (s *RecordStore) BackfillPendingVersion(kt *kit.Kit, m register.Migration) error {
	if register.IsPending(m.Version) {
		return fmt.Errorf("backfill migration %s failed, registered version is %s", m.ID, m.Version)
	}

	expr := fmt.Sprintf("UPDATE `%s` SET `version` = :version "+
		"WHERE `migration_id` = :migration_id AND `status` = :status AND `version` = :pending",
		constant.MigrationRecordTable)
	arg := map[string]interface{}{
		"migration_id": m.ID,
		"version":      m.Version,
		"status":       enumor.MigrationStatusSuccess,
		"pending":      constant.MigrationPendingVersion,
	}
	affected, err := s.o.Do().Update(kt.Ctx, expr, arg)
	if err != nil {
		logs.Errorf("backfill pending migration version failed, err: %v, id: %s, version: %s, rid: %s",
			err, m.ID, m.Version, kt.Rid)
		return fmt.Errorf("backfill migration %s version %s failed, err: %v", m.ID, m.Version, err)
	}
	if affected > 0 {
		logs.Infof("backfill pending migration version success, id: %s, version: %s, rid: %s",
			m.ID, m.Version, kt.Rid)
	}
	return nil
}

// update runs a status update of m's row. Zero affected rows means the row
// is missing.
func (s *RecordStore) update(kt *kit.Kit, m register.Migration, expr string, arg map[string]interface{}) error {
	affected, err := s.o.Do().Update(kt.Ctx, expr, arg)
	if err != nil {
		logs.Errorf("mark migration failed, err: %v, status: %s, id: %s, rid: %s", err, arg["status"], m.ID, kt.Rid)
		return fmt.Errorf("mark migration %s %s failed, err: %v", m.ID, arg["status"], err)
	}
	if affected == 0 {
		logs.Errorf("mark migration affected no record, status: %s, id: %s, rid: %s", arg["status"], m.ID, kt.Rid)
		return fmt.Errorf("mark migration %s %s failed, record not found", m.ID, arg["status"])
	}
	return nil
}

// CurrentVersion returns the highest version among the success records,
// compared with register.Compare. PENDING records ran but belong to no
// release, so they are skipped. ok is false when no success record is left.
// An unparsable version returns ErrPrecondition.
func CurrentVersion(records Records) (current register.Version, ok bool, err error) {
	for _, r := range records {
		if r.Status != enumor.MigrationStatusSuccess || register.IsPending(r.Version) {
			continue
		}
		v, err := register.Parse(r.Version)
		if err != nil {
			return register.Version{}, false, fmt.Errorf("%w: migration record %s has unparsable version %q, err: %v",
				ErrPrecondition, r.MigrationID, r.Version, err)
		}
		// Equal versions keep the smaller Raw.
		if !ok || register.Compare(v, current) > 0 ||
			(register.Compare(v, current) == 0 && v.Raw < current.Raw) {
			current, ok = v, true
		}
	}
	return current, ok, nil
}

// WarnVersionDrift logs a WARN and returns true when rec is a success record
// whose version differs from m's registered version. A PENDING record is
// backfilled instead and is not drift. It never fails the run.
func WarnVersionDrift(kt *kit.Kit, rec Record, m register.Migration) bool {
	if rec.Status != enumor.MigrationStatusSuccess || rec.Version == m.Version || register.IsPending(rec.Version) {
		return false
	}
	logs.Warnf("migration version drift, recorded: %s, registered: %s, id: %s, skip as success, rid: %s",
		rec.Version, m.Version, m.ID, kt.Rid)
	return true
}

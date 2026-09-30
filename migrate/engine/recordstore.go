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
	"strings"
	"time"
	"unicode/utf8"

	"hcm/migrate/register"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// Status is the execution status of a migration record.
type Status string

const (
	// StatusRunning is written right before a migration runs.
	StatusRunning Status = "running"
	// StatusSuccess is written after a migration succeeds, or by init --mode=adopt.
	StatusSuccess Status = "success"
	// StatusFailed is written after a migration fails.
	StatusFailed Status = "failed"
)

// maxMessageBytes is the byte budget of the message column. The column is
// VARCHAR(1024) in utf8mb4, which counts characters, so a message within 1024
// bytes always fits.
const maxMessageBytes = 1024

// Record is one row of the migration record table.
type Record struct {
	ID          uint64    `db:"id"`
	MigrationID string    `db:"migration_id"`
	Version     string    `db:"version"`
	Status      Status    `db:"status"`
	Message     string    `db:"message"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// Records holds the migration records of one database, keyed by migration ID.
type Records map[string]Record

// RecordStore reads and writes the migration record table of one database.
// Writes go through the bare orm outside any migration transaction, so a
// failed record survives the rollback of the migration it describes.
type RecordStore struct {
	o orm.Interface
}

// NewRecordStore returns the record store of the database behind o.
func NewRecordStore(o orm.Interface) *RecordStore {
	return &RecordStore{o: o}
}

// Load reads every record of the database. A status other than running,
// success or failed, or a version that does not parse, means the table was
// written by something else, and is reported as a precondition failure
// whatever the row's status.
func (s *RecordStore) Load(kt *kit.Kit) (Records, error) {
	expr := fmt.Sprintf("SELECT `id`, `migration_id`, `version`, `status`, `message`, `created_at`, `updated_at` "+
		"FROM `%s`", RecordTable)
	rows := make([]Record, 0)
	if err := s.o.Do().Select(kt.Ctx, &rows, expr, map[string]interface{}{}); err != nil {
		logs.Errorf("list migration records failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list migration records failed, err: %v", err)
	}

	records := make(Records, len(rows))
	for _, r := range rows {
		switch r.Status {
		case StatusRunning, StatusSuccess, StatusFailed:
		default:
			return nil, fmt.Errorf("%w: migration record %s has unknown status %q",
				ErrPrecondition, r.MigrationID, r.Status)
		}
		if _, err := register.Parse(r.Version); err != nil {
			return nil, fmt.Errorf("%w: migration record %s has unparsable version %q, err: %v",
				ErrPrecondition, r.MigrationID, r.Version, err)
		}
		records[r.MigrationID] = r
	}
	return records, nil
}

// MarkRunning records that m is about to run. A new row is inserted for a
// migration never seen before; an existing row goes back to running with its
// message cleared and its version left for MarkSuccess to refresh.
func (s *RecordStore) MarkRunning(kt *kit.Kit, m register.Migration) error {
	expr := fmt.Sprintf("INSERT INTO `%s` (`migration_id`, `version`, `status`, `message`) "+
		"VALUES (:migration_id, :version, :status, '') "+
		"ON DUPLICATE KEY UPDATE `status` = :status, `message` = '', `updated_at` = NOW()", RecordTable)
	arg := map[string]interface{}{
		"migration_id": m.ID,
		"version":      m.Version,
		"status":       StatusRunning,
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
		"status":       StatusSuccess,
	}
	expr := fmt.Sprintf("UPDATE `%s` SET `status` = :status, `message` = '', `version` = :version "+
		"WHERE `migration_id` = :migration_id", RecordTable)
	return s.update(kt, m, expr, arg)
}

// MarkFailed records that m failed with runErr. The stored message is cut to
// maxMessageBytes; the full error is logged here.
func (s *RecordStore) MarkFailed(kt *kit.Kit, m register.Migration, runErr error) error {
	msg := ""
	if runErr != nil {
		msg = runErr.Error()
	}
	logs.Errorf("migration failed, err: %s, id: %s, version: %s, rid: %s", msg, m.ID, m.Version, kt.Rid)

	arg := map[string]interface{}{
		"migration_id": m.ID,
		"status":       StatusFailed,
		"message":      truncateMessage(msg, maxMessageBytes),
	}
	expr := fmt.Sprintf("UPDATE `%s` SET `status` = :status, `message` = :message "+
		"WHERE `migration_id` = :migration_id", RecordTable)
	return s.update(kt, m, expr, arg)
}

// update runs a status update of m's row. MarkRunning always runs first and
// the status always changes, so zero affected rows means the row is missing.
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

// truncateMessage cuts msg to at most limit bytes without splitting a
// multi-byte character. Invalid UTF-8 is replaced first, since utf8mb4
// columns reject it.
func truncateMessage(msg string, limit int) string {
	msg = strings.ToValidUTF8(msg, "?")
	if len(msg) <= limit {
		return msg
	}
	cut := limit
	for cut > 0 && !utf8.RuneStart(msg[cut]) {
		cut--
	}
	return msg[:cut]
}

// CurrentVersion returns the highest version among the success records,
// compared with register.Compare rather than as strings. ok is false when
// there is no success record. Load already rejects unparsable versions; the
// check is kept here for records built without Load, and still fails rather
// than skipping the row.
func CurrentVersion(records Records) (current register.Version, ok bool, err error) {
	for _, r := range records {
		if r.Status != StatusSuccess {
			continue
		}
		v, err := register.Parse(r.Version)
		if err != nil {
			return register.Version{}, false, fmt.Errorf("%w: migration record %s has unparsable version %q, err: %v",
				ErrPrecondition, r.MigrationID, r.Version, err)
		}
		// Equal versions (v1.9.3 and v1.9.3.0) keep the smaller raw text, so map
		// iteration order does not change the result.
		if !ok || register.Compare(v, current) > 0 ||
			(register.Compare(v, current) == 0 && v.Raw < current.Raw) {
			current, ok = v, true
		}
	}
	return current, ok, nil
}

// WarnVersionDrift logs a WARN and returns true when rec is a success record
// whose version differs from the version m is registered with now. This is
// expected after a feature branch is archived, so it never fails the run.
func WarnVersionDrift(kt *kit.Kit, rec Record, m register.Migration) bool {
	if rec.Status != StatusSuccess || rec.Version == m.Version {
		return false
	}
	logs.Warnf("migration version drift, recorded: %s, registered: %s, id: %s, skip as success, rid: %s",
		rec.Version, m.Version, m.ID, kt.Rid)
	return true
}

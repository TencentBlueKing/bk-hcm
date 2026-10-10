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
	"context"
	"fmt"
	"strings"
	"testing"

	"hcm/migrate/register"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/migrate"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Readable migration IDs. These tests must not touch register.Main.
const (
	migA = "20260101-1200-A-0001"
	migB = "20260101-1200-B-0002"
	migC = "20260101-1200-C-0003"
	migD = "20260101-1200-D-0004"
	migE = "20260101-1200-E-0005"
	migF = "20260101-1200-F-0006"
)

func noopUp(_ context.Context, _ orm.Interface) error {
	return nil
}

// pkgTag keeps only lowercase letters and digits for a migration directory tag.
func pkgTag(raw string) string {
	raw = strings.ToLower(raw)
	var b strings.Builder
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '_' || r == ' ' || r == '-':
			b.WriteByte('_')
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "x"
	}
	return out
}

// mustMigration builds one migration via NewMigration with ParsedVersion
// populated. A composite literal leaves the unexported parsed field zero, so
// baseline selection would treat it as v0.0.0. Engine tests cannot Regist
// because types here live under hcm/migrate/engine, not migrations/.
func mustMigration(t *testing.T, id, version, timestamp, tag string) register.Migration {
	t.Helper()
	safe := pkgTag(tag)
	var pkgPath string
	if register.IsPending(version) {
		pkgPath = fmt.Sprintf("%smain/pending/%s_%s", constant.MigrationPkgPrefix, timestamp, safe)
	} else {
		pkgPath = fmt.Sprintf("%smain/%s/%s_%s_%s", constant.MigrationPkgPrefix, version, version, timestamp, safe)
	}
	m, err := register.NewMigration("main", pkgPath, id, version, timestamp, noopUp)
	require.NoError(t, err)
	return m
}

func mustVersion(t *testing.T, raw string) register.Version {
	t.Helper()
	v, err := register.Parse(raw)
	require.NoError(t, err)
	return v
}

// assertNoSentinel reports that err is a plain failure: it wraps none of the
// exit-code sentinels.
func assertNoSentinel(t *testing.T, err error) {
	t.Helper()
	assert.Error(t, err)
	assert.NotErrorIs(t, err, migrate.ErrUsage)
	assert.NotErrorIs(t, err, migrate.ErrPrecondition)
	assert.NotErrorIs(t, err, migrate.ErrMissed)
	assert.NotErrorIs(t, err, migrate.ErrRegistry)
	assert.NotErrorIs(t, err, migrate.ErrIDReuse)
}

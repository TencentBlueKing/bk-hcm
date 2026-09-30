#!/usr/bin/env bash
# new-migrate.sh creates one migration directory with its migrate.go and adds
# its blank import to migrations/imports.go.
set -euo pipefail

usage() {
	cat <<'EOF'
Usage:
  new-migrate.sh --database main|obs --desc <name> [--version <version>]

  --database   main or obs
  --desc       lowercase words joined by "_", e.g. add_bk_asset_id
  --version    vX.Y.Z, vX.Y.Z.N or vX.Y.Z-<label>.N; omit to create it in pending/

Three-segment versions go to <database>/vX.Y.Z/, four-segment ones to
<database>/vX.Y.Z.x/, and migrations without a version to <database>/pending/.
EOF
}

database=""
desc=""
version=""
while [[ $# -gt 0 ]]; do
	case "$1" in
	--database | --desc | --version)
		if [[ $# -lt 2 ]]; then
			echo "$1 needs a value" >&2
			exit 2
		fi
		case "$1" in
		--database) database="$2" ;;
		--desc) desc="$2" ;;
		--version) version="$2" ;;
		esac
		shift 2
		;;
	-h | --help) usage; exit 0 ;;
	*) echo "unknown argument: $1" >&2; usage >&2; exit 2 ;;
	esac
done

if [[ "$database" != "main" && "$database" != "obs" ]]; then
	echo "--database must be main or obs" >&2
	exit 2
fi
if [[ ! "$desc" =~ ^[a-z][a-z0-9]*(_[a-z0-9]+)*$ ]]; then
	echo "--desc must be lowercase words joined by \"_\", got: $desc" >&2
	exit 2
fi

root="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=lib.sh
source "$root/scripts/lib.sh"

# 不给版本就进 pending/，注册版本用常量。给了版本则三位进 vX.Y.Z/，
# 四位进 vX.Y.Z.x/，目录名前面加版本前缀。
if [[ -z "$version" ]]; then
	group="pending"
	prefix=""
	version_expr="constant.MigrationPendingVersion"
else
	group="$(version_group "$version")" || exit 2
	prefix="${version}_"
	version_expr="\"$version\""
fi

migrations="$root/migrations"
imports="$migrations/imports.go"
if [[ "$database" == "main" ]]; then
	registry="register.Main"
else
	registry="register.Obs"
fi

# 同一秒再建一条会撞上同一个「时间戳_名称」，所以等到下一秒。
ts="$(date +%Y%m%d%H%M%S)"
while find "$migrations/$database" -mindepth 2 -maxdepth 2 -type d -name "*${ts}_*" 2>/dev/null | grep -q .; do
	sleep 1
	ts="$(date +%Y%m%d%H%M%S)"
done

# ID 和目录时间戳取自同一时刻：前 13 位是 yyyyMMdd-HHmm，tag 由 --desc 转成大写，
# 末尾 4 位十六进制随机。合入之后这串不能再改。
tag="$(echo "$desc" | tr 'a-z_' 'A-Z-')"
rand="$(od -An -N2 -tx1 /dev/urandom | tr -d ' \n' | tr 'a-f' 'A-F')"
id="${ts:0:8}-${ts:8:4}-${tag}-${rand}"
if [[ ${#id} -gt 64 ]]; then
	echo "migration id $id is longer than 64 characters, use a shorter --desc" >&2
	exit 2
fi

if [[ ! -f "$imports" ]]; then
	echo "$imports does not exist" >&2
	exit 1
fi
rel="$database/$group/${prefix}${ts}_${desc}"
dir="$migrations/$rel"
if [[ -e "$dir" ]]; then
	echo "$dir already exists" >&2
	exit 1
fi
mkdir -p "$dir"

# 只有 pending 模板要写常量 PENDING，才需要 import constant 包。
constant_import=""
if [[ "$group" == "pending" ]]; then
	constant_import=$'\t"hcm/pkg/criteria/constant"\n'
fi

cat >"$dir/migrate.go" <<EOF
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

// Package migration is the migration ${desc}.
package migration

import (
	"context"

	"hcm/migrate/register"
	"hcm/migrate/util"
${constant_import}	"hcm/pkg/dal/dao/orm"
)

// ID is the unique identifier of this migration file, in the form
// <date>-<time>-<desc>-<random>. It must never change once executed.
// Migrations that share an ID run only the earlier one in execution order.
const ID = "${id}"

func init() {
	${registry}.Regist(ID, ${version_expr}, "${ts}", &migration{})
}

type migration struct{}

// Up applies the migration. Replace the example block below.
func (m *migration) Up(ctx context.Context, o orm.Interface) error {
	/********************** example usage, replace this block **********************/
	_, err := util.CreateTableIfNotExists(ctx, o, "example", "CREATE TABLE \`example\` ("+
		"\`id\` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT, "+
		"\`name\` VARCHAR(64) NOT NULL DEFAULT '', "+
		"PRIMARY KEY (\`id\`))")
	if err != nil {
		return err
	}

	if err = util.AddColumn(ctx, o, util.AddColumnOpt{
		Table:   "example",
		Column:  "note",
		Type:    "varchar(255)",
		NotNull: true,
		Default: util.StringDefault(""),
	}); err != nil {
		return err
	}

	err = util.AddIndex(ctx, o, "example", "idx_name", []string{"name"}, false)
	/****************************** example usage end ******************************/
	return err
}
EOF

# 在已有的 import 行上加一行再整体写回；import 块被删光时也能重建。
{
	imported_paths "$imports"
	echo "$rel"
} | write_imports "$imports"
gofmt -w "$dir/migrate.go"

say_title "new-migrate"
say_ok "已创建" "$rel/migrate.go"
if [[ -z "$version" ]]; then
	say_ok "版本" "pending，定版前不会在默认 up 里执行"
else
	say_ok "版本" "$version"
fi
say_ok "ID" "$id" "执行过后不能再改"
say_ok "imports.go" "已写入空白 import"
"$root/scripts/check-migrate.sh"

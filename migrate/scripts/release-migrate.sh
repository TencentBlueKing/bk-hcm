#!/usr/bin/env bash
# release-migrate.sh moves every migration in <database>/pending/ into the
# group directory of a released version. The directory gets the version as
# its prefix and Regist gets the version instead of PENDING. The ID and the
# migration suffix stay the same, so a pending migration already run is
# skipped and backfilled, not run again.
set -euo pipefail

usage() {
	cat <<'EOF'
Usage:
  release-migrate.sh --version <version> [--database main]...

  --version    vX.Y.Z, vX.Y.Z.N or vX.Y.Z-<label>.N
  --database   repeatable; default main
EOF
}

version=""
databases=()
while [[ $# -gt 0 ]]; do
	case "$1" in
	--version | --database)
		if [[ $# -lt 2 ]]; then
			echo "$1 needs a value" >&2
			exit 2
		fi
		if [[ "$1" == "--version" ]]; then version="$2"; else databases+=("$2"); fi
		shift 2
		;;
	-h | --help) usage; exit 0 ;;
	*) echo "unknown argument: $1" >&2; usage >&2; exit 2 ;;
	esac
done
if [[ ${#databases[@]} -eq 0 ]]; then
	databases=(main)
fi

root="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=lib.sh
source "$root/scripts/lib.sh"
migrations="$root/migrations"
imports="$migrations/imports.go"

if [[ -z "$version" ]]; then
	echo "--version is required" >&2
	exit 2
fi
group="$(version_group "$version")" || exit 2
for db in "${databases[@]}"; do
	if [[ "$db" != "main" ]]; then
		echo "--database must be main, got: $db" >&2
		exit 2
	fi
done

# 起点必须是一棵检查通过的树，否则定版会在坏数据上继续改。
# 通过时不重复打印；失败时把预检查清单打出来。
if ! check_out="$("$root/scripts/check-migrate.sh" 2>&1)"; then
	printf '%s\n' "$check_out"
	exit 1
fi

# 先确认每条都能定版，再动文件。有一条目标已存在就一条都不移动。
moves=()
released=()
for db in "${databases[@]}"; do
	for dir in "$migrations/$db/pending"/*; do
		[[ -d "$dir" ]] || continue
		name="$(basename "$dir")"
		target="$migrations/$db/$group/${version}_${name}"
		if [[ -e "$target" ]]; then
			echo "$db/$group/${version}_${name} already exists" >&2
			exit 1
		fi
		moves+=("$db $name")
	done
done
if [[ ${#moves[@]} -eq 0 ]]; then
	say_title "release-migrate" "$version"
	say_ok "没有待定版的迁移" "库: ${databases[*]}"
	exit 0
fi

paths="$(mktemp)"
trap 'rm -f "$paths" "$paths.tmp"' EXIT
imported_paths "$imports" >"$paths"
for move in "${moves[@]}"; do
	db="${move%% *}"
	name="${move#* }"
	target="$migrations/$db/$group/${version}_${name}"
	mkdir -p "$migrations/$db/$group"
	mv "$migrations/$db/pending/$name" "$target"

	file="$target/migrate.go"
	# 注册版本从常量 PENDING 换成版本字符串。ID 和时间戳不动，「时间戳_名称」也不变，
	# 所以已经跑过的这条会被跳过并回填，不会再执行。
	replace_in_file "$file" "constant.MigrationPendingVersion" "\"$version\""
	# PENDING 是 constant 包唯一的用处时，定版后这个 import 不再被使用，Go 编译会报错，所以删掉。
	if ! grep -q 'constant\.' "$file"; then
		grep -v '^[[:space:]]*"hcm/pkg/criteria/constant"$' "$file" >"$file.tmp"
		mv "$file.tmp" "$file"
	fi
	gofmt -w "$file"

	# 按整行替换 import 路径，不用子串替换，避免一个目录名是另一个的前缀时误改。
	awk -v old="$db/pending/$name" -v new="$db/$group/${version}_${name}" \
		'{ print ($0 == old ? new : $0) }' "$paths" >"$paths.tmp"
	mv "$paths.tmp" "$paths"
	released+=("$db"$'\t'"pending/$name"$'\t'"$group/${version}_$name")
done
write_imports "$imports" <"$paths"

say_title "release-migrate" "$version"
for item in "${released[@]}"; do
	db="${item%%$'\t'*}"
	rest="${item#*$'\t'}"
	from="${rest%%$'\t'*}"
	to="${rest#*$'\t'}"
	say_ok "$db" "$from" "→ $to"
done
say_done "${#released[@]} 条已定版"
"$root/scripts/check-migrate.sh"

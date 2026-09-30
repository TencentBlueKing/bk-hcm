#!/usr/bin/env bash
# check-migrate.sh 是 migrations/ 目录的预检查，不编译 Go，也不连库。
# 结束时六项都会列出来：通过打勾，失败打叉，原因写在叉下面。有失败则退出码 1。
# 同一个 ID、「时间戳_名称」相同只警告，不打叉。
#
# 第一次用，先看这六项在查什么：
#
#   1. 目录结构
#      migrations/ 下只能有 imports.go、main/。
#      每个库下面的分组只能是 pending、vX.Y.Z、vX.Y.Z.x。
#      分组里只能放迁移目录，迁移目录里不能再有子目录。
#
#   2. 目录名
#      pending 是 <timestamp>_<tag>。
#      已定版是 <version>_<timestamp>_<tag>。
#      三位版本进 vX.Y.Z/，第四段（.N 或 -label.N）进 vX.Y.Z.x/。
#
#   3. 注册
#      每个迁移目录有 migrate.go，并且只注册一次。
#      注册的库、版本、时间戳必须和目录一致。
#      pending 的版本必须是常量 constant.MigrationPendingVersion。
#      ID 不超过 64 个字符，格式是 <date>-<time>-<desc>-<random>。
#
#   4. imports.go
#      每个迁移目录有一行空白 import。
#      不能重复，也不能 import 一个不存在的目录。
#
#   5. migration ID
#      目录名里从 14 位时间戳起到结尾，是「时间戳_名称」。
#      例如 v1.9.3_20260927135045_create_migrate_sample
#      这一段是 20260927135045_create_migrate_sample。
#      同一个库、同一个 ID，这一段不同就失败。
#      只有前面的版本号不同（pending 定版、内部四位和外部三位）只警告。
#
#   6. 依赖
#      迁移文件只能依赖 register、util 和 constant。
#      不能依赖 engine、schema、cli，也不能依赖别的迁移。
set -euo pipefail

root="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=lib.sh
source "$root/scripts/lib.sh"
migrations="$root/migrations"
imports="$migrations/imports.go"
export LC_ALL=C

# fail_item 把一条失败记到对应检查项上，最后和通过项一起打出来。
fail_item() {
	local key=$1
	shift
	printf '%s\n' "$*" >>"$work/err.$key"
}

# print_result 列出六项。通过打勾，失败打叉，原因跟在该项下面。
print_result() {
	local passed=0 failed_n=0 key title detail line
	echo
	printf '%scheck-migrate%s  migrations/ 预检查\n' "$c_bold" "$c_reset"
	echo
	while IFS=$'\t' read -r key title detail; do
		[[ -n "$key" ]] || continue
		if [[ -s "$work/err.$key" ]]; then
			failed_n=$((failed_n + 1))
			printf '  %s✗%s  %s\n' "$c_red" "$c_reset" "$title"
			printf '     %s%s%s\n' "$c_dim" "$detail" "$c_reset"
			while IFS= read -r line; do
				printf '     %s%s%s\n' "$c_red" "$line" "$c_reset"
			done <"$work/err.$key"
			echo
			continue
		fi
		passed=$((passed + 1))
		printf '  %s✓%s  %s\n' "$c_green" "$c_reset" "$title"
		printf '     %s%s%s\n' "$c_dim" "$detail" "$c_reset"
		if [[ -s "$work/warn.$key" ]]; then
			while IFS= read -r line; do
				printf '     %s%s%s\n' "$c_yellow" "$line" "$c_reset"
			done <"$work/warn.$key"
		fi
		echo
	done <<'EOF'
layout	目录结构	migrations/ 下只有 imports.go、main/，分组和迁移目录合法
name	目录名	pending 与已定版的目录名、版本分组、时间戳一致
regist	注册	每个目录一份 migrate.go，只注册一次，库、版本、时间戳、ID 与目录一致
imports	imports.go	每个迁移目录有且只有一行空白 import
id	migration ID	同一个库里，同一个 ID 不能出现在两份「时间戳_名称」不同的目录里
deps	依赖	迁移不依赖 engine、schema、cli，也不依赖别的迁移
EOF
	if [[ "$failed_n" -eq 0 ]]; then
		printf '  %s%d passed%s\n' "$c_green" "$passed" "$c_reset"
	else
		printf '  %s%d passed%s, %s%d failed%s\n' "$c_green" "$passed" "$c_reset" "$c_red" "$failed_n" "$c_reset"
	fi
	echo
}

num='(0|[1-9][0-9]*)'
# 目录名里从 14 位时间戳起到结尾是「时间戳_名称」。pending 只有这一段；已定版的前面还有版本。
ts_tag='([0-9]{14})_[a-z0-9]+(_[a-z0-9]+)*'
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
: >"$work/dirs"
: >"$work/ids"

if [[ ! -f "$imports" ]]; then
	fail_item imports "migrations/imports.go 不存在"
fi

# Only imports.go and main/ live directly under migrations/.
for entry in "$migrations"/*; do
	case "$(basename "$entry")" in
	imports.go | main) ;;
	*) fail_item layout "migrations/ 下有多余项 $(basename "$entry")" ;;
	esac
done

for db in main; do
	[[ -d "$migrations/$db" ]] || continue
	registry="Main"

	for group in "$migrations/$db"/*; do
		[[ -e "$group" ]] || continue
		gname="${group##*/}"
		# 分组目录只允许 pending、vX.Y.Z、vX.Y.Z.x 三种，里面只能再放迁移目录。
		if [[ ! -d "$group" || ! "$gname" =~ ^(pending|v$num\.$num\.$num(\.x)?)$ ]]; then
			fail_item layout "$db/$gname 不是分组目录，只能是 pending、vX.Y.Z 或 vX.Y.Z.x"
			continue
		fi

		for dir in "$group"/*; do
			[[ -e "$dir" ]] || continue
			name="${dir##*/}"
			rel="$db/$gname/$name"
			if [[ ! -d "$dir" ]]; then
				fail_item layout "$rel 不是迁移目录"
				continue
			fi
			# 迁移目录里只允许放文件。隐藏子目录也算。
			for child in "$dir"/* "$dir"/.[!.]* "$dir"/..?*; do
				if [[ -d "$child" ]]; then
					fail_item layout "$rel 里还有子目录"
					break
				fi
			done

			# 目录名不合法就不再检查这个目录，避免同一处问题连报好几条。
			if [[ "$gname" == "pending" ]]; then
				if [[ ! "$name" =~ ^$ts_tag$ ]]; then
					fail_item name "$rel 的目录名必须是 <timestamp>_<tag>"
					continue
				fi
				dir_version=""
				dir_ts="${BASH_REMATCH[1]}"
			else
				if [[ ! "$name" =~ ^(v[^_/]+)_$ts_tag$ ]]; then
					fail_item name "$rel 的目录名必须是 <version>_<timestamp>_<tag>"
					continue
				fi
				dir_version="${BASH_REMATCH[1]}"
				dir_ts="${BASH_REMATCH[2]}"
				if ! version_group_of "$dir_version"; then
					fail_item name "$rel 的版本 $dir_version 不合法"
					continue
				fi
				if [[ "$group_of" != "$gname" ]]; then
					fail_item name "$rel 的版本 $dir_version 应该放在 $group_of/"
				fi
			fi

			if [[ ! -f "$dir/migrate.go" ]]; then
				fail_item regist "$rel 没有 migrate.go"
				continue
			fi
			echo "$rel" >>"$work/dirs"

			# 每个 .go 只读一遍：数 .Regist( 次数，记下注册那一行、const ID 和跨库引用。
			# 按出现次数计，同一行写两次也算两次；行注释里提到的不算。
			total=0
			in_main=0
			call=""
			id=""
			for go_file in "$dir"/*.go; do
				[[ -f "$go_file" ]] || continue
				while IFS= read -r line || [[ -n "$line" ]]; do
					[[ "$line" =~ ^[[:space:]]*// ]] && continue
					rest="$line"
					while [[ "$rest" == *".Regist("* ]]; do
						total=$((total + 1))
						[[ "$go_file" == "$dir/migrate.go" ]] && in_main=$((in_main + 1))
						rest="${rest#*.Regist(}"
					done
					if [[ "$go_file" == "$dir/migrate.go" ]]; then
						[[ -z "$call" && "$line" == *".Regist("* ]] && call="$line"
						if [[ -z "$id" && "$line" =~ $id_line_re ]]; then
							id="${BASH_REMATCH[1]}"
						fi
					fi
				done <"$go_file"
			done
			if [[ "$total" != "1" || "$in_main" != "1" ]]; then
				fail_item regist "$rel 必须在 migrate.go 里且只注册一次"
				continue
			fi

			# 注册调用必须在一行里，顺序是 register.<库>.Regist(ID, 版本, "时间戳", ...)。
			# pending 的版本必须是常量；已定版的版本和时间戳必须和目录名一致。
			if [[ ! "$call" =~ register\.([A-Za-z]+)\.Regist\(ID,[[:space:]]*([^,]+),[[:space:]]*\"([0-9]*)\", ]]; then
				fail_item regist "$rel 的注册必须写成一行：register.$registry.Regist(ID, <version>, \"<timestamp>\", ...)"
				continue
			fi
			call_registry="${BASH_REMATCH[1]}"
			call_version="${BASH_REMATCH[2]}"
			call_ts="${BASH_REMATCH[3]}"
			if [[ "$call_registry" != "$registry" ]]; then
				fail_item regist "$rel 在 $db/ 下，却注册到了 register.$call_registry"
			fi
			if [[ -z "$dir_version" && "$call_version" != "constant.MigrationPendingVersion" ]]; then
				fail_item regist "$rel 在 pending/ 下，版本必须是 constant.MigrationPendingVersion，实际是 $call_version"
			fi
			if [[ -n "$dir_version" && "$call_version" != "\"$dir_version\"" ]]; then
				fail_item regist "$rel 注册的版本是 $call_version，目录名是 \"$dir_version\""
			fi
			if [[ "$call_ts" != "$dir_ts" ]]; then
				fail_item regist "$rel 注册的时间戳是 $call_ts，目录名是 $dir_ts"
			fi

			if [[ -z "$id" ]]; then
				fail_item regist "$rel 没有 const ID"
				continue
			fi
			if [[ ${#id} -gt 64 || ! "$id" =~ $id_re ]]; then
				fail_item regist "$rel 的 ID $id 不合法，必须是 <date>-<time>-<desc>-<random>，且不超过 64 个字符"
				continue
			fi
			echo "$db $id ${name#"$dir_version"_} $rel" >>"$work/ids"
		done
	done
done

# 目录集合和 import 集合对差：只在目录里的是漏 import，只在 import 里的是多余 import。
# 管道右边的 while 跑在子 shell 里，失败写进文件，最后和别的项一起打出来。
if [[ -f "$imports" ]]; then
	imported_paths "$imports" | sort >"$work/imported"
	uniq -d "$work/imported" | while read -r rel; do
		fail_item imports "migrations/imports.go 重复 import 了 $rel"
	done
	sort -u "$work/imported" >"$work/imported.uniq"
	sort "$work/dirs" >"$work/dirs.sorted"
	comm -23 "$work/dirs.sorted" "$work/imported.uniq" | while read -r rel; do
		fail_item imports "$rel 没有写进 migrations/imports.go"
	done
	comm -13 "$work/dirs.sorted" "$work/imported.uniq" | while read -r rel; do
		fail_item imports "migrations/imports.go 引用了不存在的迁移目录 $rel"
	done
fi

# 同一个库里同一个 ID 出现多次：时间戳_名称不同是 ID 复用，失败；
# 这一段相同、只是前面的版本号不同，是定版和内外版本的正常形态，只警告。
sort "$work/ids" | awk -v errfile="$work/err.id" -v warnfile="$work/warn.id" '
	{
		key = $1 " " $2
		dirs[key] = dirs[key] " " $4
		if (!((key, $3) in seen)) {
			seen[key, $3] = 1
			suffixes[key]++
		}
		count[key]++
	}
	END {
		for (key in count) {
			if (count[key] < 2) continue
			split(key, parts, " ")
			if (suffixes[key] > 1) {
				printf "%s 的 ID %s 出现在多份迁移里，时间戳_名称不同:%s\n", parts[1], parts[2], dirs[key] > errfile
			} else {
				printf "%s 的 ID %s 出现在多份迁移里，时间戳_名称相同:%s\n", parts[1], parts[2], dirs[key] > warnfile
			}
		}
	}
'

# 迁移只能依赖 register、util 和 constant，不能依赖执行层，也不能依赖别的迁移。
for db in main; do
	[[ -d "$migrations/$db" ]] || continue
	grep -rlE '"hcm/migrate/(engine|schema|cli|migrations)(/|")' "$migrations/$db" >"$work/bad_deps" || true
	while read -r file; do
		[[ -n "$file" ]] || continue
		fail_item deps "${file#"$root/"} 依赖了 engine、schema、cli 或其他迁移"
	done <"$work/bad_deps"
done

print_result
for key in layout name regist imports id deps; do
	if [[ -s "$work/err.$key" ]]; then
		exit 1
	fi
done

#!/usr/bin/env bash
# archive-feat-migrate.sh 把一个特性分支的迁移收进主线版本。
# 特性分支合回主线、或自己转成正式线时，出包的人先跑它，再发一次带 --catch-up 的包。
#
# 例：
#   main/v1.9.3.x/v1.9.3-tenant.1_20260928110002_add_account_tenant_id  注册版本 v1.9.3-tenant.1
#   → main/v1.9.3/v1.9.3_20260928110002_add_account_tenant_id          注册版本 v1.9.3
#
# 只改目录名、所在目录和注册版本。Migration ID、时间戳和 Up 都不动，
# 所以特性环境上跑过的，归档后按 ID 跳过，只多一条「版本和记录不一致」的警告。
#
# 只动已定版、标签是 --label 的迁移。pending/、主线三位版本、内部的 .N、别的标签都不动。
#
# 下面任一情况都会停下，一个文件都不改：
#   - 目标目录已存在，且 ID 不同
#   - 两条迁移会落到同一个目标目录
#   - 归档后的执行顺序和归档前不一样。收成三位版本后同一版本内只靠时间戳排，
#     原来靠 -tenant.1、-tenant.2 排的先后可能因此变掉
# 目标目录已存在且 ID 相同，视为已经归档过，跳过并在报告里列出。
#
# 先加 --dry-run 看报告，确认后去掉 --dry-run 正式归档。
set -euo pipefail

usage() {
	cat <<'EOF'
Usage:
  archive-feat-migrate.sh --label <label> [--database main|obs]... [--dry-run]

  --label      特性分支标签，例如 tenant，对应版本 vX.Y.Z-tenant.N
  --database   可重复；不写时主库和 OBS 库一起处理
  --dry-run    只打印报告，不改任何文件
EOF
}

label=""
databases=()
dry_run=0
while [[ $# -gt 0 ]]; do
	case "$1" in
	--label | --database)
		if [[ $# -lt 2 ]]; then
			echo "$1 needs a value" >&2
			exit 2
		fi
		if [[ "$1" == "--label" ]]; then label="$2"; else databases+=("$2"); fi
		shift 2
		;;
	--dry-run) dry_run=1; shift ;;
	-h | --help) usage; exit 0 ;;
	*) echo "unknown argument: $1" >&2; usage >&2; exit 2 ;;
	esac
done
if [[ ${#databases[@]} -eq 0 ]]; then
	databases=(main obs)
fi

if [[ ! "$label" =~ ^[a-z][a-z0-9]*$ ]]; then
	echo "--label must be lowercase letters and digits starting with a letter, got: $label" >&2
	exit 2
fi
for db in "${databases[@]}"; do
	if [[ "$db" != "main" && "$db" != "obs" ]]; then
		echo "--database must be main or obs, got: $db" >&2
		exit 2
	fi
done

root="$(cd "$(dirname "$0")/.." && pwd)"
# shellcheck source=lib.sh
source "$root/scripts/lib.sh"
migrations="$root/migrations"
imports="$migrations/imports.go"
export LC_ALL=C

# 起点必须是一棵检查通过的树，否则归档会在坏数据上继续改。
# 通过时不重复打印；失败时把预检查清单打出来。
if ! check_out="$("$root/scripts/check-migrate.sh" 2>&1)"; then
	printf '%s\n' "$check_out"
	exit 1
fi

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT
# plan: 要移动的迁移，每行 db、源、目标、源版本、目标版本、ID，用 tab 分隔。
# skip: 已归档过、跳过的迁移。conflict: 会导致停下的冲突说明。
# before / after: 归档前后每条迁移的排序键，用来比较执行顺序。
: >"$work/plan"
: >"$work/skip"
: >"$work/conflict"
: >"$work/before"
: >"$work/after"
: >"$work/drop"

num='(0|[1-9][0-9]*)'
# 特性分支目录名：<vX.Y.Z>-<label>.<N>_<时间戳_名称>。第 1 组是主线版本，第 6 组是时间戳_名称。
feat_re="^(v${num}\.${num}\.${num})-${label}\.${num}_([0-9]{14}_.+)$"

# version_key 把版本拆成六列排序键，写进 key：三段数字、有没有第四段、标签、序号。
# 没有标签的第四段记成 "-"，C locale 下它排在任何小写字母前面，和 register.Compare 一致。
version_key() {
	if [[ "$1" =~ ^v${num}\.${num}\.${num}$ ]]; then
		key="${BASH_REMATCH[1]}"$'\t'"${BASH_REMATCH[2]}"$'\t'"${BASH_REMATCH[3]}"$'\t0\t-\t0'
	elif [[ "$1" =~ ^v${num}\.${num}\.${num}\.${num}$ ]]; then
		key="${BASH_REMATCH[1]}"$'\t'"${BASH_REMATCH[2]}"$'\t'"${BASH_REMATCH[3]}"$'\t1\t-\t'"${BASH_REMATCH[4]}"
	elif [[ "$1" =~ ^v${num}\.${num}\.${num}-([a-z][a-z0-9]*)\.${num}$ ]]; then
		key="${BASH_REMATCH[1]}"$'\t'"${BASH_REMATCH[2]}"$'\t'"${BASH_REMATCH[3]}"$'\t1\t'"${BASH_REMATCH[4]}"$'\t'"${BASH_REMATCH[5]}"
	else
		return 1
	fi
}

for db in "${databases[@]}"; do
	[[ -d "$migrations/$db" ]] || continue
	# 只看已定版的分组。pending/ 不以 v 开头，自然不在这里。
	for group in "$migrations/$db"/v*; do
		[[ -d "$group" ]] || continue
		gname="${group##*/}"
		total=0
		moved=0
		for dir in "$group"/*; do
			[[ -f "$dir/migrate.go" ]] || continue
			total=$((total + 1))
			name="${dir##*/}"
			ver="${name%%_*}"
			rest="${name#*_}"
			ts="${rest%%_*}"
			read_id "$dir/migrate.go"
			id="$mig_id"
			new_ver="$ver"

			if [[ "$gname" == *.x && "$name" =~ $feat_re ]]; then
				base="${BASH_REMATCH[1]}"
				from_rel="$db/$gname/$name"
				to_rel="$db/$base/${base}_${BASH_REMATCH[6]}"
				target="$migrations/$to_rel"
				if [[ -e "$target" ]]; then
					target_id=""
					if [[ -f "$target/migrate.go" ]]; then
						read_id "$target/migrate.go"
						target_id="$mig_id"
					fi
					if [[ -n "$target_id" && "$target_id" == "$id" ]]; then
						printf '%s\t%s\t%s\t%s\n' "$db" "$from_rel" "$to_rel" "$id" >>"$work/skip"
					else
						printf '%s 要移到 %s，但目标已存在，ID 是 %s，不是 %s\n' \
							"$from_rel" "$to_rel" "${target_id:-未知}" "$id" >>"$work/conflict"
					fi
				else
					printf '%s\t%s\t%s\t%s\t%s\t%s\n' "$db" "$from_rel" "$to_rel" "$ver" "$base" "$id" >>"$work/plan"
					new_ver="$base"
					moved=$((moved + 1))
				fi
			fi

			version_key "$ver"
			printf '%s\t%s\t%s\t%s\n' "$db" "$key" "$ts" "$id" >>"$work/before"
			version_key "$new_ver"
			printf '%s\t%s\t%s\t%s\n' "$db" "$key" "$ts" "$id" >>"$work/after"
		done
		# 这个 .x 分组里的迁移全部移走后，目录就空了，归档时一起删掉。
		if [[ "$gname" == *.x && "$total" -gt 0 && "$moved" -eq "$total" ]]; then
			echo "$db/$gname" >>"$work/drop"
		fi
	done
done

# 两条迁移落到同一个目标目录，比如 -tenant.1 和 -tenant.2 里时间戳和名称都一样。
cut -f3 "$work/plan" | sort | uniq -d | while read -r rel; do
	echo "有两条迁移都要移到 $rel" >>"$work/conflict"
done

# 按库、版本、时间戳、ID 排序，和 register 里的全序一致。列：库、六列版本键、时间戳、ID。
sort_keys() {
	sort -t$'\t' -k1,1 -k2,2n -k3,3n -k4,4n -k5,5n -k6,6 -k7,7n -k8,8 -k9,9 "$1" | cut -f1,9
}
sort_keys "$work/before" >"$work/before.sorted"
sort_keys "$work/after" >"$work/after.sorted"
# 同一个位置上的迁移前后不一样，就是顺序变了。
paste "$work/before.sorted" "$work/after.sorted" |
	awk -F'\t' '$2 != $4 { printf "%s 第 %d 条：归档前是 %s，归档后是 %s\n", $1, NR, $2, $4 }' >"$work/order"

planned=0
while IFS= read -r _; do planned=$((planned + 1)); done <"$work/plan"

detail="标签 -$label，库 ${databases[*]}"
if [[ "$dry_run" == 1 ]]; then
	detail="$detail，dry-run"
fi
say_title "archive-feat-migrate" "$detail"

if [[ "$planned" -eq 0 && ! -s "$work/skip" && ! -s "$work/conflict" ]]; then
	say_ok "没有要归档的迁移" "没有找到已定版、标签是 -$label 的迁移"
	exit 0
fi

while IFS=$'\t' read -r db from_rel to_rel from_ver base id; do
	say_ok "$id" "$from_rel" "→ $to_rel" "版本 $from_ver → $base"
done <"$work/plan"

while IFS=$'\t' read -r db from_rel to_rel id; do
	say_warn "已归档过，跳过  $id" "$from_rel" "目标 $to_rel 已存在，ID 相同"
done <"$work/skip"

failed=0
if [[ -s "$work/conflict" ]]; then
	failed=1
	lines=()
	while IFS= read -r line; do lines+=("$line"); done <"$work/conflict"
	say_fail "目标目录冲突" "${lines[@]}"
fi

if [[ -s "$work/order" ]]; then
	failed=1
	lines=()
	while IFS= read -r line; do lines+=("$line"); done <"$work/order"
	say_fail "执行顺序会变化" "收成主线版本后同一版本内只按时间戳排，先调整时间戳再归档" "${lines[@]}"
else
	say_ok "执行顺序不变" "按版本和时间戳排序，归档前后一致"
fi

if [[ -s "$work/drop" ]]; then
	lines=()
	while IFS= read -r line; do lines+=("$line/"); done <"$work/drop"
	say_ok "会删除空分组" "${lines[@]}"
fi

if [[ "$failed" == 1 ]]; then
	say_stop "有冲突，没有改动任何文件"
	exit 1
fi
if [[ "$dry_run" == 1 ]]; then
	say_done "dry-run：$planned 条待归档，没有改动任何文件"
	exit 0
fi
if [[ "$planned" -eq 0 ]]; then
	say_done "没有需要移动的迁移"
	exit 0
fi

# 真正动文件：移目录，只改 Regist 那一行里的版本字符串，再同步 imports.go。
imported_paths "$imports" >"$work/paths"
: >"$work/rename"
while IFS=$'\t' read -r db from_rel to_rel from_ver base id; do
	mkdir -p "$migrations/${to_rel%/*}"
	mv "$migrations/$from_rel" "$migrations/$to_rel"
	file="$migrations/$to_rel/migrate.go"
	awk -v old="\"$from_ver\"" -v new="\"$base\"" '
		index($0, ".Regist(") && (i = index($0, old)) > 0 {
			$0 = substr($0, 1, i - 1) new substr($0, i + length(old))
		}
		{ print }
	' "$file" >"$file.tmp"
	mv "$file.tmp" "$file"
	gofmt -w "$file"
	printf '%s\t%s\n' "$from_rel" "$to_rel" >>"$work/rename"
done <"$work/plan"

# 按整行替换 import 路径，不用子串替换，避免一个目录名是另一个的前缀时误改。
awk -F'\t' 'NR == FNR { to[$1] = $2; next } { print (($0 in to) ? to[$0] : $0) }' \
	"$work/rename" "$work/paths" >"$work/paths.new"
write_imports "$imports" <"$work/paths.new"

while IFS= read -r rel; do
	# macOS 可能在目录里留下 .DS_Store，它不是迁移，删掉才能把空分组删干净。
	rm -f "$migrations/$rel/.DS_Store"
	if ! rmdir "$migrations/$rel" 2>/dev/null; then
		say_warn "空分组没删掉" "$rel/ 里还有别的文件，请手动处理"
	fi
done <"$work/drop"

say_done "$planned 条已归档"
"$root/scripts/check-migrate.sh"

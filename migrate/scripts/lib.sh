#!/usr/bin/env bash
# lib.sh holds the helpers shared by the migration scripts. Source it; do not run it.

# id_re matches a migration ID, yyyyMMdd-HHmm-TAG-XXXX, as pkg/migrate does.
id_re='^[0-9]{8}-[0-9]{4}-[A-Z][A-Z0-9]*(-[A-Z0-9]+)*-[0-9A-F]{4}$'

# id_line_re 匹配 migrate.go 里的 const ID 那一行，行尾允许带注释。
id_line_re='^const ID = "([^"]*)"[[:space:]]*(//.*)?$'

# read_id 从 migrate.go 读出 const ID，结果写进 mig_id，不起子进程。没找到时 mig_id 为空。
read_id() {
	local line
	mig_id=""
	while IFS= read -r line || [[ -n "$line" ]]; do
		if [[ "$line" =~ $id_line_re ]]; then
			mig_id="${BASH_REMATCH[1]}"
			return 0
		fi
	done <"$1"
}

# version_group_of 把已定版版本映射到分组目录，结果写进 group_of，不起子 shell。
# 三位版本目录就叫这个版本，带第四段（.N 或 -label.N）的进 vX.Y.Z.x/。
# vX.Y.Z.0 和前导零都拒绝。版本不合法时返回 1，group_of 为空。
version_group_of() {
	local num='(0|[1-9][0-9]*)'
	group_of=""
	if [[ ${#1} -le 64 && "$1" =~ ^v$num\.$num\.$num$ ]]; then
		group_of="$1"
	elif [[ ${#1} -le 64 && "$1" =~ ^(v$num\.$num\.$num)(\.[1-9][0-9]*|-[a-z][a-z0-9]*\.$num)$ ]]; then
		group_of="${BASH_REMATCH[1]}.x"
	else
		return 1
	fi
}

# version_group 打印版本对应的分组目录。版本不合法时报错并返回 1。
version_group() {
	if ! version_group_of "$1"; then
		echo "version must be vX.Y.Z, vX.Y.Z.N (N > 0) or vX.Y.Z-<label>.N, at most 64 characters, got: $1" >&2
		return 1
	fi
	echo "$group_of"
}

# imported_paths 抽出 imports.go 里的空白 import，输出 migrations/ 下的相对路径。
# 行尾注释、import 写在同一行都认。
imported_paths() {
	sed -nE 's#^[[:space:]]*(import[[:space:]]+)?_[[:space:]]+"hcm/migrate/migrations/([^"]*)".*$#\2#p' "$1"
}

# write_imports 用标准输入给出的路径重写 imports.go 的 import 声明，
# 声明前后的内容保留。没有 import 块时会新建一块，所以迁移被删光后仍能再追加。
# 只按现有 import 行重写，不扫描目录，漏写的 import 仍由 check-migrate 报出来。
write_imports() {
	local tmp
	tmp="$(mktemp)"
	sort -u >"$tmp.paths"
	awk -v paths="$tmp.paths" '
		BEGIN { state = "head" }
		state == "head" && /^import[[:space:]]*\(/ {
			state = /\)[[:space:]]*$/ ? "tail" : "block"
			next
		}
		state == "head" && /^import[[:space:]]/ { state = "tail"; next }
		state == "block" { if (/^\)/) state = "tail"; next }
		state == "head" { head = head $0 "\n"; next }
		{ tail = tail $0 "\n" }
		END {
			printf "%s", head
			n = 0
			while ((getline line < paths) > 0) {
				if (n++ == 0) print "import ("
				printf "\t_ \"hcm/migrate/migrations/%s\"\n", line
			}
			if (n > 0) print ")"
			printf "%s", tail
		}
	' "$1" >"$tmp"
	mv "$tmp" "$1"
	rm -f "$tmp.paths"
	gofmt -w "$1"
}

# 终端里打勾用绿色、打叉用红色、警告用黄色。写到文件或管道时去掉颜色，避免控制符进日志。
if [[ -t 1 ]]; then
	c_green=$'\033[32m'
	c_red=$'\033[31m'
	c_yellow=$'\033[33m'
	c_bold=$'\033[1m'
	c_dim=$'\033[2m'
	c_reset=$'\033[0m'
else
	c_green=""
	c_red=""
	c_yellow=""
	c_bold=""
	c_dim=""
	c_reset=""
fi

# say_title 打印一块输出的标题。detail 是标题后面的补充，可以空着。
say_title() {
	local title=$1 detail=${2:-}
	printf '\n%s%s%s' "$c_bold" "$title" "$c_reset"
	if [[ -n "$detail" ]]; then
		printf '  %s' "$detail"
	fi
	printf '\n\n'
}

# say_ok 打一行绿勾。后面的每个参数各占一行浅色说明。
say_ok() {
	local title=$1 line
	shift
	printf '  %s✓%s  %s\n' "$c_green" "$c_reset" "$title"
	for line in "$@"; do
		printf '     %s%s%s\n' "$c_dim" "$line" "$c_reset"
	done
	echo
}

# say_warn 打一行黄色感叹号，用于跳过但不失败的项。后面的参数各占一行说明。
say_warn() {
	local title=$1 line
	shift
	printf '  %s!%s  %s\n' "$c_yellow" "$c_reset" "$title"
	for line in "$@"; do
		printf '     %s%s%s\n' "$c_yellow" "$line" "$c_reset"
	done
	echo
}

# say_fail 打一行红叉。后面的参数各占一行原因。
say_fail() {
	local title=$1 line
	shift
	printf '  %s✗%s  %s\n' "$c_red" "$c_reset" "$title"
	for line in "$@"; do
		printf '     %s%s%s\n' "$c_red" "$line" "$c_reset"
	done
	echo
}

# say_done 在清单最后写通过的数量。
say_done() {
	printf '  %s%s%s\n\n' "$c_green" "$1" "$c_reset"
}

# say_stop 在清单最后写失败的结论。
say_stop() {
	printf '  %s%s%s\n\n' "$c_red" "$1" "$c_reset"
}

# replace_in_file 做字面替换，不把参数当正则，所以版本号里的点不会被解释成任意字符。
replace_in_file() {
	local tmp
	tmp="$(mktemp)"
	awk -v old="$2" -v new="$3" '
		{
			out = ""
			while ((i = index($0, old)) > 0) {
				out = out substr($0, 1, i - 1) new
				$0 = substr($0, i + length(old))
			}
			print out $0
		}
	' "$1" >"$tmp"
	mv "$tmp" "$1"
}

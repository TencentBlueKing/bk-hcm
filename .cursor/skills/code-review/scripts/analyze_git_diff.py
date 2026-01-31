#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Git 差异分析脚本

分析指定提交范围的代码变更，支持多种输出格式和配置过滤
"""

import sys
import re
import json
import argparse
import subprocess
import fnmatch
from pathlib import Path
from dataclasses import dataclass, field
from enum import Enum
from typing import List, Dict, Any, Tuple, Optional

# ============================================================================
# 数据模型定义
# ============================================================================

class Language(Enum):
    """编程语言枚举"""
    GO = "go"
    PYTHON = "python"
    JAVA = "java"
    CPP = "cpp"
    C = "c"
    JAVASCRIPT = "javascript"
    TYPESCRIPT = "typescript"
    CSHARP = "csharp"
    SQL = "sql"
    RUST = "rust"
    RUBY = "ruby"
    PHP = "php"
    KOTLIN = "kotlin"
    SWIFT = "swift"
    SHELL = "shell"
    UNKNOWN = "unknown"


@dataclass
class CodeChange:
    """代码变更信息"""
    file_path: str
    language: Language
    added_lines: int = 0
    deleted_lines: int = 0
    hunks: List[Dict[str, Any]] = field(default_factory=list)
    
    @property
    def total_changes(self) -> int:
        return self.added_lines + self.deleted_lines


# ============================================================================
# 常量定义
# ============================================================================

# 文件扩展名到语言的映射
EXTENSION_LANGUAGE_MAP = {
    '.go': Language.GO,
    '.py': Language.PYTHON,
    '.java': Language.JAVA,
    '.kt': Language.KOTLIN,
    '.kts': Language.KOTLIN,
    '.cpp': Language.CPP,
    '.cc': Language.CPP,
    '.cxx': Language.CPP,
    '.hpp': Language.CPP,
    '.hh': Language.CPP,
    '.c': Language.C,
    '.h': Language.C,
    '.js': Language.JAVASCRIPT,
    '.jsx': Language.JAVASCRIPT,
    '.ts': Language.TYPESCRIPT,
    '.tsx': Language.TYPESCRIPT,
    '.cs': Language.CSHARP,
    '.sql': Language.SQL,
    '.rs': Language.RUST,
    '.rb': Language.RUBY,
    '.php': Language.PHP,
    '.swift': Language.SWIFT,
    '.sh': Language.SHELL,
    '.bash': Language.SHELL,
    '.zsh': Language.SHELL,
}

# 代码文件扩展名集合
CODE_EXTENSIONS = {
    '.go', '.py', '.java', '.kt', '.kts',
    '.cpp', '.cc', '.cxx', '.hpp', '.hh', '.c', '.h',
    '.js', '.ts', '.jsx', '.tsx', '.cs', '.sql',
    '.rs', '.rb', '.php', '.swift', '.sh', '.bash', '.zsh'
}

# 二进制文件扩展名（需要排除）
BINARY_EXTENSIONS = {
    '.exe', '.dll', '.so', '.dylib', '.a', '.o', '.obj',
    '.png', '.jpg', '.jpeg', '.gif', '.bmp', '.ico', '.svg',
    '.pdf', '.zip', '.tar', '.gz', '.bz2', '.xz', '.7z',
    '.mp3', '.mp4', '.avi', '.mov', '.wav',
    '.ttf', '.otf', '.woff', '.woff2', '.eot'
}

# 默认排除路径模式
DEFAULT_EXCLUDE_PATTERNS = [
    'vendor/*',
    'node_modules/*',
    '*.pb.go',
    '*.pb.*.go',
    '*_test.go',
    'test/*',
    'tests/*',
    '__pycache__/*',
    '*.pyc',
    '.git/*',
    'dist/*',
    'build/*',
    'target/*',
]


# ============================================================================
# 工具函数
# ============================================================================

def run_command(cmd: List[str], cwd: Optional[str] = None) -> Tuple[int, str, str]:
    """执行命令并返回结果
    
    Args:
        cmd: 命令列表
        cwd: 工作目录
    
    Returns:
        (返回码, stdout, stderr)
    """
    try:
        result = subprocess.run(
            cmd,
            cwd=cwd,
            capture_output=True,
            text=True
        )
        return result.returncode, result.stdout, result.stderr
    except Exception as e:
        return 1, "", str(e)


def get_git_root() -> Optional[str]:
    """获取 Git 仓库根目录
    
    Returns:
        Git 仓库根目录路径，如果不在 Git 仓库中则返回 None
    """
    code, stdout, _ = run_command(['git', 'rev-parse', '--show-toplevel'])
    if code == 0:
        return stdout.strip()
    return None


def detect_language(file_path: str) -> Language:
    """根据文件路径检测编程语言
    
    Args:
        file_path: 文件路径
    
    Returns:
        Language 枚举值
    """
    suffix = Path(file_path).suffix.lower()
    return EXTENSION_LANGUAGE_MAP.get(suffix, Language.UNKNOWN)


def load_codereview_config(git_root: str) -> Dict[str, Any]:
    """加载 .codereview 配置文件
    
    Args:
        git_root: Git 仓库根目录
    
    Returns:
        配置字典
    """
    config_path = Path(git_root) / '.codereview'
    config = {
        'exclude_paths': [],
        'ignore_categories': [],
        'ignore_rules': [],
        'severity': 'normal'
    }
    
    if not config_path.exists():
        return config
    
    try:
        with open(config_path, 'r', encoding='utf-8') as f:
            for line in f:
                line = line.strip()
                if not line or line.startswith('#'):
                    continue
                
                if '=' in line:
                    key, value = line.split('=', 1)
                    key = key.strip()
                    value = value.strip()
                    
                    if key == 'exclude_paths':
                        config['exclude_paths'] = [p.strip() for p in value.split(',')]
                    elif key == 'ignore_categories':
                        config['ignore_categories'] = [c.strip() for c in value.split(',')]
                    elif key == 'ignore_rules':
                        config['ignore_rules'] = [r.strip() for r in value.split(',')]
                    elif key == 'severity':
                        config['severity'] = value
    except Exception as e:
        print(f"警告: 读取 .codereview 配置失败: {e}", file=sys.stderr)
    
    return config


def should_exclude_file(file_path: str, exclude_patterns: List[str]) -> bool:
    """判断文件是否应该被排除
    
    Args:
        file_path: 文件路径
        exclude_patterns: 排除模式列表
    
    Returns:
        是否应该排除
    """
    for pattern in exclude_patterns:
        if fnmatch.fnmatch(file_path, pattern):
            return True
    return False


def is_binary_file(file_path: str) -> bool:
    """判断是否为二进制文件
    
    Args:
        file_path: 文件路径
    
    Returns:
        是否为二进制文件
    """
    suffix = Path(file_path).suffix.lower()
    return suffix in BINARY_EXTENSIONS


# ============================================================================
# 核心业务逻辑
# ============================================================================

def parse_diff_output(diff_output: str, verbose: bool = False) -> List[CodeChange]:
    """解析 git diff 输出
    
    Args:
        diff_output: git diff 命令的输出
        verbose: 是否输出详细信息
    
    Returns:
        CodeChange 对象列表
    """
    changes = []
    current_file = None
    current_hunks = []
    added_lines = 0
    deleted_lines = 0
    
    lines = diff_output.split('\n')
    i = 0
    
    while i < len(lines):
        line = lines[i]
        
        # 新文件开始
        if line.startswith('diff --git'):
            # 保存之前的文件
            if current_file:
                changes.append(CodeChange(
                    file_path=current_file,
                    language=detect_language(current_file),
                    added_lines=added_lines,
                    deleted_lines=deleted_lines,
                    hunks=current_hunks
                ))
                if verbose:
                    print(f"  解析文件: {current_file} (+{added_lines}/-{deleted_lines})", file=sys.stderr)
            
            # 解析文件路径
            match = re.search(r'diff --git a/(.+) b/(.+)', line)
            if match:
                current_file = match.group(2)
            
            current_hunks = []
            added_lines = 0
            deleted_lines = 0
        
        # 解析 hunk 头部
        elif line.startswith('@@'):
            match = re.search(r'@@ -(\d+)(?:,(\d+))? \+(\d+)(?:,(\d+))? @@', line)
            if match:
                old_start = int(match.group(1))
                old_count = int(match.group(2)) if match.group(2) else 1
                new_start = int(match.group(3))
                new_count = int(match.group(4)) if match.group(4) else 1
                
                hunk = {
                    'old_start': old_start,
                    'old_count': old_count,
                    'new_start': new_start,
                    'new_count': new_count,
                    'content': []
                }
                current_hunks.append(hunk)
        
        # 添加的行
        elif line.startswith('+') and not line.startswith('+++'):
            added_lines += 1
            if current_hunks:
                current_hunks[-1]['content'].append({
                    'type': 'add',
                    'content': line[1:]
                })
        
        # 删除的行
        elif line.startswith('-') and not line.startswith('---'):
            deleted_lines += 1
            if current_hunks:
                current_hunks[-1]['content'].append({
                    'type': 'delete',
                    'content': line[1:]
                })
        
        # 上下文行
        elif line.startswith(' '):
            if current_hunks:
                current_hunks[-1]['content'].append({
                    'type': 'context',
                    'content': line[1:]
                })
        
        i += 1
    
    # 保存最后一个文件
    if current_file:
        changes.append(CodeChange(
            file_path=current_file,
            language=detect_language(current_file),
            added_lines=added_lines,
            deleted_lines=deleted_lines,
            hunks=current_hunks
        ))
        if verbose:
            print(f"  解析文件: {current_file} (+{added_lines}/-{deleted_lines})", file=sys.stderr)
    
    return changes


def analyze_git_diff(
    commit_range: str,
    files: Optional[List[str]] = None,
    exclude_patterns: Optional[List[str]] = None,
    verbose: bool = False
) -> List[CodeChange]:
    """分析 Git 差异
    
    Args:
        commit_range: 提交范围，如 "HEAD~3..HEAD"
        files: 可选，指定文件列表
        exclude_patterns: 排除模式列表
        verbose: 是否输出详细信息
    
    Returns:
        CodeChange 对象列表
    """
    git_root = get_git_root()
    if not git_root:
        print("错误: 当前目录不是 Git 仓库", file=sys.stderr)
        return []
    
    if verbose:
        print(f"Git 仓库根目录: {git_root}", file=sys.stderr)
        print(f"分析提交范围: {commit_range}", file=sys.stderr)
    
    # 构建 git diff 命令
    cmd = ['git', 'diff', '--unified=3', commit_range]
    
    if files:
        cmd.append('--')
        cmd.extend(files)
    
    if verbose:
        print(f"执行命令: {' '.join(cmd)}", file=sys.stderr)
    
    code, stdout, stderr = run_command(cmd, cwd=git_root)
    
    if code != 0:
        print(f"错误: git diff 命令失败: {stderr}", file=sys.stderr)
        return []
    
    if not stdout.strip():
        print("没有发现代码变更")
        return []
    
    if verbose:
        print("开始解析 diff 输出...", file=sys.stderr)
    
    changes = parse_diff_output(stdout, verbose)
    
    # 合并排除模式
    all_exclude_patterns = DEFAULT_EXCLUDE_PATTERNS.copy()
    if exclude_patterns:
        all_exclude_patterns.extend(exclude_patterns)
    
    # 过滤文件
    original_count = len(changes)
    changes = [
        c for c in changes
        if (Path(c.file_path).suffix.lower() in CODE_EXTENSIONS and
            not is_binary_file(c.file_path) and
            not should_exclude_file(c.file_path, all_exclude_patterns))
    ]
    
    if verbose and original_count > len(changes):
        print(f"过滤后保留 {len(changes)}/{original_count} 个文件", file=sys.stderr)
    
    return changes


def format_as_json(changes: List[CodeChange]) -> str:
    """格式化为 JSON
    
    Args:
        changes: CodeChange 对象列表
    
    Returns:
        JSON 字符串
    """
    result = []
    for change in changes:
        result.append({
            'file_path': change.file_path,
            'language': change.language.value,
            'added_lines': change.added_lines,
            'deleted_lines': change.deleted_lines,
            'total_changes': change.total_changes,
            'hunk_count': len(change.hunks)
        })
    return json.dumps(result, indent=2, ensure_ascii=False)


def format_as_markdown(changes: List[CodeChange]) -> str:
    """格式化为 Markdown
    
    Args:
        changes: CodeChange 对象列表
    
    Returns:
        Markdown 字符串
    """
    if not changes:
        return "没有代码变更"
    
    # 统计信息
    total_files = len(changes)
    total_added = sum(c.added_lines for c in changes)
    total_deleted = sum(c.deleted_lines for c in changes)
    total_changes = total_added + total_deleted
    
    # 按语言分组统计
    lang_stats = {}
    for change in changes:
        lang = change.language.value
        if lang not in lang_stats:
            lang_stats[lang] = {'files': 0, 'added': 0, 'deleted': 0}
        lang_stats[lang]['files'] += 1
        lang_stats[lang]['added'] += change.added_lines
        lang_stats[lang]['deleted'] += change.deleted_lines
    
    lines = []
    lines.append("## 代码变更统计\n")
    lines.append(f"- **总文件数**: {total_files}")
    lines.append(f"- **新增行数**: +{total_added}")
    lines.append(f"- **删除行数**: -{total_deleted}")
    lines.append(f"- **总变更行数**: {total_changes}\n")
    
    lines.append("### 按语言统计\n")
    lines.append("| 语言 | 文件数 | 新增 | 删除 | 总计 |")
    lines.append("|------|--------|------|------|------|")
    for lang, stats in sorted(lang_stats.items()):
        total = stats['added'] + stats['deleted']
        lines.append(f"| {lang} | {stats['files']} | +{stats['added']} | -{stats['deleted']} | {total} |")
    
    lines.append("\n### 文件变更详情\n")
    lines.append("| 文件路径 | 语言 | 新增 | 删除 | 变更块 |")
    lines.append("|----------|------|------|------|--------|")
    for change in sorted(changes, key=lambda c: c.total_changes, reverse=True):
        lines.append(
            f"| `{change.file_path}` | {change.language.value} | "
            f"+{change.added_lines} | -{change.deleted_lines} | {len(change.hunks)} |"
        )
    
    return '\n'.join(lines)


def print_summary(changes: List[CodeChange]) -> None:
    """打印摘要信息
    
    Args:
        changes: CodeChange 对象列表
    """
    if not changes:
        print("没有代码变更")
        return
    
    total_files = len(changes)
    total_added = sum(c.added_lines for c in changes)
    total_deleted = sum(c.deleted_lines for c in changes)
    
    print(f"\n变更摘要:")
    print(f"  文件数: {total_files}")
    print(f"  新增行: +{total_added}")
    print(f"  删除行: -{total_deleted}")
    print(f"  总变更: {total_added + total_deleted}")


# ============================================================================
# 主函数
# ============================================================================

def main() -> int:
    """主函数"""
    parser = argparse.ArgumentParser(
        description='分析 Git 代码变更',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
示例:
  # 分析最近一次提交
  %(prog)s --range HEAD~1..HEAD
  
  # 分析指定范围
  %(prog)s --range v1.0.0..HEAD
  
  # 分析指定文件
  %(prog)s --range HEAD~3..HEAD --files file1.go file2.py
  
  # 输出为 JSON
  %(prog)s --range HEAD~1..HEAD --output changes.json
  
  # 输出为 Markdown
  %(prog)s --range HEAD~1..HEAD --format markdown --output changes.md
  
  # 详细模式
  %(prog)s --range HEAD~1..HEAD --verbose
        """
    )
    parser.add_argument('--range', default='HEAD~1..HEAD', help='提交范围 (默认: HEAD~1..HEAD)')
    parser.add_argument('--files', nargs='*', help='指定文件列表')
    parser.add_argument('--output', help='输出文件路径')
    parser.add_argument('--format', choices=['json', 'markdown'], default='json', 
                       help='输出格式 (默认: json)')
    parser.add_argument('--verbose', '-v', action='store_true', help='详细输出模式')
    parser.add_argument('--no-config', action='store_true', help='忽略 .codereview 配置文件')
    
    args = parser.parse_args()
    
    # 加载配置
    exclude_patterns = None
    if not args.no_config:
        git_root = get_git_root()
        if git_root:
            config = load_codereview_config(git_root)
            exclude_patterns = config.get('exclude_paths')
            if args.verbose and exclude_patterns:
                print(f"从 .codereview 加载排除模式: {exclude_patterns}", file=sys.stderr)
    
    # 分析变更
    changes = analyze_git_diff(
        args.range,
        args.files,
        exclude_patterns,
        args.verbose
    )
    
    if not changes:
        return 0
    
    # 格式化输出
    if args.format == 'markdown':
        output = format_as_markdown(changes)
    else:
        output = format_as_json(changes)
    
    # 输出结果
    if args.output:
        try:
            with open(args.output, 'w', encoding='utf-8') as f:
                f.write(output)
            print(f"结果已保存到: {args.output}")
            if not args.verbose:
                print_summary(changes)
        except Exception as e:
            print(f"错误: 写入文件失败: {e}", file=sys.stderr)
            return 1
    else:
        print(output)
    
    return 0


if __name__ == '__main__':
    sys.exit(main())

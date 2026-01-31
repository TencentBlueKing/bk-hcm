#!/usr/bin/env python3
"""
同步外部编码标准仓库

从指定的Git仓库克隆或更新编码标准文档到本地references目录
"""

import os
import sys
import subprocess
import argparse
from pathlib import Path

# 标准仓库配置
STANDARDS_REPOS = {
    'sql': {
        'url': 'https://git.woa.com/standards/sql.git',
        'file': 'README.md',
        'target': 'sql/standard.md'
    },
    'csharp': {
        'url': 'https://git.woa.com/standards/csharp.git',
        'file': 'README.md',
        'target': 'csharp/standard.md'
    },
    'protobuf': {
        'url': 'https://git.woa.com/standards/protobuf.git',
        'file': 'README.md',
        'target': 'protobuf/standard.md'
    },
    'lua': {
        'url': 'https://git.woa.com/standards/Lua.git',
        'file': 'README.md',
        'target': 'lua/standard.md'
    },
    'css': {
        'url': 'https://git.woa.com/standards/css.git',
        'file': 'README.md',
        'target': 'css/standard.md'
    }
}


def run_command(cmd, cwd=None, check=True):
    """执行shell命令"""
    try:
        result = subprocess.run(
            cmd,
            shell=True,
            cwd=cwd,
            capture_output=True,
            text=True,
            check=check
        )
        return result.returncode == 0, result.stdout, result.stderr
    except subprocess.CalledProcessError as e:
        return False, e.stdout, e.stderr


def clone_or_update_repo(repo_url, temp_dir):
    """克隆或更新Git仓库"""
    repo_name = repo_url.split('/')[-1].replace('.git', '')
    repo_path = temp_dir / repo_name
    
    if repo_path.exists():
        print(f"更新仓库: {repo_name}")
        success, stdout, stderr = run_command('git pull', cwd=repo_path)
        if not success:
            print(f"⚠️  更新失败: {stderr}")
            return None
    else:
        print(f"克隆仓库: {repo_name}")
        success, stdout, stderr = run_command(
            f'git clone {repo_url}',
            cwd=temp_dir
        )
        if not success:
            print(f"⚠️  克隆失败: {stderr}")
            return None
    
    return repo_path


def sync_standard(lang, config, references_dir, temp_dir):
    """同步单个语言的标准"""
    print(f"\n{'='*60}")
    print(f"同步 {lang.upper()} 标准...")
    print(f"{'='*60}")
    
    # 克隆或更新仓库
    repo_path = clone_or_update_repo(config['url'], temp_dir)
    if not repo_path:
        return False
    
    # 检查源文件是否存在
    source_file = repo_path / config['file']
    if not source_file.exists():
        print(f"⚠️  源文件不存在: {source_file}")
        return False
    
    # 创建目标目录
    target_path = references_dir / config['target']
    target_path.parent.mkdir(parents=True, exist_ok=True)
    
    # 复制文件
    try:
        import shutil
        shutil.copy2(source_file, target_path)
        print(f"✅ 成功同步到: {target_path}")
        return True
    except Exception as e:
        print(f"⚠️  复制失败: {e}")
        return False


def sync_all_standards(languages=None, force=False):
    """同步所有或指定的标准"""
    # 获取脚本所在目录
    script_dir = Path(__file__).parent
    skill_dir = script_dir.parent
    references_dir = skill_dir / 'references' / 'coding-standards'
    temp_dir = skill_dir / '.temp_repos'
    
    # 创建临时目录
    temp_dir.mkdir(exist_ok=True)
    
    # 确定要同步的语言
    langs_to_sync = languages if languages else list(STANDARDS_REPOS.keys())
    
    # 同步每个语言
    results = {}
    for lang in langs_to_sync:
        if lang not in STANDARDS_REPOS:
            print(f"⚠️  未知语言: {lang}")
            results[lang] = False
            continue
        
        config = STANDARDS_REPOS[lang]
        target_path = references_dir / config['target']
        
        # 如果文件已存在且不强制更新，跳过
        if target_path.exists() and not force:
            print(f"⏭️  {lang.upper()} 标准已存在，跳过（使用 --force 强制更新）")
            results[lang] = True
            continue
        
        results[lang] = sync_standard(lang, config, references_dir, temp_dir)
    
    # 打印总结
    print(f"\n{'='*60}")
    print("同步完成")
    print(f"{'='*60}")
    success_count = sum(1 for v in results.values() if v)
    total_count = len(results)
    print(f"成功: {success_count}/{total_count}")
    
    for lang, success in results.items():
        status = "✅" if success else "❌"
        print(f"  {status} {lang.upper()}")
    
    return all(results.values())


def list_standards():
    """列出所有可用的标准"""
    print("可用的编码标准:")
    print(f"{'='*60}")
    for lang, config in STANDARDS_REPOS.items():
        print(f"  • {lang.upper():<10} - {config['url']}")
    print(f"{'='*60}")


def main():
    parser = argparse.ArgumentParser(
        description='同步外部编码标准仓库',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
示例:
  # 同步所有标准
  python3 sync_standards.py --all
  
  # 同步指定语言
  python3 sync_standards.py --languages sql csharp
  
  # 强制更新已存在的标准
  python3 sync_standards.py --all --force
  
  # 列出所有可用标准
  python3 sync_standards.py --list
        """
    )
    
    parser.add_argument(
        '--all',
        action='store_true',
        help='同步所有标准'
    )
    
    parser.add_argument(
        '--languages', '-l',
        nargs='+',
        choices=list(STANDARDS_REPOS.keys()),
        help='指定要同步的语言'
    )
    
    parser.add_argument(
        '--force', '-f',
        action='store_true',
        help='强制更新已存在的标准'
    )
    
    parser.add_argument(
        '--list',
        action='store_true',
        help='列出所有可用的标准'
    )
    
    args = parser.parse_args()
    
    # 列出标准
    if args.list:
        list_standards()
        return 0
    
    # 检查参数
    if not args.all and not args.languages:
        parser.print_help()
        return 1
    
    # 同步标准
    languages = args.languages if args.languages else None
    success = sync_all_standards(languages, args.force)
    
    return 0 if success else 1


if __name__ == '__main__':
    sys.exit(main())

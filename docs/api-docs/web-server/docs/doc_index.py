#!/usr/bin/env python3.11

from pathlib import Path
import argparse
import sys

parser = argparse.ArgumentParser(
    'docindex', description="a simple script for gather document index to markdown  readme file.")
parser.add_argument('-s', '--source_dir', default='.', required=False)
parser.add_argument('-t', '--target', default='readme.md', required=False)

args = parser.parse_args()
source_dir = args.source_dir
target_file = args.target


class DocInfo:
    version: str
    permission: str
    decsription: str
    method: str
    url: str
    filename: str

    def __init__(self, version: str = '', permission: str = '', decsription: str = '', method: str = '', url: str = ''):
        self.version = version
        self.permission = permission
        self.decsription = decsription
        self.method = method
        self.url = url
        return

    def __str__(self) -> str:
        return f'|{self.method}|{self.url}|{self.decsription}|[{self.filename}]({self.filename})|'

    def __repr__(self) -> str:
        return str(self)


methods = ('GET', 'POST', 'PATCH', 'PUT', 'DELETE', 'HEAD', 'OPTION')


def extract(doc_file: Path) -> DocInfo:
    info = DocInfo()
    info.filename = str(doc_file)
    with doc_file.open() as f:
        for line in f:
            if line.startswith("- 该接口提供版本："):
                info.version = line[10:].replace('。', '')
            elif line.startswith("- 该接口所需权限："):
                info.permission = line[10:].replace('。', '')
            elif line.startswith("- 该接口功能描述："):
                info.decsription = line[10:].replace('。', '').strip()
            else:
                for method in methods:
                    if line.startswith(method):
                        info.method = method
                        info.url = line[len(method):].strip()
                        break
            if info.method != '':
                return info
    if info.url == '':
        print("url not found!", doc_file, file=sys.stderr)
    return info


def extract_all(source_dir: Path) -> list[DocInfo]:
    infolist = list()
    for p in source_dir.glob('**/*.md'):
        if str(p).lower().endswith('readme.md'):
            continue
        info = extract(p)
        if info != None:
            infolist.append(info)

    return infolist


def main():
    s = Path(source_dir)
    t = Path(target_file)
    if not s.exists() or not s.is_dir():
        print(f"source dir({source_dir}) not correct", file=sys.stderr)
        exit(1)
    out = t.open(mode='w', encoding='utf-8')

    # 按文件夹分类
    for subdir in s.iterdir():
        if not subdir.is_dir():
            continue
        print(f"## [{subdir}]({subdir})\n", file=out)
        print('| 方法| URL| 描述| 文档|', file=out)
        print('|--------|--------|-----------------|--------|', file=out)
        infos = extract_all(subdir)
        infos.sort(key=lambda x: x.url)
        for info in infos:
            print(info, file=out)
    out.close()


main()

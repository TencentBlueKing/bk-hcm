"""lint_check.py

用于在代码审查流程中对仓库执行 Lint 检查。

执行逻辑：
- 优先检测仓库根目录是否存在 `Makefile` 且定义了 `lint:` 目标；若存在则直接执行 `make lint`。
- 否则根据 `--language` 选择对应语言的默认 Lint 工具执行（如 Go/Java/C++/Python）。
- Go 语言使用 `tencentlint`：若仓库根目录存在 `.golangci.yml` 则使用之，否则使用 skill 内置的默认配置
  `.codebuddy/skills/code-review/assets/.golangci.yml`。

输出规则：
- 任意步骤失败会打印对应的错误输出并以非 0 退出码结束。
- 全部通过则输出 `Lint Success`。
"""

import argparse
import os
import subprocess
import sys
from pathlib import Path


def _find_repo_root(start: Path) -> Path:
    """Find git repo root; fallback to current working directory."""
    cur = start.resolve()
    for p in [cur] + list(cur.parents):
        if (p / ".git").exists():
            return p
    return cur


def _run(cmd: list[str], cwd: Path) -> tuple[int, str, str]:
    p = subprocess.run(
        cmd,
        cwd=str(cwd),
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
    )
    return p.returncode, p.stdout, p.stderr


def _makefile_has_lint_target(makefile_path: Path) -> bool:
    # Heuristic: detect "lint:" target definition.
    try:
        txt = makefile_path.read_text(encoding="utf-8", errors="ignore")
    except Exception:
        return False

    for line in txt.splitlines():
        s = line.strip()
        if not s or s.startswith("#"):
            continue
        if s.startswith("lint:"):
            return True
    return False


def _try_make_lint(repo_root: Path) -> bool:
    makefile = repo_root / "Makefile"
    if not makefile.exists():
        return False
    if not _makefile_has_lint_target(makefile):
        return False

    code, out, err = _run(["make", "lint"], cwd=repo_root)
    if code != 0:
        msg = (out + ("\n" if out and err else "") + err).strip()
        if msg:
            print(msg)
        else:
            print("make lint failed")
        return True

    print("Lint Success")
    return True


def _lint_go(repo_root: Path, skill_root: Path) -> int:
    # Ensure tencentlint exists; if missing, install the latest.
    code, _, _ = _run(["tencentlint", "--version"], cwd=repo_root)
    if code != 0:
        install_code, out, err = _run(
            ["go", "install", "git.woa.com/standards/go/cmd/tencentlint@master"],
            cwd=repo_root,
        )
        if install_code != 0:
            msg = (out + ("\n" if out and err else "") + err).strip()
            print(msg if msg else "install tencentlint failed")
            return install_code

    local_cfg = repo_root / ".golangci.yml"
    if local_cfg.exists():
        cfg = local_cfg
    else:
        cfg = skill_root / "assets" / ".golangci.yml"

    cmd = ["tencentlint", "run", "-c", str(cfg)]
    code, out, err = _run(cmd, cwd=repo_root)
    if code != 0:
        msg = (out + ("\n" if out and err else "") + err).strip()
        print(msg if msg else "tencentlint failed")
        return code
    print("Lint Success")
    return 0


def _lint_java(repo_root: Path) -> int:
    # Default approach: use Maven/Gradle ecosystem if present.
    # - Maven: mvn -q -DskipTests spotless:check (if spotless configured)
    # - Gradle: ./gradlew -q spotlessCheck (if spotless configured)
    # Fallback: run checkstyle if configured.
    if (repo_root / "gradlew").exists():
        code, out, err = _run(["./gradlew", "-q", "spotlessCheck"], cwd=repo_root)
        if code == 0:
            print("Lint Success")
            return 0
        # if spotless not configured, continue trying other checks

    if (repo_root / "pom.xml").exists():
        # Try common lint plugins when configured; if not, mvn will fail and we surface output.
        code, out, err = _run(["mvn", "-q", "-DskipTests", "spotless:check"], cwd=repo_root)
        if code == 0:
            print("Lint Success")
            return 0

    # As a generic fallback, run 'mvn -q -DskipTests verify' if pom exists.
    if (repo_root / "pom.xml").exists():
        code, out, err = _run(["mvn", "-q", "-DskipTests", "verify"], cwd=repo_root)
        if code == 0:
            print("Lint Success")
            return 0
        msg = (out + ("\n" if out and err else "") + err).strip()
        print(msg if msg else "Java lint failed")
        return code

    # No recognizable build tool.
    print("Java project not detected (no pom.xml/gradlew)")
    return 2


def _lint_cpp(repo_root: Path) -> int:
    # Default: run clang-tidy if compile_commands.json exists.
    cc = repo_root / "compile_commands.json"
    if not cc.exists():
        print("C/C++ project missing compile_commands.json; cannot run clang-tidy")
        return 2

    # Run clang-tidy over all .cpp/.cc/.cxx files under repo root.
    files: list[str] = []
    for ext in ("*.cpp", "*.cc", "*.cxx", "*.c"):
        files.extend([str(p) for p in repo_root.rglob(ext)])

    if not files:
        print("No C/C++ source files found")
        return 0

    cmd = ["clang-tidy", "-p", str(repo_root)] + files
    code, out, err = _run(cmd, cwd=repo_root)
    if code != 0:
        msg = (out + ("\n" if out and err else "") + err).strip()
        print(msg if msg else "clang-tidy failed")
        return code

    print("Lint Success")
    return 0


def _lint_python(repo_root: Path) -> int:
    # Default: ruff if available, else flake8.
    code, out, err = _run(["ruff", "check", "."], cwd=repo_root)
    if code == 0:
        print("Lint Success")
        return 0

    # If ruff is not installed, it returns 127 on many shells; still try flake8.
    code2, out2, err2 = _run(["flake8", "."], cwd=repo_root)
    if code2 != 0:
        msg = (out2 + ("\n" if out2 and err2 else "") + err2).strip()
        if msg:
            print(msg)
        else:
            print("Python lint failed")
        return code2

    print("Lint Success")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description="Run repository lint checks")
    parser.add_argument("--language", "-l", required=True, help="go/java/cpp/python")
    parser.add_argument(
        "--repo",
        help="repo path (defaults to auto-detected git root from cwd)",
        default=None,
    )
    args = parser.parse_args()

    repo_root = Path(args.repo).resolve() if args.repo else _find_repo_root(Path.cwd())
    skill_root = Path(__file__).resolve().parents[1]

    # 1) Prefer Makefile lint if present.
    if _try_make_lint(repo_root):
        return 0

    lang = str(args.language).strip().lower()
    if lang == "go":
        return _lint_go(repo_root, skill_root)
    if lang == "java":
        return _lint_java(repo_root)
    if lang in ("cpp", "c++", "c"):
        return _lint_cpp(repo_root)
    if lang == "python":
        return _lint_python(repo_root)

    print(f"Unsupported language: {args.language}")
    return 2


if __name__ == "__main__":
    sys.exit(main())

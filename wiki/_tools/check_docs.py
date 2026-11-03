"""检查 wiki 链接和围栏，并在临时模块中编译每个 Go 示例。"""

import argparse
import json
import os
from pathlib import Path
import re
import subprocess
import tempfile
from urllib.parse import unquote, urlsplit


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--format", action="store_true", help="使用 gofmt 格式化 Markdown 中的 Go 示例")
    parser.add_argument("--links-only", action="store_true", help="只检查链接、围栏和示例完整性")
    args = parser.parse_args()
    root = Path(__file__).resolve().parents[2]
    wiki = root / "wiki"
    pages = sorted(wiki.rglob("*.md"))
    errors = []
    samples = []
    for page in pages:
        source = page.read_text(encoding="utf-8-sig")
        if len(re.findall(r"^# [^\n]+", source, re.M)) != 1:
            errors.append(f"{page.relative_to(root)}: 必须有且仅有一个一级标题")
        if len(re.findall(r"^```", source, re.M)) % 2:
            errors.append(f"{page.relative_to(root)}: 代码围栏未闭合")
        # 先去掉代码块，避免把 JSON、数组和代码示意误当作链接。
        prose = re.sub(r"^```[^\n]*\n.*?^```[ \t]*$", "", source, flags=re.M | re.S)
        for target in re.findall(r"!?\[[^\]]*\]\(([^)]+)\)", prose):
            target = target.strip().strip("<>")
            parts = urlsplit(target)
            if parts.scheme or target.startswith("#"):
                continue
            destination = (page.parent / unquote(parts.path)).resolve()
            if not destination.exists():
                errors.append(f"{page.relative_to(root)}: 链接不存在 {target}")
        index = 0

        def sample(match):
            nonlocal index
            index += 1
            code = match[1]
            if not re.match(r"\s*package\s+\w+", code):
                errors.append(f"{page.relative_to(root)}: Go 示例 {index} 不是完整编译单元")
            if re.search(r'"github.com/ProgramCX/GoQQBot/internal/', code):
                errors.append(f"{page.relative_to(root)}: Go 示例 {index} 导入了 internal")
            if args.format:
                result = subprocess.run(["gofmt"], input=code, text=True, encoding="utf-8", capture_output=True)
                if result.returncode:
                    errors.append(f"{page.relative_to(root)}: 示例 {index} gofmt 失败\n{result.stderr}")
                else:
                    code = result.stdout
            samples.append((str(page.relative_to(root)), index, code))
            return "```go\n" + code.rstrip() + "\n```"

        updated = re.sub(r"^```go[ \t]*\n(.*?)^```[ \t]*$", sample, source, flags=re.M | re.S)
        if args.format and updated != source:
            page.write_text(updated, encoding="utf-8")
    if errors:
        raise SystemExit("\n".join(errors))
    print(f"链接与围栏检查通过：{len(pages)} 页，{len(samples)} 个 Go 示例。", flush=True)
    if args.links_only:
        return
    # 每个代码块使用独立包，防止同名演示函数相互影响。
    with tempfile.TemporaryDirectory(prefix="goqqbot-wiki-") as directory:
        project = Path(directory)
        module = (
            "module wiki-examples\n\ngo 1.24.4\n\n"
            "require github.com/ProgramCX/GoQQBot v0.0.0\n\n"
            "replace github.com/ProgramCX/GoQQBot => " + json.dumps(root.as_posix()) + "\n"
        )
        (project / "go.mod").write_text(module, encoding="utf-8")
        for number, (page, index, code) in enumerate(samples, 1):
            path = project / f"sample{number:03}"
            path.mkdir()
            filename = "snippet_test.go" if re.search(r"func (Test|Example)\w*\(", code) else "snippet.go"
            (path / filename).write_text(code, encoding="utf-8")
            print(f"sample{number:03}: {page} / Go {index}", flush=True)
        env = dict(os.environ)
        env.setdefault("GOMAXPROCS", "2")
        result = subprocess.run(
            ["go", "test", "-mod=mod", "-p", "1", "-run", "^$", "./..."],
            cwd=project, env=env,
        )
        if result.returncode:
            raise SystemExit(result.returncode)
    print("所有 Go 示例编译通过；没有运行机器人连接或发送请求。")


if __name__ == "__main__":
    main()

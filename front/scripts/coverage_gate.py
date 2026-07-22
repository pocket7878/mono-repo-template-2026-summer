#!/usr/bin/env python3
"""Enforce scoped frontend coverage and stricter coverage for changed high-risk files."""

from __future__ import annotations

import fnmatch
import json
import sys
from pathlib import Path


FRONT_DIR = Path(__file__).resolve().parents[1]
SUMMARY_PATH = FRONT_DIR / "coverage" / "coverage-summary.json"
EXCLUDED = (
    "app/lib/backend-api/openapi/**",
    "**/*.test.ts",
    "**/*.test.tsx",
    "**/mocks/**",
)
HIGH_RISK = ("app/routes/**", "app/features/*/api/server.ts")


def matches(path: str, patterns: tuple[str, ...]) -> bool:
    return any(fnmatch.fnmatch(path, pattern) for pattern in patterns)


def main() -> int:
    if not SUMMARY_PATH.exists():
        print("coverage summary is missing; run npm run coverage first", file=sys.stderr)
        return 1
    summary = json.loads(SUMMARY_PATH.read_text())
    covered = total = 0
    for raw_path, metrics in summary.items():
        if raw_path == "total":
            continue
        path = Path(raw_path).resolve().relative_to(FRONT_DIR).as_posix()
        if matches(path, EXCLUDED):
            continue
        lines = metrics["lines"]
        covered += lines["covered"]
        total += lines["total"]
    percentage = 0 if total == 0 else round(covered * 100 / total, 1)
    print(f"frontend scoped coverage: {percentage:.1f}% ({covered}/{total} lines)")
    if percentage < 60.0:
        print("coverage gate failed: scoped coverage is below 60.0%", file=sys.stderr)
        return 1
    changed_files = [Path(argument).as_posix().removeprefix("front/") for argument in sys.argv[1:]]
    failing = []
    for path in changed_files:
        if not matches(path, HIGH_RISK) or matches(path, EXCLUDED):
            continue
        metrics = summary.get(str((FRONT_DIR / path).resolve()))
        percentage_for_file = 0 if metrics is None else metrics["lines"]["pct"]
        if percentage_for_file < 80.0:
            failing.append(f"{path}: {percentage_for_file:.1f}%")
    if failing:
        print("coverage gate failed for high-risk files: " + ", ".join(failing), file=sys.stderr)
        return 1
    print("frontend coverage gate passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

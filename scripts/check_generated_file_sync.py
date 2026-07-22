#!/usr/bin/env python3
"""Regenerate API artifacts and fail when their committed contents drift."""

from __future__ import annotations

import subprocess
import sys


def main() -> int:
    subprocess.run(["make", "gen"], check=True)
    result = subprocess.run(
        [
            "git",
            "diff",
            "--exit-code",
            "--",
            "shared/backend_api/openapi",
            "backend/internal/handler/api/api.gen.go",
            "front/app/lib/backend-api/openapi/types.ts",
        ],
        check=False,
    )
    if result.returncode:
        print("generated API artifacts are out of sync; run 'make gen' and commit the result.", file=sys.stderr)
    return result.returncode


if __name__ == "__main__":
    sys.exit(main())

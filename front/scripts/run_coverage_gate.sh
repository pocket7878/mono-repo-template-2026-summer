#!/usr/bin/env bash
set -euo pipefail

npm run coverage
python3 ./scripts/coverage_gate.py "$@"

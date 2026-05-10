#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"
COMPOSE="sed-service/e2e_tests/docker-compose.test.yml"
docker compose -f "$COMPOSE" up --build -d
cleanup() {
  docker compose -f "$COMPOSE" down -v
}
trap cleanup EXIT
python3 -m venv .venv-sed-e2e
# shellcheck disable=SC1091
source .venv-sed-e2e/bin/activate
pip install -r sed-service/e2e_tests/requirements.txt
pytest -q sed-service/e2e_tests/tests

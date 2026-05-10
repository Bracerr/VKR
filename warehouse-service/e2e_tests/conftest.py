import os

import pytest

BASE_URL = os.getenv("WAREHOUSE_URL", "http://localhost:8090").rstrip("/")


@pytest.fixture
def base_url():
    return BASE_URL

import requests


def test_production_health(prod_api):
    r = requests.get(f"{prod_api}/health", timeout=10)
    assert r.status_code == 200


def test_production_ready(prod_api):
    r = requests.get(f"{prod_api}/ready", timeout=10)
    assert r.status_code == 200

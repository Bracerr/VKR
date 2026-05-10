import requests


def test_procurement_health(proc_api):
    r = requests.get(f"{proc_api}/health", timeout=10)
    assert r.status_code == 200


def test_procurement_ready(proc_api):
    r = requests.get(f"{proc_api}/ready", timeout=10)
    assert r.status_code == 200


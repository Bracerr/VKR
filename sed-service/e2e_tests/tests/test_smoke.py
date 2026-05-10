import requests


def test_sed_health(sed_api):
    r = requests.get(f"{sed_api}/health", timeout=10)
    assert r.status_code == 200


def test_warehouse_health(wh_api):
    r = requests.get(f"{wh_api}/health", timeout=10)
    assert r.status_code == 200

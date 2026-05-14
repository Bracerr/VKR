import requests


def test_sales_health(sales_api):
    r = requests.get(f"{sales_api}/health", timeout=10)
    assert r.status_code == 200


def test_sales_ready(sales_api):
    r = requests.get(f"{sales_api}/ready", timeout=10)
    assert r.status_code == 200


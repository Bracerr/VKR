import requests


def test_health(base_url: str):
    r = requests.get(f"{base_url}/health", timeout=5)
    assert r.status_code == 200
    assert r.json().get("status") == "ok"

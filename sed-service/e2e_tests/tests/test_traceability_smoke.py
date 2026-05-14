import requests


def test_trace_health(trace_api):
    r = requests.get(f"{trace_api}/health", timeout=10)
    assert r.status_code == 200


def test_trace_ready(trace_api):
    r = requests.get(f"{trace_api}/ready", timeout=10)
    assert r.status_code == 200


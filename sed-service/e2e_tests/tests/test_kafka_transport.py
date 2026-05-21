"""Smoke: стенд с EVENT_TRANSPORT=kafka поднимается и health OK."""

import os

import requests


def test_event_transport_kafka_default():
    assert os.environ.get("EVENT_TRANSPORT", "kafka").lower() in ("kafka", "dual")


def test_services_health(auth_api, wh_api, sed_api, trace_api):
    for url in (f"{auth_api}/ready", f"{wh_api}/health", f"{sed_api}/health", f"{trace_api}/health"):
        r = requests.get(url, timeout=10)
        assert r.status_code == 200, url + " " + r.text

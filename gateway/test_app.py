import pytest
from app import app, queues


@pytest.fixture
def client():
    app.config["TESTING"] = True
    queues.clear()
    with app.test_client() as client:
        yield client


def test_health(client):
    resp = client.get("/health")
    assert resp.status_code == 200
    data = resp.get_json()
    assert data["status"] == "healthy"
    assert data["service"] == "gateway"


def test_create_queue(client):
    resp = client.post("/queues", json={"name": "test-queue"})
    assert resp.status_code == 201
    data = resp.get_json()
    assert data["name"] == "test-queue"
    assert data["message_count"] == 0


def test_create_queue_missing_name(client):
    resp = client.post("/queues", json={})
    assert resp.status_code == 400


def test_create_queue_duplicate(client):
    client.post("/queues", json={"name": "dup"})
    resp = client.post("/queues", json={"name": "dup"})
    assert resp.status_code == 409


def test_list_queues(client):
    client.post("/queues", json={"name": "q1"})
    client.post("/queues", json={"name": "q2"})
    resp = client.get("/queues")
    assert resp.status_code == 200
    data = resp.get_json()
    assert len(data) == 2


def test_publish_message(client):
    client.post("/queues", json={"name": "mq"})
    resp = client.post("/queues/mq/messages", json={"body": "hello"})
    assert resp.status_code == 201
    data = resp.get_json()
    assert data["body"] == "hello"
    assert data["status"] == "pending"


def test_publish_message_queue_not_found(client):
    resp = client.post("/queues/nope/messages", json={"body": "hello"})
    assert resp.status_code == 404


def test_publish_message_missing_body(client):
    client.post("/queues", json={"name": "mq2"})
    resp = client.post("/queues/mq2/messages", json={})
    assert resp.status_code == 400


def test_get_messages(client):
    client.post("/queues", json={"name": "mq3"})
    client.post("/queues/mq3/messages", json={"body": "msg1"})
    resp = client.get("/queues/mq3/messages")
    assert resp.status_code == 200
    data = resp.get_json()
    assert len(data) == 1
    assert data[0]["body"] == "msg1"


def test_queue_stats(client):
    client.post("/queues", json={"name": "sq"})
    client.post("/queues/sq/messages", json={"body": "a"})
    client.post("/queues/sq/messages", json={"body": "b"})
    resp = client.get("/queues/sq/stats")
    assert resp.status_code == 200
    data = resp.get_json()
    assert data["total"] == 2
    assert data["pending"] == 2
    assert data["processed"] == 0

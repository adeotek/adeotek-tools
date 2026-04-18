from fastapi.testclient import TestClient

from app.main import create_app


def test_app_factory_returns_fastapi_instance():
    app = create_app()
    assert app.title == "homelab-chatbot"


def test_health_endpoint_exists():
    app = create_app()
    client = TestClient(app)
    response = client.get("/api/health")
    assert response.status_code == 200
    assert response.json() == {"status": "ok"}

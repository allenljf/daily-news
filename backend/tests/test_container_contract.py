"""Go container and Cloud Run command contracts."""

from __future__ import annotations

from pathlib import Path

BACKEND_ROOT = Path(__file__).resolve().parents[1]
REPOSITORY_ROOT = BACKEND_ROOT.parent


def test_dockerfile_builds_non_root_go_api_and_job_binaries() -> None:
    dockerfile = (BACKEND_ROOT / "Dockerfile").read_text()

    assert "FROM golang:" in dockerfile
    assert "go build" in dockerfile
    assert "USER daily-news" in dockerfile
    assert 'CMD ["/app/api"]' in dockerfile
    assert "python" not in dockerfile.lower()


def test_cloud_run_job_executes_go_job_binary() -> None:
    manifest = (REPOSITORY_ROOT / "infra/cloud-run/job.yaml").read_text()

    assert "command:" in manifest
    assert "- /app/daily-news-job" in manifest
    assert "scripts/run-job.sh" not in manifest


def test_cloud_run_service_declares_public_invocation_for_flutter_auth() -> None:
    manifest = (REPOSITORY_ROOT / "infra/cloud-run/service.yaml").read_text()

    assert "run.googleapis.com/invoker-iam-disabled: 'true'" in manifest


def test_ci_runs_go_verification_instead_of_python_backend_checks() -> None:
    workflow = (REPOSITORY_ROOT / ".github/workflows/ci.yml").read_text()

    assert "go vet ./..." in workflow
    assert "go test ./..." in workflow
    assert "setup-go" in workflow
    assert "setup-uv" not in workflow

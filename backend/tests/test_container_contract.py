"""Container entrypoint contracts."""

from __future__ import annotations

import os
import subprocess
from pathlib import Path

BACKEND_ROOT = Path(__file__).resolve().parents[1]


def _write_command_recorder(bin_dir: Path, command: str) -> None:
    recorder = bin_dir / command
    recorder.write_text(
        "#!/bin/sh\n"
        "printf '%s\\n' \"$@\" > \"$COMMAND_CAPTURE\"\n",
    )
    recorder.chmod(0o755)


def _run_entrypoint(
    tmp_path: Path,
    script_name: str,
    command: str,
    *,
    arguments: list[str] | None = None,
    environment: dict[str, str] | None = None,
) -> list[str]:
    bin_dir = tmp_path / "bin"
    bin_dir.mkdir()
    _write_command_recorder(bin_dir, command)
    capture = tmp_path / "command.txt"

    completed = subprocess.run(
        [str(BACKEND_ROOT / "scripts" / script_name), *(arguments or [])],
        check=True,
        cwd=BACKEND_ROOT,
        env={
            **os.environ,
            "PATH": f"{bin_dir}:{os.environ['PATH']}",
            "COMMAND_CAPTURE": str(capture),
            **(environment or {}),
        },
        text=True,
        capture_output=True,
    )

    assert completed.stdout == ""
    assert completed.stderr == ""
    return capture.read_text().splitlines()


def test_service_entrypoint_starts_uvicorn_on_cloud_run_port(tmp_path: Path) -> None:
    """Removing the API module, public bind address, or PORT forwarding breaks this contract."""
    assert _run_entrypoint(
        tmp_path,
        "serve.sh",
        "uvicorn",
        environment={"PORT": "9090"},
    ) == ["app.main:app", "--host", "0.0.0.0", "--port", "9090"]


def test_job_entrypoint_runs_daily_news_module_and_forwards_arguments(tmp_path: Path) -> None:
    """Removing the Job module or argument forwarding breaks a Cloud Run Job execution."""
    assert _run_entrypoint(
        tmp_path,
        "run-job.sh",
        "python",
        arguments=["--run-id", "00000000-0000-0000-0000-000000000001"],
    ) == [
        "-m",
        "app.jobs.daily_news",
        "--run-id",
        "00000000-0000-0000-0000-000000000001",
    ]

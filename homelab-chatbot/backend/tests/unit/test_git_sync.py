import subprocess
from pathlib import Path

import pytest

from app.config import RepoConfig
from app.ingestion.git_sync import GitSync, SyncResult


def _init_remote_repo(tmp: Path) -> Path:
    remote = tmp / "remote.git"
    work = tmp / "work"
    work.mkdir()
    subprocess.run(["git", "init", "--bare", str(remote)], check=True)
    subprocess.run(["git", "init", str(work)], check=True)
    subprocess.run(["git", "-C", str(work), "config", "user.email", "t@t"], check=True)
    subprocess.run(["git", "-C", str(work), "config", "user.name", "t"], check=True)
    subprocess.run(["git", "-C", str(work), "config", "commit.gpgsign", "false"], check=True)
    (work / "README.md").write_text("# hello\n")
    subprocess.run(["git", "-C", str(work), "add", "."], check=True)
    subprocess.run(["git", "-C", str(work), "commit", "-m", "init"], check=True)
    subprocess.run(["git", "-C", str(work), "branch", "-M", "main"], check=True)
    subprocess.run(
        ["git", "-C", str(work), "remote", "add", "origin", str(remote)], check=True
    )
    subprocess.run(["git", "-C", str(work), "push", "-u", "origin", "main"], check=True)
    return remote


@pytest.fixture
def local_repo(tmp_path: Path) -> Path:
    return _init_remote_repo(tmp_path)


def test_initial_clone_when_missing(tmp_path: Path, local_repo: Path):
    cfg = RepoConfig(
        name="r", url=f"file://{local_repo}", branch="main", token_env="UNUSED"
    )
    sync = GitSync(clone_root=tmp_path / "repos", get_token=lambda _: None)
    result = sync.sync(cfg)
    assert result.cloned is True
    assert result.changed_files == []
    assert (tmp_path / "repos" / "r" / "README.md").exists()


def test_pull_detects_changed_files(tmp_path: Path, local_repo: Path):
    cfg = RepoConfig(
        name="r", url=f"file://{local_repo}", branch="main", token_env="UNUSED"
    )
    sync = GitSync(clone_root=tmp_path / "repos", get_token=lambda _: None)
    sync.sync(cfg)  # initial clone

    work_dir = tmp_path / "upstream_work"
    subprocess.run(
        ["git", "clone", f"file://{local_repo}", str(work_dir)], check=True
    )
    subprocess.run(["git", "-C", str(work_dir), "config", "user.email", "t@t"], check=True)
    subprocess.run(["git", "-C", str(work_dir), "config", "user.name", "t"], check=True)
    subprocess.run(["git", "-C", str(work_dir), "config", "commit.gpgsign", "false"], check=True)
    (work_dir / "docs.md").write_text("# docs\n")
    subprocess.run(["git", "-C", str(work_dir), "add", "."], check=True)
    subprocess.run(["git", "-C", str(work_dir), "commit", "-m", "add"], check=True)
    subprocess.run(["git", "-C", str(work_dir), "push"], check=True)

    result = sync.sync(cfg)
    assert result.cloned is False
    names = [c.path for c in result.changed_files]
    assert "docs.md" in names


def test_no_change_returns_empty(tmp_path: Path, local_repo: Path):
    cfg = RepoConfig(
        name="r", url=f"file://{local_repo}", branch="main", token_env="UNUSED"
    )
    sync = GitSync(clone_root=tmp_path / "repos", get_token=lambda _: None)
    sync.sync(cfg)
    result = sync.sync(cfg)
    assert result.cloned is False
    assert result.changed_files == []


def test_include_globs_filter(tmp_path: Path, local_repo: Path):
    cfg = RepoConfig(
        name="r",
        url=f"file://{local_repo}",
        branch="main",
        token_env="UNUSED",
        include_globs=["**/*.xlsx"],
    )
    sync = GitSync(clone_root=tmp_path / "repos", get_token=lambda _: None)
    result = sync.sync(cfg)
    assert result.matched_files == []  # no .xlsx in repo


def test_token_injected_into_https_url(tmp_path: Path):
    cfg = RepoConfig(
        name="r",
        url="https://github.com/user/repo.git",
        branch="main",
        token_env="T",
    )
    sync = GitSync(clone_root=tmp_path / "repos", get_token=lambda v: "tok123" if v == "T" else None)
    assert (
        sync._auth_url(cfg)
        == "https://x-access-token:tok123@github.com/user/repo.git"
    )

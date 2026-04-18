"""Clone/pull configured repos and report changed files."""

import fnmatch
import subprocess
from dataclasses import dataclass, field
from pathlib import Path
from typing import Callable, Literal
from urllib.parse import urlparse, urlunparse

from app.config import RepoConfig

ChangeStatus = Literal["A", "M", "D", "R"]


@dataclass
class ChangedFile:
    path: str
    status: ChangeStatus


@dataclass
class SyncResult:
    repo: str
    cloned: bool
    old_sha: str | None
    new_sha: str
    changed_files: list[ChangedFile] = field(default_factory=list)
    matched_files: list[ChangedFile] = field(default_factory=list)


class GitSync:
    """Performs clone/pull operations and reports changes matching configured globs."""

    def __init__(
        self,
        clone_root: Path,
        get_token: Callable[[str], str | None],
    ) -> None:
        self._root = Path(clone_root)
        self._get_token = get_token
        self._root.mkdir(parents=True, exist_ok=True)

    def repo_path(self, cfg: RepoConfig) -> Path:
        return self._root / cfg.name

    def _auth_url(self, cfg: RepoConfig) -> str:
        token = self._get_token(cfg.token_env)
        if not token:
            return cfg.url
        parsed = urlparse(cfg.url)
        if parsed.scheme not in ("http", "https"):
            return cfg.url
        netloc = f"x-access-token:{token}@{parsed.hostname}"
        if parsed.port:
            netloc += f":{parsed.port}"
        return urlunparse(parsed._replace(netloc=netloc))

    def _git(self, *args: str, cwd: Path | None = None) -> str:
        result = subprocess.run(
            ["git", *args],
            cwd=str(cwd) if cwd else None,
            check=True,
            capture_output=True,
            text=True,
        )
        return result.stdout

    def _head_sha(self, path: Path) -> str:
        return self._git("rev-parse", "HEAD", cwd=path).strip()

    @staticmethod
    def _glob_match(rel: str, glob: str) -> bool:
        """Return True if rel matches glob, handling ** patterns."""
        if fnmatch.fnmatch(rel, glob):
            return True
        # For patterns like **/*.md, also match files in the repo root
        if glob.startswith("**/"):
            tail = glob[3:]
            if fnmatch.fnmatch(rel, tail):
                return True
        return False

    def _match(self, paths: list[ChangedFile], globs: list[str]) -> list[ChangedFile]:
        out = []
        for cf in paths:
            if any(self._glob_match(cf.path, g) for g in globs):
                out.append(cf)
        return out

    def sync(self, cfg: RepoConfig) -> SyncResult:
        path = self.repo_path(cfg)
        url = self._auth_url(cfg)

        if not (path / ".git").exists():
            if path.exists():
                raise RuntimeError(f"{path} exists but is not a git repo")
            self._git("clone", "--branch", cfg.branch, url, str(path))
            new_sha = self._head_sha(path)
            matched = self._list_matching(path, cfg.include_globs)
            return SyncResult(
                repo=cfg.name,
                cloned=True,
                old_sha=None,
                new_sha=new_sha,
                changed_files=[],
                matched_files=matched,
            )

        old_sha = self._head_sha(path)
        self._git("fetch", "origin", cfg.branch, cwd=path)
        self._git("reset", "--hard", f"origin/{cfg.branch}", cwd=path)
        new_sha = self._head_sha(path)

        changed: list[ChangedFile] = []
        if old_sha != new_sha:
            diff = self._git("diff", "--name-status", f"{old_sha}..{new_sha}", cwd=path)
            for line in diff.strip().splitlines():
                parts = line.split("\t")
                if len(parts) >= 2:
                    status_char = parts[0][0]
                    if status_char in ("A", "M", "D", "R"):
                        changed.append(ChangedFile(path=parts[-1], status=status_char))

        matched = self._match(changed, cfg.include_globs)
        return SyncResult(
            repo=cfg.name,
            cloned=False,
            old_sha=old_sha,
            new_sha=new_sha,
            changed_files=changed,
            matched_files=matched,
        )

    def _list_matching(self, repo_path: Path, globs: list[str]) -> list[ChangedFile]:
        out = []
        for p in repo_path.rglob("*"):
            if not p.is_file():
                continue
            rel = p.relative_to(repo_path).as_posix()
            if ".git/" in rel or rel.startswith(".git/"):
                continue
            if any(self._glob_match(rel, g) for g in globs):
                out.append(ChangedFile(path=rel, status="A"))
        return out

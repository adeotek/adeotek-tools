"""Chunk Markdown files into embedding-ready pieces with metadata."""

import hashlib
import re
from dataclasses import dataclass
from pathlib import Path

HEADING_RE = re.compile(r"^(#{1,6})\s+(.*?)\s*$")
MAX_CHUNK_CHARS = 2000  # approx 500 tokens
OVERLAP_CHARS = 200


@dataclass
class Chunk:
    id: str
    text: str
    repo: str
    file_path: str
    line_start: int
    line_end: int
    commit_sha: str
    heading_path: str


def _chunk_id(repo: str, file_path: str, line_start: int, text: str) -> str:
    h = hashlib.sha256()
    h.update(repo.encode())
    h.update(b"\0")
    h.update(file_path.encode())
    h.update(b"\0")
    h.update(str(line_start).encode())
    h.update(b"\0")
    h.update(text.encode())
    return h.hexdigest()[:24]


def _walk_sections(lines: list[str]) -> list[tuple[int, int, list[str], str]]:
    """Split lines into (start_line, end_line, body_lines, heading_path) sections."""
    sections: list[tuple[int, int, list[str], str]] = []
    stack: list[tuple[int, str]] = []  # (level, title)
    current_start = 1
    current_lines: list[str] = []

    def heading_path() -> str:
        return " > ".join(title for _, title in stack)

    for idx, line in enumerate(lines, start=1):
        m = HEADING_RE.match(line)
        if m:
            if current_lines:
                sections.append(
                    (current_start, idx - 1, current_lines, heading_path())
                )
            level = len(m.group(1))
            title = m.group(2).strip()
            while stack and stack[-1][0] >= level:
                stack.pop()
            stack.append((level, title))
            current_start = idx
            current_lines = [line]
        else:
            current_lines.append(line)
    if current_lines:
        sections.append(
            (current_start, len(lines), current_lines, heading_path())
        )
    return sections


def _split_long(text: str, max_chars: int, overlap: int) -> list[str]:
    if len(text) <= max_chars:
        return [text]
    parts: list[str] = []
    start = 0
    while start < len(text):
        end = min(len(text), start + max_chars)
        parts.append(text[start:end])
        if end == len(text):
            break
        start = end - overlap
    return parts


def chunk_markdown_file(
    path: Path,
    *,
    repo: str,
    file_path: str,
    commit_sha: str,
) -> list[Chunk]:
    """Return a list of Chunks for the given Markdown file."""
    text = path.read_text(encoding="utf-8")
    lines = text.splitlines(keepends=False)

    chunks: list[Chunk] = []
    for start, end, body_lines, heading_path in _walk_sections(lines):
        body = "\n".join(body_lines).strip()
        if not body:
            continue
        pieces = _split_long(body, MAX_CHUNK_CHARS, OVERLAP_CHARS)
        for piece in pieces:
            chunk_text = piece.strip()
            if not chunk_text:
                continue
            chunks.append(
                Chunk(
                    id=_chunk_id(repo, file_path, start, chunk_text),
                    text=chunk_text,
                    repo=repo,
                    file_path=file_path,
                    line_start=start,
                    line_end=end,
                    commit_sha=commit_sha,
                    heading_path=heading_path,
                )
            )
    return chunks

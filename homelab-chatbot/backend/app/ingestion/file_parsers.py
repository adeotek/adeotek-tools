"""Parsers for directly-uploaded files (txt, md, docx, pdf, image, xlsx).

All non-Excel parsers return list[Chunk] ready for embedding and vector-store
upsert. Excel returns ([], sheet_map) because its data goes to SQLite via
ExcelLoader, not the vector store.
"""

import tempfile
from io import BytesIO
from pathlib import Path

from app.ingestion.markdown import (
    Chunk,
    _chunk_id,
    _split_long,
    MAX_CHUNK_CHARS,
    OVERLAP_CHARS,
    chunk_markdown_file,
)

UPLOAD_REPO = "_uploads"

ACCEPTED_EXTENSIONS = {
    ".xlsx", ".docx", ".txt", ".md",
    ".png", ".jpg", ".jpeg", ".gif", ".webp",
    ".pdf",
}


def chunk_text_content(text: str, filename: str) -> list[Chunk]:
    """Split plain text into Chunks using the same parameters as the markdown pipeline."""
    if not text.strip():
        return []
    lines = text.splitlines(keepends=False)
    total_lines = len(lines)
    chunks: list[Chunk] = []
    for piece in _split_long(text, MAX_CHUNK_CHARS, OVERLAP_CHARS):
        piece = piece.strip()
        if not piece:
            continue
        chunks.append(
            Chunk(
                id=_chunk_id(UPLOAD_REPO, filename, 1, piece),
                text=piece,
                repo=UPLOAD_REPO,
                file_path=filename,
                line_start=1,
                line_end=total_lines or 1,
                commit_sha="",
                heading_path="",
            )
        )
    return chunks


def parse_txt(content: bytes, filename: str) -> list[Chunk]:
    return chunk_text_content(content.decode("utf-8", errors="replace"), filename)


def parse_md(content: bytes, filename: str) -> list[Chunk]:
    tmp = tempfile.NamedTemporaryFile(suffix=".md", delete=False)
    try:
        tmp.write(content)
        tmp.flush()
        tmp.close()
        return chunk_markdown_file(
            Path(tmp.name),
            repo=UPLOAD_REPO,
            file_path=filename,
            commit_sha="",
        )
    finally:
        Path(tmp.name).unlink(missing_ok=True)


def parse_docx(content: bytes, filename: str) -> list[Chunk]:
    from docx import Document  # type: ignore[import-untyped]

    doc = Document(BytesIO(content))
    text = "\n".join(p.text for p in doc.paragraphs if p.text.strip())
    return chunk_text_content(text, filename)


def parse_pdf(content: bytes, filename: str) -> list[Chunk]:
    from pypdf import PdfReader  # type: ignore[import-untyped]

    reader = PdfReader(BytesIO(content))
    pages = [page.extract_text() or "" for page in reader.pages]
    text = "\n\n".join(pages)
    return chunk_text_content(text, filename)


def parse_image(content: bytes, filename: str) -> list[Chunk]:
    import pytesseract  # type: ignore[import-untyped]
    from PIL import Image

    image = Image.open(BytesIO(content))
    text = pytesseract.image_to_string(image).strip()
    if text:
        return chunk_text_content(text, filename)
    # Fallback: index the filename so the file is at least discoverable
    fallback = f"Image file: {filename}"
    return [
        Chunk(
            id=_chunk_id(UPLOAD_REPO, filename, 1, fallback),
            text=fallback,
            repo=UPLOAD_REPO,
            file_path=filename,
            line_start=1,
            line_end=1,
            commit_sha="",
            heading_path="",
        )
    ]


def parse_xlsx(content: bytes, _filename: str, excel_loader) -> tuple[list[Chunk], dict]:
    """Load an Excel file into SQLite via ExcelLoader. Returns ([], sheet_map)."""
    tmp = tempfile.NamedTemporaryFile(suffix=".xlsx", delete=False)
    try:
        tmp.write(content)
        tmp.flush()
        tmp.close()
        sheet_map = excel_loader.load(Path(tmp.name))
        return [], sheet_map
    finally:
        Path(tmp.name).unlink(missing_ok=True)


def parse_file(
    content: bytes,
    filename: str,
    excel_loader,
) -> tuple[list[Chunk], dict]:
    """Dispatch to the correct parser based on file extension.

    Returns (chunks, sheet_map). sheet_map is non-empty only for .xlsx files.
    Raises ValueError for unsupported extensions.
    """
    ext = Path(filename).suffix.lower()
    if ext not in ACCEPTED_EXTENSIONS:
        raise ValueError(f"unsupported file type: {ext!r}")

    if ext == ".txt":
        return parse_txt(content, filename), {}
    if ext == ".md":
        return parse_md(content, filename), {}
    if ext == ".docx":
        return parse_docx(content, filename), {}
    if ext == ".pdf":
        return parse_pdf(content, filename), {}
    if ext in {".png", ".jpg", ".jpeg", ".gif", ".webp"}:
        return parse_image(content, filename), {}
    if ext == ".xlsx":
        return parse_xlsx(content, filename, excel_loader)

    raise ValueError(f"unhandled extension: {ext!r}")  # unreachable

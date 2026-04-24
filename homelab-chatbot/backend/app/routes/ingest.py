"""File upload and ingestion endpoint."""

import logging
from pathlib import Path

from fastapi import APIRouter, Depends, HTTPException, Request, UploadFile
from pydantic import BaseModel

from app.deps import require_session
from app.ingestion.file_parsers import ACCEPTED_EXTENSIONS, UPLOAD_REPO, parse_file
from app.retrieval.vector_tool import VectorSearchTool
from app.storage.upload_log_db import UploadLogDB

logger = logging.getLogger(__name__)

router = APIRouter(
    prefix="/api/ingest",
    tags=["ingest"],
    dependencies=[Depends(require_session)],
)

MAX_FILES = 10
MAX_FILE_BYTES = 20 * 1024 * 1024  # 20 MB


class FileResult(BaseModel):
    filename: str
    status: str
    chunks_created: int
    replaced: bool
    error_message: str | None = None


class UploadResponse(BaseModel):
    results: list[FileResult]


@router.post("/upload")
async def upload_files(request: Request, files: list[UploadFile]) -> UploadResponse:
    if len(files) > MAX_FILES:
        raise HTTPException(status_code=422, detail=f"Too many files — maximum is {MAX_FILES}.")

    vector_tool: VectorSearchTool = request.app.state.vector_tool
    store = vector_tool._store
    embedder = vector_tool._embedder
    excel_loader = request.app.state.excel_loader
    upload_log: UploadLogDB = request.app.state.upload_log_db

    results: list[FileResult] = []

    for upload in files:
        filename = Path(upload.filename or "unknown").name
        ext = Path(filename).suffix.lower()
        content = await upload.read()

        if len(content) > MAX_FILE_BYTES:
            msg = f"File exceeds 20 MB limit ({len(content) / 1_048_576:.1f} MB)."
            upload_log.log_upload(
                filename=filename,
                file_size=len(content),
                mime_type=upload.content_type or "application/octet-stream",
                status="error",
                chunks_created=0,
                replaced=False,
                error_message=msg,
            )
            results.append(FileResult(filename=filename, status="error", chunks_created=0, replaced=False, error_message=msg))
            continue

        if ext not in ACCEPTED_EXTENSIONS:
            msg = f"Unsupported file type: {ext!r}."
            upload_log.log_upload(
                filename=filename,
                file_size=len(content),
                mime_type=upload.content_type or "application/octet-stream",
                status="error",
                chunks_created=0,
                replaced=False,
                error_message=msg,
            )
            results.append(FileResult(filename=filename, status="error", chunks_created=0, replaced=False, error_message=msg))
            continue

        replaced = store.has_file(UPLOAD_REPO, filename)
        try:
            chunks, _sheet_map = parse_file(content, filename, excel_loader)

            if replaced:
                store.delete_by_file(UPLOAD_REPO, filename)

            chunks_created = 0
            if chunks:
                vectors = embedder.embed_batch([c.text for c in chunks])
                store.upsert(chunks, vectors)
                chunks_created = len(chunks)

            upload_log.log_upload(
                filename=filename,
                file_size=len(content),
                mime_type=upload.content_type or "application/octet-stream",
                status="ok",
                chunks_created=chunks_created,
                replaced=replaced,
                error_message=None,
            )
            results.append(FileResult(filename=filename, status="ok", chunks_created=chunks_created, replaced=replaced))
        except Exception as exc:
            logger.exception("Failed to ingest uploaded file %r", filename)
            msg = str(exc)
            upload_log.log_upload(
                filename=filename,
                file_size=len(content),
                mime_type=upload.content_type or "application/octet-stream",
                status="error",
                chunks_created=0,
                replaced=replaced,
                error_message=msg,
            )
            results.append(FileResult(filename=filename, status="error", chunks_created=0, replaced=replaced, error_message=msg))

    return UploadResponse(results=results)

# Context: file-upload-ingestion-feature
**Saved**: 2026-04-25 | **Directory**: /home/georg/projects/adeotek-tools
**Branch**: feat/homelab-chatbot | **Recent commits**: 557bf03 [feat:homelab-chatbot] files upload, ea2794e [fix:homelab-chatbot] fixes, dba7c71 [fix:homelab-chatbot] fixes

## Goal
Add direct file upload + ingestion to the homelab-chatbot's Data Ingestion page. Users can upload Excel, Word, text, Markdown, image (OCR via pytesseract), and PDF files (max 10 files, 20 MB each). Files are parsed, chunked, embedded into the LanceDB vector store (repo `_uploads`), and never persisted after ingestion. Excel sheets go into SQLite `kb.db` (same as git-repo-sourced Excel). A persistent `upload_log` table in `kb.db` tracks all uploads and is queryable by the LLM via the existing `SQLTool`.

## Current State
All backend and frontend work is complete. The following bugs/improvements were applied post-implementation:
- `python-multipart` missing from `pyproject.toml` (added)
- `LanceTable.to_pandas()` called with unsupported `columns=` kwarg (fixed to `tbl.to_pandas()[["repo", "file_path"]]`)
- `ExcelLoader.normalize_snake_case()` crashed on integer column headers (fixed with `str(c)` cast — pre-existing bug)
- `SQLTool.schema_summary()` improved: now includes sample rows per table, refreshes on every query call, excluded `upload_log` from UI `table_stats()`
- `SQLTool._run()` now returns schema + hint when query returns empty rows, enabling the ReAct agent to self-correct and retry rather than immediately giving up

Ongoing issue: the chatbot sometimes fails to find data from uploaded Excel files despite the data being present. Root cause is likely a value format mismatch (e.g. `ollama1` vs `ollama1.lan`) combined with weak SQL generation by local Ollama models. The empty-result hint fix is the latest attempt to address this — not yet confirmed resolved.

## Key Files
- `backend/app/storage/upload_log_db.py` — new; raw sqlite3 wrapper for `upload_log` table in `kb.db`
- `backend/app/ingestion/file_parsers.py` — new; dispatches to per-type parsers (txt/md/docx/pdf/image/xlsx)
- `backend/app/routes/ingest.py` — new; `POST /api/ingest/upload` multipart endpoint
- `backend/app/storage/lance.py` — added `has_file(repo, file_path) -> bool` for dedup detection
- `backend/app/ingestion/excel.py` — fixed `str(c)` cast in `normalize_snake_case` call
- `backend/app/routes/stats.py` — `StatsOut` extended with `uploaded_files`, `uploaded_chunks`, `upload_log`
- `backend/app/main.py` — `ExcelLoader` and `UploadLogDB` constructed unconditionally; ingest router included
- `backend/pyproject.toml` — added `python-multipart`, `python-docx`, `pypdf`, `pytesseract`, `Pillow`
- `Dockerfile` — added `tesseract-ocr` to apt install in runtime stage
- `backend/app/retrieval/sql_tool.py` — schema_summary shows sample rows + refreshes per-query; table_stats excludes system tables; _run returns schema+hint on empty results
- `frontend/lib/api.ts` — new types `UploadLogEntry`, `UploadFileResult`, `UploadResponse`; `Stats` extended; `uploadFiles()` using raw `FormData` fetch
- `frontend/components/stats/UploadModal.tsx` — new; drag-drop modal with validation, per-file results, Escape-to-close
- `frontend/components/stats/UploadLogTable.tsx` — new; presentational upload history table
- `frontend/app/stats/page.tsx` — Upload button between Refresh and Sync now; uploaded files/chunks cards in Overview grid; Uploaded Files section with log table

## Decisions Made
- **`upload_log` in `kb.db` not `chat.db`**: User explicitly chose this so the LLM's existing `SQLTool` can answer "what files were uploaded" questions without any new tooling.
- **Virtual repo `_uploads`**: Underscore prefix prevents collision with user-defined git repo names; all uploaded file chunks land here in the LanceDB vector store.
- **Image OCR via pytesseract**: User chose this over vision-LLM description; fallback chunk `"Image file: {filename}"` stored when OCR returns empty.
- **`schema_summary()` includes `upload_log`**: Kept in LLM-visible schema so chatbot can query upload history, but excluded from the UI `table_stats()` display.
- **No file persistence**: Uploaded files processed in memory (or temp file for xlsx/md), immediately deleted after ingestion.
- **`uploadFiles()` uses raw `fetch` not `request()`**: The existing `request()` helper hardcodes `Content-Type: application/json` which breaks multipart boundary; FormData requires no explicit Content-Type.
- **Return schema+hint on empty SQL results**: When `_run()` gets empty rows, it now returns the schema and a LIKE-retry hint so the ReAct agent can self-correct instead of giving up after one attempt.

## Next Steps
- [ ] Confirm the empty-result hint fix resolves the chatbot inventory lookup issue in practice
- [ ] Test image upload end-to-end (requires `tesseract-ocr` installed on host for dev)
- [ ] Consider adding a "Delete uploaded file" action to the upload log table (removes chunks from vector store and kb tables)
- [x] `uv sync` to install new dependencies
- [x] Fix `python-multipart` missing dep
- [x] Fix LanceDB `to_pandas(columns=...)` kwarg error
- [x] Fix integer column headers in ExcelLoader

## Notes
- `tesseract-ocr` system package must be present at runtime; in dev mode install separately (`apt install tesseract-ocr` or equivalent)
- The LanceDB `has_file()` method loads the full table into pandas to check existence — acceptable for small vector stores but may be slow with millions of chunks
- `upload_log` will appear in kb.db alongside Excel-sourced tables; the chatbot can query it with natural language ("list my recent uploads")
- The integer column header bug in `ExcelLoader` was pre-existing and also affected git-repo-synced Excel files
- If chatbot still fails to find Excel data after all fixes, the most likely remaining cause is the local Ollama model ignoring schema constraints; switching to Anthropic/Google provider for testing would confirm this

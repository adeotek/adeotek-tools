# Session Naming — Design Spec

**Date:** 2026-05-10
**Status:** Approved

## Overview

Add optional names to sessions so users can label them meaningfully. A name can be set at creation time or changed later from the active-session header.

---

## Data Layer

### DB migration (`backend/src/db/schema.ts`)

Add a `name TEXT` column to the `sessions` table via an inline migration (same pattern as the existing `claude_session_id` migration):

```sql
ALTER TABLE sessions ADD COLUMN name TEXT;
```

`name` is nullable — sessions without an explicit name fall back to the workdir's last path segment in the UI.

### Backend API (`backend/src/routes/sessions.ts`)

| Method | Path | Change |
|--------|------|--------|
| `POST` | `/api/sessions` | Accept optional `name` in JSON body; insert into DB when non-empty |
| `PATCH` | `/api/sessions/:id` | **New.** Accept `{ name: string }`, update the row, return `{ ok: true }` |
| `GET` | `/api/sessions` | No change — `SELECT s.*` already includes `name` |

`PATCH /api/sessions/:id` validates that `name` is a string and that the session exists (404 if not).

### Frontend types

**`useDashboard.ts` — `Session` interface:**
```ts
name: string | null   // add this field
```

**`SessionContext.tsx` — `SessionState`:**
```ts
name: string | null   // add this field (null in initial state)
```

**`SessionContext.tsx` — actions:**
- `SESSION_CREATED` gains optional `name?: string`
- `RESUME_SESSION` gains optional `name?: string`
- New action: `{ type: 'SESSION_RENAMED'; name: string | null }` (null clears the name)

---

## UI

### NewSessionModal

Add a `Name` field **above** the workdir input. The field is optional — the label reads `Name (optional)`. If left blank the `name` key is omitted from the POST body entirely.

Tab order: Name → Workdir → Submit.

### SessionList

Change the primary title of each row from:
```ts
lastSegment(session.workdir)
```
to:
```ts
session.name ?? lastSegment(session.workdir)
```

No other changes to the list. The full workdir path below the title is unchanged.

### SessionHeader

**Name display:** When a session has a name, show it as `text-text-primary` text **before** the workdir chip in the status row (same horizontal flex line). When no name is set, nothing extra is shown — the workdir chip stays as is.

**Rename button:** A small Pencil (`Pencil` icon, 11px) button is added to the **left** of the existing action-button group (before New session / Stop session). It uses the same muted style as those buttons.

**Edit mode:** Clicking the Pencil button replaces the name display (or workdir chip if no name) with:
- An `<input>` pre-filled with the current name (or empty if none)
- A `Save` button
- An `✕` Cancel button

Keyboard: `Enter` saves, `Escape` cancels. Clicking Cancel reverts with no API call.

**Save flow:**
1. Call `PATCH /api/sessions/:id` with `{ name }`.
2. On success, dispatch `SESSION_RENAMED` to update in-memory state immediately (no re-fetch needed).
3. On error, show a brief inline error message and return to edit mode.

### DashboardView

- `handleSessionStart` receives the optional name from `NewSessionModal` and includes it in the `SESSION_CREATED` dispatch.
- `handleResume` reads `session.name` from the session object and includes it in the `RESUME_SESSION` dispatch.
- `SessionHeader` receives `sessionName: string | null` and `onRename: (name: string) => Promise<void>` props.

---

## Error Handling

- PATCH 404 (session not found): show inline error in edit mode.
- PATCH network error: show inline error, stay in edit mode, allow retry.
- Empty name on save: treat as clearing the name (store `null` / empty string is normalised to `null` on the backend).

---

## Out of Scope

- Unique name enforcement (names are cosmetic labels only)
- Renaming from the session list (header only)
- Name history / undo

import Database from 'better-sqlite3'
import path from 'path'
import fs from 'fs'

const DB_DIR = process.env.DATA_DIR ?? path.join(process.env.HOME ?? '/root', '.claude', 'dashboard')
const DB_PATH = path.join(DB_DIR, 'dashboard.db')

function initDb(): Database.Database {
  fs.mkdirSync(DB_DIR, { recursive: true })
  const db = new Database(DB_PATH)
  db.pragma('journal_mode = WAL')
  db.pragma('foreign_keys = ON')

  db.prepare(`
    CREATE TABLE IF NOT EXISTS sessions (
      id TEXT PRIMARY KEY,
      workdir TEXT NOT NULL,
      model TEXT,
      started_at INTEGER NOT NULL,
      ended_at INTEGER
    )
  `).run()

  db.prepare(`
    CREATE TABLE IF NOT EXISTS messages (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      session_id TEXT NOT NULL REFERENCES sessions(id),
      role TEXT NOT NULL CHECK(role IN ('user', 'assistant')),
      content TEXT NOT NULL,
      tokens INTEGER,
      created_at INTEGER NOT NULL
    )
  `).run()

  db.prepare(`
    CREATE TABLE IF NOT EXISTS usage_cache (
      date TEXT PRIMARY KEY,
      input_tokens INTEGER NOT NULL DEFAULT 0,
      output_tokens INTEGER NOT NULL DEFAULT 0,
      cost_usd REAL NOT NULL DEFAULT 0,
      source TEXT NOT NULL CHECK(source IN ('local', 'api', 'merged')),
      cached_at INTEGER NOT NULL
    )
  `).run()

  db.prepare(`
    CREATE TABLE IF NOT EXISTS oauth_cache (
      key TEXT PRIMARY KEY,
      data TEXT NOT NULL,
      cached_at INTEGER NOT NULL
    )
  `).run()

  // Migrate: add claude_session_id to sessions if missing
  const sessionCols = db.pragma('table_info(sessions)') as Array<{ name: string }>
  if (!sessionCols.find((c) => c.name === 'claude_session_id')) {
    db.prepare('ALTER TABLE sessions ADD COLUMN claude_session_id TEXT').run()
  }

  db.prepare(`
    CREATE TABLE IF NOT EXISTS settings (
      key   TEXT PRIMARY KEY,
      value TEXT NOT NULL
    )
  `).run()

  return db
}

export const db = initDb()

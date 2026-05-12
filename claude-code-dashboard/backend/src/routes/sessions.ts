import type { FastifyInstance } from 'fastify'
import { v4 as uuidv4 } from 'uuid'
import fs from 'fs'
import os from 'os'
import path from 'path'
import { db } from '../db/schema'
import { sessionManager } from '../ws/session'

export async function sessionRoutes(fastify: FastifyInstance) {
  fastify.get<{ Querystring: { path?: string } }>('/api/directories', async (req, reply) => {
    const rawPath = req.query.path ?? os.homedir()
    const resolved = rawPath.startsWith('~')
      ? path.join(os.homedir(), rawPath.slice(1))
      : rawPath
    try {
      const entries = fs.readdirSync(resolved, { withFileTypes: true })
      const dirs = entries
        .filter((e) => e.isDirectory() && !e.name.startsWith('.'))
        .map((e) => path.join(resolved, e.name))
        .sort()
      return reply.send(dirs)
    } catch {
      return reply.send([])
    }
  })

  fastify.get('/api/sessions', async (_req, reply) => {
    const rows = db
      .prepare(
        `SELECT s.*,
          CASE WHEN s.ended_at IS NULL THEN 1 ELSE 0 END as is_active,
          COUNT(m.id) as message_count
        FROM sessions s
        LEFT JOIN messages m ON m.session_id = s.id
        GROUP BY s.id
        ORDER BY s.started_at DESC
        LIMIT 50`,
      )
      .all()
    return reply.send(rows)
  })

  fastify.post<{ Body: { workdir: string; name?: string } }>('/api/sessions', async (req, reply) => {
    const { workdir, name } = req.body
    if (!workdir || typeof workdir !== 'string') {
      return reply.status(400).send({ error: 'workdir is required' })
    }

    const id = uuidv4()
    db.prepare(
      'INSERT INTO sessions (id, workdir, name, started_at) VALUES (?, ?, ?, ?)',
    ).run(id, workdir, name?.trim() || null, Date.now())

    return reply.status(201).send({ sessionId: id })
  })

  fastify.post<{ Params: { id: string } }>('/api/sessions/:id/stop', async (req, reply) => {
    const { id } = req.params
    sessionManager.kill(id)
    db.prepare('UPDATE sessions SET ended_at = ? WHERE id = ?').run(Date.now(), id)
    return reply.send({ ok: true })
  })

  fastify.patch<{ Params: { id: string }; Body: { name: string } }>('/api/sessions/:id', async (req, reply) => {
    const { id } = req.params
    const { name } = req.body

    const session = db.prepare('SELECT id FROM sessions WHERE id = ?').get(id)
    if (!session) {
      return reply.status(404).send({ error: 'Session not found' })
    }

    db.prepare('UPDATE sessions SET name = ? WHERE id = ?').run(
      typeof name === 'string' ? name.trim() || null : null,
      id,
    )
    return reply.send({ ok: true })
  })

  fastify.get<{ Params: { id: string } }>('/api/sessions/:id/messages', async (req, reply) => {
    const { id } = req.params
    const session = db.prepare('SELECT id FROM sessions WHERE id = ?').get(id)
    if (!session) {
      return reply.status(404).send({ error: 'Session not found' })
    }
    const messages = db
      .prepare('SELECT * FROM messages WHERE session_id = ? ORDER BY created_at ASC')
      .all(id)
    return reply.send({ messages })
  })

  fastify.delete<{ Params: { id: string } }>('/api/sessions/:id', async (req, reply) => {
    const { id } = req.params
    const session = db.prepare('SELECT id FROM sessions WHERE id = ?').get(id)
    if (!session) {
      return reply.status(404).send({ error: 'Session not found' })
    }
    // Kill the session if it's running
    sessionManager.kill(id)
    db.prepare('DELETE FROM messages WHERE session_id = ?').run(id)
    db.prepare('DELETE FROM sessions WHERE id = ?').run(id)
    return reply.send({ ok: true })
  })
}

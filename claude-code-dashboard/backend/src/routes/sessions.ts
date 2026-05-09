import type { FastifyInstance } from 'fastify'
import { v4 as uuidv4 } from 'uuid'
import { db } from '../db/schema'
import { sessionManager } from '../ws/session'

export async function sessionRoutes(fastify: FastifyInstance) {
  fastify.get('/api/sessions', async (_req, reply) => {
    const rows = db
      .prepare('SELECT * FROM sessions ORDER BY started_at DESC LIMIT 50')
      .all()
    return reply.send(rows)
  })

  fastify.post<{ Body: { workdir: string } }>('/api/sessions', async (req, reply) => {
    const { workdir } = req.body
    if (!workdir || typeof workdir !== 'string') {
      return reply.status(400).send({ error: 'workdir is required' })
    }

    const id = uuidv4()
    db.prepare(
      'INSERT INTO sessions (id, workdir, started_at) VALUES (?, ?, ?)',
    ).run(id, workdir, Date.now())

    return reply.status(201).send({ sessionId: id })
  })

  fastify.post<{ Params: { id: string } }>('/api/sessions/:id/stop', async (req, reply) => {
    const { id } = req.params
    sessionManager.kill(id)
    db.prepare('UPDATE sessions SET ended_at = ? WHERE id = ?').run(Date.now(), id)
    return reply.send({ ok: true })
  })
}

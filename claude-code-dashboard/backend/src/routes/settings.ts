import type { FastifyInstance } from 'fastify'
import { db } from '../db/schema'

const DEFAULTS: Record<string, string> = {
  bypass_permissions: 'true',
}

const ALLOWED = new Set(Object.keys(DEFAULTS))

export async function settingsRoutes(fastify: FastifyInstance) {
  fastify.get('/api/settings', async (_req, reply) => {
    const rows = db.prepare('SELECT key, value FROM settings').all() as Array<{
      key: string
      value: string
    }>
    const result: Record<string, string> = { ...DEFAULTS }
    for (const row of rows) result[row.key] = row.value
    return reply.send(result)
  })

  fastify.post<{ Body: Record<string, string> }>(
    '/api/settings',
    {
      schema: {
        body: {
          type: 'object',
          additionalProperties: { type: 'string' },
        },
      },
    },
    async (req, reply) => {
      const entries = Object.entries(req.body).filter(([key]) => ALLOWED.has(key))
      if (entries.length === 0) return reply.status(400).send({ error: 'no valid keys' })
      const upsert = db.prepare('INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)')
      const tx = db.transaction(() => {
        for (const [key, value] of entries) upsert.run(key, value)
      })
      tx()
      return reply.send({ ok: true })
    },
  )
}

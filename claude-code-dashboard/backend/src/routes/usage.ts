import type { FastifyInstance } from 'fastify'
import { getUsage } from '../services/usageCache'

export async function usageRoutes(fastify: FastifyInstance) {
  fastify.get<{ Querystring: { month?: string } }>('/api/usage', async (req, reply) => {
    const month = req.query.month ?? new Date().toISOString().slice(0, 7)
    if (!/^\d{4}-\d{2}$/.test(month)) {
      return reply.status(400).send({ error: 'month must be YYYY-MM' })
    }
    const result = await getUsage(month)
    return reply.send(result)
  })
}

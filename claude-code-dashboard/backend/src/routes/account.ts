import type { FastifyInstance } from 'fastify'
import { getAccountInfo } from '../services/claudeAccount'

export async function accountRoutes(fastify: FastifyInstance) {
  fastify.get('/api/account', async (_req, reply) => {
    const info = await getAccountInfo()
    return reply.send(info)
  })
}

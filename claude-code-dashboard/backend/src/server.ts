import Fastify from 'fastify'
import websocket from '@fastify/websocket'
import cors from '@fastify/cors'
import { accountRoutes } from './routes/account'
import { usageRoutes } from './routes/usage'
import { sessionRoutes } from './routes/sessions'
import { sessionWsRoutes } from './ws/session'

// TODO: add bearer token auth — add @fastify/bearer-auth plugin here
// and set token via DASHBOARD_TOKEN env var
// fastify.addHook('onRequest', async (request, reply) => { ... })

const fastify = Fastify({ logger: true })

async function start() {
  await fastify.register(cors, {
    origin: process.env.FRONTEND_ORIGIN ?? 'http://localhost:5173',
  })
  await fastify.register(websocket)

  await fastify.register(accountRoutes)
  await fastify.register(usageRoutes)
  await fastify.register(sessionRoutes)
  await fastify.register(sessionWsRoutes)

  fastify.get('/health', async () => ({ status: 'ok' }))

  const PORT = Number(process.env.PORT ?? 3001)
  const HOST = process.env.HOST ?? '0.0.0.0'

  try {
    await fastify.listen({ port: PORT, host: HOST })
  } catch (err) {
    fastify.log.error(err)
    process.exit(1)
  }
}

start()

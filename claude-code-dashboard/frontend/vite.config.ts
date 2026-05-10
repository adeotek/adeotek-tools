import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { hostname as getHostname, networkInterfaces } from 'os'

const hostname = getHostname()
const localIps = (Object.values(networkInterfaces()).flat() as { internal: boolean; address: string }[])
  .filter(iface => iface && !iface.internal)
  .map(iface => iface.address)

export default defineConfig({
  plugins: [react()],
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'vendor-react': ['react', 'react-dom', 'react-router-dom'],
          'vendor-xterm': ['@xterm/xterm', '@xterm/addon-fit'],
          'vendor-charts': ['recharts'],
          'vendor-markdown': ['react-markdown', 'react-syntax-highlighter'],
        },
      },
    },
  },
  server: {
    host: true,
    port: 9999,
    allowedHosts: ['localhost', hostname, `${hostname}.lan`, ...localIps],
    proxy: {
      '/api': 'http://localhost:9998',
      '/ws': {
        target: 'ws://localhost:9998',
        ws: true,
      },
    },
  },
})

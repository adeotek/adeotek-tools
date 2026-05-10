import type { Config } from 'tailwindcss'

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        bg: {
          base: '#0a0a0a',
          surface: '#0f0f0f',
          elevated: '#141414',
          panel: '#1a1a1a',
        },
        border: {
          subtle: '#1e1e1e',
          DEFAULT: '#2a2a2a',
          strong: '#333333',
        },
        accent: {
          DEFAULT: '#d97706',
          hover: '#f59e0b',
          muted: '#92400e',
        },
        text: {
          primary: '#f2f2f2',
          secondary: '#aaaaaa',
          muted: '#777777',
          dim: '#555555',
        },
        status: {
          green: '#22c55e',
          blue: '#93c5fd',
          purple: '#c084fc',
          red: '#f87171',
        },
      },
      fontFamily: {
        mono: ['JetBrains Mono', 'Fira Code', 'Cascadia Code', 'monospace'],
      },
    },
  },
  plugins: [],
} satisfies Config

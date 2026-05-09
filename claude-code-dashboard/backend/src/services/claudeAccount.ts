import fs from 'fs'
import path from 'path'
import os from 'os'
import { execFileNoThrow } from '../utils/execFileNoThrow'

const CLAUDE_HOME = path.join(os.homedir(), '.claude')
const SETTINGS_PATH = path.join(CLAUDE_HOME, 'settings.json')

const claudeBin = () => process.env.CLAUDE_BIN ?? 'claude'

export interface AccountInfo {
  claudeInstalled: boolean
  version: string | null
  model: string | null
  os: string
  nodeVersion: string
  authStatus: 'authenticated' | 'unauthenticated' | 'unknown'
}

export async function getAccountInfo(): Promise<AccountInfo> {
  const base: AccountInfo = {
    claudeInstalled: false,
    version: null,
    model: null,
    os: `${os.platform()} ${os.arch()}`,
    nodeVersion: process.version,
    authStatus: 'unknown',
  }

  const versionResult = await execFileNoThrow(claudeBin(), ['--version'])
  if (versionResult.status !== 0) return base

  base.claudeInstalled = true
  const versionMatch = versionResult.stdout.match(/[\d]+\.[\d]+\.[\d]+/)
  base.version = versionMatch ? versionMatch[0] : versionResult.stdout.trim()

  try {
    const raw = fs.readFileSync(SETTINGS_PATH, 'utf8')
    const settings = JSON.parse(raw) as Record<string, unknown>
    if (typeof settings.model === 'string') base.model = settings.model
  } catch {
    // settings.json absent or malformed — model stays null
  }

  const authResult = await execFileNoThrow(claudeBin(), ['config', 'get', 'oauthToken'])
  base.authStatus = authResult.status === 0 && authResult.stdout.trim().length > 0
    ? 'authenticated'
    : 'unauthenticated'

  return base
}

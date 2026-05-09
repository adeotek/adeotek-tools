import { execFile } from 'child_process'
import { promisify } from 'util'

const execFileAsync = promisify(execFile)

export async function execFileNoThrow(
  cmd: string,
  args: string[] = [],
): Promise<{ stdout: string; stderr: string; status: number }> {
  try {
    const { stdout, stderr } = await execFileAsync(cmd, args, { timeout: 10_000 })
    return { stdout, stderr, status: 0 }
  } catch (err: unknown) {
    const e = err as { stdout?: string; stderr?: string; code?: number }
    return { stdout: e.stdout ?? '', stderr: e.stderr ?? '', status: e.code ?? 1 }
  }
}

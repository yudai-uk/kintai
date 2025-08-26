import { promises as fs } from 'fs';
import path from 'path';

function getLogPath() {
  // Project root/log/frontend.log
  const rootLogDir = path.resolve(process.cwd(), '..', 'log');
  const file = path.join(rootLogDir, 'frontend.log');
  return { rootLogDir, file };
}

export async function logServer(level: 'info'|'warn'|'error', message: string, context?: Record<string, any>) {
  const { rootLogDir, file } = getLogPath();
  try {
    await fs.mkdir(rootLogDir, { recursive: true });
    const line = JSON.stringify({
      ts: new Date().toISOString(),
      level,
      message,
      context: context ?? null,
    }) + '\n';
    await fs.appendFile(file, line, { encoding: 'utf8' });
  } catch (e) {
    // Swallow logging errors to avoid breaking main flow
  }
}


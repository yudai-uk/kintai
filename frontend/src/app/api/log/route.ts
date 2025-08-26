import { NextResponse } from 'next/server';
import { logServer } from '@/utils/logger';

export const runtime = 'nodejs';

export async function POST(req: Request) {
  try {
    const body = await req.json().catch(() => ({}));
    const level = (body.level as 'info'|'warn'|'error') || 'error';
    const message = (body.message as string) || 'client-error';
    const context = (body.context as Record<string, any>) || undefined;
    await logServer(level, message, context);
    return NextResponse.json({ ok: true });
  } catch (e: any) {
    return NextResponse.json({ ok: false, error: e?.message || 'log_failed' }, { status: 500 });
  }
}


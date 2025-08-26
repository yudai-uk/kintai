'use server';

import { redirect } from 'next/navigation';
import { createClient } from '@/utils/supabase/server';
import { logServer } from '@/utils/logger';

export type LoginState = {
  error?: string;
};

export async function login(_prevState: LoginState, formData: FormData): Promise<LoginState | void> {
  const email = String(formData.get('email') || '').trim();
  const password = String(formData.get('password') || '');
  const redirectTo = String(formData.get('redirect') || '/employee') || '/employee';

  if (!email || !password) {
    return { error: 'メールアドレスとパスワードを入力してください。' };
  }

  const supabase = await createClient();
  const { error } = await supabase.auth.signInWithPassword({ email, password });

  if (error) {
    await logServer('error', 'login_failed', { email, error: error.message });
    return { error: error.message };
  }

  redirect(redirectTo);
}

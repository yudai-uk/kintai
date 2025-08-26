'use client';

import { useActionState } from 'react';
import { useFormStatus } from 'react-dom';
import { login, type LoginState } from './actions';

function SubmitButton() {
  const { pending } = useFormStatus();
  return (
    <button
      type="submit"
      disabled={pending}
      className="flex min-w-[84px] max-w-[480px] items-center justify-center overflow-hidden rounded-full h-10 px-4 flex-1 bg-[#999999] text-[#141414] text-sm font-bold tracking-[0.015em] disabled:opacity-60"
    >
      <span className="truncate">{pending ? 'Logging in...' : 'Login'}</span>
    </button>
  );
}

export function LoginForm({ redirect }: { redirect?: string }) {
  const initialState: LoginState = {};
  const [state, formAction] = useActionState(login, initialState);

  return (
    <form action={formAction} noValidate>
      <input type="hidden" name="redirect" value={redirect || '/'} />

      <div className="flex max-w-[480px] flex-wrap items-end gap-4 px-4 py-3">
        <label className="flex flex-col min-w-40 flex-1" htmlFor="email">
          <span className="text-[#141414] text-base font-medium leading-normal pb-2">Email</span>
          <input
            id="email"
            name="email"
            type="email"
            autoComplete="email"
            placeholder="Email"
            className="form-input w-full rounded-xl text-[#141414] focus:outline-0 focus:ring-0 border border-[#dbdbdb] bg-neutral-50 focus:border-[#dbdbdb] h-14 placeholder:text-neutral-500 p-[15px] text-base"
            required
          />
        </label>
      </div>

      <div className="flex max-w-[480px] flex-wrap items-end gap-4 px-4 py-3">
        <label className="flex flex-col min-w-40 flex-1" htmlFor="password">
          <span className="text-[#141414] text-base font-medium leading-normal pb-2">Password</span>
          <input
            id="password"
            name="password"
            type="password"
            autoComplete="current-password"
            placeholder="Password"
            className="form-input w-full rounded-xl text-[#141414] focus:outline-0 focus:ring-0 border border-[#dbdbdb] bg-neutral-50 focus:border-[#dbdbdb] h-14 placeholder:text-neutral-500 p-[15px] text-base"
            required
          />
        </label>
      </div>

      <div className="px-4 pb-2">
        {state?.error && (
          <p className="text-sm text-red-600">{state.error}</p>
        )}
      </div>

      <div className="flex px-4 py-3">
        <SubmitButton />
      </div>

      <p className="text-neutral-500 text-sm leading-normal pb-3 pt-1 px-4 text-center underline">
        Forgot password?
      </p>
      <p className="text-neutral-500 text-sm leading-normal pb-3 pt-1 px-4 text-center underline">
        Don&apos;t have an account? Sign Up
      </p>
    </form>
  );
}

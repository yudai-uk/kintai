'use client';

import { useEffect, useMemo, useState } from 'react';
import { apiClient } from '@/lib/api';
import { supabase } from '@/utils/supabase/client';
import { useRouter } from 'next/navigation';
import { Header } from '@/components/Header';

type Attendance = {
  id: number;
  date: string; // ISO date only
  clock_in?: string | null; // ISO datetime
  clock_out?: string | null; // ISO datetime
  break_time?: number | null;
  note?: string | null;
  break_start?: string | null;
  break_end?: string | null;
  out_start?: string | null;
  out_end?: string | null;
  work_mode?: 'office' | 'remote' | null;
};

function formatTime(iso?: string | null) {
  if (!iso) return '-';
  try {
    const d = new Date(iso);
    return d.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' });
  } catch {
    return '-';
  }
}

export default function EmployeeDashboard() {
  const [items, setItems] = useState<Attendance[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    (async () => {
      try {
        const res = await apiClient.get<{ data: Attendance[] }>(
          `/api/v1/attendance/me?limit=20&_ts=${Date.now()}`,
          { auth: true },
        );
        setItems(res.data || []);
      } catch (e: any) {
        setError(e?.message || 'Failed to load');
        // fire-and-forget client log
        fetch('/api/log', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ level: 'error', message: 'employee_fetch_failed', context: { error: e?.message } }),
        }).catch(() => {});
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  const today = useMemo(() => {
    const todayKey = new Date().toLocaleDateString('sv-SE');
    return items.find((a) => {
      try {
        const recKey = new Date(a.date).toLocaleDateString('sv-SE');
        return recKey === todayKey;
      } catch {
        return false;
      }
    });
  }, [items]);

  const statusLabel = today?.clock_in
    ? today?.clock_out
      ? '退勤済'
      : today?.break_start && !today?.break_end
        ? '休憩中'
        : today?.out_start && !today?.out_end
          ? '外出中'
          : '勤務中'
    : '勤務前';

  const refresh = async () => {
    const res = await apiClient.get<{ data: Attendance[] }>(
      '/api/v1/attendance/me?limit=20',
      { auth: true },
    );
    setItems(res.data || []);
  };

  const doClock = async (action: 'in' | 'out') => {
    try {
      setLoading(true);
      const qs = action === 'out' ? '?action=clock_out' : '';
      const updated = await apiClient.post<Attendance>(`/api/v1/attendance${qs}`, {}, { auth: true });
      // Merge/update local state to avoid cache issues
      setItems((prev) => {
        const next = [...prev];
        const idx = next.findIndex((a) => new Date(a.date).toLocaleDateString('sv-SE') === new Date(updated.date).toLocaleDateString('sv-SE'));
        if (idx >= 0) next[idx] = { ...next[idx], ...updated } as Attendance;
        else next.unshift(updated as any);
        return next;
      });
    } catch (e: any) {
      setError(e?.message || 'Failed to update attendance');
      fetch('/api/log', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ level: 'error', message: 'employee_clock_failed', context: { action, error: e?.message } }),
      }).catch(() => {});
    } finally {
      setLoading(false);
    }
  };

  const doAction = async (action: 'break_start' | 'break_end' | 'out' | 'return') => {
    try {
      setLoading(true);
      const updated = await apiClient.post<Attendance>(`/api/v1/attendance?action=${action}`, {}, { auth: true });
      setItems((prev) => {
        const next = [...prev];
        const idx = next.findIndex((a) => new Date(a.date).toLocaleDateString('sv-SE') === new Date(updated.date).toLocaleDateString('sv-SE'));
        if (idx >= 0) next[idx] = { ...next[idx], ...updated } as Attendance;
        else next.unshift(updated as any);
        return next;
      });
    } catch (e: any) {
      setError(e?.message || 'Failed to update attendance');
      fetch('/api/log', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ level: 'error', message: 'employee_action_failed', context: { action, error: e?.message } }),
      }).catch(() => {});
    } finally {
      setLoading(false);
    }
  };

  const toggleWorkMode = async () => {
    try {
      setLoading(true);
      const next = today?.work_mode === 'remote' ? 'office' : 'remote';
      const updated = await apiClient.post<Attendance>('/api/v1/attendance?action=workmode', { mode: next }, { auth: true });
      setItems((prev) => {
        const nextArr = [...prev];
        const idx = nextArr.findIndex((a) => new Date(a.date).toLocaleDateString('sv-SE') === new Date(updated.date).toLocaleDateString('sv-SE'));
        if (idx >= 0) nextArr[idx] = { ...nextArr[idx], ...updated } as Attendance;
        else nextArr.unshift(updated as any);
        return nextArr;
      });
    } catch (e: any) {
      setError(e?.message || 'Failed to change work mode');
      fetch('/api/log', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ level: 'error', message: 'employee_workmode_failed', context: { error: e?.message } }),
      }).catch(() => {});
    } finally {
      setLoading(false);
    }
  };

  const logout = async () => {
    await supabase.auth.signOut();
    router.push('/login');
  };

  return (
    <div className="relative flex min-h-screen flex-col bg-neutral-50 text-[#141414]">
      <Header />
      <div className="px-6 md:px-10 flex flex-1 justify-center py-5">
        <div className="flex flex-col max-w-[960px] w-full">
          <div className="flex flex-wrap justify-between gap-3 p-4">
            <p className="text-[#141414] text-[32px] font-bold leading-tight">Dashboard</p>
            <button onClick={logout} className="rounded-full h-10 px-4 bg-[#ededed] text-[#141414] text-sm font-bold">Logout</button>
          </div>

          <h2 className="text-[#141414] text-[22px] font-bold leading-tight tracking-[-0.015em] px-4 pb-3 pt-5">Today's Clock In/Out</h2>
          <div className="flex items-center gap-4 bg-neutral-50 px-4 min-h-[72px] py-2">
            <div className="text-[#141414] flex items-center justify-center rounded-lg bg-[#ededed] shrink-0 size-12">
              <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" fill="currentColor" viewBox="0 0 256 256">
                <path d="M128,24A104,104,0,1,0,232,128,104.11,104.11,0,0,0,128,24Zm0,192a88,88,0,1,1,88-88A88.1,88.1,0,0,1,128,216Zm64-88a8,8,0,0,1-8,8H128a8,8,0,0,1-8-8V72a8,8,0,0,1,16,0v48h48A8,8,0,0,1,192,128Z" />
              </svg>
            </div>
            <div className="flex flex-col justify-center">
              <p className="text-[#141414] text-base font-medium leading-normal line-clamp-1">
                {today?.clock_in ? formatTime(today.clock_in) : '-'}
              </p>
              <p className="text-neutral-500 text-sm leading-normal line-clamp-2">{statusLabel}</p>
            </div>
          </div>

          <div className="flex justify-stretch">
            <div className="flex flex-1 gap-3 flex-wrap px-4 py-3 justify-between">
              {(!today || !today.clock_in) && (
                <button onClick={() => doClock('in')} disabled={loading}
                  className="flex min-w-[84px] max-w-[480px] items-center justify-center overflow-hidden rounded-full h-10 px-4 bg-[#999999] text-[#141414] text-sm font-bold">
                  <span className="truncate">出勤</span>
                </button>
              )}
              {/* 現在休憩・外出中でなく、かつ未完了（1回のみ） */}
              {today?.clock_in && !today.clock_out && !(today?.break_start && !today?.break_end) && !(today?.out_start && !today?.out_end) && !today?.break_end && (
                <button onClick={() => doAction('break_start')} disabled={loading}
                  className="flex min-w-[84px] max-w-[480px] items-center justify-center overflow-hidden rounded-full h-10 px-4 bg-[#ededed] text-[#141414] text-sm font-bold">
                  <span className="truncate">休憩開始</span>
                </button>
              )}
              {today?.clock_in && !today.clock_out && today?.break_start && !today?.break_end && (
                <button onClick={() => doAction('break_end')} disabled={loading}
                  className="flex min-w-[84px] max-w-[480px] items-center justify-center overflow-hidden rounded-full h-10 px-4 bg-[#ededed] text-[#141414] text-sm font-bold">
                  <span className="truncate">休憩終了</span>
                </button>
              )}
              {today?.clock_in && !today.clock_out && !(today?.break_start && !today?.break_end) && !(today?.out_start && !today?.out_end) && !today?.out_end && (
                <button onClick={() => doAction('out')} disabled={loading}
                  className="flex min-w-[84px] max-w-[480px] items-center justify-center overflow-hidden rounded-full h-10 px-4 bg-[#ededed] text-[#141414] text-sm font-bold">
                  <span className="truncate">外出</span>
                </button>
              )}
              {today?.clock_in && !today.clock_out && today?.out_start && !today?.out_end && (
                <button onClick={() => doAction('return')} disabled={loading}
                  className="flex min-w-[84px] max-w-[480px] items-center justify-center overflow-hidden rounded-full h-10 px-4 bg-[#ededed] text-[#141414] text-sm font-bold">
                  <span className="truncate">戻り</span>
                </button>
              )}
              {today?.clock_in && !today.clock_out && (
                <button onClick={() => doClock('out')} disabled={loading || (today?.break_start && !today?.break_end) || (today?.out_start && !today?.out_end)}
                  className="flex min-w-[84px] max-w-[480px] items-center justify-center overflow-hidden rounded-full h-10 px-4 bg-[#999999] text-[#141414] text-sm font-bold disabled:opacity-60">
                  <span className="truncate">退勤</span>
                </button>
              )}
            </div>
          </div>

          <div className="px-4 py-2">
            <button onClick={toggleWorkMode} disabled={loading}
              className="rounded-full h-9 px-4 bg-[#ededed] text-[#141414] text-sm font-bold">
              勤務形態: {today?.work_mode === 'remote' ? '在宅' : '出社'}（切替）
            </button>
          </div>

          <h2 className="text-[#141414] text-[22px] font-bold leading-tight tracking-[-0.015em] px-4 pb-3 pt-5">Clocked Events</h2>
          <div className="px-4 py-3">
            <div className="overflow-hidden rounded-xl border border-[#dbdbdb] bg-neutral-50">
              <table className="w-full">
                <thead>
                  <tr className="bg-neutral-50">
                    <th className="px-4 py-3 text-left text-[#141414] text-sm font-medium leading-normal w-1/2">Time</th>
                    <th className="px-4 py-3 text-left text-[#141414] text-sm font-medium leading-normal w-1/2">Event</th>
                  </tr>
                </thead>
                <tbody>
                  {today?.clock_in && (
                    <tr className="border-t border-t-[#dbdbdb]">
                      <td className="h-[56px] px-4 py-2 text-neutral-500 text-sm">{formatTime(today.clock_in)}</td>
                      <td className="h-[56px] px-4 py-2 text-neutral-500 text-sm">Clock In</td>
                    </tr>
                  )}
                  {today?.break_start && (
                    <tr className="border-t border-t-[#dbdbdb]">
                      <td className="h-[56px] px-4 py-2 text-neutral-500 text-sm">{formatTime(today.break_start)}</td>
                      <td className="h-[56px] px-4 py-2 text-neutral-500 text-sm">Break Start</td>
                    </tr>
                  )}
                  {today?.break_end && (
                    <tr className="border-t border-t-[#dbdbdb]">
                      <td className="h-[56px] px-4 py-2 text-neutral-500 text-sm">{formatTime(today.break_end)}</td>
                      <td className="h-[56px] px-4 py-2 text-neutral-500 text-sm">Break End</td>
                    </tr>
                  )}
                  {today?.out_start && (
                    <tr className="border-t border-t-[#dbdbdb]">
                      <td className="h-[56px] px-4 py-2 text-neutral-500 text-sm">{formatTime(today.out_start)}</td>
                      <td className="h-[56px] px-4 py-2 text-neutral-500 text-sm">Out</td>
                    </tr>
                  )}
                  {today?.out_end && (
                    <tr className="border-t border-t-[#dbdbdb]">
                      <td className="h-[56px] px-4 py-2 text-neutral-500 text-sm">{formatTime(today.out_end)}</td>
                      <td className="h-[56px] px-4 py-2 text-neutral-500 text-sm">Return</td>
                    </tr>
                  )}
                  {today?.clock_out && (
                    <tr className="border-t border-t-[#dbdbdb]">
                      <td className="h-[56px] px-4 py-2 text-neutral-500 text-sm">{formatTime(today.clock_out)}</td>
                      <td className="h-[56px] px-4 py-2 text-neutral-500 text-sm">Clock Out</td>
                    </tr>
                  )}
                  {!today && !loading && (
                    <tr className="border-t border-t-[#dbdbdb]">
                      <td className="h-[56px] px-4 py-2 text-neutral-500 text-sm" colSpan={2}>No records for today</td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>

          {error && (
            <div className="px-4 pb-4">
              <p className="text-sm text-red-600">{error}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

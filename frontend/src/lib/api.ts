const API_BASE_URL = process.env.NODE_ENV === 'development'
  ? 'http://localhost:8080'
  : process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

import { supabase } from '@/utils/supabase/client';

async function authHeader() {
  try {
    const { data } = await supabase.auth.getSession();
    const token = data.session?.access_token;
    return token ? { Authorization: `Bearer ${token}` } : {};
  } catch {
    return {};
  }
}

export class ApiClient {
  private baseUrl: string;

  constructor() {
    this.baseUrl = API_BASE_URL;
  }

  async get<T>(endpoint: string, opts: { auth?: boolean } = {}): Promise<T> {
    const auth = opts.auth ? await authHeader() : {};
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      method: 'GET',
      cache: 'no-store',
      headers: {
        'Content-Type': 'application/json',
        ...auth,
      },
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response.json();
  }

  async post<T>(endpoint: string, data: any, opts: { auth?: boolean } = {}): Promise<T> {
    const auth = opts.auth ? await authHeader() : {};
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...auth,
      },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response.json();
  }

  async put<T>(endpoint: string, data: any, opts: { auth?: boolean } = {}): Promise<T> {
    const auth = opts.auth ? await authHeader() : {};
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        ...auth,
      },
      body: JSON.stringify(data),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response.json();
  }

  async delete<T>(endpoint: string, opts: { auth?: boolean } = {}): Promise<T> {
    const auth = opts.auth ? await authHeader() : {};
    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        ...auth,
      },
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response.json();
  }
}

export const apiClient = new ApiClient();

'use client';

import { useState } from 'react';
import { apiClient } from '@/lib/api';

export default function Home() {
  const [connectionStatus, setConnectionStatus] = useState<string>('未確認');
  const [isLoading, setIsLoading] = useState(false);

  const testConnection = async () => {
    setIsLoading(true);
    try {
      const response = await apiClient.get('/health');
      setConnectionStatus('接続成功: ' + JSON.stringify(response));
    } catch (error) {
      setConnectionStatus('接続エラー: ' + (error as Error).message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="font-sans min-h-screen p-8 bg-gray-50">
      <div className="max-w-4xl mx-auto">
        <header className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            勤怠管理システム
          </h1>
          <p className="text-gray-600">
            Frontend (Next.js) と Backend (Go) の接続テスト
          </p>
        </header>

        <main className="space-y-6">
          <div className="bg-white rounded-lg shadow-md p-6">
            <h2 className="text-xl font-semibold mb-4">バックエンド接続テスト</h2>
            
            <div className="flex items-center gap-4 mb-4">
              <button
                onClick={testConnection}
                disabled={isLoading}
                className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {isLoading ? '接続中...' : '接続テスト'}
              </button>
            </div>

            <div className="bg-gray-100 rounded-md p-4">
              <p className="font-mono text-sm">
                <strong>接続状況:</strong> {connectionStatus}
              </p>
            </div>
          </div>

          <div className="grid md:grid-cols-2 gap-6">
            <div className="bg-white rounded-lg shadow-md p-6">
              <h3 className="text-lg font-semibold mb-3">システム情報</h3>
              <ul className="space-y-2 text-sm">
                <li><strong>Frontend:</strong> Next.js 15 + React 19</li>
                <li><strong>Backend:</strong> Go + Echo Framework</li>
                <li><strong>Database:</strong> PostgreSQL (Supabase Local)</li>
                <li><strong>Frontend Port:</strong> 3000</li>
                <li><strong>Backend Port:</strong> 8080</li>
                <li><strong>Database Port:</strong> 54322</li>
              </ul>
            </div>

            <div className="bg-white rounded-lg shadow-md p-6">
              <h3 className="text-lg font-semibold mb-3">実行手順</h3>
              <ol className="space-y-2 text-sm list-decimal list-inside">
                <li>Supabaseローカル開発環境を起動</li>
                <li>Goバックエンドサーバーを起動</li>
                <li>Next.jsフロントエンドを起動</li>
                <li>接続テストボタンをクリック</li>
              </ol>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}

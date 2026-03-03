'use client';

import { useEffect, useState } from 'react';

export default function LoginPage() {
  const [authUrl, setAuthUrl] = useState<string>('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchAuthUrl = async () => {
      try {
        const response = await fetch('http://localhost:8080/api/auth/login', {
          method: 'POST',
        });

        if (!response.ok) {
          throw new Error(`HTTP error! status: ${response.status}`);
        }

        const data = await response.json();
        setAuthUrl(data.auth_url);
      } catch (error) {
        console.error('Failed to fetch auth URL:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchAuthUrl();
  }, []);

  if (loading) {
    return <div className="flex justify-center items-center h-screen bg-slate-900 text-white">読み込み中...</div>;
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 flex flex-col">
      {/* Header */}
      <header className="bg-black/50 backdrop-blur-sm border-b border-slate-700">
        <div className="max-w-6xl mx-auto px-6 py-4 flex justify-between items-center">
          <div className="flex items-center gap-2">
            <span className="text-xl font-bold text-white">DevMetrics Hub</span>
          </div>
        </div>
      </header>

      {/* Login Form */}
      <main className="flex-1 flex justify-center items-center">
        <div className="w-full max-w-md px-6">
          <div className="bg-slate-800/50 backdrop-blur border border-slate-700 rounded-lg p-8">
            <h1 className="text-3xl font-bold text-white mb-2">DevMetrics Hub</h1>
            <p className="text-slate-400 mb-8">GitHub activity を分析・可視化</p>

            <a
              href={authUrl || '#'}
              className="w-full bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 text-white py-3 rounded-lg font-semibold transition flex items-center justify-center gap-2"
            >
              <svg
                className="w-5 h-5"
                fill="currentColor"
                viewBox="0 0 20 20"
              >
                <path d="M10 0C4.477 0 0 4.477 0 10c0 4.42 2.865 8.17 6.839 9.49.5.092.682-.217.682-.482 0-.237-.008-.868-.013-1.703-2.782.603-3.369-1.343-3.369-1.343-.454-1.156-1.11-1.463-1.11-1.463-.908-.62.069-.608.069-.608 1.003.07 1.531 1.03 1.531 1.03.891 1.529 2.341 1.544 2.914 1.186.09-.923.349-1.543.635-1.897-2.22-.253-4.555-1.11-4.555-4.943 0-1.091.39-1.984 1.029-2.683-.103-.253-.446-1.27.098-2.647 0 0 .84-.269 2.75 1.025A9.578 9.578 0 0110 4.836c.85.004 1.705.114 2.504.336 1.909-1.294 2.747-1.025 2.747-1.025.546 1.377.203 2.394.1 2.647.64.699 1.028 1.592 1.028 2.683 0 3.842-2.339 4.687-4.566 4.935.359.309.678.919.678 1.852 0 1.336-.012 2.415-.012 2.743 0 .267.18.578.688.48C17.137 18.167 20 14.418 20 10c0-5.523-4.477-10-10-10z" />
              </svg>
              GitHub でログイン
            </a>

            <div className="mt-6 pt-6 border-t border-slate-700">
              <p className="text-slate-400 text-sm text-center">
                初めてですか？ GitHub アカウントでログインして開始してください
              </p>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
}
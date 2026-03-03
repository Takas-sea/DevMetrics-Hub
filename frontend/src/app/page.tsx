'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';

export default function Home() {
  const [authUrl, setAuthUrl] = useState<string>('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchAuthUrl = async () => {
      try {
        const response = await fetch('http://localhost:8080/api/auth/login', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
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
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900">
      {/* Header */}
      <header className="bg-black/50 backdrop-blur-sm border-b border-slate-700">
        <div className="max-w-6xl mx-auto px-6 py-4 flex justify-between items-center">
          <div className="flex items-center gap-2">
            <span className="text-xl font-bold text-white">DevMetrics Hub</span>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <main className="max-w-6xl mx-auto px-6">
        <section className="py-20 text-center">
          <h1 className="text-5xl md:text-6xl font-bold text-white mb-6">
            GitHub Activity を
            <span className="block mt-2 bg-gradient-to-r from-blue-400 to-purple-500 bg-clip-text text-transparent">
              可視化・分析
            </span>
          </h1>

          <p className="text-xl text-slate-300 mb-8 max-w-2xl mx-auto">
            GitHub の活動を自動で収集・分析し、あなたのエンジニアとしての成長を可視化するダッシュボード
          </p>

          <a
            href={authUrl || '#'}
            className="inline-block bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 text-white px-8 py-4 rounded-lg text-lg font-semibold transition"
          >
            GitHub でログイン
          </a>
        </section>
      </main>
    </div>
  );
}

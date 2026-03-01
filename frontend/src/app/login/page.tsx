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
    return <div className="flex justify-center items-center h-screen">読み込み中...</div>;
  }

  return (
    <div className="flex justify-center items-center h-screen bg-gray-100">
      <div className="bg-white p-8 rounded-lg shadow-lg">
        <h1 className="text-2xl font-bold mb-4">DevMetrics Hub</h1>
        <p className="text-gray-600 mb-6">GitHub で ログイン</p>
        <a
          href={authUrl}
          className="bg-black text-white px-4 py-2 rounded hover:bg-gray-800"
        >
          GitHub でログイン
        </a>
      </div>
    </div>
  );
}
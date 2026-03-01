'use client';

import { useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';

export default function CallbackPage() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string>('');

  useEffect(() => {
    const handleCallback = async () => {
      try {
        const code = searchParams.get('code');
        if (!code) {
          setError('認可コードがありません');
          setLoading(false);
          return;
        }

        const response = await fetch(
          `http://localhost:8080/api/auth/callback?code=${code}`,
          { method: 'GET' }
        );

        if (!response.ok) {
          throw new Error('認証に失敗しました');
        }

        const data = await response.json();
        localStorage.setItem('token', data.token);
        localStorage.setItem('user', JSON.stringify(data.user));

        router.push('/dashboard');
      } catch (err) {
        setError(err instanceof Error ? err.message : '認証エラー');
        setLoading(false);
      }
    };

    handleCallback();
  }, [searchParams, router]);

  if (loading) {
    return <div className="flex justify-center items-center h-screen">認証処理中...</div>;
  }

  if (error) {
    return (
      <div className="flex justify-center items-center h-screen bg-gray-100">
        <div className="bg-white p-8 rounded-lg shadow-lg">
          <h1 className="text-2xl font-bold mb-4 text-red-600">エラー</h1>
          <p className="text-gray-600 mb-6">{error}</p>
          <a href="/login" className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700">
            ログインに戻る
          </a>
        </div>
      </div>
    );
  }

  return null;
}
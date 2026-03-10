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
        const oauthError = searchParams.get('error');
        if (oauthError) {
          setError('GitHub認証がキャンセルされました');
          setLoading(false);
          return;
        }

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
        const apiUser = data.user ?? {};
        const normalizedUser = {
          id: apiUser.id ?? apiUser.ID ?? '',
          username: apiUser.username ?? apiUser.Username ?? '',
          email: apiUser.email ?? apiUser.Email ?? '',
          avatar_url: apiUser.avatar_url ?? apiUser.AvatarURL ?? '',
          bio: apiUser.bio ?? apiUser.Bio ?? '',
        };

        localStorage.setItem('token', data.token);
        localStorage.setItem('user', JSON.stringify(normalizedUser));

        router.replace('/dashboard');
      } catch (err) {
        setError(err instanceof Error ? err.message : '認証エラー');
        setLoading(false);
      }
    };

    handleCallback();
  }, [searchParams, router]);

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 flex justify-center items-center px-6">
        <div className="bg-slate-800/50 backdrop-blur border border-slate-700 rounded-lg p-8 w-full max-w-md text-center">
          <h1 className="text-2xl font-bold text-white mb-3">認証処理中...</h1>
          <p className="text-slate-300">GitHubアカウントを確認しています</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 flex justify-center items-center px-6">
        <div className="bg-slate-800/50 backdrop-blur border border-slate-700 rounded-lg p-8 w-full max-w-md">
          <h1 className="text-2xl font-bold mb-4 text-red-400">認証エラー</h1>
          <p className="text-slate-300 mb-6">{error}</p>
          <a
            href="/login"
            className="inline-flex items-center justify-center w-full bg-gradient-to-r from-blue-500 to-purple-600 hover:from-blue-600 hover:to-purple-700 text-white py-3 rounded-lg font-semibold transition"
          >
            ログインに戻る
          </a>
        </div>
      </div>
    );
  }

  return null;
}
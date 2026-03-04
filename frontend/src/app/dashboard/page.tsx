'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Image from 'next/image';

interface User {
  id: string;
  username: string;
  email: string;
  avatar_url: string;
  bio: string;
}

interface Activity {
  date: string;
  count: number;
}

export default function DashboardPage() {
  const router = useRouter();
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const userStr = localStorage.getItem('user');
    const token = localStorage.getItem('token');

    if (!token || !userStr) {
      router.replace('/login');
      return;
    }

    try {
      const userData = JSON.parse(userStr);
      setUser(userData);
    } catch {
      router.replace('/login');
    } finally {
      setLoading(false);
    }
  }, [router]);

  const handleLogout = () => {
    localStorage.removeItem('user');
    localStorage.removeItem('token');
    router.replace('/login');
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 flex justify-center items-center">
        <div className="bg-slate-800/50 backdrop-blur border border-slate-700 rounded-lg p-8">
          <h1 className="text-2xl font-bold text-white">読み込み中...</h1>
        </div>
      </div>
    );
  }

  if (!user) {
    return null;
  }

  const sampleActivities: Activity[] = [
    { date: '2024-03-04', count: 15 },
    { date: '2024-03-03', count: 8 },
    { date: '2024-03-02', count: 12 },
    { date: '2024-03-01', count: 6 },
    { date: '2024-02-28', count: 20 },
  ];

  const totalContributions = sampleActivities.reduce((sum, a) => sum + a.count, 0);
  const maxActivity = Math.max(...sampleActivities.map(a => a.count));

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900">
      {/* Header */}
      <header className="bg-black/50 backdrop-blur-sm border-b border-slate-700 sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-6 py-4 flex justify-between items-center">
          <span className="text-xl font-bold text-white">DevMetrics Hub</span>
          <button
            onClick={handleLogout}
            className="text-slate-300 hover:text-white transition"
          >
            ログアウト
          </button>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-6 py-8">
        {/* User Info Card */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
          <div className="md:col-span-1 bg-slate-800/50 backdrop-blur border border-slate-700 rounded-lg p-6">
            <div className="flex flex-col items-center">
              <Image
                src={user.avatar_url}
                alt={user.username}
                width={96}
                height={96}
                className="rounded-full mb-4 border-2 border-slate-600"
              />
              <h2 className="text-2xl font-bold text-white mb-2">{user.username}</h2>
              <p className="text-slate-400 text-sm mb-4 text-center">{user.bio || 'プロフィール情報なし'}</p>
              <p className="text-slate-400 text-xs">{user.email}</p>
            </div>
          </div>

          {/* Stats Cards */}
          <div className="md:col-span-2 grid grid-cols-2 gap-4">
            <div className="bg-gradient-to-br from-blue-500/10 to-purple-500/10 border border-blue-500/20 rounded-lg p-6">
              <p className="text-slate-400 text-sm mb-2">今週の貢献</p>
              <p className="text-4xl font-bold text-blue-400">{totalContributions}</p>
            </div>
            <div className="bg-gradient-to-br from-purple-500/10 to-pink-500/10 border border-purple-500/20 rounded-lg p-6">
              <p className="text-slate-400 text-sm mb-2">過去5日の平均</p>
              <p className="text-4xl font-bold text-purple-400">
                {Math.round(totalContributions / sampleActivities.length)}
              </p>
            </div>
          </div>
        </div>

        {/* Activity Chart */}
        <div className="bg-slate-800/50 backdrop-blur border border-slate-700 rounded-lg p-6 mb-8">
          <h3 className="text-xl font-bold text-white mb-6">アクティビティ</h3>

          <div className="flex items-end justify-between gap-2 mb-6" style={{ height: '200px' }}>
            {sampleActivities.map((activity) => (
              <div
                key={activity.date}
                className="flex-1 flex flex-col items-center"
                title={`${activity.date}: ${activity.count} 件`}
              >
                <div className="w-full bg-blue-400/5 border border-blue-400/20 rounded-t-lg relative group hover:bg-blue-400/10 transition"
                  style={{
                    height: `${(activity.count / maxActivity) * 180}px`,
                  }}
                >
                  <div className="absolute -top-8 left-1/2 -translate-x-1/2 opacity-0 group-hover:opacity-100 transition whitespace-nowrap bg-slate-900 px-3 py-1 rounded text-sm text-white border border-slate-600">
                    {activity.count}
                  </div>
                </div>
                <p className="text-xs text-slate-400 mt-3">{activity.date.slice(5)}</p>
              </div>
            ))}
          </div>
        </div>

        {/* Recent Repositories (Placeholder) */}
        <div className="bg-slate-800/50 backdrop-blur border border-slate-700 rounded-lg p-6">
          <h3 className="text-xl font-bold text-white mb-6">最近のアクティビティ</h3>

          <div className="space-y-4">
            {['Repository 1', 'Repository 2', 'Repository 3'].map((repo, idx) => (
              <div
                key={idx}
                className="flex items-center justify-between p-4 bg-slate-700/30 rounded-lg border border-slate-600 hover:border-slate-500 transition"
              >
                <div>
                  <p className="text-white font-semibold">{repo}</p>
                  <p className="text-slate-400 text-sm">
                    {sampleActivities[idx]?.count || 0} commits
                  </p>
                </div>
                <span className="px-3 py-1 bg-blue-500/20 text-blue-300 text-xs rounded-full border border-blue-400/30">
                  Active
                </span>
              </div>
            ))}
          </div>
        </div>
      </main>
    </div>
  );
}

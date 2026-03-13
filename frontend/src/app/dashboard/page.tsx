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

interface RepoActivity {
  name: string;
  commits: number;
}

interface ActivityApiResponse {
  daily: Activity[];
  summary: {
    total_contributions: number;
    average_daily: number;
  };
  recent_repositories: RepoActivity[];
}

const periodOptions = [5, 7, 30] as const;

function formatDateKey(date: Date): string {
  const year = date.getFullYear();
  const month = `${date.getMonth() + 1}`.padStart(2, '0');
  const day = `${date.getDate()}`.padStart(2, '0');
  return `${year}-${month}-${day}`;
}

function createFallbackActivities(days: number): Activity[] {
  const today = new Date();
  const result: Activity[] = [];

  for (let i = days - 1; i >= 0; i -= 1) {
    const date = new Date(today);
    date.setDate(today.getDate() - i);
    result.push({
      date: formatDateKey(date),
      count: 0,
    });
  }

  return result;
}

const fallbackRepositories: RepoActivity[] = [
  { name: 'Repository 1', commits: 15 },
  { name: 'Repository 2', commits: 8 },
  { name: 'Repository 3', commits: 12 },
];

export default function DashboardPage() {
  const router = useRouter();
  const [selectedDays, setSelectedDays] = useState<number>(5);
  const [user, setUser] = useState<User | null>(null);
  const [activities, setActivities] = useState<Activity[]>(createFallbackActivities(5));
  const [repositories, setRepositories] = useState<RepoActivity[]>(fallbackRepositories);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const userStr = localStorage.getItem('user');
    const token = localStorage.getItem('token');

    if (!token || !userStr) {
      router.replace('/login');
      return;
    }

    try {
      const rawUser = JSON.parse(userStr);
      const normalizedUser: User = {
        id: rawUser.id ?? rawUser.ID ?? '',
        username: rawUser.username ?? rawUser.Username ?? '',
        email: rawUser.email ?? rawUser.Email ?? '',
        avatar_url: rawUser.avatar_url ?? rawUser.AvatarURL ?? '',
        bio: rawUser.bio ?? rawUser.Bio ?? '',
      };
      setUser(normalizedUser);

      setActivities(createFallbackActivities(selectedDays));
      setRepositories(fallbackRepositories);

      fetch(`http://localhost:8080/api/activities/me?days=${selectedDays}`, {
        method: 'GET',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      })
        .then((response) => {
          if (!response.ok) {
            throw new Error('failed to fetch activities');
          }
          return response.json() as Promise<ActivityApiResponse>;
        })
        .then((data) => {
          if (Array.isArray(data.daily) && data.daily.length > 0) {
            setActivities(data.daily);
          }
          if (Array.isArray(data.recent_repositories) && data.recent_repositories.length > 0) {
            setRepositories(data.recent_repositories);
          }
        })
        .catch((error) => {
          console.error('Failed to fetch activities:', error);
        });
    } catch {
      router.replace('/login');
    } finally {
      setLoading(false);
    }
  }, [router, selectedDays]);

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

  const displayName = user.username?.trim() || 'GitHub User';
  const displayEmail = user.email?.trim() || 'メール未設定';
  const avatarUrl = user.avatar_url?.trim() || null;
  const avatarAlt = `${displayName} のアバター`;
  const initials = displayName.slice(0, 1).toUpperCase();

  const totalContributions = activities.reduce((sum, a) => sum + a.count, 0);
  const averageDaily = (totalContributions / Math.max(activities.length, 1)).toFixed(1);
  const maxActivity = Math.max(...activities.map((a) => a.count), 1);

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
              {avatarUrl ? (
                <Image
                  src={avatarUrl}
                  alt={avatarAlt}
                  width={96}
                  height={96}
                  className="rounded-full mb-4 border-2 border-slate-600"
                />
              ) : (
                <div className="w-24 h-24 rounded-full mb-4 border-2 border-slate-600 bg-slate-700 flex items-center justify-center text-white text-2xl font-bold">
                  {initials}
                </div>
              )}
              <h2 className="text-2xl font-bold text-white mb-2">{displayName}</h2>
              <p className="text-slate-400 text-sm mb-4 text-center">{user.bio || 'プロフィール情報なし'}</p>
              <p className="text-slate-400 text-xs">{displayEmail}</p>
            </div>
          </div>

          {/* Stats Cards */}
          <div className="md:col-span-2 space-y-4">
            <div className="flex justify-end">
              <label className="text-slate-400 text-sm flex items-center gap-2">
                期間
                <select
                  value={selectedDays}
                  onChange={(event) => setSelectedDays(Number(event.target.value))}
                  className="bg-slate-800 border border-slate-600 text-white rounded px-2 py-1"
                >
                  {periodOptions.map((days) => (
                    <option key={days} value={days}>
                      {days}日
                    </option>
                  ))}
                </select>
              </label>
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="bg-gradient-to-br from-blue-500/10 to-purple-500/10 border border-blue-500/20 rounded-lg p-6">
                <p className="text-slate-400 text-sm mb-2">過去{selectedDays}日の貢献</p>
                <p className="text-4xl font-bold text-blue-400">{totalContributions}</p>
              </div>
              <div className="bg-gradient-to-br from-purple-500/10 to-pink-500/10 border border-purple-500/20 rounded-lg p-6">
                <p className="text-slate-400 text-sm mb-2">過去{selectedDays}日の平均</p>
                <p className="text-4xl font-bold text-purple-400">{averageDaily}</p>
              </div>
            </div>
          </div>
        </div>

        {/* Activity Chart */}
        <div className="bg-slate-800/50 backdrop-blur border border-slate-700 rounded-lg p-6 mb-8">
          <h3 className="text-xl font-bold text-white mb-6">アクティビティ（過去{selectedDays}日）</h3>

          <div className="flex items-end justify-between gap-2 mb-6" style={{ height: '200px' }}>
            {activities.map((activity) => (
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
            {repositories.map((repo) => (
              <div
                key={repo.name}
                className="flex items-center justify-between p-4 bg-slate-700/30 rounded-lg border border-slate-600 hover:border-slate-500 transition"
              >
                <div>
                  <p className="text-white font-semibold">{repo.name}</p>
                  <p className="text-slate-400 text-sm">
                    {repo.commits} commits
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

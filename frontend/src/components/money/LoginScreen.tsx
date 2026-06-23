import { FormEvent, useState } from 'react';
import { Loader2, Lock, Mountain, ShieldCheck } from 'lucide-react';
import { useAuth } from '../../contexts/AuthContext';

export function LoginScreen() {
  const { login, isLoading, error } = useAuth();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [localError, setLocalError] = useState<string | null>(null);

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    setLocalError(null);
    if (!email.trim() || !password) {
      setLocalError('Email and password are required.');
      return;
    }
    try {
      await login(email.trim(), password);
    } catch {
      setLocalError('Unable to sign in.');
    }
  };

  return (
    <main className="min-h-screen bg-slate-950 text-white">
      <div className="absolute inset-0 bg-[radial-gradient(circle_at_top_left,_rgba(34,197,94,0.25),_transparent_32rem),radial-gradient(circle_at_bottom_right,_rgba(14,165,233,0.18),_transparent_28rem)]" />
      <div className="relative mx-auto flex min-h-screen max-w-5xl flex-col justify-center px-5 py-10">
        <div className="mb-8 max-w-xl">
          <div className="mb-5 inline-flex items-center gap-2 rounded-full border border-emerald-300/20 bg-emerald-400/10 px-3 py-1 text-sm text-emerald-100">
            <ShieldCheck className="h-4 w-4" />
            Private developer toolkit
          </div>
          <div className="flex items-center gap-3">
            <div className="rounded-3xl bg-white/10 p-4 shadow-2xl shadow-emerald-950/40 ring-1 ring-white/10">
              <Mountain className="h-9 w-9 text-emerald-200" />
            </div>
            <div>
              <p className="text-sm uppercase tracking-[0.35em] text-emerald-200">Money Creek</p>
              <h1 className="text-4xl font-black tracking-tight sm:text-6xl">Crag Toolkit</h1>
            </div>
          </div>
          <p className="mt-5 text-lg leading-8 text-slate-300">
            Trails, topographic sketches, points of interest, notes, and images for trusted crag development work.
          </p>
        </div>

        <form onSubmit={handleSubmit} className="w-full max-w-md rounded-[2rem] border border-white/10 bg-white/10 p-5 shadow-2xl shadow-black/40 backdrop-blur-xl sm:p-7">
          <div className="mb-6 flex items-center gap-3">
            <div className="rounded-2xl bg-slate-950/60 p-3">
              <Lock className="h-5 w-5 text-emerald-200" />
            </div>
            <div>
              <h2 className="text-xl font-bold">Sign in</h2>
              <p className="text-sm text-slate-300">Use the app auth credentials seeded for the toolkit.</p>
            </div>
          </div>

          <label className="mb-4 block">
            <span className="mb-2 block text-sm font-medium text-slate-200">Email</span>
            <input
              autoComplete="email"
              type="email"
              value={email}
              onChange={event => setEmail(event.target.value)}
              className="w-full rounded-2xl border border-white/10 bg-slate-950/60 px-4 py-3 text-white outline-none ring-emerald-300/60 transition placeholder:text-slate-500 focus:ring-2"
              placeholder="developer@example.com"
            />
          </label>

          <label className="mb-5 block">
            <span className="mb-2 block text-sm font-medium text-slate-200">Password</span>
            <input
              autoComplete="current-password"
              type="password"
              value={password}
              onChange={event => setPassword(event.target.value)}
              className="w-full rounded-2xl border border-white/10 bg-slate-950/60 px-4 py-3 text-white outline-none ring-emerald-300/60 transition placeholder:text-slate-500 focus:ring-2"
              placeholder="••••••••"
            />
          </label>

          {(localError || error) && (
            <div className="mb-4 rounded-2xl border border-red-300/30 bg-red-500/10 px-4 py-3 text-sm text-red-100">
              {localError || error}
            </div>
          )}

          <button
            type="submit"
            disabled={isLoading}
            className="flex w-full items-center justify-center gap-2 rounded-2xl bg-emerald-300 px-4 py-3 font-bold text-slate-950 shadow-lg shadow-emerald-950/30 transition hover:bg-emerald-200 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {isLoading && <Loader2 className="h-4 w-4 animate-spin" />}
            Enter toolkit
          </button>
        </form>
      </div>
    </main>
  );
}

import { ReactNode } from 'react';
import { BookOpen, Eye, EyeOff, LogOut, Map as MapIcon, RefreshCw, Users } from 'lucide-react';
import { BinderTabs } from './BinderTabs';
import { BinderSection } from './binderSections';
import { BinderBackground } from './BinderBackground';
import { MoneyCurrentUser } from '../../types/money';

interface FieldBinderShellProps {
  user?: MoneyCurrentUser | null;
  activeSection: BinderSection;
  counts: Partial<Record<BinderSection, number>>;
  isFetching: boolean;
  focusMode: boolean;
  onSectionChange: (section: BinderSection) => void;
  onRefresh: () => void;
  onLogout: () => void;
  onToggleFocus: () => void;
  children: ReactNode;
}

export function FieldBinderShell({ user, activeSection, counts, isFetching, focusMode, onSectionChange, onRefresh, onLogout, onToggleFocus, children }: FieldBinderShellProps) {
  return (
    <main className="relative min-h-screen overflow-x-hidden p-3 text-slate-950 sm:p-4 lg:p-6">
      <BinderBackground />
      <div className="relative z-10 mx-auto max-w-[108rem]">
        <header className="mb-4 rounded-[1.75rem] border border-stone-200 bg-white/88 p-4 shadow-sm backdrop-blur sm:p-5">
          <div className="flex flex-col gap-4 xl:flex-row xl:items-center xl:justify-between">
            <div className="flex min-w-0 items-start gap-3">
              <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-2xl bg-teal-900 text-white shadow-sm sm:h-14 sm:w-14">
                <BookOpen className="h-7 w-7" />
              </div>
              <div className="min-w-0">
                <p className="flex items-center gap-2 text-xs font-semibold uppercase tracking-[0.2em] text-teal-900">Money Creek workspace</p>
                <h1 className="mt-1 text-2xl font-semibold tracking-tight text-slate-950 sm:text-4xl">Map, notes, photos, and trail records in one place</h1>
                <p className="mt-2 flex flex-wrap items-center gap-2 text-sm text-slate-600"><Users className="h-4 w-4" />{user?.display_name} · {user?.role}</p>
              </div>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <button onClick={onToggleFocus} className="rounded-2xl border border-stone-300 bg-white px-4 py-3 text-sm font-semibold text-slate-800 shadow-sm hover:border-teal-700 hover:text-teal-900" title="Toggle focus mode">
                {focusMode ? <EyeOff className="mr-2 inline h-4 w-4" /> : <Eye className="mr-2 inline h-4 w-4" />}{focusMode ? 'Full workspace' : 'Focus mode'}
              </button>
              <button onClick={onRefresh} className="rounded-2xl border border-stone-300 bg-white px-4 py-3 text-sm font-semibold text-slate-800 shadow-sm hover:border-teal-700 hover:text-teal-900" title="Refresh data">
                <RefreshCw className={`mr-2 inline h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />Refresh
              </button>
              <a href="/" className="rounded-2xl border border-stone-300 bg-white px-4 py-3 text-sm font-semibold text-slate-800 shadow-sm hover:border-teal-700 hover:text-teal-900" title="Dashboard"><MapIcon className="mr-2 inline h-4 w-4" />Home</a>
              <button onClick={onLogout} className="rounded-2xl border border-stone-300 bg-white px-4 py-3 text-sm font-semibold text-slate-800 shadow-sm hover:border-red-400 hover:text-red-700" title="Sign out"><LogOut className="mr-2 inline h-4 w-4" />Sign out</button>
            </div>
          </div>
        </header>

        <section className="overflow-hidden rounded-[1.75rem] border border-stone-200 bg-white/82 shadow-sm backdrop-blur lg:flex lg:min-h-[calc(100vh-11rem)]">
          <BinderTabs activeSection={activeSection} counts={counts} onChange={onSectionChange} />
          <div className="relative min-w-0 flex-1 bg-[#f7f5ee] p-3 sm:p-4">
            <div className="relative min-h-full">{children}</div>
          </div>
        </section>
      </div>
    </main>
  );
}

import { BinderSection, BINDER_SECTIONS } from './binderSections';

interface BinderTabsProps {
  activeSection: BinderSection;
  onChange: (section: BinderSection) => void;
  counts: Partial<Record<BinderSection, number>>;
}

export function BinderTabs({ activeSection, onChange, counts }: BinderTabsProps) {
  return (
    <nav className="sticky top-0 z-30 -mx-3 mb-3 overflow-x-auto border-b border-stone-200 bg-stone-50/95 px-3 py-2 backdrop-blur lg:static lg:mx-0 lg:mb-0 lg:w-56 lg:shrink-0 lg:overflow-visible lg:border-b-0 lg:border-r lg:border-stone-200 lg:bg-[#f8f6ef] lg:p-4">
      <div className="flex min-w-max gap-2 lg:min-w-0 lg:flex-col">
        {BINDER_SECTIONS.map(section => {
          const Icon = section.icon;
          const active = activeSection === section.id;
          return (
            <button
              key={section.id}
              type="button"
              onClick={() => onChange(section.id)}
              className={`flex min-w-28 items-center gap-3 rounded-2xl border px-3 py-2.5 text-left transition focus:outline-none focus:ring-2 focus:ring-teal-700/25 lg:min-w-0 ${active ? 'border-teal-800 bg-white text-slate-950 shadow-sm' : 'border-transparent text-slate-600 hover:border-stone-200 hover:bg-white/75 hover:text-slate-950'}`}
              aria-current={active ? 'page' : undefined}
            >
              <span className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-xl ${active ? 'bg-teal-800 text-white' : 'bg-stone-200 text-slate-700'}`}>
                <Icon className="h-4 w-4" />
              </span>
              <span className="min-w-0 flex-1">
                <span className="block text-sm font-semibold">{section.label}</span>
                <span className="hidden text-xs text-slate-500 lg:block">{section.kicker}</span>
              </span>
              <span className="rounded-full bg-stone-100 px-2 py-0.5 text-xs font-semibold text-slate-600">{counts[section.id] ?? 0}</span>
            </button>
          );
        })}
      </div>
    </nav>
  );
}

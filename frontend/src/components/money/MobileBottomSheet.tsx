import { ReactNode } from 'react';
import { ChevronDown, ChevronUp } from 'lucide-react';

interface MobileBottomSheetProps {
  title: string;
  subtitle?: string;
  collapsed: boolean;
  onToggle: () => void;
  children: ReactNode;
}

export function MobileBottomSheet({ title, subtitle, collapsed, onToggle, children }: MobileBottomSheetProps) {
  return (
    <aside className={`absolute inset-x-0 bottom-0 z-20 rounded-t-[2rem] border border-white/20 bg-slate-950/92 text-white shadow-2xl shadow-black/50 backdrop-blur-xl transition-all duration-300 md:inset-y-4 md:left-auto md:right-4 md:w-[27rem] md:rounded-[2rem] ${collapsed ? 'h-24 md:h-auto' : 'h-[72vh] md:h-auto'}`}>
      <div className="flex h-full flex-col overflow-hidden">
        <button onClick={onToggle} className="flex w-full items-center justify-between gap-4 px-5 py-4 text-left md:hidden">
          <div>
            <h2 className="font-bold">{title}</h2>
            {subtitle && <p className="mt-0.5 text-xs text-slate-400">{subtitle}</p>}
          </div>
          {collapsed ? <ChevronUp className="h-5 w-5 text-slate-300" /> : <ChevronDown className="h-5 w-5 text-slate-300" />}
        </button>
        <div className="hidden border-b border-white/10 px-5 py-4 md:block">
          <h2 className="font-bold">{title}</h2>
          {subtitle && <p className="mt-0.5 text-xs text-slate-400">{subtitle}</p>}
        </div>
        <div className={`${collapsed ? 'hidden md:block' : 'block'} min-h-0 flex-1 overflow-y-auto custom-scrollbar px-4 pb-5 md:px-5 md:pt-4`}>
          {children}
        </div>
      </div>
    </aside>
  );
}

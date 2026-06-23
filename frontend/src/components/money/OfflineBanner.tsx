import { WifiOff } from 'lucide-react';

export function OfflineBanner() {
  return (
    <div className="absolute left-3 right-3 top-20 z-30 rounded-2xl border border-amber-300 bg-amber-50/95 px-4 py-3 text-sm text-amber-900 shadow-lg backdrop-blur md:left-auto md:right-4 md:w-80">
      <div className="flex items-center gap-2 font-medium">
        <WifiOff className="h-4 w-4" />
        Offline mode
      </div>
      <p className="mt-1 text-xs text-amber-800">Viewing cached toolkit data. Edits and uploads need a network connection.</p>
    </div>
  );
}

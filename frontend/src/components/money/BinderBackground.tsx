export function BinderBackground() {
  return (
    <div className="pointer-events-none fixed inset-0 overflow-hidden bg-[#f4f1e8] text-slate-950" aria-hidden="true">
      <div className="absolute inset-0 bg-[linear-gradient(180deg,#f7f4ec_0%,#eee9dd_54%,#e5dfd1_100%)]" />
      <div className="absolute inset-x-0 top-0 h-64 bg-[radial-gradient(circle_at_25%_0%,rgba(78,107,91,0.16),transparent_32rem),radial-gradient(circle_at_82%_12%,rgba(36,88,87,0.12),transparent_26rem)]" />
      <div className="absolute inset-0 bg-[linear-gradient(90deg,rgba(31,41,55,0.035)_1px,transparent_1px),linear-gradient(rgba(31,41,55,0.028)_1px,transparent_1px)] bg-[size:48px_48px] opacity-45" />
    </div>
  );
}

import { Outlet, Link, useLocation } from 'react-router-dom';
import { useAuthStore } from '../features/auth/store';
import { SystemHaltButton } from '../features/system/components/SystemHaltButton';
import { useSystemMetrics } from '../features/system/hooks/useSystemControl';
import { StatusBadge } from '../shared/components/StatusBadge';

const MIN_ACTIVE_ACCOUNTS_TG = Number(import.meta.env.VITE_MIN_ACTIVE_ACCOUNTS ?? 3);
const MEMORY_PRESSURE_THRESHOLD = 0.80;

const ObservabilityBanners = () => {
  const { data: metrics } = useSystemMetrics();

  if (!metrics) return null;

  const isMemoryPressure = metrics.cascade_memory_usage_ratio > MEMORY_PRESSURE_THRESHOLD;
  const isPoolCritical = metrics.active_tg_accounts < MIN_ACTIVE_ACCOUNTS_TG;

  if (!isMemoryPressure && !isPoolCritical) return null;

  return (
    <div className="flex flex-col gap-0">
      {isMemoryPressure && (
        <div
          className="flex items-center gap-3 px-6 py-2.5 bg-orange-500/10 border-b border-orange-500/30 text-orange-400 text-sm font-mono"
          role="alert"
          aria-live="polite"
        >
          <span className="banner-pulse inline-block w-2.5 h-2.5 rounded-full bg-orange-400 shrink-0" />
          <span>
            <strong>ВНИМАНИЕ:</strong> Высокое потребление памяти (cgroup) —{' '}
            {Math.round(metrics.cascade_memory_usage_ratio * 100)}% использовано
          </span>
        </div>
      )}
      {isPoolCritical && (
        <div
          className="flex items-center gap-3 px-6 py-2.5 bg-danger/10 border-b border-danger/30 text-red-400 text-sm font-mono"
          role="alert"
          aria-live="assertive"
        >
          <span className="banner-pulse inline-block w-2.5 h-2.5 rounded-full bg-red-400 shrink-0" />
          <span>
            <strong>КРИТИЧЕСКИЙ СБОЙ ПУЛА:</strong> Нехватка активных аккаунтов —{' '}
            {metrics.active_tg_accounts} / {MIN_ACTIVE_ACCOUNTS_TG} минимально необходимых
          </span>
        </div>
      )}
    </div>
  );
};

export const DashboardLayout = () => {
  const location = useLocation();
  const logout = useAuthStore((state) => state.logout);
  const metrics = useSystemMetrics();

  const navItems = [
    { path: '/', label: 'Обзор', icon: '◈' },
    { path: '/campaigns', label: 'Кампании', icon: '⚡' },
    { path: '/pool', label: 'Пул Аккаунтов', icon: '◉' },
    { path: '/leads', label: 'Лиды (152-ФЗ)', icon: '◆' },
  ];

  const systemStatus = metrics.data?.system_status ?? 'OPERATIONAL';

  return (
    <div className="flex h-screen overflow-hidden text-neutral-100 bg-bg-base">
      {/* Sidebar */}
      <aside className="w-64 bg-bg-surface border-r border-neutral-800 flex flex-col shrink-0">
        <div className="p-6 border-b border-neutral-800">
          <h1 className="text-xl font-bold tracking-widest font-mono">
            CASCADE <span className="text-accent-cyan">PRO</span>
          </h1>
          <p className="text-xs text-neutral-500 mt-1 font-mono">Waterfall Routing System</p>
        </div>

        <nav className="flex-1 px-3 py-4 space-y-1" aria-label="Main navigation">
          {navItems.map((item) => {
            const isActive = item.path === '/'
              ? location.pathname === '/'
              : location.pathname.startsWith(item.path);
            return (
              <Link
                key={item.path}
                to={item.path}
                className={`flex items-center gap-3 px-4 py-2.5 rounded-lg transition-all text-sm ${
                  isActive
                    ? 'bg-accent-cyan/10 text-accent-cyan border border-accent-cyan/20 font-medium'
                    : 'text-neutral-400 hover:bg-neutral-800 hover:text-white border border-transparent'
                }`}
              >
                <span className="text-xs opacity-70">{item.icon}</span>
                {item.label}
              </Link>
            );
          })}
        </nav>

        <div className="p-4 border-t border-neutral-800">
          <div className="flex items-center gap-2 mb-3 px-2">
            <StatusBadge status={systemStatus} />
          </div>
          <button
            id="sidebar-logout-btn"
            onClick={logout}
            className="w-full text-left px-4 py-2 text-sm text-neutral-400 hover:text-white
              hover:bg-neutral-800 rounded-md transition-colors"
          >
            → Выйти
          </button>
        </div>
      </aside>

      {/* Main content area */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Observability Banners — above header */}
        <ObservabilityBanners />

        {/* Top Header */}
        <header className="h-16 bg-bg-surface border-b border-neutral-800 flex items-center justify-between px-8 shrink-0">
          <div className="text-xs font-mono text-neutral-500 flex items-center gap-4">
            <span>
              СТАТУС:{' '}
              <span className={
                systemStatus === 'OPERATIONAL' ? 'text-green-400' :
                systemStatus === 'DEGRADED' ? 'text-orange-400' : 'text-danger'
              }>
                {systemStatus}
              </span>
            </span>
            {metrics.data && (
              <span className="text-neutral-600">
                MEM: {Math.round(metrics.data.cascade_memory_usage_ratio * 100)}% |{' '}
                TG: {metrics.data.active_tg_accounts} акк. | Q: {metrics.data.queue_depth}
              </span>
            )}
          </div>
          <SystemHaltButton />
        </header>

        {/* Page content */}
        <main className="flex-1 overflow-auto p-8">
          <Outlet />
        </main>
      </div>
    </div>
  );
};

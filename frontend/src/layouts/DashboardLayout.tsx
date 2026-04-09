import { Outlet, Link, useLocation } from 'react-router-dom';
import { useAuthStore } from '../features/auth/store';
import { SystemHaltButton } from '../features/system/components/SystemHaltButton';

export const DashboardLayout = () => {
  const location = useLocation();
  const logout = useAuthStore((state) => state.logout);

  const navItems = [
    { path: '/', label: 'Обзор' },
    { path: '/campaigns', label: 'Кампании' },
    { path: '/pool', label: 'Пул Аккаунтов' },
  ];

  return (
    <div className="flex h-screen overflow-hidden text-neutral-100 bg-bg-base">
      <aside className="w-64 bg-bg-surface border-r border-neutral-800 flex flex-col">
        <div className="p-6">
          <h1 className="text-2xl font-bold tracking-wider">
            CASCADE <span className="text-accent-cyan">PRO</span>
          </h1>
        </div>
        <nav className="flex-1 px-4 space-y-2">
          {navItems.map((item) => {
            const isActive = location.pathname === item.path;
            return (
              <Link
                key={item.path}
                to={item.path}
                className={`block px-4 py-2 rounded-md transition-colors ${
                  isActive 
                    ? 'bg-accent-cyan/10 text-accent-cyan border border-accent-cyan/20' 
                    : 'text-neutral-400 hover:bg-neutral-800 hover:text-white'
                }`}
              >
                {item.label}
              </Link>
            );
          })}
        </nav>
        <div className="p-4 border-t border-neutral-800">
          <button onClick={logout} className="w-full text-left px-4 py-2 text-neutral-400 hover:text-white">
            Выйти
          </button>
        </div>
      </aside>

      <div className="flex-1 flex flex-col">
        <header className="h-16 bg-bg-surface border-b border-neutral-800 flex items-center justify-between px-8">
          <div className="text-sm font-mono text-neutral-500">
            STATUS: <span className="text-green-500">SYSTEM_OPERATIONAL</span>
          </div>
          <SystemHaltButton />
        </header>

        <main className="flex-1 overflow-auto p-8">
          <Outlet /> 
        </main>
      </div>
    </div>
  );
};

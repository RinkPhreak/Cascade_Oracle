import { useState } from 'react';
import { AccountGrid } from '../features/pool/components/AccountGrid';
import { AccountEventDrawer } from '../features/pool/components/AccountEventDrawer';
import { ImportAccountModal } from '../features/pool/components/AddAccountModal';
import { ProxyTab } from '../features/pool/components/ProxyTab';
import { useAccounts } from '../features/pool/hooks/useAccounts';
import { useProxies } from '../features/pool/hooks/useProxies';
import type { DtoTgAccount } from '../api/generated';

type Tab = 'accounts' | 'proxies';

export const PoolPage = () => {
  const [tab, setTab] = useState<Tab>('accounts');
  const [addModalOpen, setAddModalOpen] = useState(false);
  const [selectedAccount, setSelectedAccount] = useState<DtoTgAccount | null>(null);

  const { data: accounts = [], isLoading: accountsLoading } = useAccounts();
  const { data: proxies = [] } = useProxies();

  const activeCount = accounts.filter((a) => a.status === 'ACTIVE').length;
  const bannedCount = accounts.filter((a) => a.status === 'BANNED').length;

  return (
    <div className="flex flex-col gap-6">
      {/* Page Header */}
      <div className="flex items-start justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white tracking-tight">Пул Аккаунтов</h2>
          <p className="text-neutral-400 text-sm mt-1">
            Управление TG аккаунтами и мобильными прокси
          </p>
        </div>
        <button
          id="add-account-btn"
          onClick={() => setAddModalOpen(true)}
          className="flex items-center gap-2 px-4 py-2.5 bg-accent-cyan text-black font-bold text-sm
            rounded-lg hover:bg-accent-cyan-hover transition-colors"
        >
          <span>+</span> Добавить аккаунт
        </button>
      </div>

      {/* Stats row */}
      <div className="grid grid-cols-3 gap-4">
        {[
          { label: 'Всего аккаунтов', value: accounts.length, color: 'text-white' },
          { label: 'Активных', value: activeCount, color: 'text-green-400' },
          { label: 'Заблокированных', value: bannedCount, color: bannedCount > 0 ? 'text-danger' : 'text-neutral-400' },
        ].map(({ label, value, color }) => (
          <div key={label} className="bg-bg-surface border border-neutral-800 rounded-xl p-4">
            <p className="text-xs text-neutral-500 font-mono uppercase tracking-wider mb-1">{label}</p>
            <p className={`text-2xl font-bold font-mono ${color}`}>{value}</p>
          </div>
        ))}
      </div>

      {/* Tab navigation */}
      <div className="flex gap-1 border-b border-neutral-800">
        {(['accounts', 'proxies'] as const).map((t) => (
          <button
            key={t}
            id={`pool-tab-${t}`}
            onClick={() => setTab(t)}
            className={`px-5 py-3 text-sm font-medium transition-colors border-b-2 -mb-px ${
              tab === t
                ? 'border-accent-cyan text-accent-cyan'
                : 'border-transparent text-neutral-400 hover:text-white'
            }`}
          >
            {t === 'accounts' ? '◉ Аккаунты' : '◈ Прокси'}
          </button>
        ))}
      </div>

      {/* Tab content */}
      {tab === 'accounts' && (
        <AccountGrid
          accounts={accounts}
          proxies={proxies}
          isLoading={accountsLoading}
          onViewEvents={setSelectedAccount}
        />
      )}
      {tab === 'proxies' && <ProxyTab />}

      {/* Modals & Drawers */}
      <ImportAccountModal
        isOpen={addModalOpen}
        onClose={() => setAddModalOpen(false)}
      />
      <AccountEventDrawer
        account={selectedAccount}
        onClose={() => setSelectedAccount(null)}
      />
    </div>
  );
};

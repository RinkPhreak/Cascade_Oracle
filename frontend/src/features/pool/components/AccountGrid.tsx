import type { TgAccount } from '../../../api/extended-types';
import { StatusBadge } from '../../../shared/components/StatusBadge';

interface AccountGridProps {
  accounts: TgAccount[];
  proxies: { id: string; host: string; port: number; is_healthy: boolean }[];
  isLoading: boolean;
  onViewEvents: (account: TgAccount) => void;
}

const DailyProgress = ({ count, limit }: { count: number; limit: number }) => {
  const pct = Math.min((count / limit) * 100, 100);
  const isNearLimit = pct >= 80;
  return (
    <div className="flex items-center gap-2">
      <div className="w-20 bg-neutral-800 rounded-full h-1.5 overflow-hidden">
        <div
          className={`h-full rounded-full transition-all ${isNearLimit ? 'bg-orange-400' : 'bg-accent-cyan'}`}
          style={{ width: `${pct}%` }}
        />
      </div>
      <span className={`text-xs font-mono tabular-nums ${isNearLimit ? 'text-orange-400' : 'text-neutral-400'}`}>
        {count}/{limit}
      </span>
    </div>
  );
};

const ProxyCell = ({
  proxyId,
  proxies,
}: {
  proxyId: string | null;
  proxies: AccountGridProps['proxies'];
}) => {
  if (!proxyId) {
    return <span className="text-xs text-neutral-600 italic">Не привязан</span>;
  }
  const proxy = proxies.find((p) => p.id === proxyId);
  if (!proxy) {
    return <span className="text-xs text-neutral-600 font-mono">{proxyId.slice(0, 8)}…</span>;
  }
  return (
    <div className="flex items-center gap-2">
      <span
        className={`inline-block w-2 h-2 rounded-full shrink-0 ${proxy.is_healthy ? 'bg-green-400' : 'bg-danger'}`}
        title={proxy.is_healthy ? 'Healthy' : 'Unhealthy'}
      />
      <span className="text-xs font-mono text-neutral-300">{proxy.host}:{proxy.port}</span>
    </div>
  );
};

const SkeletonRow = () => (
  <tr className="border-t border-neutral-800">
    {[1, 2, 3, 4, 5].map((i) => (
      <td key={i} className="px-4 py-3">
        <div className="h-4 bg-neutral-800 rounded animate-pulse" style={{ width: `${40 + i * 12}%` }} />
      </td>
    ))}
  </tr>
);

export const AccountGrid = ({ accounts, proxies, isLoading, onViewEvents }: AccountGridProps) => {
  return (
    <div className="overflow-hidden rounded-xl border border-neutral-800">
      <table className="w-full text-sm" aria-label="Таблица TG аккаунтов">
        <thead>
          <tr className="bg-neutral-900/60 text-left">
            <th className="px-4 py-3 text-xs font-mono font-semibold text-neutral-400 uppercase tracking-wider">
              Телефон
            </th>
            <th className="px-4 py-3 text-xs font-mono font-semibold text-neutral-400 uppercase tracking-wider">
              Статус
            </th>
            <th className="px-4 py-3 text-xs font-mono font-semibold text-neutral-400 uppercase tracking-wider">
              Сообщения / сут
            </th>
            <th className="px-4 py-3 text-xs font-mono font-semibold text-neutral-400 uppercase tracking-wider">
              Прокси
            </th>
            <th className="px-4 py-3 text-xs font-mono font-semibold text-neutral-400 uppercase tracking-wider">
              Добавлен
            </th>
            <th className="px-4 py-3" />
          </tr>
        </thead>
        <tbody>
          {isLoading && (
            <>
              <SkeletonRow />
              <SkeletonRow />
              <SkeletonRow />
            </>
          )}

          {!isLoading && accounts.length === 0 && (
            <tr>
              <td colSpan={6} className="px-4 py-12 text-center text-neutral-500 text-sm">
                Нет аккаунтов в пуле. Добавьте первый аккаунт.
              </td>
            </tr>
          )}

          {!isLoading &&
            accounts.map((account) => (
              <tr
                key={account.id}
                className="border-t border-neutral-800 hover:bg-neutral-900/40 transition-colors"
              >
                <td className="px-4 py-3 font-mono text-white">{account.phone}</td>
                <td className="px-4 py-3">
                  <StatusBadge status={account.status} />
                </td>
                <td className="px-4 py-3">
                  <DailyProgress count={account.daily_count} limit={account.daily_limit} />
                </td>
                <td className="px-4 py-3">
                  <ProxyCell proxyId={account.proxy_id} proxies={proxies} />
                </td>
                <td className="px-4 py-3 text-neutral-500 text-xs">
                  {new Date(account.created_at).toLocaleDateString('ru-RU')}
                </td>
                <td className="px-4 py-3 text-right">
                  <button
                    id={`account-events-${account.id}`}
                    onClick={() => onViewEvents(account)}
                    className="text-xs text-neutral-400 hover:text-accent-cyan border border-neutral-700
                      hover:border-accent-cyan/50 px-3 py-1 rounded-md transition-colors"
                  >
                    События →
                  </button>
                </td>
              </tr>
            ))}
        </tbody>
      </table>
    </div>
  );
};

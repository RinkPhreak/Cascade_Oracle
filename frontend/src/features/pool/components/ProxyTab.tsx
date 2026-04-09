import { useState } from 'react';
import type { Proxy, CreateProxyRequest } from '../../../api/extended-types';
import { useProxies, useAddProxy, useDeleteProxy } from '../hooks/useProxies';
import { useAccounts } from '../hooks/useAccounts';
import { StatusBadge } from '../../../shared/components/StatusBadge';
import { ReassignProxyDialog } from './ReassignProxyDialog';

const AddProxyForm = ({ onDone }: { onDone: () => void }) => {
  const [form, setForm] = useState<CreateProxyRequest>({
    host: '',
    port: 1080,
    username: '',
    password: '',
    protocol: 'socks5',
  });
  const addProxy = useAddProxy();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    addProxy.mutate(form, { onSuccess: onDone });
  };

  const set = <K extends keyof CreateProxyRequest>(key: K, value: CreateProxyRequest[K]) =>
    setForm((prev) => ({ ...prev, [key]: value }));

  return (
    <form
      onSubmit={handleSubmit}
      className="mt-4 p-5 bg-neutral-900 border border-neutral-800 rounded-xl flex flex-col gap-4"
    >
      <h3 className="text-sm font-semibold text-white">Новый прокси</h3>
      <div className="grid grid-cols-3 gap-3">
        <div className="col-span-2 flex flex-col gap-1">
          <label htmlFor="proxy-host" className="text-xs text-neutral-400 font-mono">Хост *</label>
          <input
            id="proxy-host"
            required
            value={form.host}
            onChange={(e) => set('host', e.target.value)}
            placeholder="proxy.example.com"
            className="bg-neutral-800 border border-neutral-700 focus:border-accent-cyan rounded-md
              px-3 py-2 text-sm text-white outline-none transition-colors"
          />
        </div>
        <div className="flex flex-col gap-1">
          <label htmlFor="proxy-port" className="text-xs text-neutral-400 font-mono">Порт *</label>
          <input
            id="proxy-port"
            type="number"
            required
            min={1}
            max={65535}
            value={form.port}
            onChange={(e) => set('port', Number(e.target.value))}
            className="bg-neutral-800 border border-neutral-700 focus:border-accent-cyan rounded-md
              px-3 py-2 text-sm text-white outline-none transition-colors"
          />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-3">
        <div className="flex flex-col gap-1">
          <label htmlFor="proxy-user" className="text-xs text-neutral-400 font-mono">Логин</label>
          <input
            id="proxy-user"
            value={form.username}
            onChange={(e) => set('username', e.target.value)}
            className="bg-neutral-800 border border-neutral-700 focus:border-accent-cyan rounded-md
              px-3 py-2 text-sm text-white outline-none transition-colors"
          />
        </div>
        <div className="flex flex-col gap-1">
          <label htmlFor="proxy-pass" className="text-xs text-neutral-400 font-mono">Пароль</label>
          <input
            id="proxy-pass"
            type="password"
            value={form.password}
            onChange={(e) => set('password', e.target.value)}
            className="bg-neutral-800 border border-neutral-700 focus:border-accent-cyan rounded-md
              px-3 py-2 text-sm text-white outline-none transition-colors"
          />
        </div>
      </div>

      <div className="flex flex-col gap-1">
        <label htmlFor="proxy-protocol" className="text-xs text-neutral-400 font-mono">Протокол</label>
        <select
          id="proxy-protocol"
          value={form.protocol}
          onChange={(e) => set('protocol', e.target.value as CreateProxyRequest['protocol'])}
          className="bg-neutral-800 border border-neutral-700 focus:border-accent-cyan rounded-md
            px-3 py-2 text-sm text-white outline-none transition-colors"
        >
          <option value="socks5">SOCKS5</option>
          <option value="http">HTTP</option>
          <option value="mtproto">MTProto</option>
        </select>
      </div>

      <div className="flex gap-3 justify-end">
        <button
          type="button"
          onClick={onDone}
          className="px-4 py-2 text-sm text-neutral-400 hover:text-white border border-neutral-700 rounded-md transition-colors"
        >
          Отмена
        </button>
        <button
          type="submit"
          id="add-proxy-submit-btn"
          disabled={!form.host || addProxy.isPending}
          className="px-4 py-2 text-sm font-bold bg-accent-cyan/10 hover:bg-accent-cyan/20
            border border-accent-cyan text-accent-cyan rounded-md transition-colors
            disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-2"
        >
          {addProxy.isPending && (
            <span className="w-4 h-4 border-2 border-accent-cyan/30 border-t-accent-cyan rounded-full animate-spin" />
          )}
          Добавить прокси
        </button>
      </div>
    </form>
  );
};

export const ProxyTab = () => {
  const { data: proxies = [], isLoading } = useProxies();
  const { data: accounts = [] } = useAccounts();
  const deleteProxy = useDeleteProxy();

  const [showForm, setShowForm] = useState(false);
  const [reassignProxy, setReassignProxy] = useState<Proxy | null>(null);

  return (
    <div>
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <div>
          <h3 className="text-base font-semibold text-white">Мобильные прокси</h3>
          <p className="text-xs text-neutral-500 mt-0.5">Sticky Binding — каждый прокси жёстко привязан к аккаунту</p>
        </div>
        <button
          id="add-proxy-btn"
          onClick={() => setShowForm((v) => !v)}
          className="text-sm px-4 py-2 bg-accent-cyan/10 hover:bg-accent-cyan/20 border border-accent-cyan/50
            text-accent-cyan rounded-lg transition-colors font-medium"
        >
          {showForm ? 'Отмена' : '+ Добавить прокси'}
        </button>
      </div>

      {showForm && <AddProxyForm onDone={() => setShowForm(false)} />}

      {/* Proxy table */}
      <div className="mt-4 overflow-hidden rounded-xl border border-neutral-800">
        <table className="w-full text-sm" aria-label="Таблица прокси">
          <thead>
            <tr className="bg-neutral-900/60 text-left">
              {['Хост:Порт', 'Протокол', 'Привязан к аккаунту', 'Здоровье', 'Задержка', ''].map((h) => (
                <th key={h} className="px-4 py-3 text-xs font-mono font-semibold text-neutral-400 uppercase tracking-wider">
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {isLoading && (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center">
                  <div className="w-6 h-6 border-2 border-neutral-700 border-t-accent-cyan rounded-full animate-spin mx-auto" />
                </td>
              </tr>
            )}
            {!isLoading && proxies.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-12 text-center text-neutral-500 text-sm">
                  Прокси не добавлены.
                </td>
              </tr>
            )}
            {!isLoading &&
              proxies.map((proxy) => {
                const boundAccount = accounts.find((a) => a.id === proxy.bound_account_id);
                return (
                  <tr key={proxy.id} className="border-t border-neutral-800 hover:bg-neutral-900/40 transition-colors">
                    <td className="px-4 py-3 font-mono text-white">
                      {proxy.host}:{proxy.port}
                    </td>
                    <td className="px-4 py-3">
                      <span className="text-xs font-mono text-neutral-400 uppercase">{proxy.protocol}</span>
                    </td>
                    <td className="px-4 py-3">
                      {boundAccount ? (
                        <div className="flex items-center gap-2">
                          <span className="text-xs bg-accent-cyan/10 border border-accent-cyan/20 text-accent-cyan px-2 py-0.5 rounded font-mono">
                            🔒 {boundAccount.phone}
                          </span>
                        </div>
                      ) : (
                        <span className="text-xs text-neutral-600 italic">Не привязан</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      <StatusBadge status={proxy.is_healthy ? 'healthy' : 'unhealthy'} />
                    </td>
                    <td className="px-4 py-3 font-mono text-xs text-neutral-400">
                      {proxy.latency_ms != null ? `${proxy.latency_ms} мс` : '—'}
                    </td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2 justify-end">
                        <button
                          id={`reassign-proxy-${proxy.id}`}
                          onClick={() => setReassignProxy(proxy)}
                          className="text-xs text-neutral-400 hover:text-accent-cyan border border-neutral-700
                            hover:border-accent-cyan/50 px-2.5 py-1 rounded-md transition-colors"
                        >
                          Перепривязать
                        </button>
                        <button
                          id={`delete-proxy-${proxy.id}`}
                          onClick={() => deleteProxy.mutate(proxy.id)}
                          disabled={deleteProxy.isPending}
                          className="text-xs text-neutral-600 hover:text-danger border border-neutral-700
                            hover:border-danger/50 px-2.5 py-1 rounded-md transition-colors"
                        >
                          ✕
                        </button>
                      </div>
                    </td>
                  </tr>
                );
              })}
          </tbody>
        </table>
      </div>

      <ReassignProxyDialog
        isOpen={!!reassignProxy}
        onClose={() => setReassignProxy(null)}
        proxy={reassignProxy}
        accounts={accounts.map((a) => ({ id: a.id, phone: a.phone }))}
      />
    </div>
  );
};

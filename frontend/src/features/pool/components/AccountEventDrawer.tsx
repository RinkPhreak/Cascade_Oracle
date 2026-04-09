import type { TgAccount, AccountEvent, AccountEventType } from '../../../api/extended-types';
import { useAccountEvents } from '../hooks/useAccounts';
import { Drawer } from '../../../shared/components/Drawer';

const eventTypeColors: Record<AccountEventType, string> = {
  BANNED: 'bg-danger text-white',
  SERVICE_NOTICE: 'bg-yellow-500 text-black',
  RECONNECT: 'bg-blue-500 text-white',
  SESSION_CREATED: 'bg-green-500 text-white',
  SESSION_EXPIRED: 'bg-orange-500 text-white',
  PROXY_CHANGED: 'bg-purple-500 text-white',
  FLOOD_WAIT: 'bg-orange-400 text-black',
  STATUS_CHANGE: 'bg-accent-cyan text-black',
};

const EventTypeLabels: Record<AccountEventType, string> = {
  BANNED: 'БАН',
  SERVICE_NOTICE: 'УВЕДОМЛЕНИЕ',
  RECONNECT: 'ПЕРЕПОДКЛЮЧЕНИЕ',
  SESSION_CREATED: 'СЕССИЯ СОЗДАНА',
  SESSION_EXPIRED: 'СЕССИЯ ИСТЕКЛА',
  PROXY_CHANGED: 'СМЕНА ПРОКСИ',
  FLOOD_WAIT: 'FLOOD WAIT',
  STATUS_CHANGE: 'СМЕНА СТАТУСА',
};

const EventItem = ({ event }: { event: AccountEvent }) => {
  const colorClass = eventTypeColors[event.type] ?? 'bg-neutral-600 text-white';
  const label = EventTypeLabels[event.type] ?? event.type;

  return (
    <li className="flex gap-4 group">
      {/* Timeline line */}
      <div className="flex flex-col items-center">
        <div className={`w-2.5 h-2.5 rounded-full shrink-0 mt-1 ${colorClass.split(' ')[0]}`} />
        <div className="w-px flex-1 bg-neutral-800 mt-1 group-last:hidden" />
      </div>

      {/* Content */}
      <div className="pb-5 flex-1 min-w-0">
        <div className="flex items-start justify-between gap-2 flex-wrap">
          <span
            className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-mono font-bold ${colorClass}`}
          >
            {label}
          </span>
          <time className="text-xs text-neutral-600 font-mono shrink-0">
            {new Date(event.occurred_at).toLocaleString('ru-RU', {
              day: '2-digit', month: '2-digit', year: '2-digit',
              hour: '2-digit', minute: '2-digit',
            })}
          </time>
        </div>
        <p className="mt-1.5 text-sm text-neutral-300 leading-relaxed">{event.description}</p>
      </div>
    </li>
  );
};

interface AccountEventDrawerProps {
  account: TgAccount | null;
  onClose: () => void;
}

export const AccountEventDrawer = ({ account, onClose }: AccountEventDrawerProps) => {
  const { data: events, isLoading, isError } = useAccountEvents(account?.id ?? null);

  return (
    <Drawer
      isOpen={!!account}
      onClose={onClose}
      title={`События: ${account?.phone ?? '—'}`}
      width="w-[520px]"
    >
      {isLoading && (
        <div className="flex flex-col gap-3">
          {[1, 2, 3, 4].map((i) => (
            <div key={i} className="flex gap-4">
              <div className="w-2.5 h-2.5 rounded-full bg-neutral-800 animate-pulse mt-1 shrink-0" />
              <div className="flex-1 space-y-1.5">
                <div className="h-4 bg-neutral-800 rounded animate-pulse w-1/3" />
                <div className="h-3 bg-neutral-800 rounded animate-pulse w-3/4" />
              </div>
            </div>
          ))}
        </div>
      )}

      {isError && (
        <p className="text-danger text-sm">
          Не удалось загрузить события аккаунта.
        </p>
      )}

      {!isLoading && !isError && events && (
        <>
          {events.length === 0 ? (
            <p className="text-neutral-500 text-sm text-center py-8">
              Для этого аккаунта событий не зафиксировано.
            </p>
          ) : (
            <ul className="flex flex-col">
              {events.map((event) => (
                <EventItem key={event.id} event={event} />
              ))}
            </ul>
          )}

          <p className="mt-4 text-xs text-neutral-600 font-mono text-center">
            {events.length} событий
          </p>
        </>
      )}
    </Drawer>
  );
};

import type { DtoContact as Contact, DtoContactTrace as ContactTrace } from '../../../api/generated';
import { Modal } from '../../../shared/components/Modal';
import { useContactTrace } from '../hooks/useLeads';
import { AnonymiseButton } from './AnonymiseButton';

const CHANNEL_ICONS: Record<string, string> = {
  telegram: '✈',
  sms: '📱',
  system: '⚙',
};

const STATUS_STYLES: Record<string, string> = {
  ENQUEUED: 'bg-neutral-700 text-neutral-300',
  ATTEMPTED: 'bg-blue-500/20 text-blue-400',
  RATE_LIMITED: 'bg-orange-500/20 text-orange-400',
  NOT_FOUND: 'bg-red-500/20 text-red-400',
  DELIVERED: 'bg-green-500/20 text-green-400',
  REPLIED: 'bg-purple-500/20 text-purple-400',
  FAILED: 'bg-danger/20 text-danger',
};

const TraceStep = ({ trace }: { trace: ContactTrace }) => {
  const statusStyle = STATUS_STYLES[trace.status] ?? 'bg-neutral-700 text-neutral-400';
  const channelIcon = CHANNEL_ICONS[trace.channel] ?? '●';

  return (
    <li className="flex gap-4">
      {/* Timeline */}
      <div className="flex flex-col items-center">
        <div className={`flex items-center justify-center w-7 h-7 rounded-full text-sm border
          ${trace.status === 'REPLIED' ? 'border-purple-500 bg-purple-500/20 text-purple-400' :
            trace.status === 'DELIVERED' ? 'border-green-500 bg-green-500/20 text-green-400' :
            trace.status === 'FAILED' || trace.status === 'NOT_FOUND' ? 'border-danger bg-danger/20 text-danger' :
            'border-neutral-600 bg-neutral-800 text-neutral-400'}`}
        >
          {channelIcon}
        </div>
        <div className="w-px flex-1 bg-neutral-800 mt-1" />
      </div>

      {/* Content */}
      <div className="pb-5 flex-1 min-w-0">
        <div className="flex items-center gap-2 flex-wrap">
          <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-mono font-bold ${statusStyle}`}>
            {trace.status}
          </span>
          <span className="text-xs text-neutral-600 font-mono">
            Шаг {trace.step + 1} · {trace.channel.toUpperCase()}
          </span>
          <time className="text-xs text-neutral-600 font-mono ml-auto">
            {new Date(trace.occurred_at).toLocaleString('ru-RU', {
              day: '2-digit', month: '2-digit',
              hour: '2-digit', minute: '2-digit',
            })}
          </time>
        </div>
        <p className="mt-1 text-sm text-neutral-300">{trace.description}</p>
        {trace.error_code && (
          <span className="mt-1 inline-block text-xs font-mono text-orange-400 bg-orange-500/10 px-2 py-0.5 rounded">
            {trace.error_code}
          </span>
        )}
      </div>
    </li>
  );
};

interface ContactHistoryModalProps {
  contact: Contact | null;
  onClose: () => void;
}

export const ContactHistoryModal = ({ contact, onClose }: ContactHistoryModalProps) => {
  const { data: traceRaw, isLoading, isError } = useContactTrace(contact?.id ?? null);
  const traces = (traceRaw as ContactTrace[] | undefined) ?? [];

  return (
    <Modal
      isOpen={!!contact}
      onClose={onClose}
      title={`История: ${contact?.phone ?? '—'}`}
      width="max-w-xl"
    >
      {contact && (
        <div className="flex flex-col">
          {/* Contact profile header */}
          <div className="px-6 py-4 bg-neutral-900/60 border-b border-neutral-800">
            <div className="flex items-start justify-between gap-4">
              <div>
                <p className="text-sm font-semibold text-white">{contact.name || <span className="text-neutral-500 italic">Без имени</span>}</p>
                <p className="text-xs font-mono text-neutral-400 mt-0.5">{contact.phone}</p>
                {contact.reply_at && (
                  <p className="text-xs text-neutral-500 mt-1">
                    Ответил: {new Date(contact.reply_at).toLocaleString('ru-RU')}
                  </p>
                )}
              </div>
              <AnonymiseButton contact={contact} onSuccess={onClose} />
            </div>

            {contact.reply_text && (
              <div className="mt-3 bg-purple-500/10 border border-purple-500/20 rounded-lg px-4 py-3">
                <p className="text-xs text-purple-400 font-mono mb-1">ОТВЕТ КОНТАКТА</p>
                <p className="text-sm text-neutral-200 leading-relaxed">{contact.reply_text}</p>
                {contact.reply_account_phone && (
                  <p className="text-xs text-neutral-500 mt-2">
                    Аккаунт получения: <span className="font-mono text-neutral-400">{contact.reply_account_phone}</span>
                  </p>
                )}
              </div>
            )}
          </div>

          {/* Waterfall trace */}
          <div className="px-6 py-5">
            <p className="text-xs font-mono text-neutral-500 uppercase tracking-wider mb-4">Waterfall Trace</p>

            {isLoading && (
              <div className="flex flex-col gap-4">
                {[1, 2, 3].map((i) => (
                  <div key={i} className="flex gap-4">
                    <div className="w-7 h-7 rounded-full bg-neutral-800 animate-pulse shrink-0" />
                    <div className="flex-1 space-y-2">
                      <div className="h-4 bg-neutral-800 rounded animate-pulse w-1/3" />
                      <div className="h-3 bg-neutral-800 rounded animate-pulse w-2/3" />
                    </div>
                  </div>
                ))}
              </div>
            )}

            {isError && (
              <p className="text-danger text-sm">Не удалось загрузить историю доставки.</p>
            )}

            {!isLoading && !isError && (
              <>
                {traces.length === 0 ? (
                  <p className="text-neutral-500 text-sm text-center py-4">
                    История событий недоступна.
                  </p>
                ) : (
                  <ul className="flex flex-col">
                    {traces.map((trace: ContactTrace) => (
                      <TraceStep key={trace.id} trace={trace} />
                    ))}
                  </ul>
                )}
              </>
            )}
          </div>
        </div>
      )}
    </Modal>
  );
};

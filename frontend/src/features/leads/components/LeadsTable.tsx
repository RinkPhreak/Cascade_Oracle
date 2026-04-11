import type { DtoContact } from '../../../api/generated';
import { AnonymiseButton } from './AnonymiseButton';

interface LeadsTableProps {
  contacts: DtoContact[];
  isLoading: boolean;
  onSelect: (contact: DtoContact) => void;
}

const SkeletonRow = () => (
  <tr className="border-t border-neutral-800">
    {[1, 2, 3, 4, 5].map((i) => (
      <td key={i} className="px-4 py-4">
        <div className="h-4 bg-neutral-800 rounded animate-pulse" style={{ width: `${50 + i * 10}%` }} />
      </td>
    ))}
  </tr>
);

export const LeadsTable = ({ contacts, isLoading, onSelect }: LeadsTableProps) => {
  return (
    <div className="overflow-hidden rounded-xl border border-neutral-800">
      <table className="w-full text-sm" aria-label="Таблица лидов (ответивших контактов)">
        <thead>
          <tr className="bg-neutral-900/60 text-left">
            {['Телефон', 'Имя', 'Текст ответа', 'Время ответа', 'Аккаунт получения', ''].map((h) => (
              <th key={h} className="px-4 py-3 text-xs font-mono font-semibold text-neutral-400 uppercase tracking-wider">
                {h}
              </th>
            ))}
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

          {!isLoading && contacts.length === 0 && (
            <tr>
              <td colSpan={6} className="px-4 py-16 text-center text-neutral-500 text-sm">
                Нет ответивших контактов.
              </td>
            </tr>
          )}

          {!isLoading &&
            contacts.map((contact) => (
              <tr
                key={contact.id}
                className="border-t border-neutral-800 hover:bg-neutral-900/40 transition-colors cursor-pointer"
                onClick={() => onSelect(contact)}
              >
                <td className="px-4 py-3 font-mono text-white">{contact.phone}</td>
                <td className="px-4 py-3 text-neutral-300">
                  {contact.name || <span className="text-neutral-600 italic">—</span>}
                </td>
                <td className="px-4 py-3 max-w-xs">
                  {contact.reply_text ? (
                    <p className="text-sm text-neutral-200 truncate" title={contact.reply_text}>
                      {contact.reply_text}
                    </p>
                  ) : (
                    <span className="text-neutral-600 italic text-xs">—</span>
                  )}
                </td>
                <td className="px-4 py-3 text-neutral-500 text-xs font-mono whitespace-nowrap">
                  {contact.reply_at
                    ? new Date(contact.reply_at).toLocaleString('ru-RU', {
                        day: '2-digit', month: '2-digit', year: '2-digit',
                        hour: '2-digit', minute: '2-digit',
                      })
                    : '—'}
                </td>
                <td className="px-4 py-3 font-mono text-xs text-neutral-400">
                  {contact.reply_account_phone ?? '—'}
                </td>
                <td className="px-4 py-3" onClick={(e) => e.stopPropagation()}>
                  <AnonymiseButton contact={contact} />
                </td>
              </tr>
            ))}
        </tbody>
      </table>
    </div>
  );
};

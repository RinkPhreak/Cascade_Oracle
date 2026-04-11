import { useState } from 'react';
import { LeadsTable } from '../features/leads/components/LeadsTable';
import { ContactHistoryModal } from '../features/leads/components/ContactHistoryModal';
import { useLeads } from '../features/leads/hooks/useLeads';
import type { DtoContact } from '../api/generated';

export const LeadsPage = () => {
  const { data: contacts = [], isLoading, isError } = useLeads();
  const [selectedContact, setSelectedContact] = useState<DtoContact | null>(null);
  const [search, setSearch] = useState('');

  const filtered = search.trim()
    ? contacts.filter(
        (c) =>
          c.phone.includes(search) ||
          c.name.toLowerCase().includes(search.toLowerCase()) ||
          (c.reply_text ?? '').toLowerCase().includes(search.toLowerCase())
      )
    : contacts;

  return (
    <div className="flex flex-col gap-6">
      {/* Page Header */}
      <div>
        <h2 className="text-2xl font-bold text-white tracking-tight">Лиды и 152-ФЗ</h2>
        <p className="text-neutral-400 text-sm mt-1">
          Ответившие контакты · Waterfall трассировка · Право на забвение
        </p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-2 gap-4">
        <div className="bg-bg-surface border border-neutral-800 rounded-xl p-4">
          <p className="text-xs text-neutral-500 font-mono uppercase tracking-wider mb-1">Всего лидов</p>
          <p className="text-2xl font-bold font-mono text-purple-400">{contacts.length}</p>
        </div>
        <div className="bg-bg-surface border border-neutral-800 rounded-xl p-4">
          <p className="text-xs text-neutral-500 font-mono uppercase tracking-wider mb-1">152-ФЗ: Анонимизировано</p>
          <p className="text-2xl font-bold font-mono text-neutral-400">
            — <span className="text-sm font-normal text-neutral-600">(не отображаются)</span>
          </p>
        </div>
      </div>

      {/* Search */}
      <div className="relative">
        <span className="absolute left-4 top-1/2 -translate-y-1/2 text-neutral-500 text-sm">🔍</span>
        <input
          id="leads-search"
          type="search"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Поиск по телефону, имени или тексту ответа..."
          className="w-full bg-bg-surface border border-neutral-800 focus:border-accent-cyan
            rounded-xl pl-10 pr-4 py-3 text-sm text-white outline-none transition-colors
            placeholder:text-neutral-600"
        />
      </div>

      {/* 152-FZ info banner */}
      <div className="flex items-start gap-3 text-xs bg-purple-500/5 border border-purple-500/15 rounded-xl px-4 py-4">
        <span className="text-purple-400 text-base shrink-0">§</span>
        <div className="text-neutral-400 leading-relaxed">
          <strong className="text-purple-400">152-ФЗ "Право на забвение"</strong> — Чтобы анонимизировать контакт,
          нажмите кнопку <span className="font-mono text-danger bg-danger/10 px-1 py-0.5 rounded">Анонимизировать (152-ФЗ)</span> в таблице или
          в профиле контакта. Действие необратимо и требует двойного подтверждения.
          После анонимизации контакт исчезает из данного списка.
        </div>
      </div>

      {isError && (
        <div className="flex items-center gap-2 text-danger bg-danger/5 border border-danger/20 rounded-xl px-4 py-3 text-sm">
          <span>⚠</span>
          <span>Не удалось загрузить список контактов.</span>
        </div>
      )}

      <LeadsTable
        contacts={filtered}
        isLoading={isLoading}
        onSelect={setSelectedContact}
      />

      <ContactHistoryModal
        contact={selectedContact}
        onClose={() => setSelectedContact(null)}
      />
    </div>
  );
};

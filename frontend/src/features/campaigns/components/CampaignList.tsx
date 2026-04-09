import type { Campaign } from '../../../api/extended-types';
import { StatusBadge } from '../../../shared/components/StatusBadge';

interface CampaignListProps {
  campaigns: Campaign[];
  isLoading: boolean;
  onSelect: (campaign: Campaign) => void;
  onNew: () => void;
}

const SkeletonRow = () => (
  <div className="flex items-center gap-4 p-4 border-b border-neutral-800">
    {[1, 2, 3, 4].map((i) => (
      <div key={i} className="h-4 bg-neutral-800 rounded animate-pulse" style={{ width: `${60 + i * 15}px` }} />
    ))}
  </div>
);

export const CampaignList = ({ campaigns, isLoading, onSelect, onNew }: CampaignListProps) => {
  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <h3 className="text-base font-semibold text-white">Все кампании</h3>
        <button
          id="new-campaign-btn"
          onClick={onNew}
          className="flex items-center gap-2 px-4 py-2 bg-accent-cyan text-black font-bold text-sm
            rounded-lg hover:bg-accent-cyan-hover transition-colors"
        >
          + Новая кампания
        </button>
      </div>

      <div className="overflow-hidden rounded-xl border border-neutral-800">
        {/* Table header */}
        <div className="grid grid-cols-[1fr_auto_auto_auto] gap-4 px-5 py-3 bg-neutral-900/60 border-b border-neutral-800">
          {['Название', 'Статус', 'Контакты', 'Создана'].map((h) => (
            <span key={h} className="text-xs font-mono font-semibold text-neutral-400 uppercase tracking-wider">{h}</span>
          ))}
        </div>

        {isLoading && (
          <>
            <SkeletonRow />
            <SkeletonRow />
            <SkeletonRow />
          </>
        )}

        {!isLoading && campaigns.length === 0 && (
          <div className="px-5 py-16 text-center">
            <p className="text-neutral-500 text-sm">Кампаний нет. Создайте первую.</p>
          </div>
        )}

        {!isLoading && campaigns.map((c) => {
          const progressPct = c.total_contacts > 0
            ? Math.round(((c.completed + c.replied) / c.total_contacts) * 100)
            : 0;

          return (
            <button
              key={c.id}
              id={`campaign-row-${c.id}`}
              onClick={() => onSelect(c)}
              className="w-full grid grid-cols-[1fr_auto_auto_auto] gap-4 items-center px-5 py-4
                border-b border-neutral-800 last:border-0 hover:bg-neutral-900/40 transition-colors text-left"
            >
              <div className="min-w-0">
                <p className="text-sm font-medium text-white truncate">{c.name}</p>
                {c.scheduled_at && (
                  <p className="text-xs text-neutral-500 mt-0.5">
                    📅 {new Date(c.scheduled_at).toLocaleString('ru-RU', { dateStyle: 'short', timeStyle: 'short' })}
                  </p>
                )}
                {c.status === 'RUNNING' && c.total_contacts > 0 && (
                  <div className="w-full max-w-xs bg-neutral-800 rounded-full h-1 mt-2 overflow-hidden">
                    <div className="h-full bg-accent-cyan rounded-full" style={{ width: `${progressPct}%` }} />
                  </div>
                )}
              </div>
              <StatusBadge status={c.status} />
              <span className="text-sm font-mono text-neutral-300">
                {c.total_contacts > 0 ? c.total_contacts.toLocaleString('ru-RU') : '—'}
              </span>
              <span className="text-xs text-neutral-500 font-mono">
                {new Date(c.created_at).toLocaleDateString('ru-RU')}
              </span>
            </button>
          );
        })}
      </div>
    </div>
  );
};

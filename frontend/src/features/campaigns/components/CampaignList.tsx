import type { DtoCampaign as Campaign } from '../../../api/generated';
import { StatusBadge } from '../../../shared/components/StatusBadge';
import { useDeleteCampaign } from '../hooks/useCampaigns';

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
  const deleteCampaign = useDeleteCampaign();

  const handleDelete = (e: React.MouseEvent, id: string, name: string) => {
    e.stopPropagation();
    if (window.confirm(`Вы уверены, что хотите удалить кампанию "${name}"? Все связанные данные будут стерты.`)) {
      deleteCampaign.mutate(id);
    }
  };

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
            <div key={c.id} className="relative group">
              <button
                id={`campaign-row-${c.id}`}
                onClick={() => onSelect(c)}
                className="w-full grid grid-cols-[1fr_auto_auto_auto] gap-4 items-center px-5 py-4
                  border-b border-neutral-800 last:border-0 hover:bg-neutral-900/40 transition-colors text-left"
              >
                <div className="min-w-0 pr-8">
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
              
              <button
                id={`campaign-delete-${c.id}`}
                onClick={(e) => handleDelete(e, c.id, c.name)}
                className="absolute right-3 top-1/2 -translate-y-1/2 p-2 text-neutral-600 hover:text-danger
                  bg-neutral-900/0 hover:bg-neutral-800 rounded-lg opacity-0 group-hover:opacity-100 transition-all"
                title="Удалить кампанию"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
          );
        })}
      </div>
    </div>
  );
};

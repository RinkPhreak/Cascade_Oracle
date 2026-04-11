import { useState } from 'react';
import type { DtoCampaign as Campaign } from '../../../api/generated';
import { useCampaignStats } from '../hooks/useCampaignTasks';
import { useStartCampaign } from '../hooks/useStartCampaign';
import { usePauseCampaign } from '../hooks/usePauseCampaign';
import { WaterfallFunnel } from './WaterfallFunnel';
import { ErrorTaxonomyChart } from './ErrorTaxonomyChart';
import { StuckAttemptsPanel } from './StuckAttemptsPanel';
import { StatusBadge } from '../../../shared/components/StatusBadge';
import { Modal } from '../../../shared/components/Modal';

interface PauseCampaignModalProps {
  isOpen: boolean;
  onClose: () => void;
  campaignId: string;
}

const PauseCampaignModal = ({ isOpen, onClose, campaignId }: PauseCampaignModalProps) => {
  const [password, setPassword] = useState('');
  const [reason, setReason] = useState('');
  const pause = usePauseCampaign(onClose);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!password || !reason) return;
    pause.mutate({ campaignId, password, reason });
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="⏸ Аварийная пауза кампании" width="max-w-md">
      <form onSubmit={handleSubmit} className="px-6 py-5 flex flex-col gap-4">
        <p className="text-sm text-neutral-300 leading-relaxed">
          Кампания будет немедленно приостановлена. Требуется повторная аутентификация.
        </p>
        <div className="flex flex-col gap-1.5">
          <label htmlFor="pause-password" className="text-xs font-mono text-neutral-400">Пароль *</label>
          <input
            id="pause-password"
            type="password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="bg-neutral-900 border border-danger/50 focus:border-danger rounded-md px-3 py-2 text-sm text-white outline-none transition-colors"
          />
        </div>
        <div className="flex flex-col gap-1.5">
          <label htmlFor="pause-reason" className="text-xs font-mono text-neutral-400">Причина *</label>
          <textarea
            id="pause-reason"
            required
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            rows={3}
            className="bg-neutral-900 border border-neutral-700 focus:border-danger rounded-md px-3 py-2 text-sm text-white outline-none transition-colors resize-none"
          />
        </div>
        <div className="flex gap-3 justify-end">
          <button type="button" onClick={onClose} className="px-4 py-2 text-sm text-neutral-400 hover:text-white border border-neutral-700 rounded-md transition-colors">Отмена</button>
          <button type="submit" id="pause-campaign-submit" disabled={!password || !reason || pause.isPending}
            className="px-4 py-2 text-sm font-bold bg-danger/10 hover:bg-danger/20 border border-danger text-danger rounded-md transition-colors disabled:opacity-40 flex items-center gap-2">
            {pause.isPending && <span className="w-4 h-4 border-2 border-danger/30 border-t-danger rounded-full animate-spin" />}
            Поставить на паузу
          </button>
        </div>
      </form>
    </Modal>
  );
};

interface CampaignMonitorProps {
  campaign: Campaign;
  onBack: () => void;
}

export const CampaignMonitor = ({ campaign, onBack }: CampaignMonitorProps) => {
  const { data: stats, isLoading } = useCampaignStats(campaign.id);
  const startCampaign = useStartCampaign();
  const [pauseModalOpen, setPauseModalOpen] = useState(false);

  const isRunning = campaign.status === 'RUNNING';
  const isDraft = campaign.status === 'DRAFT';

  const progressPct = stats && stats.total > 0
    ? Math.round(((stats.completed + stats.replied) / stats.total) * 100)
    : 0;

  return (
    <div className="flex flex-col gap-6">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div>
          <button onClick={onBack} className="text-xs text-neutral-500 hover:text-accent-cyan transition-colors mb-2 flex items-center gap-1">
            ← Все кампании
          </button>
          <h3 className="text-xl font-bold text-white">{campaign.name}</h3>
          <div className="flex items-center gap-3 mt-1.5">
            <StatusBadge status={campaign.status} />
            {campaign.scheduled_at && (
              <span className="text-xs text-neutral-500 font-mono">
                📅 {new Date(campaign.scheduled_at).toLocaleString('ru-RU')}
              </span>
            )}
          </div>
        </div>
        <div className="flex gap-2">
          {isDraft && (
            <button
              id="start-campaign-btn"
              onClick={() => startCampaign.mutate(campaign.id)}
              disabled={startCampaign.isPending}
              className="px-4 py-2 text-sm font-bold bg-green-500/10 hover:bg-green-500/20 border border-green-500/50 text-green-400 rounded-lg transition-colors disabled:opacity-40 flex items-center gap-2"
            >
              {startCampaign.isPending && <span className="w-4 h-4 border-2 border-green-500/30 border-t-green-500 rounded-full animate-spin" />}
              ▶ Запустить
            </button>
          )}
          {isRunning && (
            <button
              id="pause-campaign-btn"
              onClick={() => setPauseModalOpen(true)}
              className="px-4 py-2 text-sm font-bold bg-danger/10 hover:bg-danger/20 border border-danger text-danger rounded-lg transition-colors"
            >
              ⏸ Аварийный стоп
            </button>
          )}
        </div>
      </div>

      {/* Progress */}
      {stats && (
        <div className="bg-bg-surface border border-neutral-800 rounded-xl p-5 flex flex-col gap-3">
          <div className="flex items-center justify-between">
            <span className="text-sm font-semibold text-white">Прогресс доставки</span>
            <span className="text-2xl font-bold font-mono text-accent-cyan">{progressPct}%</span>
          </div>
          <div className="w-full bg-neutral-800 rounded-full h-3 overflow-hidden">
            <div
              className="h-full bg-accent-cyan rounded-full transition-all duration-500"
              style={{ width: `${progressPct}%` }}
            />
          </div>
          <div className="grid grid-cols-4 gap-4 mt-1">
            {[
              { label: 'Всего', value: stats.total, color: 'text-white' },
              { label: 'Доставлено', value: stats.completed, color: 'text-green-400' },
              { label: 'Ответили', value: stats.replied, color: 'text-purple-400' },
              { label: 'Ошибка', value: stats.failed, color: 'text-danger' },
            ].map(({ label, value, color }) => (
              <div key={label} className="text-center">
                <p className={`text-lg font-bold font-mono ${color}`}>{value.toLocaleString('ru-RU')}</p>
                <p className="text-xs text-neutral-500 mt-0.5">{label}</p>
              </div>
            ))}
          </div>
        </div>
      )}

      {isLoading && (
        <div className="bg-bg-surface border border-neutral-800 rounded-xl p-8 flex items-center justify-center">
          <div className="w-8 h-8 border-2 border-neutral-700 border-t-accent-cyan rounded-full animate-spin" />
        </div>
      )}

      {/* Waterfall Funnel */}
      {stats && (
        <div className="bg-bg-surface border border-neutral-800 rounded-xl p-5">
          <h4 className="text-sm font-semibold text-white mb-4">Воронка Waterfall</h4>
          <WaterfallFunnel stats={stats} />
        </div>
      )}

      {/* Error Taxonomy */}
      {stats && (
        <div className="bg-bg-surface border border-neutral-800 rounded-xl p-5">
          <h4 className="text-sm font-semibold text-white mb-4">Таксономия ошибок</h4>
          <ErrorTaxonomyChart stats={stats} />
        </div>
      )}

      {/* Stuck Attempts */}
      <div className="bg-bg-surface border border-neutral-800 rounded-xl p-5">
        <h4 className="text-sm font-semibold text-white mb-4">Мониторинг зависших задач</h4>
        <StuckAttemptsPanel campaignId={campaign.id} />
      </div>

      <PauseCampaignModal
        isOpen={pauseModalOpen}
        onClose={() => setPauseModalOpen(false)}
        campaignId={campaign.id}
      />
    </div>
  );
};

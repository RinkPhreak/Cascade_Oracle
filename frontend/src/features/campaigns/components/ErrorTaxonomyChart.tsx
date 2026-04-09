import type { CampaignStats } from '../../../api/extended-types';

interface ErrorTaxonomyChartProps {
  stats: CampaignStats;
}

const ERROR_LABELS: Record<string, { label: string; color: string; bar: string }> = {
  USER_NOT_FOUND:    { label: 'Пользователь не найден', color: 'text-red-400',    bar: 'bg-red-500/60' },
  RATE_LIMITED:      { label: 'Rate Limited',            color: 'text-orange-400', bar: 'bg-orange-500/60' },
  FLOOD_WAIT:        { label: 'Flood Wait',              color: 'text-yellow-400', bar: 'bg-yellow-500/60' },
  PEER_FLOOD:        { label: 'Peer Flood',              color: 'text-red-400',    bar: 'bg-red-400/60' },
  PHONE_NOT_FOUND:   { label: 'Телефон не найден',       color: 'text-orange-400', bar: 'bg-orange-400/60' },
  PRIVACY_RESTRICTED:{ label: 'Приватность',             color: 'text-purple-400', bar: 'bg-purple-500/60' },
  BANNED:            { label: 'Аккаунт заблокирован',    color: 'text-danger',     bar: 'bg-danger/60' },
  SMS_FAILED:        { label: 'SMS Ошибка',              color: 'text-blue-400',   bar: 'bg-blue-500/60' },
};

export const ErrorTaxonomyChart = ({ stats }: ErrorTaxonomyChartProps) => {
  const { error_breakdown } = stats;
  const totalErrors = Object.values(error_breakdown).reduce((s, c) => s + c, 0);

  if (totalErrors === 0) {
    return (
      <p className="text-neutral-600 text-sm italic text-center py-4">
        Ошибок не зафиксировано
      </p>
    );
  }

  const sorted = Object.entries(error_breakdown)
    .sort(([, a], [, b]) => b - a)
    .slice(0, 8);

  const maxCount = sorted[0]?.[1] ?? 1;

  return (
    <div className="flex flex-col gap-2.5">
      {sorted.map(([code, count]) => {
        const cfg = ERROR_LABELS[code] ?? { label: code, color: 'text-neutral-400', bar: 'bg-neutral-600' };
        const pct = Math.round((count / totalErrors) * 100);
        const barWidth = Math.round((count / maxCount) * 100);

        return (
          <div key={code} className="flex items-center gap-3">
            {/* Label */}
            <div className="w-40 shrink-0 text-right">
              <span className={`text-xs font-mono ${cfg.color}`}>{cfg.label}</span>
            </div>
            {/* Bar */}
            <div className="flex-1 bg-neutral-800 rounded-full h-5 overflow-hidden">
              <div
                className={`h-full ${cfg.bar} rounded-full flex items-center pl-2 transition-all`}
                style={{ width: `${barWidth}%` }}
              >
                <span className="text-xs font-mono font-bold text-white/90 whitespace-nowrap">
                  {count.toLocaleString('ru-RU')}
                </span>
              </div>
            </div>
            {/* Percentage */}
            <div className="w-10 shrink-0 text-right">
              <span className="text-xs font-mono text-neutral-400">{pct}%</span>
            </div>
          </div>
        );
      })}
      <p className="text-xs text-neutral-600 text-right mt-1">
        Всего ошибок: {totalErrors.toLocaleString('ru-RU')}
      </p>
    </div>
  );
};

import type { CampaignStats } from '../../../api/extended-types';

interface WaterfallFunnelProps {
  stats: CampaignStats;
}

interface FunnelLayer {
  label: string;
  sublabel: string;
  count: number;
  pct: number;
  color: string;
  bgColor: string;
}

export const WaterfallFunnel = ({ stats }: WaterfallFunnelProps) => {
  const { total, tg_attempted, sms_attempted, completed, replied } = stats;
  if (total === 0) return null;

  const pct = (n: number) => Math.round((n / total) * 100);

  const layers: FunnelLayer[] = [
    {
      label: 'Всего контактов',
      sublabel: 'Загружено в кампанию',
      count: total,
      pct: 100,
      color: 'text-white',
      bgColor: 'bg-neutral-700',
    },
    {
      label: 'TG Попытки',
      sublabel: 'Отправлено через Telegram',
      count: tg_attempted,
      pct: pct(tg_attempted),
      color: 'text-accent-cyan',
      bgColor: 'bg-accent-cyan/40',
    },
    {
      label: 'SMS Попытки',
      sublabel: 'Escalated to SMS fallback',
      count: sms_attempted,
      pct: pct(sms_attempted),
      color: 'text-blue-400',
      bgColor: 'bg-blue-500/40',
    },
    {
      label: 'Доставлено',
      sublabel: 'Confirmed delivery',
      count: completed,
      pct: pct(completed),
      color: 'text-green-400',
      bgColor: 'bg-green-500/40',
    },
    {
      label: 'Ответили',
      sublabel: 'Contact replied',
      count: replied,
      pct: pct(replied),
      color: 'text-purple-400',
      bgColor: 'bg-purple-500/40',
    },
  ];

  return (
    <div className="flex flex-col gap-1">
      {layers.map((layer, i) => {
        const trapezoidWidth = 100 - i * 10;
        return (
          <div key={layer.label} className="flex items-center gap-4">
            {/* Funnel bar */}
            <div className="flex-1 flex flex-col items-center">
              <div
                className={`${layer.bgColor} rounded transition-all flex items-center justify-center py-2.5`}
                style={{ width: `${trapezoidWidth}%` }}
              >
                <span className={`text-xs font-mono font-bold ${layer.color}`}>
                  {layer.count.toLocaleString('ru-RU')}
                </span>
              </div>
            </div>
            {/* Labels */}
            <div className="w-40 shrink-0">
              <p className={`text-xs font-semibold ${layer.color}`}>
                {layer.label}
              </p>
              <p className="text-xs text-neutral-600">{layer.pct}% — {layer.sublabel}</p>
            </div>
          </div>
        );
      })}
    </div>
  );
};

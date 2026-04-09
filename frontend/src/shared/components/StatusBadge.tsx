interface StatusBadgeProps {
  status: string;
  className?: string;
}

const statusMap: Record<string, { label: string; classes: string }> = {
  // TG Account statuses
  WARMING_UP:   { label: 'Прогрев',    classes: 'bg-yellow-500/15 text-yellow-400 border-yellow-500/30' },
  ACTIVE:       { label: 'Активен',    classes: 'bg-green-500/15 text-green-400 border-green-500/30' },
  COOLING_DOWN: { label: 'Охлаждение', classes: 'bg-blue-500/15 text-blue-400 border-blue-500/30' },
  SUSPENDED:    { label: 'Приостановлен', classes: 'bg-orange-500/15 text-orange-400 border-orange-500/30' },
  BANNED:       { label: 'Заблокирован', classes: 'bg-danger/15 text-red-400 border-danger/30' },

  // Campaign statuses
  DRAFT:        { label: 'Черновик',   classes: 'bg-neutral-700/50 text-neutral-400 border-neutral-600' },
  SCHEDULED:    { label: 'Запланирована', classes: 'bg-blue-500/15 text-blue-400 border-blue-500/30' },
  RUNNING:      { label: 'Выполняется', classes: 'bg-accent-cyan/15 text-accent-cyan border-accent-cyan/30' },
  PAUSED:       { label: 'Пауза',      classes: 'bg-orange-500/15 text-orange-400 border-orange-500/30' },
  COMPLETED:    { label: 'Завершена',  classes: 'bg-green-500/15 text-green-400 border-green-500/30' },
  FAILED:       { label: 'Ошибка',     classes: 'bg-danger/15 text-red-400 border-danger/30' },

  // Task statuses
  pending:      { label: 'Ожидание',   classes: 'bg-neutral-700/50 text-neutral-400 border-neutral-600' },
  in_progress:  { label: 'В работе',   classes: 'bg-accent-cyan/15 text-accent-cyan border-accent-cyan/30' },
  completed:    { label: 'Доставлено', classes: 'bg-green-500/15 text-green-400 border-green-500/30' },
  replied:      { label: 'Ответ',      classes: 'bg-purple-500/15 text-purple-400 border-purple-500/30' },
  failed:       { label: 'Ошибка',     classes: 'bg-danger/15 text-red-400 border-danger/30' },

  // Generic
  healthy:      { label: 'Здоров',     classes: 'bg-green-500/15 text-green-400 border-green-500/30' },
  unhealthy:    { label: 'Недоступен', classes: 'bg-danger/15 text-red-400 border-danger/30' },
  OPERATIONAL:  { label: 'Operational', classes: 'bg-green-500/15 text-green-400 border-green-500/30' },
  DEGRADED:     { label: 'Degraded',   classes: 'bg-orange-500/15 text-orange-400 border-orange-500/30' },
  HALTED:       { label: 'HALTED',     classes: 'bg-danger/15 text-red-400 border-danger/30' },
};

export const StatusBadge = ({ status, className = '' }: StatusBadgeProps) => {
  const config = statusMap[status] ?? {
    label: status,
    classes: 'bg-neutral-700/50 text-neutral-400 border-neutral-600',
  };

  return (
    <span
      className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-mono font-medium
        border ${config.classes} ${className}`}
    >
      {config.label}
    </span>
  );
};

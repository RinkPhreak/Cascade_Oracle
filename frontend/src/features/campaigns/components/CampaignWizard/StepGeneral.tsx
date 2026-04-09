import { useState } from 'react';

interface StepGeneralProps {
  name: string;
  scheduledAt: string;
  onNameChange: (v: string) => void;
  onScheduledAtChange: (v: string) => void;
}

export const StepGeneral = ({
  name,
  scheduledAt,
  onNameChange,
  onScheduledAtChange,
}: StepGeneralProps) => {
  // Compute minimum datetime (now) for the picker
  const [minDateTime] = useState(() =>
    new Date(Date.now() + 60_000).toISOString().slice(0, 16)
  );

  return (
    <div className="flex flex-col gap-5">
      <div className="flex flex-col gap-1.5">
        <label htmlFor="campaign-name" className="text-xs font-mono text-neutral-400 uppercase tracking-wider">
          Название кампании *
        </label>
        <input
          id="campaign-name"
          type="text"
          required
          value={name}
          onChange={(e) => onNameChange(e.target.value)}
          placeholder="Промо-рассылка Апрель 2026"
          maxLength={200}
          className="bg-neutral-900 border border-neutral-700 focus:border-accent-cyan rounded-lg
            px-4 py-3 text-white text-sm outline-none transition-colors placeholder:text-neutral-600"
        />
      </div>

      <div className="flex flex-col gap-1.5">
        <label htmlFor="campaign-schedule" className="text-xs font-mono text-neutral-400 uppercase tracking-wider">
          Запуск (по расписанию)
          <span className="text-neutral-600 normal-case ml-1">— необязательно, оставьте пустым для немедленного запуска</span>
        </label>
        <input
          id="campaign-schedule"
          type="datetime-local"
          value={scheduledAt}
          min={minDateTime}
          onChange={(e) => onScheduledAtChange(e.target.value)}
          className="bg-neutral-900 border border-neutral-700 focus:border-accent-cyan rounded-lg
            px-4 py-3 text-white text-sm outline-none transition-colors
            [color-scheme:dark]"
        />
      </div>

      {scheduledAt && (
        <div className="flex items-center gap-2 text-xs text-accent-cyan bg-accent-cyan/5 border border-accent-cyan/20 rounded-lg px-4 py-3">
          <span>🕐</span>
          <span>
            Кампания будет запущена:{' '}
            <strong>
              {new Date(scheduledAt).toLocaleString('ru-RU', {
                dateStyle: 'long',
                timeStyle: 'short',
              })}
            </strong>
          </span>
        </div>
      )}
    </div>
  );
};

interface StepTemplatesProps {
  telegramTemplate: string;
  smsTemplate: string;
  onTelegramChange: (v: string) => void;
  onSmsChange: (v: string) => void;
}

const TELEGRAM_MAX = 4096;
const SMS_MAX = 160;

const TemplateArea = ({
  id,
  label,
  description,
  value,
  onChange,
  maxLength,
  placeholder,
  icon,
}: {
  id: string;
  label: string;
  description: string;
  value: string;
  onChange: (v: string) => void;
  maxLength: number;
  placeholder: string;
  icon: string;
}) => {
  const remaining = maxLength - value.length;
  const isNearLimit = remaining < maxLength * 0.1;

  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-start justify-between">
        <div>
          <label htmlFor={id} className="flex items-center gap-2 text-sm font-medium text-white">
            <span>{icon}</span>
            {label}
          </label>
          <p className="text-xs text-neutral-500 mt-0.5">{description}</p>
        </div>
        <span className={`text-xs font-mono mt-1 ${isNearLimit ? 'text-orange-400' : 'text-neutral-600'}`}>
          {value.length}/{maxLength}
        </span>
      </div>
      <textarea
        id={id}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        maxLength={maxLength}
        rows={5}
        placeholder={placeholder}
        className="bg-neutral-900 border border-neutral-700 focus:border-accent-cyan rounded-lg
          px-4 py-3 text-white text-sm outline-none transition-colors resize-y font-mono
          placeholder:text-neutral-600 placeholder:font-sans leading-relaxed"
      />
      {value.length === 0 && (
        <p className="text-xs text-neutral-600 italic">
          Переменные: {'{'}name{'}'}, {'{'}phone{'}'}, {'{'}extra_data.key{'}'}
        </p>
      )}
    </div>
  );
};

export const StepTemplates = ({
  telegramTemplate,
  smsTemplate,
  onTelegramChange,
  onSmsChange,
}: StepTemplatesProps) => {
  const isValid = telegramTemplate.trim().length > 0 || smsTemplate.trim().length > 0;

  return (
    <div className="flex flex-col gap-6">
      {!isValid && (
        <div className="text-xs text-orange-400 bg-orange-500/10 border border-orange-500/20 rounded-lg px-4 py-3">
          ⚠ Необходимо заполнить хотя бы один канал — Telegram или SMS.
        </div>
      )}

      <TemplateArea
        id="template-telegram"
        label="Telegram"
        description="Шаблон сообщения для TG канала (Markdown поддерживается)"
        value={telegramTemplate}
        onChange={onTelegramChange}
        maxLength={TELEGRAM_MAX}
        placeholder={`Привет, {name}!\n\nЭто ваше персональное предложение...\n\nПодробности: https://example.com`}
        icon="✈"
      />

      <div className="border-t border-neutral-800" />

      <TemplateArea
        id="template-sms"
        label="SMS"
        description="Шаблон для SMS (запасной канал). Рекомендуется до 160 символов."
        value={smsTemplate}
        onChange={onSmsChange}
        maxLength={SMS_MAX}
        placeholder={`Привет, {name}! Предложение для вас: https://ex.co`}
        icon="📱"
      />

      {smsTemplate.length > SMS_MAX * 0.9 && (
        <p className="text-xs text-orange-400">
          SMS свыше {SMS_MAX} символов будут тарифицироваться как составные (×2 стоимость).
        </p>
      )}
    </div>
  );
};

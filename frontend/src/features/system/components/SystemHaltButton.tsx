import { useState } from 'react';
import { Modal } from '../../../shared/components/Modal';
import { useSystemHalt, useSystemResume } from '../hooks/useSystemControl';

type Mode = 'halt' | 'resume';

interface BreakGlassModalProps {
  isOpen: boolean;
  onClose: () => void;
  mode: Mode;
}

const BreakGlassModal = ({ isOpen, onClose, mode }: BreakGlassModalProps) => {
  const [password, setPassword] = useState('');
  const [reason, setReason] = useState('');

  const halt = useSystemHalt(onClose);
  const resume = useSystemResume(onClose);

  const mutation = mode === 'halt' ? halt : resume;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!password || !reason) return;
    mutation.mutate({ password, reason });
  };

  const handleClose = () => {
    setPassword('');
    setReason('');
    onClose();
  };

  const isHalt = mode === 'halt';

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title={isHalt ? '⚠ АВАРИЙНАЯ ОСТАНОВКА СИСТЕМЫ' : '▶ ВОЗОБНОВЛЕНИЕ СИСТЕМЫ'}
      width="max-w-md"
    >
      <form onSubmit={handleSubmit} className="px-6 py-5 flex flex-col gap-4">
        <p className="text-sm text-neutral-300 leading-relaxed">
          {isHalt
            ? 'Это действие немедленно остановит все активные кампании и заморозит очередь сообщений. Все действия логируются и требуют повторной аутентификации.'
            : 'Возобновление системы восстановит очередь сообщений. Убедитесь, что причина инцидента устранена.'}
        </p>

        <div className="flex flex-col gap-1.5">
          <label htmlFor="bg-password" className="text-xs font-mono text-neutral-400">
            Пароль оператора (повторная аутентификация) *
          </label>
          <input
            id="bg-password"
            type="password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="••••••••"
            className={`bg-neutral-900 border rounded-md px-3 py-2 text-sm text-white outline-none transition-colors
              ${isHalt ? 'border-danger/50 focus:border-danger' : 'border-neutral-700 focus:border-accent-cyan'}`}
          />
        </div>

        <div className="flex flex-col gap-1.5">
          <label htmlFor="bg-reason" className="text-xs font-mono text-neutral-400">
            Причина действия * <span className="text-neutral-600">(будет записана в журнал аудита)</span>
          </label>
          <textarea
            id="bg-reason"
            required
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            rows={3}
            placeholder="Опишите причину..."
            className={`bg-neutral-900 border rounded-md px-3 py-2 text-sm text-white outline-none transition-colors resize-none
              ${isHalt ? 'border-danger/50 focus:border-danger' : 'border-neutral-700 focus:border-accent-cyan'}`}
          />
        </div>

        {mutation.isError && (
          <p className="text-danger text-xs font-mono">
            {(mutation.error as Error).message}
          </p>
        )}

        <div className="flex gap-3 justify-end pt-1">
          <button
            type="button"
            onClick={handleClose}
            disabled={mutation.isPending}
            className="px-4 py-2 text-sm text-neutral-400 hover:text-white border border-neutral-700
              hover:border-neutral-500 rounded-md transition-colors disabled:opacity-50"
          >
            Отмена
          </button>
          <button
            id={`break-glass-${mode}-submit`}
            type="submit"
            disabled={!password || !reason || mutation.isPending}
            className={`px-5 py-2 text-sm font-bold rounded-md transition-colors
              flex items-center gap-2 disabled:opacity-40 disabled:cursor-not-allowed
              ${isHalt
                ? 'bg-danger/10 hover:bg-danger/20 border border-danger text-danger'
                : 'bg-accent-cyan/10 hover:bg-accent-cyan/20 border border-accent-cyan text-accent-cyan'
              }`}
          >
            {mutation.isPending && (
              <span className={`inline-block w-4 h-4 border-2 rounded-full animate-spin
                ${isHalt ? 'border-danger/30 border-t-danger' : 'border-accent-cyan/30 border-t-accent-cyan'}`}
              />
            )}
            {isHalt ? 'ОСТАНОВИТЬ СИСТЕМУ' : 'ВОЗОБНОВИТЬ СИСТЕМУ'}
          </button>
        </div>
      </form>
    </Modal>
  );
};

export const SystemHaltButton = () => {
  const [modalOpen, setModalOpen] = useState(false);
  const [mode, setMode] = useState<Mode>('halt');

  const openHalt = () => { setMode('halt'); setModalOpen(true); };
  const openResume = () => { setMode('resume'); setModalOpen(true); };

  return (
    <>
      <div className="flex gap-2">
        <button
          id="system-resume-btn"
          onClick={openResume}
          className="bg-accent-cyan/10 hover:bg-accent-cyan/20 border border-accent-cyan/50
            text-accent-cyan px-3 py-1.5 rounded-md font-bold text-xs transition-colors uppercase tracking-wider"
        >
          ▶ Resume
        </button>
        <button
          id="system-halt-btn"
          onClick={openHalt}
          className="bg-danger/10 hover:bg-danger/20 border border-danger text-danger
            px-3 py-1.5 rounded-md font-bold text-xs transition-colors uppercase tracking-wider"
        >
          ⛔ Break Glass Halt
        </button>
      </div>

      <BreakGlassModal
        isOpen={modalOpen}
        onClose={() => setModalOpen(false)}
        mode={mode}
      />
    </>
  );
};

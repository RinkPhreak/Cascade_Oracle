import { useState } from 'react';
import { Modal } from '../../../shared/components/Modal';
import { useInitAccountSession, useSessionStatus, useSubmitPairingCode } from '../hooks/useAccounts';

type Tab = 'qr' | 'code';

interface AddAccountModalProps {
  isOpen: boolean;
  onClose: () => void;
}

const QrPanel = ({ sessionId, qrUrl }: { sessionId: string; qrUrl: string | null }) => {
  const { data: session } = useSessionStatus(sessionId);
  const activeQrUrl = session?.qr_url ?? qrUrl;
  const status = session?.status;

  if (status === 'AUTHENTICATED') {
    return (
      <div className="flex flex-col items-center gap-4 py-8">
        <div className="w-16 h-16 rounded-full bg-green-500/20 border border-green-500/40 flex items-center justify-center">
          <span className="text-2xl text-green-400">✓</span>
        </div>
        <p className="text-green-400 font-semibold text-center">
          Аккаунт успешно добавлен в пул!
        </p>
      </div>
    );
  }

  if (status === 'FAILED') {
    return (
      <div className="flex flex-col items-center gap-4 py-8">
        <div className="w-16 h-16 rounded-full bg-danger/20 border border-danger/40 flex items-center justify-center">
          <span className="text-2xl text-red-400">✕</span>
        </div>
        <p className="text-danger text-sm text-center">
          Авторизация не удалась. Закройте диалог и попробуйте снова.
        </p>
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center gap-5 py-4">
      <div className="relative">
        {activeQrUrl ? (
          <div className="w-52 h-52 bg-white rounded-xl p-2 flex items-center justify-center">
            <img
              src={activeQrUrl}
              alt="Telegram QR код для авторизации"
              className="w-full h-full object-contain"
            />
          </div>
        ) : (
          <div className="w-52 h-52 bg-neutral-900 border border-neutral-700 rounded-xl flex items-center justify-center">
            <div className="w-10 h-10 border-2 border-accent-cyan/30 border-t-accent-cyan rounded-full animate-spin" />
          </div>
        )}
        <div className="absolute -bottom-2 left-1/2 -translate-x-1/2 flex items-center gap-1.5 bg-bg-surface px-3 py-1 rounded-full border border-neutral-700">
          <span className="w-1.5 h-1.5 rounded-full bg-accent-cyan animate-pulse" />
          <span className="text-xs text-neutral-400 font-mono">Ожидание сканирования...</span>
        </div>
      </div>

      <div className="text-center space-y-1.5 pt-3">
        <p className="text-sm text-neutral-300">
          Откройте Telegram на телефоне → <span className="text-accent-cyan">Настройки → Устройства → Подключить устройство</span>
        </p>
        <p className="text-xs text-neutral-600">QR-код обновляется автоматически каждые 30 сек.</p>
      </div>
    </div>
  );
};

const PairingCodePanel = ({ sessionId }: { sessionId: string }) => {
  const [code, setCode] = useState('');
  const submitCode = useSubmitPairingCode();
  const { data: session } = useSessionStatus(sessionId);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!code.trim()) return;
    submitCode.mutate({ sessionId, code: code.trim() });
  };

  if (session?.status === 'AUTHENTICATED') {
    return (
      <div className="flex flex-col items-center gap-4 py-8">
        <span className="text-2xl text-green-400">✓</span>
        <p className="text-green-400 font-semibold">Аккаунт успешно добавлен!</p>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-5 py-4">
      <p className="text-sm text-neutral-300 leading-relaxed">
        Введите код сопряжения из приложения Telegram. Код появится в настройках активных сессий.
      </p>
      <div className="flex flex-col gap-2">
        <label htmlFor="pairing-code-input" className="text-xs font-mono text-neutral-400">
          Код сопряжения
        </label>
        <input
          id="pairing-code-input"
          type="text"
          value={code}
          onChange={(e) => setCode(e.target.value)}
          placeholder="Например: ABCDE-12345"
          maxLength={20}
          className="bg-neutral-900 border border-neutral-700 focus:border-accent-cyan rounded-lg
            px-4 py-3 text-white text-center text-lg font-mono tracking-widest outline-none transition-colors"
          autoComplete="off"
          autoFocus
        />
      </div>

      {submitCode.isError && (
        <p className="text-danger text-xs font-mono text-center">
          {(submitCode.error as Error).message}
        </p>
      )}

      <button
        type="submit"
        id="submit-pairing-code-btn"
        disabled={!code.trim() || submitCode.isPending}
        className="w-full bg-accent-cyan/10 hover:bg-accent-cyan/20 border border-accent-cyan text-accent-cyan
          font-bold py-2.5 rounded-lg transition-colors disabled:opacity-40 disabled:cursor-not-allowed
          flex items-center justify-center gap-2"
      >
        {submitCode.isPending && (
          <span className="w-4 h-4 border-2 border-accent-cyan/30 border-t-accent-cyan rounded-full animate-spin" />
        )}
        Подтвердить
      </button>
    </form>
  );
};

export const AddAccountModal = ({ isOpen, onClose }: AddAccountModalProps) => {
  const [tab, setTab] = useState<Tab>('qr');
  const initSession = useInitAccountSession();

  // Trigger session init when modal opens
  const handleClose = () => {
    initSession.reset();
    onClose();
  };

  // Init session once on open
  if (isOpen && !initSession.data && !initSession.isPending && !initSession.isError) {
    initSession.mutate();
  }

  const session = initSession.data;

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Добавить TG аккаунт"
      width="max-w-md"
    >
      <div className="px-6 py-5 flex flex-col gap-4">
        {/* Tab switcher */}
        <div className="flex gap-1 bg-neutral-900 rounded-lg p-1">
          {(['qr', 'code'] as const).map((t) => (
            <button
              key={t}
              id={`tab-${t}`}
              onClick={() => setTab(t)}
              className={`flex-1 py-2 text-sm font-medium rounded-md transition-colors ${
                tab === t
                  ? 'bg-accent-cyan/10 text-accent-cyan border border-accent-cyan/20'
                  : 'text-neutral-400 hover:text-white'
              }`}
            >
              {t === 'qr' ? '📱 QR Код' : '🔑 Код сопряжения'}
            </button>
          ))}
        </div>

        {/* Session loading / error */}
        {initSession.isPending && (
          <div className="flex items-center justify-center py-12">
            <div className="w-8 h-8 border-2 border-accent-cyan/30 border-t-accent-cyan rounded-full animate-spin" />
          </div>
        )}

        {initSession.isError && (
          <div className="text-center py-8 flex flex-col gap-3">
            <p className="text-danger text-sm">{(initSession.error as Error).message}</p>
            <button
              onClick={() => initSession.mutate()}
              className="text-accent-cyan text-xs hover:underline"
            >
              Повторить попытку
            </button>
          </div>
        )}

        {session && tab === 'qr' && (
          <QrPanel sessionId={session.session_id} qrUrl={session.qr_url} />
        )}

        {session && tab === 'code' && (
          <PairingCodePanel sessionId={session.session_id} />
        )}
      </div>
    </Modal>
  );
};

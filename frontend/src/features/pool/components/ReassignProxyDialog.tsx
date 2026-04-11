import { useState } from 'react';
import type { DtoProxy } from '../../../api/generated';
import { useReassignProxy } from '../hooks/useProxies';
import { ConfirmDialog } from '../../../shared/components/ConfirmDialog';
import { Modal } from '../../../shared/components/Modal';

interface ReassignProxyDialogProps {
  isOpen: boolean;
  onClose: () => void;
  proxy: DtoProxy | null;
  accounts: { id: string; phone: string }[];
}

export const ReassignProxyDialog = ({
  isOpen,
  onClose,
  proxy,
  accounts,
}: ReassignProxyDialogProps) => {
  const [selectedAccountId, setSelectedAccountId] = useState('');
  const [reason, setReason] = useState('');
  const [step, setStep] = useState<'form' | 'confirm'>('form');

  const reassign = useReassignProxy();

  const handleClose = () => {
    setSelectedAccountId('');
    setReason('');
    setStep('form');
    onClose();
  };

  const handleFormSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedAccountId || !reason.trim()) return;
    setStep('confirm');
  };

  const handleConfirm = () => {
    if (!proxy || !selectedAccountId) return;
    reassign.mutate(
      { proxyId: proxy.id, account_id: selectedAccountId, reason },
      { onSuccess: handleClose }
    );
  };

  if (step === 'confirm') {
    return (
      <ConfirmDialog
        isOpen={isOpen}
        onClose={() => setStep('form')}
        onConfirm={handleConfirm}
        title="Подтвердите смену привязки"
        description={`Прокси ${proxy?.host}:${proxy?.port} будет перепривязан к новому аккаунту. Причина: "${reason}". Это действие будет записано в журнал оператора.`}
        confirmLabel="Выполнить перепривязку"
        isLoading={reassign.isPending}
      />
    );
  }

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title={`Перепривязать прокси: ${proxy?.host ?? '—'}`}
      width="max-w-md"
    >
      <form
        id="reassign-proxy-form"
        onSubmit={handleFormSubmit}
        className="px-6 py-5 flex flex-col gap-4"
      >
        <p className="text-sm text-neutral-400 leading-relaxed">
          Текущая привязка:{' '}
          <span className="text-white font-mono">
            {accounts.find((a) => a.id === proxy?.bound_account_id)?.phone ?? 'Нет'}
          </span>
        </p>

        <div className="flex flex-col gap-1.5">
          <label htmlFor="reassign-account-select" className="text-xs text-neutral-400 font-mono">
            Привязать к аккаунту *
          </label>
          <select
            id="reassign-account-select"
            required
            value={selectedAccountId}
            onChange={(e) => setSelectedAccountId(e.target.value)}
            className="bg-neutral-900 border border-neutral-700 focus:border-accent-cyan rounded-md
              px-3 py-2 text-sm text-white outline-none transition-colors"
          >
            <option value="">Выберите аккаунт...</option>
            {accounts.map((a) => (
              <option key={a.id} value={a.id}>
                {a.phone}
              </option>
            ))}
          </select>
        </div>

        <div className="flex flex-col gap-1.5">
          <label htmlFor="reassign-reason" className="text-xs text-neutral-400 font-mono">
            Причина перепривязки * <span className="text-neutral-600">(журнал оператора)</span>
          </label>
          <textarea
            id="reassign-reason"
            required
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            rows={3}
            placeholder="Например: смена IP, ротация прокси..."
            className="bg-neutral-900 border border-neutral-700 focus:border-accent-cyan rounded-md
              px-3 py-2 text-sm text-white outline-none transition-colors resize-none"
          />
        </div>

        <div className="flex gap-3 justify-end">
          <button
            type="button"
            onClick={handleClose}
            className="px-4 py-2 text-sm text-neutral-400 hover:text-white border border-neutral-700
              hover:border-neutral-500 rounded-md transition-colors"
          >
            Отмена
          </button>
          <button
            type="submit"
            id="reassign-proxy-next-btn"
            disabled={!selectedAccountId || !reason.trim()}
            className="px-4 py-2 text-sm font-bold bg-accent-cyan/10 hover:bg-accent-cyan/20
              border border-accent-cyan text-accent-cyan rounded-md transition-colors
              disabled:opacity-40 disabled:cursor-not-allowed"
          >
            Далее →
          </button>
        </div>
      </form>
    </Modal>
  );
};

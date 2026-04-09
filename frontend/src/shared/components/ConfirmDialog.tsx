import { useState } from 'react';
import { Modal } from './Modal';

interface ConfirmDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: () => void;
  title: string;
  description: string;
  /**
   * If provided, the user must type this exact string to enable the confirm button.
   * Used for double-confirmation on destructive operations.
   */
  confirmPhrase?: string;
  confirmLabel?: string;
  isLoading?: boolean;
}

export const ConfirmDialog = ({
  isOpen,
  onClose,
  onConfirm,
  title,
  description,
  confirmPhrase,
  confirmLabel = 'Подтвердить',
  isLoading = false,
}: ConfirmDialogProps) => {
  const [phraseInput, setPhraseInput] = useState('');

  const isConfirmEnabled = confirmPhrase
    ? phraseInput === confirmPhrase
    : true;

  const handleConfirm = () => {
    if (!isConfirmEnabled || isLoading) return;
    onConfirm();
    setPhraseInput('');
  };

  const handleClose = () => {
    setPhraseInput('');
    onClose();
  };

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title={title} width="max-w-md">
      <div className="px-6 py-5 flex flex-col gap-5">
        <p className="text-neutral-300 text-sm leading-relaxed">{description}</p>

        {confirmPhrase && (
          <div className="flex flex-col gap-2">
            <label className="text-xs text-neutral-400 font-mono">
              Введите <span className="text-danger font-bold">{confirmPhrase}</span> для подтверждения:
            </label>
            <input
              type="text"
              id="confirm-phrase-input"
              value={phraseInput}
              onChange={(e) => setPhraseInput(e.target.value)}
              placeholder={confirmPhrase}
              className="w-full bg-neutral-900 border border-neutral-700 focus:border-danger
                rounded-md px-3 py-2 text-sm text-white outline-none font-mono
                transition-colors placeholder:text-neutral-600"
              autoComplete="off"
            />
          </div>
        )}

        <div className="flex gap-3 justify-end pt-1">
          <button
            onClick={handleClose}
            disabled={isLoading}
            className="px-4 py-2 text-sm text-neutral-400 hover:text-white border border-neutral-700
              hover:border-neutral-500 rounded-md transition-colors disabled:opacity-50"
          >
            Отмена
          </button>
          <button
            id="confirm-dialog-submit"
            onClick={handleConfirm}
            disabled={!isConfirmEnabled || isLoading}
            className="px-4 py-2 text-sm font-bold bg-danger/10 hover:bg-danger/20 border border-danger
              text-danger rounded-md transition-colors disabled:opacity-40 disabled:cursor-not-allowed
              flex items-center gap-2"
          >
            {isLoading && (
              <span className="inline-block w-4 h-4 border-2 border-danger/30 border-t-danger rounded-full animate-spin" />
            )}
            {confirmLabel}
          </button>
        </div>
      </div>
    </Modal>
  );
};

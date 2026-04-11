import { useState } from 'react';
import type { DtoContact as Contact } from '../../../api/generated';
import { useAnonymiseContact } from '../hooks/useAnonymiseContact';
import { ConfirmDialog } from '../../../shared/components/ConfirmDialog';

interface AnonymiseButtonProps {
  contact: Contact;
  onSuccess?: () => void;
}

const CONFIRM_PHRASE = 'АНОНИМИЗИРОВАТЬ';

export const AnonymiseButton = ({ contact, onSuccess }: AnonymiseButtonProps) => {
  const [step, setStep] = useState<'idle' | 'confirm'>('idle');
  const anonymise = useAnonymiseContact(() => {
    setStep('idle');
    onSuccess?.();
  });

  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    setStep('confirm');
  };

  return (
    <>
      <button
        id={`anonymise-btn-${contact.id}`}
        onClick={handleClick}
        className="px-3 py-1.5 text-xs font-bold font-mono bg-danger/10 hover:bg-danger/20
          border border-danger/40 hover:border-danger text-danger rounded-md transition-all
          uppercase tracking-wide"
        title="Анонимизировать контакт (152-ФЗ)"
      >
        Анонимизировать (152-ФЗ)
      </button>

      <ConfirmDialog
        isOpen={step === 'confirm'}
        onClose={() => setStep('idle')}
        onConfirm={() => anonymise.mutate(contact.id)}
        title="Право на забвение (152-ФЗ)"
        description={`Все персональные данные контакта ${contact.phone} (имя, текст ответа, дополнительные данные) будут необратимо уничтожены. Это действие нельзя отменить.`}
        confirmPhrase={CONFIRM_PHRASE}
        confirmLabel="Анонимизировать навсегда"
        isLoading={anonymise.isPending}
      />
    </>
  );
};

import { useState } from 'react';
import { Modal } from '../../../../shared/components/Modal';
import { StepGeneral } from './StepGeneral';
import { StepAudience, type ParsedContact } from './StepAudience';
import { StepTemplates } from './StepTemplates';
import { useCreateCampaign } from '../../hooks/useCreateCampaign';
import { useImportContacts } from '../../hooks/useImportContacts';
import type { DtoCampaignResponse } from '../../../../api/generated';

interface CampaignWizardProps {
  isOpen: boolean;
  onClose: () => void;
  onCreated?: (campaign: DtoCampaignResponse) => void;
}

const STEPS = ['Общие', 'Аудитория', 'Шаблоны'] as const;

export const CampaignWizard = ({ isOpen, onClose, onCreated }: CampaignWizardProps) => {
  const [step, setStep] = useState(0);

  // Step 1 state
  const [name, setName] = useState('');
  const [scheduledAt, setScheduledAt] = useState('');

  // Step 2 state
  const [csvFile, setCsvFile] = useState<File | null>(null);
  const [parsedContacts, setParsedContacts] = useState<ParsedContact[]>([]);

  // Step 3 state
  const [telegramTemplate, setTelegramTemplate] = useState('');
  const [smsTemplate, setSmsTemplate] = useState('');

  // Created campaign ID (after step 1 save)
  const [createdCampaignId, setCreatedCampaignId] = useState<string | null>(null);

  const createCampaign = useCreateCampaign((campaign) => {
    setCreatedCampaignId(campaign.id ?? null);
    setStep(1);
  });

  const importContacts = useImportContacts(createdCampaignId ?? '');

  const handleClose = () => {
    // Reset all state
    setStep(0);
    setName('');
    setScheduledAt('');
    setCsvFile(null);
    setParsedContacts([]);
    setTelegramTemplate('');
    setSmsTemplate('');
    setCreatedCampaignId(null);
    createCampaign.reset();
    importContacts.reset();
    onClose();
  };

  const canProceedStep0 = name.trim().length > 0;
  const canProceedStep1 = !!csvFile && parsedContacts.length > 0;
  const canProceedStep2 =
    telegramTemplate.trim().length > 0 || smsTemplate.trim().length > 0;

  const handleNext = async () => {
    if (step === 0) {
      // Create campaign with templates (templates required on creation per API spec)
      // We'll pass placeholders and update in final step
      const templates: Record<string, string> = {};
      if (telegramTemplate) templates.telegram = telegramTemplate;
      if (smsTemplate) templates.sms = smsTemplate;

      createCampaign.mutate({
        name: name.trim(),
        scheduled_at: scheduledAt || undefined,
        // Templates are required by the API, pass empty for now — user fills in step 3
        templates: templates,
      });
      return;
    }

    if (step === 1) {
      if (csvFile && createdCampaignId) {
        importContacts.mutate(csvFile, {
          onSuccess: () => setStep(2),
        });
      } else {
        setStep(2);
      }
      return;
    }

    if (step === 2) {
      // Final step: update campaign templates
      // Templates were passed at creation; if user changed them we'd patch here
      // For now, finalize
      if (onCreated && createdCampaignId) {
        onCreated({
          id: createdCampaignId,
          name,
          status: 'DRAFT',
          scheduled_at: scheduledAt || undefined,
          created_at: new Date().toISOString(),
        });
      }
      handleClose();
    }
  };

  const isNextLoading =
    createCampaign.isPending || importContacts.isPending;

  const isNextDisabled = (() => {
    if (step === 0) return !canProceedStep0 || isNextLoading;
    if (step === 1) return !canProceedStep1 || isNextLoading;
    if (step === 2) return !canProceedStep2 || isNextLoading;
    return false;
  })();

  return (
    <Modal
      isOpen={isOpen}
      onClose={handleClose}
      title="Новая кампания"
      width="max-w-2xl"
    >
      <div className="flex flex-col">
        {/* Step indicator */}
        <div className="px-6 pt-5 pb-4 border-b border-neutral-800">
          <div className="flex items-center gap-2">
            {STEPS.map((label, i) => (
              <div key={label} className="flex items-center gap-2">
                <div
                  className={`flex items-center justify-center w-7 h-7 rounded-full text-xs font-bold font-mono transition-colors ${
                    i < step
                      ? 'bg-green-500 text-black'
                      : i === step
                      ? 'bg-accent-cyan text-black'
                      : 'bg-neutral-800 text-neutral-500'
                  }`}
                >
                  {i < step ? '✓' : i + 1}
                </div>
                <span
                  className={`text-sm ${
                    i === step ? 'text-white font-medium' : 'text-neutral-500'
                  }`}
                >
                  {label}
                </span>
                {i < STEPS.length - 1 && (
                  <div className={`flex-1 h-px w-8 ${i < step ? 'bg-green-500/50' : 'bg-neutral-800'}`} />
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Step content */}
        <div className="px-6 py-6 min-h-[300px]">
          {step === 0 && (
            <StepGeneral
              name={name}
              scheduledAt={scheduledAt}
              onNameChange={setName}
              onScheduledAtChange={setScheduledAt}
            />
          )}
          {step === 1 && (
            <StepAudience
              file={csvFile}
              parsedContacts={parsedContacts}
              onFileChange={(f, parsed) => {
                setCsvFile(f);
                setParsedContacts(parsed);
              }}
            />
          )}
          {step === 2 && (
            <StepTemplates
              telegramTemplate={telegramTemplate}
              smsTemplate={smsTemplate}
              onTelegramChange={setTelegramTemplate}
              onSmsChange={setSmsTemplate}
            />
          )}
        </div>

        {/* Footer navigation */}
        <div className="px-6 py-4 border-t border-neutral-800 flex items-center justify-between">
          <button
            onClick={() => (step > 0 ? setStep(step - 1) : handleClose())}
            className="px-4 py-2 text-sm text-neutral-400 hover:text-white border border-neutral-700
              hover:border-neutral-500 rounded-lg transition-colors"
          >
            {step === 0 ? 'Отмена' : '← Назад'}
          </button>

          <button
            id={`wizard-next-step-${step}`}
            onClick={handleNext}
            disabled={isNextDisabled}
            className="px-6 py-2 text-sm font-bold bg-accent-cyan text-black rounded-lg
              hover:bg-accent-cyan-hover transition-colors disabled:opacity-40 disabled:cursor-not-allowed
              flex items-center gap-2"
          >
            {isNextLoading && (
              <span className="w-4 h-4 border-2 border-black/20 border-t-black rounded-full animate-spin" />
            )}
            {step === 2 ? 'Создать кампанию ✓' : 'Далее →'}
          </button>
        </div>
      </div>
    </Modal>
  );
};

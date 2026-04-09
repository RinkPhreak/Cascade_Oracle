import { useState } from 'react';
import { CampaignList } from '../features/campaigns/components/CampaignList';
import { CampaignMonitor } from '../features/campaigns/components/CampaignMonitor';
import { CampaignWizard } from '../features/campaigns/components/CampaignWizard/CampaignWizard';
import { useCampaigns } from '../features/campaigns/hooks/useCampaigns';
import type { Campaign } from '../api/extended-types';

export const CampaignsPage = () => {
  const { data: campaigns = [], isLoading } = useCampaigns();
  const [selectedCampaign, setSelectedCampaign] = useState<Campaign | null>(null);
  const [wizardOpen, setWizardOpen] = useState(false);

  // Stats
  const runningCount = campaigns.filter((c) => c.status === 'RUNNING').length;

  return (
    <div className="flex flex-col gap-6">
      {/* Page Header */}
      <div>
        <h2 className="text-2xl font-bold text-white tracking-tight">Кампании</h2>
        <p className="text-neutral-400 text-sm mt-1">
          Управление рассылками и мониторинг waterfall доставки
        </p>
      </div>

      {/* Stats row */}
      <div className="grid grid-cols-3 gap-4">
        {[
          { label: 'Всего кампаний', value: campaigns.length, color: 'text-white' },
          { label: 'Активных сейчас', value: runningCount, color: runningCount > 0 ? 'text-accent-cyan' : 'text-neutral-400' },
          {
            label: 'Завершённых',
            value: campaigns.filter((c) => c.status === 'COMPLETED').length,
            color: 'text-green-400',
          },
        ].map(({ label, value, color }) => (
          <div key={label} className="bg-bg-surface border border-neutral-800 rounded-xl p-4">
            <p className="text-xs text-neutral-500 font-mono uppercase tracking-wider mb-1">{label}</p>
            <p className={`text-2xl font-bold font-mono ${color}`}>{value}</p>
          </div>
        ))}
      </div>

      {/* Main content: list or monitor */}
      {selectedCampaign ? (
        <CampaignMonitor
          campaign={selectedCampaign}
          onBack={() => setSelectedCampaign(null)}
        />
      ) : (
        <CampaignList
          campaigns={campaigns}
          isLoading={isLoading}
          onSelect={setSelectedCampaign}
          onNew={() => setWizardOpen(true)}
        />
      )}

      <CampaignWizard
        isOpen={wizardOpen}
        onClose={() => setWizardOpen(false)}
        onCreated={(c) => {
          const newCampaign: Campaign = {
            id: c.id ?? '',
            name: c.name ?? '',
            status: (c.status as Campaign['status']) ?? 'DRAFT',
            scheduled_at: c.scheduled_at ?? null,
            created_at: c.created_at ?? new Date().toISOString(),
            total_contacts: 0,
            completed: 0,
            replied: 0,
            failed: 0,
          };
          setSelectedCampaign(newCampaign);
        }}
      />
    </div>
  );
};

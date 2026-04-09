import { useMutation, useQueryClient } from '@tanstack/react-query';
import { postApiV1CampaignsByIdPause } from '../../../api/generated';
import { CAMPAIGNS_QUERY_KEY } from './useCampaigns';
import { toast } from '../../../shared/hooks/useToast';

export const usePauseCampaign = (onSuccess?: () => void) => {
  const qc = useQueryClient();
  return useMutation<void, Error, { campaignId: string; password: string; reason: string }>({
    mutationFn: async ({ campaignId, password, reason }) => {
      const { error } = await postApiV1CampaignsByIdPause({
        path: { id: campaignId },
        body: { password, reason },
      });
      if (error) {
        const errMsg = (error as { message?: string }).message ?? 'Failed to pause campaign';
        throw new Error(errMsg);
      }
    },
    onSuccess: () => {
      toast.success('Кампания приостановлена. Break-Glass событие зафиксировано.');
      qc.invalidateQueries({ queryKey: CAMPAIGNS_QUERY_KEY });
      onSuccess?.();
    },
    onError: (err) => toast.error(err.message),
  });
};

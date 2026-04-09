import { useMutation, useQueryClient } from '@tanstack/react-query';
import { postApiV1CampaignsByIdStart } from '../../../api/generated';
import { CAMPAIGNS_QUERY_KEY } from './useCampaigns';
import { toast } from '../../../shared/hooks/useToast';

export const useStartCampaign = () => {
  const qc = useQueryClient();
  return useMutation<void, Error, string>({
    mutationFn: async (campaignId) => {
      const { error } = await postApiV1CampaignsByIdStart({
        path: { id: campaignId },
      });
      if (error) {
        const errMsg = (error as { message?: string }).message ?? 'Failed to start campaign';
        throw new Error(errMsg);
      }
    },
    onSuccess: () => {
      toast.success('Кампания поставлена в очередь на выполнение.');
      qc.invalidateQueries({ queryKey: CAMPAIGNS_QUERY_KEY });
    },
    onError: (err) => toast.error(err.message),
  });
};

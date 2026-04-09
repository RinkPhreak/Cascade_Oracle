import { useMutation, useQueryClient } from '@tanstack/react-query';
import { postApiV1Campaigns } from '../../../api/generated';
import type { DtoCampaignResponse, DtoCreateCampaignRequest } from '../../../api/generated';
import { CAMPAIGNS_QUERY_KEY } from './useCampaigns';
import { toast } from '../../../shared/hooks/useToast';

export const useCreateCampaign = (onSuccess?: (campaign: DtoCampaignResponse) => void) => {
  const qc = useQueryClient();
  return useMutation<DtoCampaignResponse, Error, DtoCreateCampaignRequest>({
    mutationFn: async (body) => {
      const { data, error } = await postApiV1Campaigns({ body });
      if (error) {
        const errMsg = (error as { message?: string }).message ?? 'Failed to create campaign';
        throw new Error(errMsg);
      }
      return data!;
    },
    onSuccess: (campaign) => {
      toast.success(`Кампания "${campaign.name}" создана.`);
      qc.invalidateQueries({ queryKey: CAMPAIGNS_QUERY_KEY });
      onSuccess?.(campaign);
    },
    onError: (err) => toast.error(err.message),
  });
};

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import type { DtoCampaign } from '../../../api/generated';
import { client } from '../../../api/client';
import { toast } from '../../../shared/hooks/useToast';

export const CAMPAIGNS_QUERY_KEY = ['campaigns'] as const;

/** Fetch all campaigns, polling every 15s. */
export const useCampaigns = () =>
  useQuery<DtoCampaign[]>({
    queryKey: CAMPAIGNS_QUERY_KEY,
    queryFn: async () => {
      const response = await client.get<DtoCampaign[], unknown>({
        url: '/api/v1/campaigns',
      });
      const result = response as { data?: DtoCampaign[]; error?: { message?: string } };
      if (result.error) {
        return [];
      }
      return result.data ?? [];
    },
    refetchInterval: 15_000,
  });

/** Delete a campaign and its associated data. */
export const useDeleteCampaign = () => {
  const qc = useQueryClient();
  return useMutation<void, Error, string>({
    mutationFn: async (id) => {
      const response = await client.delete<void, unknown>({
        url: `/api/v1/campaigns/${id}`,
      });
      const result = response as { error?: { message?: string } };
      if (result.error) throw new Error(result.error.message ?? 'Failed to delete campaign');
    },
    onSuccess: () => {
      toast.success('Кампания удалена');
      qc.invalidateQueries({ queryKey: CAMPAIGNS_QUERY_KEY });
    },
    onError: (err) => toast.error(err.message),
  });
};

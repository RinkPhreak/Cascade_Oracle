import { useQuery } from '@tanstack/react-query';
import type { Campaign } from '../../../api/extended-types';
import { client } from '../../../api/client';

export const CAMPAIGNS_QUERY_KEY = ['campaigns'] as const;

/** Fetch all campaigns, polling every 15s. */
export const useCampaigns = () =>
  useQuery<Campaign[]>({
    queryKey: CAMPAIGNS_QUERY_KEY,
    queryFn: async () => {
      const response = await client.get<Campaign[], unknown>({
        url: '/api/v1/campaigns',
      });
      const result = response as { data?: Campaign[]; error?: { message?: string } };
      if (result.error) throw new Error(result.error.message ?? 'Failed to fetch campaigns');
      return result.data ?? [];
    },
    refetchInterval: 15_000,
  });

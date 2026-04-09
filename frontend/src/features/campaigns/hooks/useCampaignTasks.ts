import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { CampaignStats, CampaignTask } from '../../../api/extended-types';
import { client } from '../../../api/client';
import { toast } from '../../../shared/hooks/useToast';

/** Fetch aggregated stats for a campaign, polling every 5s while running. */
export const useCampaignStats = (campaignId: string | null) =>
  useQuery<CampaignStats>({
    queryKey: ['campaigns', campaignId, 'stats'],
    queryFn: async () => {
      const response = await client.get<CampaignStats, unknown>({
        url: `/api/v1/campaigns/${campaignId}/stats`,
      });
      const result = response as { data?: CampaignStats; error?: { message?: string } };
      if (result.error) throw new Error(result.error.message ?? 'Failed to fetch stats');
      return result.data!;
    },
    enabled: !!campaignId,
    refetchInterval: 5_000,
  });

/** Fetch tasks that have been in_progress for >10 minutes (600s). */
export const useStuckTasks = (campaignId: string | null) =>
  useQuery<CampaignTask[]>({
    queryKey: ['campaigns', campaignId, 'stuck-tasks'],
    queryFn: async () => {
      const response = await client.get<CampaignTask[], unknown>({
        url: `/api/v1/campaigns/${campaignId}/tasks`,
        // Query params added via URL search string
      });
      // We filter client-side; ideally the backend supports ?status=in_progress&stuck_gt=600
      const result = response as { data?: CampaignTask[]; error?: { message?: string } };
      if (result.error) throw new Error(result.error.message ?? 'Failed to fetch tasks');
      const now = Date.now();
      return (result.data ?? []).filter((t) => {
        if (t.status !== 'in_progress' || !t.started_at) return false;
        const elapsedSeconds = (now - new Date(t.started_at).getTime()) / 1000;
        return elapsedSeconds > 600;
      });
    },
    enabled: !!campaignId,
    refetchInterval: 30_000,
  });

/** Requeue a stuck task — calls POST /api/v1/tasks/{id}/requeue. */
export const useRequeueTask = (campaignId: string) => {
  const qc = useQueryClient();
  return useMutation<void, Error, string>({
    mutationFn: async (taskId) => {
      const response = await client.post<void, unknown>({
        url: `/api/v1/tasks/${taskId}/requeue`,
        body: {},
        headers: { 'Content-Type': 'application/json' },
      });
      const result = response as { error?: { message?: string } };
      if (result.error) throw new Error(result.error.message ?? 'Failed to requeue task');
    },
    onSuccess: () => {
      toast.success('Задача поставлена в очередь повторно.');
      qc.invalidateQueries({ queryKey: ['campaigns', campaignId, 'stuck-tasks'] });
      qc.invalidateQueries({ queryKey: ['campaigns', campaignId, 'stats'] });
    },
    onError: (err) => toast.error(err.message),
  });
};

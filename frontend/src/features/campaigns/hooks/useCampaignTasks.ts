import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { DtoCampaignStats, DtoCampaignTask } from '../../../api/generated';
import { client } from '../../../api/client';
import { toast } from '../../../shared/hooks/useToast';

/** Fetch aggregated stats for a campaign, polling every 5s while running. */
export const useCampaignStats = (campaignId: string | null) =>
  useQuery<DtoCampaignStats>({
    queryKey: ['campaigns', campaignId, 'stats'],
    queryFn: async () => {
      const response = await client.get<DtoCampaignStats, unknown>({
        url: `/api/v1/campaigns/${campaignId}/stats`,
      });
      const result = response as { data?: DtoCampaignStats; error?: { message?: string } };
      if (result.error) {
        return {
          campaign_id: campaignId ?? '',
          total: 0,
          completed: 0,
          replied: 0,
          failed: 0,
          in_progress: 0,
          tg_attempted: 0,
          sms_attempted: 0,
          error_breakdown: {}
        } as DtoCampaignStats;
      }
      return result.data!;
    },
    enabled: !!campaignId,
    refetchInterval: 5_000,
  });

/** Fetch tasks that have been in_progress for >10 minutes (600s). */
export const useStuckTasks = (campaignId: string | null) =>
  useQuery<DtoCampaignTask[]>({
    queryKey: ['campaigns', campaignId, 'stuck-tasks'],
    queryFn: async () => {
      const response = await client.get<DtoCampaignTask[], unknown>({
        url: `/api/v1/campaigns/${campaignId}/tasks`,
        // Query params added via URL search string
      });
      // We filter client-side; ideally the backend supports ?status=in_progress&stuck_gt=600
      const result = response as { data?: DtoCampaignTask[]; error?: { message?: string } };
      if (result.error) {
        return [];
      }
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

import { useMutation, useQuery } from '@tanstack/react-query';
import { postApiV1SystemHalt, postApiV1SystemResume } from '../../../api/generated';
import type { SystemMetrics } from '../../../api/extended-types';
import { client } from '../../../api/client';
import { toast } from '../../../shared/hooks/useToast';

/** Fetch live system metrics: memory ratio, active account count, system status. */
export const useSystemMetrics = () =>
  useQuery<SystemMetrics>({
    queryKey: ['system', 'metrics'],
    queryFn: async () => {
      const response = await client.get<SystemMetrics, unknown>({
        url: '/api/v1/system/metrics',
      });
      // hey-api client returns { data, error, response }
      const result = response as { data?: SystemMetrics; error?: { message?: string } };
      if (result.error) throw new Error(result.error.message ?? 'Failed to fetch metrics');
      return result.data!;
    },
    refetchInterval: 10_000,
    retry: false,
    // Don't throw on error — banners are non-critical
    throwOnError: false,
  });

/** Emergency Break-Glass system halt mutation. */
export const useSystemHalt = (onSuccess?: () => void) =>
  useMutation({
    mutationFn: async (payload: { password: string; reason: string }) => {
      const { data, error } = await postApiV1SystemHalt({
        body: { password: payload.password, reason: payload.reason },
      });
      if (error) {
        const errMsg = (error as { message?: string }).message ?? 'System halt failed';
        throw new Error(errMsg);
      }
      return data;
    },
    onSuccess: () => {
      toast.success('Система остановлена. Break-Glass событие залогировано.');
      onSuccess?.();
    },
    onError: (err: Error) => {
      toast.error(err.message);
    },
  });

/** Emergency Break-Glass system resume mutation. */
export const useSystemResume = (onSuccess?: () => void) =>
  useMutation({
    mutationFn: async (payload: { password: string; reason: string }) => {
      const { data, error } = await postApiV1SystemResume({
        body: { password: payload.password, reason: payload.reason },
      });
      if (error) {
        const errMsg = (error as { message?: string }).message ?? 'System resume failed';
        throw new Error(errMsg);
      }
      return data;
    },
    onSuccess: () => {
      toast.success('Система возобновлена.');
      onSuccess?.();
    },
    onError: (err: Error) => {
      toast.error(err.message);
    },
  });

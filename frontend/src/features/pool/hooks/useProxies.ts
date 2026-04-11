import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { DtoProxy, DtoCreateProxyRequest, DtoReassignProxyRequest } from '../../../api/generated';
import { client } from '../../../api/client';
import { toast } from '../../../shared/hooks/useToast';

const PROXIES_QUERY_KEY = ['pool', 'proxies'] as const;

/** Fetch all proxies in the pool. */
export const useProxies = () =>
  useQuery<DtoProxy[]>({
    queryKey: PROXIES_QUERY_KEY,
    queryFn: async () => {
      const response = await client.get<DtoProxy[], unknown>({
        url: '/api/v1/proxies',
      });
      const result = response as { data?: DtoProxy[]; error?: { message?: string } };
      if (result.error) {
        return []; // Graceful fallback
      }
      return result.data ?? [];
    },
    refetchInterval: 30_000,
  });

/** Add a new proxy to the pool. */
export const useAddProxy = () => {
  const qc = useQueryClient();
  return useMutation<DtoProxy, Error, DtoCreateProxyRequest>({
    mutationFn: async (body) => {
      const response = await client.post<DtoProxy, unknown>({
        url: '/api/v1/proxies',
        body,
        headers: { 'Content-Type': 'application/json' },
      });
      const result = response as { data?: DtoProxy; error?: { message?: string } };
      if (result.error) throw new Error(result.error.message ?? 'Failed to add proxy');
      return result.data!;
    },
    onSuccess: () => {
      toast.success('Прокси успешно добавлен.');
      qc.invalidateQueries({ queryKey: PROXIES_QUERY_KEY });
    },
    onError: (err) => toast.error(err.message),
  });
};

/** Delete a proxy from the pool. */
export const useDeleteProxy = () => {
  const qc = useQueryClient();
  return useMutation<void, Error, string>({
    mutationFn: async (proxyId) => {
      const response = await client.delete<void, unknown>({
        url: `/api/v1/proxies/${proxyId}`,
      });
      const result = response as { error?: { message?: string } };
      if (result.error) throw new Error(result.error.message ?? 'Failed to delete proxy');
    },
    onSuccess: () => {
      toast.success('Прокси удалён.');
      qc.invalidateQueries({ queryKey: PROXIES_QUERY_KEY });
    },
    onError: (err) => toast.error(err.message),
  });
};

/** Re-assign a proxy to a different account (sticky binding change). */
export const useReassignProxy = () => {
  const qc = useQueryClient();
  return useMutation<void, Error, { proxyId: string } & DtoReassignProxyRequest>({
    mutationFn: async ({ proxyId, account_id, reason }) => {
      const response = await client.post<void, unknown>({
        url: `/api/v1/proxies/${proxyId}/reassign`,
        body: { account_id, reason },
        headers: { 'Content-Type': 'application/json' },
      });
      const result = response as { error?: { message?: string } };
      if (result.error) throw new Error(result.error.message ?? 'Failed to reassign proxy');
    },
    onSuccess: () => {
      toast.success('Привязка прокси изменена. Решение оператора записано.');
      qc.invalidateQueries({ queryKey: PROXIES_QUERY_KEY });
      qc.invalidateQueries({ queryKey: ['pool', 'accounts'] });
    },
    onError: (err) => toast.error(err.message),
  });
};

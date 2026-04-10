import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { TgAccount, AccountEvent } from '../../../api/extended-types';
import { client } from '../../../api/client';
import { toast } from '../../../shared/hooks/useToast';

const ACCOUNTS_QUERY_KEY = ['pool', 'accounts'] as const;

/** Fetch all TG accounts in the pool. */
export const useAccounts = () =>
  useQuery<TgAccount[]>({
    queryKey: ACCOUNTS_QUERY_KEY,
    queryFn: async () => {
      const response = await client.get<TgAccount[], unknown>({
        url: '/api/v1/accounts',
      });
      const result = response as { data?: TgAccount[]; error?: { message?: string } };
      if (result.error) {
        return []; // Graceful fallback
      }
      return result.data ?? [];
    },
    refetchInterval: 15_000,
  });

/** Fetch account event log for a specific account. */
export const useAccountEvents = (accountId: string | null) =>
  useQuery<AccountEvent[]>({
    queryKey: ['pool', 'accounts', accountId, 'events'],
    queryFn: async () => {
      const response = await client.get<AccountEvent[], unknown>({
        url: `/api/v1/accounts/${accountId}/events`,
      });
      const result = response as { data?: AccountEvent[]; error?: { message?: string } };
      if (result.error) {
        return [];
      }
      return result.data ?? [];
    },
    enabled: !!accountId,
  });

/** Import a TG account using session files and proxy. */
export const useImportAccount = () => {
  const qc = useQueryClient();
  return useMutation<TgAccount, Error, FormData>({
    mutationFn: async (formData) => {
      // client.post doesn't handle FormData neatly with current config, so we use native fetch via client.base?
      // Actually, I'll see if I can use client.post with custom headers.
      // But FormData usually shouldn't have 'Content-Type': 'application/json'
      
      const response = await fetch('/api/v1/accounts/import', {
        method: 'POST',
        body: formData,
        // No Content-Type header let browser set it with boundary
      });

      if (!response.ok) {
        const errData = await response.json().catch(() => ({ message: 'Import failed' }));
        throw new Error(errData.message || response.statusText);
      }
      
      return response.json();
    },
    onSuccess: () => {
      toast.success('Аккаунт успешно импортирован!');
      qc.invalidateQueries({ queryKey: ACCOUNTS_QUERY_KEY });
    },
    onError: (err) => toast.error(err.message),
  });
};

/** Delete a TG account from the pool. */
export const useDeleteAccount = () => {
  const qc = useQueryClient();
  return useMutation<void, Error, string>({
    mutationFn: async (id) => {
      const response = await client.delete<void, unknown>({
        url: `/api/v1/accounts/${id}`,
      });
      const result = response as { error?: { message?: string } };
      if (result.error) throw new Error(result.error.message ?? 'Failed to delete account');
    },
    onSuccess: () => {
      toast.success('Аккаунт удален');
      qc.invalidateQueries({ queryKey: ACCOUNTS_QUERY_KEY });
    },
    onError: (err) => toast.error(err.message),
  });
};

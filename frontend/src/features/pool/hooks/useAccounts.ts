import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { TgAccount, AccountEvent, TgAccountSession } from '../../../api/extended-types';
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

/** Initiate a new TG account registration session — returns QR/pairing info. */
export const useInitAccountSession = () => {
  return useMutation<TgAccountSession, Error>({
    mutationFn: async () => {
      const response = await client.post<TgAccountSession, unknown>({
        url: '/api/v1/accounts/register',
        body: {},
        headers: { 'Content-Type': 'application/json' },
      });
      const result = response as { data?: TgAccountSession; error?: { message?: string } };
      if (result.error) throw new Error(result.error.message ?? 'Failed to init session');
      return result.data!;
    },
    onError: (err) => toast.error(err.message),
  });
};

/** Poll a registration session status (QR scan / code auth). */
export const useSessionStatus = (sessionId: string | null) =>
  useQuery<TgAccountSession>({
    queryKey: ['pool', 'session', sessionId],
    queryFn: async () => {
      const response = await client.get<TgAccountSession, unknown>({
        url: `/api/v1/accounts/register/${sessionId}`,
      });
      const result = response as { data?: TgAccountSession; error?: { message?: string } };
      if (result.error) throw new Error(result.error.message ?? 'Failed to poll session');
      return result.data!;
    },
    enabled: !!sessionId,
    refetchInterval: (query) => {
      const status = query.state.data?.status;
      // Stop polling once finalized
      if (status === 'AUTHENTICATED' || status === 'FAILED') return false;
      return 3_000;
    },
  });

/** Submit the pairing code to complete authentication. */
export const useSubmitPairingCode = () => {
  const qc = useQueryClient();
  return useMutation<void, Error, { sessionId: string; code: string }>({
    mutationFn: async ({ sessionId, code }) => {
      const response = await client.post<void, unknown>({
        url: `/api/v1/accounts/register/${sessionId}/confirm`,
        body: { code },
        headers: { 'Content-Type': 'application/json' },
      });
      const result = response as { error?: { message?: string } };
      if (result.error) throw new Error(result.error.message ?? 'Invalid pairing code');
    },
    onSuccess: () => {
      toast.success('Аккаунт успешно добавлен в пул!');
      qc.invalidateQueries({ queryKey: ACCOUNTS_QUERY_KEY });
    },
    onError: (err) => toast.error(err.message),
  });
};

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { getApiV1Accounts, postApiV1AccountsImport, deleteApiV1AccountsById, getApiV1AccountsByIdEvents } from '../../../api/generated';
import type { DtoTgAccount, DtoAccountEvent } from '../../../api/generated';
import { toast } from '../../../shared/hooks/useToast';

const ACCOUNTS_QUERY_KEY = ['pool', 'accounts'] as const;

/** Fetch all TG accounts in the pool. */
export const useAccounts = () =>
  useQuery<DtoTgAccount[]>({
    queryKey: ACCOUNTS_QUERY_KEY,
    queryFn: async () => {
      const { data, error } = await getApiV1Accounts();
      if (error) {
        return []; // Graceful fallback
      }
      return (data as DtoTgAccount[]) ?? [];
    },
    refetchInterval: 15_000,
  });

/** Fetch account event log for a specific account. */
export const useAccountEvents = (accountId: string | null) =>
  useQuery<DtoAccountEvent[]>({
    queryKey: ['pool', 'accounts', accountId, 'events'],
    queryFn: async () => {
      if (!accountId) return [];
      const { data, error } = await getApiV1AccountsByIdEvents({
        path: { id: accountId }
      });
      if (error) {
        return [];
      }
      return (data as DtoAccountEvent[]) ?? [];
    },
    enabled: !!accountId,
  });

/** Import a TG account using session files and proxy. */
export const useImportAccount = () => {
  const qc = useQueryClient();
  return useMutation<DtoTgAccount, Error, FormData>({
    mutationFn: async (formData) => {
      const { data, error } = await postApiV1AccountsImport({
        body: formData as any,
      });

      if (error) {
        const errMsg = (error as { message?: string }).message || 'Import failed';
        throw new Error(errMsg);
      }
      
      return data as DtoTgAccount;
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
      const { error } = await deleteApiV1AccountsById({
        path: { id }
      });
      if (error) throw new Error((error as any).message ?? 'Failed to delete account');
    },
    onSuccess: () => {
      toast.success('Аккаунт удален');
      qc.invalidateQueries({ queryKey: ACCOUNTS_QUERY_KEY });
    },
    onError: (err) => toast.error(err.message),
  });
};

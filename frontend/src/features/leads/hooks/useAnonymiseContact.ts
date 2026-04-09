import { useMutation, useQueryClient } from '@tanstack/react-query';
import { postApiV1ContactsByIdAnonymise } from '../../../api/generated';
import type { DtoAnonymiseResponse } from '../../../api/generated';
import { LEADS_QUERY_KEY } from './useLeads';
import { toast } from '../../../shared/hooks/useToast';

export const useAnonymiseContact = (onSuccess?: () => void) => {
  const qc = useQueryClient();
  return useMutation<DtoAnonymiseResponse, Error, string>({
    mutationFn: async (contactId) => {
      const { data, error } = await postApiV1ContactsByIdAnonymise({
        path: { id: contactId },
      });
      if (error) {
        const errMsg = (error as { message?: string }).message ?? 'Anonymisation failed';
        throw new Error(errMsg);
      }
      return data!;
    },
    onSuccess: (result) => {
      toast.success(result.message ?? 'Контакт успешно анонимизирован (152-ФЗ).');
      qc.invalidateQueries({ queryKey: LEADS_QUERY_KEY });
      onSuccess?.();
    },
    onError: (err) => toast.error(err.message),
  });
};

import { useMutation, useQueryClient } from '@tanstack/react-query';
import { postApiV1CampaignsByIdImport } from '../../../api/generated';
import { CAMPAIGNS_QUERY_KEY } from './useCampaigns';
import { toast } from '../../../shared/hooks/useToast';

interface ImportResult {
  imported: number;
  skipped?: number;
  errors?: number;
}

export const useImportContacts = (campaignId: string) => {
  const qc = useQueryClient();
  return useMutation<ImportResult, Error, File>({
    mutationFn: async (file) => {
      const { data, error } = await postApiV1CampaignsByIdImport({
        path: { id: campaignId },
        body: { file },
      });
      if (error) {
        const errMsg = (error as { message?: string }).message ?? 'CSV import failed';
        throw new Error(errMsg);
      }
      // The SDK returns Record<string, number> — map to our ImportResult shape
      const raw = data as Record<string, number> | undefined;
      return {
        imported: raw?.imported ?? raw?.count ?? 0,
        skipped: raw?.skipped,
        errors: raw?.errors,
      };
    },
    onSuccess: (result) => {
      toast.success(`Импорт завершён: ${result.imported} контактов загружено.`);
      qc.invalidateQueries({ queryKey: CAMPAIGNS_QUERY_KEY });
    },
    onError: (err) => toast.error(err.message),
  });
};

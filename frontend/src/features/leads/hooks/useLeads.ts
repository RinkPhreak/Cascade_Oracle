import { useQuery } from '@tanstack/react-query';
import type { Contact } from '../../../api/extended-types';
import { client } from '../../../api/client';

export const LEADS_QUERY_KEY = ['leads'] as const;

/** Fetch contacts where has_replied = true. Polls every 20s. */
export const useLeads = () =>
  useQuery<Contact[]>({
    queryKey: LEADS_QUERY_KEY,
    queryFn: async () => {
      const response = await client.get<Contact[], unknown>({
        url: '/api/v1/contacts',
        // The backend should support ?has_replied=true filter
      });
      const result = response as { data?: Contact[]; error?: { message?: string } };
      if (result.error) {
        return [];
      }
      // Filter client-side in case backend doesn't support the query param yet
      return (result.data ?? []).filter((c) => c.has_replied && !c.is_anonymised);
    },
    refetchInterval: 20_000,
  });

/** Fetch the waterfall trace for a single contact. */
export const useContactTrace = (contactId: string | null) =>
  useQuery({
    queryKey: ['contacts', contactId, 'trace'],
    queryFn: async () => {
      const response = await client.get({
        url: `/api/v1/contacts/${contactId}/trace`,
      });
      const result = response as { data?: unknown; error?: { message?: string } };
      if (result.error) {
        return [];
      }
      return result.data;
    },
    enabled: !!contactId,
  });

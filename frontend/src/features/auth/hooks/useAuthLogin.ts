import { useMutation } from '@tanstack/react-query';
import { postApiV1AuthLogin } from '../../../api/generated';
import type { PostApiV1AuthLoginData, PostApiV1AuthLoginResponse } from '../../../api/generated';

export interface ApiErrorResponse {
  code: string;
  message: string;
  details?: Record<string, unknown>;
}

export const useAuthLogin = () => {
  return useMutation({
    mutationFn: async (data: Omit<PostApiV1AuthLoginData, 'url'>) => {
      const { data: responseData, error } = await postApiV1AuthLogin({
        ...data,
        url: '/api/v1/auth/login'
      } as PostApiV1AuthLoginData);
      if (error) {
        throw error as ApiErrorResponse;
      }
      return responseData as PostApiV1AuthLoginResponse;
    }
  });
};

import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { authApi } from "@/lib/api/auth";

export function useProfile() {
  return useQuery({
    queryKey: ["profile"],
    queryFn: authApi.getProfile,
  });
}

export function useUpdateProfile() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: authApi.updateProfile,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["profile"] });
    },
  });
}

export function useChangePassword() {
  return useMutation({
    mutationFn: authApi.changePassword,
  });
}

export function useDeleteAccount() {
  return useMutation({
    mutationFn: authApi.deleteAccount,
  });
}

export function useSessions() {
  return useQuery({
    queryKey: ["sessions"],
    queryFn: authApi.getSessions,
  });
}

export function useRevokeSession() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: authApi.revokeSession,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["sessions"] });
    },
  });
}

export function useLoginHistory() {
  return useQuery({
    queryKey: ["loginHistory"],
    queryFn: authApi.getLoginHistory,
  });
}

export function useSetup2FA() {
  return useMutation({
    mutationFn: authApi.setup2FA,
  });
}

export function useVerify2FA() {
  return useMutation({
    mutationFn: authApi.verify2FASetup,
  });
}

export function useDisable2FA() {
  return useMutation({
    mutationFn: authApi.disable2FA,
  });
}

export function useAuth() {
  return {
    user: null,
    isLoading: false,
    isAuthenticated: false,
    login: async () => {},
    logout: async () => {},
  };
}

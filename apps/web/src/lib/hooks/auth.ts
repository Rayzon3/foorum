import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import * as api from "../api";

const meKey = (token?: string | null) => ["me", token ?? "anon"] as const;

export function useMe(token?: string | null) {
  return useQuery({
    queryKey: meKey(token),
    queryFn: async () => {
      if (!token) {
        return null;
      }
      return api.fetchMe(token);
    },
    enabled: !!token,
  });
}

export function useLogin() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: { identifier: string; password: string }) =>
      api.login(input.identifier, input.password),
    onSuccess: () => {
      toast.success("Welcome back");
    },
    onError: (err) => {
      toast.error((err as Error).message || "login_failed");
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["me"] });
    },
  });
}

export function useRegister() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: { email: string; username: string; password: string }) =>
      api.register(input.email, input.username, input.password),
    onSuccess: () => {
      toast.success("Account created");
    },
    onError: (err) => {
      toast.error((err as Error).message || "register_failed");
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ["me"] });
    },
  });
}

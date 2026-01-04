import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

import * as api from "../api";

const feedKey = (token?: string) => ["feed", token ?? "anon"] as const;

export function useFeed(token?: string) {
  return useQuery({
    queryKey: feedKey(token),
    queryFn: () => api.fetchFeed(token),
  });
}

export function useCreatePost(token?: string | null) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (input: { title: string; body: string }) => {
      if (!token) {
        throw new Error("login_required");
      }
      return api.createPost(token, input.title, input.body);
    },
    onSuccess: (created) => {
      queryClient.setQueryData<api.Post[]>(feedKey(token ?? undefined), (prev) =>
        prev ? [created, ...prev] : [created]
      );
      toast.success("Post created");
    },
    onError: (err) => {
      toast.error((err as Error).message || "post_failed");
    },
  });
}

export function useVotePost(token?: string | null) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (input: { postID: string; value: number }) => {
      if (!token) {
        throw new Error("login_required");
      }
      return api.votePost(token, input.postID, input.value);
    },
    onMutate: async (input) => {
      const key = feedKey(token ?? undefined);
      await queryClient.cancelQueries({ queryKey: key });
      const previous = queryClient.getQueryData<api.Post[]>(key);
      if (previous) {
        const next = previous.map((post) => {
          if (post.id !== input.postID) {
            return post;
          }
          const delta = input.value - post.myVote;
          return { ...post, myVote: input.value, score: post.score + delta };
        });
        queryClient.setQueryData(key, next);
      }
      return { previous };
    },
    onError: (_err, _input, ctx) => {
      if (ctx?.previous) {
        queryClient.setQueryData(feedKey(token ?? undefined), ctx.previous);
      }
      toast.error("Vote failed");
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: feedKey(token ?? undefined) });
    },
  });
}

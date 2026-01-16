import React from "react";
import { Link } from "@tanstack/react-router";
import { ArrowDown, ArrowUp } from "lucide-react";

import { useAuth } from "../lib/auth";
import { useCreatePost, useFeed, useVotePost } from "../lib/hooks/posts";
import { Button } from "../components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "../components/ui/card";
import { Input } from "../components/ui/input";
import { Textarea } from "../components/ui/textarea";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../components/ui/dialog";

export function HomePage() {
  const auth = useAuth();
  const [title, setTitle] = React.useState("");
  const [body, setBody] = React.useState("");
  const [isPostDialogOpen, setPostDialogOpen] = React.useState(false);
  const feedQuery = useFeed(auth.token ?? undefined);
  const createPost = useCreatePost(auth.token);
  const votePost = useVotePost(auth.token);
  const postError = (feedQuery.error as Error | null)?.message ?? null;

  async function handleCreatePost(event: React.FormEvent) {
    event.preventDefault();
    try {
      await createPost.mutateAsync({ title, body });
      setTitle("");
      setBody("");
      setPostDialogOpen(false);
    } catch {
      // errors are surfaced via createPost.error
    }
  }

  async function handleVote(
    postID: string,
    currentVote: number,
    value: number
  ) {
    const nextValue = currentVote === value ? 0 : value;
    try {
      await votePost.mutateAsync({ postID, value: nextValue });
    } catch {
      // errors are surfaced via votePost.error
    }
  }

  return (
    <section className="grid gap-8 md:grid-cols-[2fr_1fr]">
      <div className="space-y-6">
        <Card className="rounded-3xl border-border/70 bg-card/90">
          <CardHeader>
            <CardTitle className="text-3xl">Start new conversations</CardTitle>
            <CardDescription className="text-base">
              Jabber is social app currently in development. Build rooms create
              posts, follow people, and keep the conversation flowing.
            </CardDescription>
          </CardHeader>
          <CardContent className="pt-0">
            {!auth.user && (
              <div className="mt-6 flex flex-wrap gap-3">
                <Button asChild>
                  <Link to="/register">Create account</Link>
                </Button>
                <Button variant="outline" asChild>
                  <Link to="/login">Log in</Link>
                </Button>
                <Dialog>
                  <DialogTrigger asChild>
                    <Button variant="ghost">How it works</Button>
                  </DialogTrigger>
                  <DialogContent>
                    <DialogHeader>
                      <DialogTitle>About this scaffold</DialogTitle>
                      <DialogDescription>
                        Gruvbox-dark UI made with React on the frontend and Go
                        API ready for posts, feeds, and realtime rooms(coming
                        soon).
                      </DialogDescription>
                    </DialogHeader>
                    <DialogFooter>
                      <DialogClose asChild>
                        <Button variant="outline">Got it</Button>
                      </DialogClose>
                    </DialogFooter>
                  </DialogContent>
                </Dialog>
              </div>
            )}
          </CardContent>
        </Card>
        <Card className="rounded-3xl border-border/70 bg-card/90">
          <CardHeader>
            <CardTitle className="text-xl">Home feed</CardTitle>
            <CardDescription>Latest posts from the community.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {!auth.user && (
              <p className="text-sm text-muted-foreground">
                Log in to view the home feed.
              </p>
            )}
            {auth.user && feedQuery.isLoading && (
              <p className="text-sm text-muted-foreground">Loading posts...</p>
            )}
            {auth.user &&
              !feedQuery.isLoading &&
              (feedQuery.data?.length ?? 0) === 0 && (
                <p className="text-sm text-muted-foreground">No posts yet.</p>
              )}
            {auth.user &&
              (feedQuery.data ?? []).map((post) => (
                <Card
                  key={post.id}
                  className="rounded-2xl border-border/70 bg-background/70"
                >
                  <CardHeader className="pb-2">
                    <CardTitle className="text-lg">{post.title}</CardTitle>
                    <CardDescription>
                      {post.author.username || post.author.email} ·{" "}
                      {new Date(post.createdAt).toLocaleString()}
                    </CardDescription>
                  </CardHeader>
                  <CardContent className="pt-0">
                    <p className="text-sm text-foreground/90 whitespace-pre-line">
                      {post.body}
                    </p>
                  </CardContent>
                  <CardFooter className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Button
                        type="button"
                        variant={post.myVote === 1 ? "success" : "outline"}
                        onClick={() => handleVote(post.id, post.myVote, 1)}
                      >
                        <ArrowUp className="h-4 w-4" />
                        Upvote
                      </Button>
                      <Button
                        type="button"
                        variant={post.myVote === -1 ? "destructive" : "outline"}
                        onClick={() => handleVote(post.id, post.myVote, -1)}
                      >
                        <ArrowDown className="h-4 w-4" />
                        Downvote
                      </Button>
                    </div>
                    <span className="text-sm text-muted-foreground">
                      Score: {post.score}
                    </span>
                  </CardFooter>
                </Card>
              ))}
            {auth.user && postError && (
              <p className="text-xs text-destructive">
                Post error: {postError}
              </p>
            )}
          </CardContent>
        </Card>
      </div>
      <aside className="space-y-6">
        <Card className="rounded-3xl border-border/70 bg-card/90">
          <CardHeader className="pb-3">
            <CardTitle className="text-lg">Session</CardTitle>
          </CardHeader>
          <CardContent className="text-sm text-muted-foreground">
            {auth.loading && "Checking session..."}
            {!auth.loading && auth.user && (
              <div className="space-y-2">
                <p className="text-foreground/80">Signed in as</p>
                <p className="font-semibold text-foreground">
                  {auth.user.email}
                </p>
                {/* <p className="text-xs text-muted-foreground">
                  User ID: {auth.user.id}
                </p> */}
              </div>
            )}
            {!auth.loading && !auth.user && (
              <p>No active session. Try signing up.</p>
            )}
            {auth.error && (
              <p className="mt-3 text-xs text-destructive">
                Auth error: {auth.error}
              </p>
            )}
          </CardContent>
        </Card>
        {auth.user && (
          <Card className="rounded-3xl border-border/70 bg-card/90">
            <CardHeader>
              <CardTitle className="text-lg">Create a post</CardTitle>
              <CardDescription>
                Share an update with the community.
              </CardDescription>
            </CardHeader>
            <CardContent>
              <Dialog open={isPostDialogOpen} onOpenChange={setPostDialogOpen}>
                <DialogTrigger asChild>
                  <Button className="w-full">New post</Button>
                </DialogTrigger>
                <DialogContent>
                  <DialogHeader>
                    <DialogTitle>Create a post</DialogTitle>
                    <DialogDescription>
                      Share an update with the community.
                    </DialogDescription>
                  </DialogHeader>
                  <form className="space-y-4" onSubmit={handleCreatePost}>
                    <label className="block text-sm text-muted-foreground">
                      Title
                      <Input
                        value={title}
                        onChange={(event) => setTitle(event.target.value)}
                        className="mt-2"
                        required
                      />
                    </label>
                    <label className="block text-sm text-muted-foreground">
                      Body
                      <Textarea
                        value={body}
                        onChange={(event) => setBody(event.target.value)}
                        className="mt-2"
                        required
                      />
                    </label>
                    <DialogFooter>
                      <DialogClose asChild>
                        <Button type="button" variant="outline">
                          Cancel
                        </Button>
                      </DialogClose>
                      <Button type="submit" disabled={createPost.isPending}>
                        {createPost.isPending ? "Posting..." : "Post"}
                      </Button>
                    </DialogFooter>
                  </form>
                </DialogContent>
              </Dialog>
            </CardContent>
          </Card>
        )}
      </aside>
    </section>
  );
}

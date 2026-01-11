import React from "react";

import { useAuth } from "../lib/auth";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { Input } from "../components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../components/ui/select";
import { Mic, MicOff } from "lucide-react";

type Role = "listener" | "speaker";
type Participant = {
  userId: string;
  role: Role;
};

export function SpacesPage() {
  const auth = useAuth();
  const [roomId, setRoomId] = React.useState("lobby");
  const [role, setRole] = React.useState<Role>("listener");
  const [status, setStatus] = React.useState("disconnected");
  const [logs, setLogs] = React.useState<string[]>([]);
  const [isMicOn, setMicOn] = React.useState(false);
  const [micError, setMicError] = React.useState<string | null>(null);
  const [remoteAudioActive, setRemoteAudioActive] = React.useState(false);
  const [participants, setParticipants] = React.useState<Participant[]>([]);
  const [micLevel, setMicLevel] = React.useState(0);
  const isSpeakingSelf = role === "speaker" && isMicOn && micLevel > 0.12;

  const wsRef = React.useRef<WebSocket | null>(null);
  const pcRef = React.useRef<RTCPeerConnection | null>(null);
  const localStreamRef = React.useRef<MediaStream | null>(null);
  const remoteAudioRef = React.useRef<HTMLAudioElement | null>(null);
  const audioContextRef = React.useRef<AudioContext | null>(null);
  const localAnalyserRef = React.useRef<AnalyserNode | null>(null);
  const rafRef = React.useRef<number | null>(null);

  function addLog(message: string) {
    setLogs((prev) => [...prev, `${new Date().toLocaleTimeString()} ${message}`].slice(-50));
  }

  function buildWsUrl() {
    const fallbackBase = `${window.location.protocol}//${window.location.hostname}:8080`;
    const base = import.meta.env.VITE_API_URL ?? fallbackBase;
    const wsBase = base.replace(/^http/, "ws").replace(/\/+$/, "");
    const token = encodeURIComponent(auth.token ?? "");
    return `${wsBase}/api/v1/rooms/${encodeURIComponent(roomId)}/ws?token=${token}`;
  }

  async function connect() {
    if (!auth.token) {
      addLog("login_required");
      return;
    }
    if (wsRef.current || pcRef.current) {
      addLog("already_connected");
      return;
    }

    setStatus("connecting");
    const wsUrl = buildWsUrl();
    addLog(`connecting ${wsUrl.replace(/token=.*$/, "token=***")}`);
    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = async () => {
      addLog("ws_open");
      ws.send(JSON.stringify({ type: "join", payload: { role } }));

      const pc = new RTCPeerConnection({
        iceServers: [{ urls: "stun:stun.l.google.com:19302" }],
      });
      pcRef.current = pc;

      pc.onicecandidate = (event) => {
        if (!event.candidate) {
          return;
        }
        ws.send(
          JSON.stringify({
            type: "candidate",
            candidate: {
              candidate: event.candidate.candidate,
              sdpMid: event.candidate.sdpMid,
              sdpMLineIndex: event.candidate.sdpMLineIndex,
            },
          })
        );
      };

      pc.ontrack = (event) => {
        if (remoteAudioRef.current) {
          remoteAudioRef.current.srcObject = event.streams[0];
          remoteAudioRef.current.muted = false;
          remoteAudioRef.current
            .play()
            .catch((err) => addLog(`audio_play_error ${(err as Error).message}`));
        }
        setRemoteAudioActive(true);

      };

      if (role === "speaker") {
        try {
          const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
          localStreamRef.current = stream;
          stream.getTracks().forEach((track) => pc.addTrack(track, stream));
          setMicOn(stream.getTracks().some((track) => track.enabled));
          setMicError(null);
          startMicVisualizer(stream);
        } catch (err) {
          const message = (err as Error).message;
          setMicError(message);
          addLog(`mic_error ${message}`);
        }
        pc.addTransceiver("audio", { direction: "sendrecv" });
      } else {
        pc.addTransceiver("audio", { direction: "recvonly" });
      }

      const offer = await pc.createOffer();
      await pc.setLocalDescription(offer);
      if (pc.localDescription?.sdp) {
        ws.send(JSON.stringify({ type: "offer", sdp: pc.localDescription.sdp }));
      }
      setStatus("connected");
    };

    ws.onmessage = async (event) => {
      const msg = JSON.parse(event.data);
      if (msg.type === "answer" && pcRef.current) {
        await pcRef.current.setRemoteDescription({ type: "answer", sdp: msg.sdp });
        addLog("answer_received");
      }
      if (msg.type === "candidate" && pcRef.current) {
        const candidate = msg.candidate;
        if (candidate?.candidate) {
          await pcRef.current.addIceCandidate({
            candidate: candidate.candidate,
            sdpMid: candidate.sdpMid ?? undefined,
            sdpMLineIndex: candidate.sdpMLineIndex ?? undefined,
          });
        }
      }
      if (msg.type === "participants") {
        const list = msg.payload?.participants ?? [];
        setParticipants(list);
      }
      if (msg.type === "participant_joined") {
        const participant = msg.payload?.participant;
        if (participant?.userId) {
          setParticipants((prev) => {
            const filtered = prev.filter((p) => p.userId !== participant.userId);
            return [...filtered, participant];
          });
        }
      }
      if (msg.type === "participant_left") {
        const userId = msg.payload?.userId;
        if (userId) {
          setParticipants((prev) => prev.filter((p) => p.userId !== userId));
        }
      }
      if (msg.type === "error") {
        addLog(`error ${msg.error}`);
      }
    };

    ws.onclose = (event) => {
      addLog(`ws_closed code=${event.code} reason=${event.reason || "none"}`);
      cleanup();
    };

    ws.onerror = () => {
      addLog("ws_error");
    };
  }

  function cleanup() {
    setStatus("disconnected");
    setMicOn(false);
    setMicError(null);
    setRemoteAudioActive(false);
    setParticipants([]);
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }
    if (pcRef.current) {
      pcRef.current.close();
      pcRef.current = null;
    }
    if (localStreamRef.current) {
      localStreamRef.current.getTracks().forEach((track) => track.stop());
      localStreamRef.current = null;
    }
    stopMicVisualizer();
  }

  async function startMicVisualizer(stream: MediaStream) {
    if (rafRef.current) {
      return;
    }
    if (!audioContextRef.current) {
      audioContextRef.current = new AudioContext();
    }
    const audioCtx = audioContextRef.current;
    if (audioCtx.state === "suspended") {
      await audioCtx.resume();
    }
    const source = audioCtx.createMediaStreamSource(stream);
    const analyser = audioCtx.createAnalyser();
    analyser.fftSize = 512;
    source.connect(analyser);
    localAnalyserRef.current = analyser;

    const buffer = new Uint8Array(analyser.frequencyBinCount);
    const tick = () => {
      analyser.getByteFrequencyData(buffer);
      const level = avgLevel(buffer);
      setMicLevel(level);
      rafRef.current = requestAnimationFrame(tick);
    };
    rafRef.current = requestAnimationFrame(tick);
  }

  function stopMicVisualizer() {
    if (rafRef.current) {
      cancelAnimationFrame(rafRef.current);
      rafRef.current = null;
    }
    setMicLevel(0);
    localAnalyserRef.current = null;
    if (audioContextRef.current) {
      audioContextRef.current.close().catch(() => {});
      audioContextRef.current = null;
    }
  }


  function toggleMic() {
    if (!localStreamRef.current) {
      setMicError("mic_unavailable");
      return;
    }
    localStreamRef.current.getTracks().forEach((track) => {
      track.enabled = !track.enabled;
      setMicOn(track.enabled);
      if (!track.enabled) {
        setMicLevel(0);
      }
    });
  }

  return (
    <div className="grid gap-6">
      <Card className="rounded-3xl border-border/70 bg-card/90">
        <CardHeader>
          <CardTitle className="text-2xl">Spaces (MVP)</CardTitle>
          <CardDescription>
            Connect to a room and publish or listen to audio.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="grid gap-4 md:grid-cols-[2fr_1fr]">
            <label className="block text-sm text-muted-foreground">
              Room ID
              <Input
                value={roomId}
                onChange={(event) => setRoomId(event.target.value)}
                className="mt-2"
              />
            </label>
            <label className="block text-sm text-muted-foreground">
              Role
              <Select value={role} onValueChange={(value) => setRole(value as Role)}>
                <SelectTrigger className="mt-2">
                  <SelectValue placeholder="Select role" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="listener">Listener</SelectItem>
                  <SelectItem value="speaker">Speaker</SelectItem>
                </SelectContent>
              </Select>
            </label>
          </div>
          <div className="flex items-center gap-3">
            <Button onClick={connect} disabled={status !== "disconnected"}>
              {status === "connecting" ? "Connecting..." : "Connect"}
            </Button>
            <Button variant="outline" onClick={cleanup} disabled={status === "disconnected"}>
              Disconnect
            </Button>
            {role === "speaker" && (
              <div className="relative">
                <span
                  className="absolute inset-0 rounded-full bg-success/40"
                  style={{
                    transform: `scale(${1 + Math.min(1.5, micLevel * 2.5)})`,
                    opacity: Math.min(0.9, micLevel * 2.2),
                    transition: "transform 120ms ease, opacity 120ms ease",
                  }}
                />
                <Button
                  variant={isMicOn ? "success" : "destructive"}
                  onClick={toggleMic}
                  disabled={status !== "connected"}
                  className="relative z-10"
                >
                  {isMicOn ? <Mic className="h-4 w-4" /> : <MicOff className="h-4 w-4" />}
                  {isMicOn ? "Mic on" : "Mic off"}
                </Button>
              </div>
            )}
            <span className="text-sm text-muted-foreground">Status: {status}</span>
          </div>
          <div className="flex flex-wrap items-center gap-2 text-xs">
            <span className="rounded-full bg-accent px-3 py-1 text-accent-foreground">
              Role: {role}
            </span>
            <span className={`rounded-full px-3 py-1 ${remoteAudioActive ? "bg-success text-success-foreground" : "bg-muted text-muted-foreground"}`}>
              Remote audio: {remoteAudioActive ? "active" : "idle"}
            </span>
            <span
              className={`rounded-full px-3 py-1 ${
                role !== "speaker"
                  ? "bg-muted text-muted-foreground"
                  : micError
                    ? "bg-destructive text-destructive-foreground"
                    : isMicOn
                      ? "bg-success text-success-foreground"
                      : "bg-destructive text-destructive-foreground"
              }`}
            >
              Mic: {role !== "speaker" ? "n/a" : micError ? "error" : isMicOn ? "on" : "off"}
            </span>
            {micError && (
              <span className="text-destructive">Mic error: {micError}</span>
            )}
          </div>
          <audio ref={remoteAudioRef} autoPlay playsInline />
        </CardContent>
      </Card>

      <Card className="rounded-3xl border-border/70 bg-card/80">
        <CardHeader>
          <CardTitle className="text-lg">Participants</CardTitle>
          <CardDescription>
            {participants.length} in room
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-2 text-sm">
          {participants.length === 0 && (
            <p className="text-muted-foreground">No one here yet.</p>
          )}
          {participants.map((participant) => (
            <div key={participant.userId} className="flex items-center justify-between rounded-2xl border border-border/60 bg-background/70 px-3 py-2">
              <div className="flex items-center gap-2">
                <span className="text-xs uppercase text-muted-foreground">
                  {participant.userId.slice(0, 8)}
                </span>
                {auth.user?.id === participant.userId && (
                  <span className="rounded-full bg-accent px-2 py-0.5 text-[10px] text-accent-foreground">
                    You
                  </span>
                )}
                {auth.user?.id === participant.userId && (
                  <span
                    className={`h-2 w-2 rounded-full ${
                      isSpeakingSelf ? "bg-success" : "bg-muted"
                    }`}
                    title={isSpeakingSelf ? "Speaking" : "Idle"}
                  />
                )}
              </div>
              <span
                className={`rounded-full px-2 py-0.5 text-[10px] ${
                  participant.role === "speaker"
                    ? "bg-success text-success-foreground"
                    : "bg-muted text-muted-foreground"
                }`}
              >
                {participant.role}
              </span>
            </div>
          ))}
        </CardContent>
      </Card>

      <Card className="rounded-3xl border-border/70 bg-card/80">
        <CardHeader>
          <CardTitle className="text-lg">Logs</CardTitle>
        </CardHeader>
        <CardContent className="space-y-1 text-xs text-muted-foreground">
          {logs.length === 0 && <p>No events yet.</p>}
          {logs.map((line, idx) => (
            <p key={idx}>{line}</p>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}

function avgLevel(buffer: Uint8Array) {
  let sum = 0;
  for (const value of buffer) {
    sum += value;
  }
  return Math.min(1, sum / buffer.length / 255);
}

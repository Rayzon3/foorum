import React from "react";

import { useAuth } from "../lib/auth";
import { Button } from "../components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card";
import { Input } from "../components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../components/ui/select";
import { Mic, MicOff } from "lucide-react";

type Role = "listener" | "speaker";

export function SpacesPage() {
  const auth = useAuth();
  const [roomId, setRoomId] = React.useState("lobby");
  const [role, setRole] = React.useState<Role>("listener");
  const [status, setStatus] = React.useState("disconnected");
  const [logs, setLogs] = React.useState<string[]>([]);
  const [isMicOn, setMicOn] = React.useState(false);
  const [micError, setMicError] = React.useState<string | null>(null);
  const [remoteAudioActive, setRemoteAudioActive] = React.useState(false);

  const wsRef = React.useRef<WebSocket | null>(null);
  const pcRef = React.useRef<RTCPeerConnection | null>(null);
  const localStreamRef = React.useRef<MediaStream | null>(null);
  const remoteAudioRef = React.useRef<HTMLAudioElement | null>(null);

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
  }


  function toggleMic() {
    if (!localStreamRef.current) {
      setMicError("mic_unavailable");
      return;
    }
    localStreamRef.current.getTracks().forEach((track) => {
      track.enabled = !track.enabled;
      setMicOn(track.enabled);
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
              <Button variant={isMicOn ? "success" : "destructive"} onClick={toggleMic} disabled={status !== "connected"}>
                {isMicOn ? <Mic className="h-4 w-4" /> : <MicOff className="h-4 w-4" />}
                {isMicOn ? "Mic on" : "Mic off"}
              </Button>
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

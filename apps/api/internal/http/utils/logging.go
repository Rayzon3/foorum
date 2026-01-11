package utils

import (
  "bufio"
  "net"
  "fmt"
  "log"
  "net/http"
  "time"
)

type statusRecorder struct {
  http.ResponseWriter
  status int
}

func (sr *statusRecorder) WriteHeader(code int) {
  sr.status = code
  sr.ResponseWriter.WriteHeader(code)
}

func (sr *statusRecorder) Flush() {
  if flusher, ok := sr.ResponseWriter.(http.Flusher); ok {
    flusher.Flush()
  }
}

func (sr *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
  hijacker, ok := sr.ResponseWriter.(http.Hijacker)
  if !ok {
    return nil, nil, http.ErrNotSupported
  }
  return hijacker.Hijack()
}

func (sr *statusRecorder) Push(target string, opts *http.PushOptions) error {
  pusher, ok := sr.ResponseWriter.(http.Pusher)
  if !ok {
    return http.ErrNotSupported
  }
  return pusher.Push(target, opts)
}

func RequestLogger(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
    next.ServeHTTP(sr, r)
    durationMs := time.Since(start).Milliseconds()
    log.Printf("method=%s path=%s status=%s duration_ms=%d", r.Method, r.URL.Path, colorStatus(sr.status), durationMs)
  })
}

func colorStatus(status int) string {
  color := "\x1b[32m" // green
  switch {
  case status >= 500:
    color = "\x1b[31m" // red
  case status >= 400:
    color = "\x1b[33m" // yellow
  case status >= 300:
    color = "\x1b[36m" // cyan
  }
  return fmt.Sprintf("%s%d\x1b[0m", color, status)
}

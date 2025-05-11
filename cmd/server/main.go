package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"

	"github.com/nimaaskarian/pomodoro/requests"
	"github.com/nimaaskarian/pomodoro/timer"
)

func main() {
  ln, err := net.Listen("tcp", ":8088")
  if err != nil {
    log.Fatal(err)
  }
  t := timer.Timer {
    Config: timer.DefaultConfig,
  }
  t.Config.OnTick = func(*timer.Timer) {
    fmt.Println(t.String())
  }
  t.Config.OnSeek = t.Config.OnTick
  t.Init()
  go t.Loop()
  buff := make([]byte, 1024)
  for {
    conn, err := ln.Accept()
    if err != nil {
      slog.Warn("connection throw error", "err", err)
      continue
    }
    conn.Write([]byte("OK PD 0.0.1\n"))
    for {
      n, err := conn.Read(buff)
      if err == io.EOF {
        break
      } else if err != nil {
        slog.Warn("read throw error", "err", err)
        continue
      }
      cmd, out, err := requests.ParseInput(&t, string(bytes.TrimSpace(buff[:n])))
      if err != nil {
        slog.Error("command throw error", "err", err)
        conn.Write(fmt.Appendf(nil, "ACK {%s} %s\n", cmd, err))
      } else {
        conn.Write([]byte(out))
        conn.Write([]byte("OK\n"))
      }
    }
  }
}

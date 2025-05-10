package main

import (
	"fmt"
	"io"
	"log"
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
    for {
      if err != nil {
        log.Fatal(err)
      }
      n, err := conn.Read(buff)
      if err == io.EOF {
        break
      } else if err != nil {
        log.Println("Error: ", err)
      }
      cmd, out, err := requests.ParseInput(&t, string(buff[:n-1]))
      if err != nil {
        fmt.Println(err)
        conn.Write(fmt.Appendf(nil, "ACK {%s} %s\n", cmd, err))
      } else {
        if out != "" {
          conn.Write([]byte(out+"\n"))
        }
        conn.Write([]byte("OK\n"))
      }
    }
  }
}

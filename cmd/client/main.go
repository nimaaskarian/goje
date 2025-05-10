package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
  conn, err := net.Dial("tcp", ":8088")
  if err != nil {
    log.Fatal(err)
  }
  defer conn.Close()
  fmt.Fprintln(conn, "I'm old and horny, like a hot pie...")
  buff := make([]byte, 1024)
  n, _ := conn.Read(buff)
  fmt.Printf("%s\n", buff[:n])
}

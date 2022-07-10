package main

import (
	"github.com/KylinMountain/StreamingServerGo/transport/rtmp"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	go startRTMPServer()
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGKILL, syscall.SIGABRT)
}

func startRTMPServer() {
	listener, err := net.Listen("tcp", ":1935")
	if err != nil {
		return
	}

	rc := rtmp.NewRTMPServer()
	rc.Serve(listener)
}

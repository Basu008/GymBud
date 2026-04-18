package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Basu008/GymBud/server"
)

func main() {
	s := server.NewServer()
	s.StartServer()

	err := sendSystemdNotification()
	if err != nil {
		s.Log.Error().Err(err).Msg("Systemd notification error")
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-c

	s.StopServer()
	s.Log.Info().Msg("Server Stopped")
}

func sendSystemdNotification() error {
	notifySocket := os.Getenv("NOTIFY_SOCKET")
	if notifySocket != "" {
		fmt.Println("notifyin")
	}
	return nil
}

package main

import (
	"github.com/ltfred/neam-protocol/server"
	"github.com/sirupsen/logrus"
)

func main() {
	s, err := server.NewServer(&server.Config{
		Host: "127.0.0.1",
		Port: 9999,
	})
	if err != nil {
		logrus.Error(err)

		return
	}

	s.Run()
}

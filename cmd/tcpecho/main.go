package main

import (
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"os"
	"strconv"
)

func main() {
	var port int
	if envPortStr, ok := os.LookupEnv("PORT"); ok {
		envPort, err := strconv.Atoi(envPortStr)
		if err != nil {
			logrus.Fatal(errors.Wrap(err, "unable to parse port from command line"))
			return
		}
		port = envPort
	} else {
		flag.IntVar(&port, "port", 10000, "port to listen on")
	}
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logrus.Fatal(errors.Wrap(err, "unable to listen for tcp"))
	}
	defer func() {
		if err := lis.Close(); err != nil {
			logrus.Error(errors.Wrap(err, "unable to close server"))
		}
	}()
	logrus.Info("proxy started")
	for {
		conn, err := lis.Accept()
		if err != nil {
			logrus.Error("cannot accept tcp connection")
			break
		}
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	if _, err := io.Copy(conn, conn); err != nil {
		if !errors.Is(err, net.ErrClosed) {
			logrus.Error(errors.Wrap(err, "unable to copy conn"))
		}
		_ = conn.Close()
		return
	}
}

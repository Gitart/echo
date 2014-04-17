package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"syscall"
)

var (
	flagSet     = flag.NewFlagSet("admin", flag.ExitOnError)
	showVersion = flag.Bool("version", false, "print version string")
	port        = flag.Int("port", 8000, "listening port")
	bind        = flag.String("bind", "0.0.0.0", "binding ip addr")
)

// tcpdump -i lo0 -Xn "port 8000"
func main() {
	addr := fmt.Sprintf("%s:%d", *bind, *port)
	fmt.Printf("INFO: bind on addr %s\n", addr)

	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		fmt.Printf("ERROR: invalid addr [%s] , err %s\n", tcpAddr, err)
		return
	}

	exitChan := make(chan int)
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		exitChan <- 1
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go start(addr)

	<-exitChan
}

func start(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("ERROR: listen on port [:%s] error %s\n", port, err)
	}
	defer func() { ln.Close() }()

	fmt.Printf("INFO: start listenn on port %d\n", *port)
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			fmt.Printf("ERROR: accept [%s] err %s\n", conn.RemoteAddr().String(), err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	fmt.Printf("INFO: handle conn %s\n", conn.RemoteAddr().String())

	r := bufio.NewReader(conn)
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Printf("INFO: [%s] client close connection.\n", conn.RemoteAddr().String())
			} else {
				fmt.Printf("ERROR: [%s] conn err %s.\n", conn.RemoteAddr().String(), err)
			}
			break
		}
		fmt.Printf("INFO: [%s]<data> %s", conn.RemoteAddr().String(), line)
		_, err = conn.Write([]byte(line))
		if err != nil {
			fmt.Printf("ERROR: [%s] conn write err %s.\n", conn.RemoteAddr().String(), err)
			break
		}
	}
}

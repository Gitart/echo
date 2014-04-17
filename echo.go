package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
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
		log.Fatalf("ERROR: invalid addr [%s] , err %s\n", tcpAddr, err)
		return
	}

	exitChan := make(chan int)
	signalChan := make(chan os.Signal, 1)
	go func() {
		<-signalChan
		exitChan <- 1
	}()
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	s := NewEchoServer(addr)
	go s.Start()
	<-exitChan
	s.Stop()
}

type EchoServer struct {
	bindAddr    string
	tcpListener net.Listener
}

func NewEchoServer(bindAddr string) *EchoServer {
	return &EchoServer{bindAddr: bindAddr}
}

func (s *EchoServer) Start() {
	tcpListener, err := net.Listen("tcp", s.bindAddr)
	if err != nil {
		log.Fatalf("ERROR: listen on [%s] error %s\n", s.bindAddr, err)
	}
	s.tcpListener = tcpListener

	fmt.Printf("INFO: start listenn on port %d\n", *port)
	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			// handle error
			fmt.Printf("ERROR: accept [%s] err %s\n", conn.RemoteAddr().String(), err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *EchoServer) Stop() {
	if s.tcpListener != nil {
		s.tcpListener.Close()
	}
	fmt.Printf("INFO: echo server stop\n")
}

func (s *EchoServer) handleConnection(conn net.Conn) {
	client := conn.RemoteAddr().String()
	fmt.Printf("INFO: handle conn %s\n", client)

	r := bufio.NewReader(conn)
	defer func() {
		if conn != nil {
			conn.Close()
		}
		fmt.Printf("INFO: close conn [%s]\n", client)
	}()

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Printf("INFO: [%s] client close connection.\n", client)
			} else {
				fmt.Printf("ERROR: [%s] conn err %s.\n", client, err)
			}
			break
		}
		fmt.Printf("INFO: [%s]<data> %s", client, line)
		_, err = conn.Write([]byte(line))
		if err != nil {
			fmt.Printf("ERROR: [%s] conn write err %s.\n", client, err)
			break
		}
	}
}

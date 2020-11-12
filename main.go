package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"golang.org/x/net/websocket"
)

func main() {
	wst := WSTerminal{}
	wst.ParseFlags()
	fmt.Printf("%+v\n", wst)

	wst.Connect()
}

// WSTerminal is used to interact with a websocket
type WSTerminal struct {
	Execute     string
	Interactive bool
	Insecure    bool
	Output      string
	Reconnect   bool
	Timeout     int
	HostAddress string

	wsConn *websocket.Conn
}

// ParseFlags parses cli flags
func (wst *WSTerminal) ParseFlags() {
	flag.BoolVar(&wst.Interactive, "i", false, "Interact with the websocket and send messages")
	flag.BoolVar(&wst.Insecure, "insecure", false, "Disable certificate verification on ssl connections")
	// flag.StringVar(&wst.Output, "o", "", "Output all communication to a file")
	// flag.StringVar(&wst.Execute, "x", "", "Execute a command on every event")
	// flag.IntVar(&wst.Timeout, "t", 0, "Wait up to t seconds to receive a message from the host")
	// flag.BoolVar(&wst.Reconnect, "r", false, "Automatically reconnect to the host if the connection is dropped")

	flag.Parse()

	if flag.NArg() != 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	wst.HostAddress = flag.Arg(0)
}

// Connect to the websocket host
func (wst *WSTerminal) Connect() {
	wst.dial()
	if wst.Interactive {
		go wst.prompt()
	}
	wst.read()
}

func (wst *WSTerminal) dial() {
	hostURL, err := url.Parse(wst.HostAddress)
	if err != nil {
		log.Fatal(err)
	}
	wsConfig, err := websocket.NewConfig(hostURL.String(), hostURL.String())
	if err != nil {
		log.Fatal(err)
	}
	if wst.Insecure {
		wsConfig.TlsConfig = &tls.Config{InsecureSkipVerify: true}
	}
	wst.wsConn, err = websocket.DialConfig(wsConfig)
	if err != nil {
		log.Fatal(err)
	}
}

func (wst *WSTerminal) read() {
	var message string
	for wst.wsConn != nil {
		readErr := websocket.Message.Receive(wst.wsConn, &message)
		if readErr != nil {
			log.Println("Failed to read message ", readErr)
			wst.wsConn.Close()
			wst.wsConn = nil
		}
		fmt.Println("<", message)
	}
}

func (wst *WSTerminal) prompt() {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		scanner.Scan()
		writeErr := websocket.Message.Send(wst.wsConn, scanner.Text())
		if writeErr != nil {
			log.Println("Failed to write message ", writeErr)
			wst.wsConn.Close()
			wst.wsConn = nil
		}
	}
}

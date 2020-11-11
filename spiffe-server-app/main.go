package main

import (
	"bufio"
	"context"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"log"
	"net"
	"net/http"
)

const (
    serverAddress = ":8080"
	socketPath = "unix:///run/spire/sockets/agent.sock"
)

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("I am doing great, thanks for asking!"))
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	req, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		log.Printf("Error reading req: %s", err.Error())
		return
	}
	log.Printf("From client: %q", req)

	_, err = conn.Write([]byte("Hello from the server!\n"))
	if err != nil {
		log.Printf("Error sending response: %s", err.Error())
	}
}

func main() {
	ctx := context.Background()
	clientID := spiffeid.Must("example.org", "ns", "default", "sa", "spiffe-client-sa")
	listener, err := spiffetls.ListenWithMode(
		// Set a context so we can cancel and timeout
		ctx,
		"tcp",
		// Set the server and port we'll listen to for clients
		serverAddress,
		// Configure the x509 source
		spiffetls.MTLSServerWithSourceOptions(
			// Configure what SVID can talk to us
			tlsconfig.AuthorizeID(clientID),
			// Configure where our workload API socket is. We could omit this and use SPIFFE_ENDPOINT_SOCKET instead.
			workloadapi.WithClientOptions(workloadapi.WithAddr(socketPath)),
		))
	if err != nil {
		log.Fatalf("Unable to create TLS listener: %v", err)
	}
	defer listener.Close()

	http.HandleFunc("/health", healthHandler)
	go http.ListenAndServe(":8081", nil)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Unable to accept connection: %s", err.Error())
		}
		go handleConnection(conn)
	}
}


package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"net/http"
)

const (
	socketPath = "unix:///run/spire/sockets/agent.sock"
)

func rootHandler(w http.ResponseWriter, _ *http.Request) {
	ctx := context.Background()
	serverID := spiffeid.Must("example.org", "ns", "default", "sa", "spiffe-server-app-sa")

	conn, err := spiffetls.DialWithMode(
		ctx,
		"tcp",
		"spiffe-server-app:8080",
		spiffetls.MTLSClientWithSourceOptions(
			tlsconfig.AuthorizeID(serverID),
			workloadapi.WithClientOptions(workloadapi.WithAddr(socketPath)),
		))
	if err != nil {
		_, _ = w.Write([]byte(fmt.Sprintf("Something went wrong: %s", err.Error())))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	_, _ = fmt.Fprintf(conn, "Hello from the client!\n")

	serverResp, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		_, _ = w.Write([]byte(fmt.Sprintf("Something went wrong: %s", err.Error())))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte(fmt.Sprintf("Backend server response: %s", serverResp)))
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("I am doing great, thanks for asking!"))
}

func whoamiHandler(w http.ResponseWriter, _ *http.Request) {
	ctx := context.Background()
	client, err := workloadapi.New(ctx, workloadapi.WithAddr(socketPath))
	if err != nil {
		_, _ = w.Write([]byte(fmt.Sprintf("Something went wrong: %s", err.Error())))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	svid, err := client.FetchX509SVID(ctx)
	if err != nil {
		_, _ = w.Write([]byte(fmt.Sprintf("Something went wrong: %s", err.Error())))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	cert, key, err := svid.Marshal()
	if err != nil {
		_, _ = w.Write([]byte(fmt.Sprintf("Something went wrong: %s", err.Error())))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, _ = w.Write([]byte(fmt.Sprintf("This is me:\n%s\n\nThis is my key:\n%s", cert, key)))
}

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/whoami", whoamiHandler)

	_ = http.ListenAndServe(":8080", nil)
}

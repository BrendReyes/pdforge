package cmd

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/brendreyes/pdforge/internal/server"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a local web GUI server",
	Long: `Serve starts a lightweight web GUI accessible from any device on the same network.
All PDF processing happens locally — no files are uploaded to any cloud service.`,
	Example: `  pdforge serve
  pdforge serve --port 9090
  pdforge serve --host 127.0.0.1 --port 7878`,
	RunE: runServe,
}

var servePort int
var serveHost string

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.SetHelpTemplate(subHelpTemplate)
	serveCmd.Flags().IntVarP(&servePort, "port", "p", 7878, "Port to listen on")
	serveCmd.Flags().StringVar(&serveHost, "host", "0.0.0.0", "Host address to bind to")
}

func runServe(cmd *cobra.Command, args []string) error {
	addr := fmt.Sprintf("%s:%d", serveHost, servePort)
	localIP := getLocalIP()

	fmt.Fprintln(cmd.OutOrStdout(), "pdforge serve")
	fmt.Fprintln(cmd.OutOrStdout(), "")
	fmt.Fprintf(cmd.OutOrStdout(), "  Local:   http://localhost:%d\n", servePort)
	if localIP != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "  Network: http://%s:%d\n", localIP, servePort)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "")
	fmt.Fprintln(cmd.OutOrStdout(), "Press Ctrl+C to stop.")
	fmt.Fprintln(cmd.OutOrStdout(), "")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	srv := server.New(addr)

	go func() {
		<-quit
		fmt.Fprintln(cmd.OutOrStdout(), "\nShutting down...")
		srv.Shutdown()
	}()

	return srv.Start()
}

// getLocalIP returns the machine's outbound LAN IP by probing a UDP connection.
// No packet is actually sent — UDP is connectionless.
func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		return ""
	}

	return localAddr.IP.String()
}

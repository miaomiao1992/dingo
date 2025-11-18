package main

import (
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/MadAppGang/dingo/pkg/lsp"
	"go.lsp.dev/jsonrpc2"
)

var logger lsp.Logger

func main() {
	// Configure logging from environment variable
	logLevel := os.Getenv("DINGO_LSP_LOG")
	if logLevel == "" {
		logLevel = "info"
	}
	logger = lsp.NewLogger(logLevel, os.Stderr)

	logger.Infof("Starting dingo-lsp server (log level: %s)", logLevel)

	// Find gopls in $PATH
	goplsPath := findGopls(logger)
	if goplsPath == "" {
		logger.Fatalf("gopls not found in $PATH. Install: go install golang.org/x/tools/gopls@latest")
	}

	// Create LSP proxy server
	server, err := lsp.NewServer(lsp.ServerConfig{
		Logger:        logger,
		GoplsPath:     goplsPath,
		AutoTranspile: true, // Default from user decision
	})
	if err != nil {
		logger.Fatalf("Failed to create server: %v", err)
	}

	// Create stdio transport using ReadWriteCloser wrapper
	logger.Infof("Creating stdin/stdout ReadWriteCloser")
	rwc := &stdinoutCloser{stdin: os.Stdin, stdout: os.Stdout, logger: logger}
	logger.Infof("Creating JSON-RPC2 stream")
	stream := jsonrpc2.NewStream(rwc)
	logger.Infof("Creating JSON-RPC2 connection")
	conn := jsonrpc2.NewConn(stream)
	logger.Infof("JSON-RPC2 connection created: %p", conn)

	// Start serving with cancellable context (Gemini fix)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// CRITICAL FIX (Sherlock): Store connection BEFORE starting handler
	// This prevents race condition where handlers try to use nil ideConn
	logger.Infof("Storing connection in server")
	server.SetConn(conn, ctx)

	// Create handler and start connection
	handler := server.Handler()
	logger.Infof("Starting JSON-RPC2 connection handler")
	conn.Go(ctx, handler)
	logger.Infof("JSON-RPC2 connection handler started")

	// Wait for connection to close
	<-conn.Done()
	logger.Infof("JSON-RPC2 connection handler finished")
	logger.Infof("Server stopped")
}

// findGopls looks for gopls binary in $PATH
func findGopls(logger lsp.Logger) string {
	path, err := exec.LookPath("gopls")
	if err != nil {
		logger.Debugf("gopls not found in $PATH: %v", err)
		return ""
	}
	logger.Infof("Found gopls at: %s", path)
	return path
}

// stdinoutCloser wraps os.Stdin and os.Stdout as ReadWriteCloser
type stdinoutCloser struct {
	stdin  *os.File
	stdout *os.File
	logger lsp.Logger
}

func (s *stdinoutCloser) Read(p []byte) (n int, err error) {
	n, err = s.stdin.Read(p)
	s.logger.Debugf("stdinoutCloser.Read: n=%d, err=%v", n, err)
	return n, err
}

func (s *stdinoutCloser) Write(p []byte) (n int, err error) {
	n, err = s.stdout.Write(p)
	s.logger.Debugf("stdinoutCloser.Write: n=%d, err=%v", n, err)
	return n, err
}

func (s *stdinoutCloser) Close() error {
	s.logger.Infof("stdinoutCloser.Close called")
	// Don't actually close stdin/stdout, but log the event
	return nil
}

var _ io.ReadWriteCloser = (*stdinoutCloser)(nil)


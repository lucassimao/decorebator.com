package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"testing"
	"time"
)

type shutdownRecorder struct {
	hadDeadline bool
	remaining   time.Duration
	shutdownErr error
	closeErr    error
	closeCalls  int
}

func (s *shutdownRecorder) Shutdown(ctx context.Context) error {
	deadline, ok := ctx.Deadline()
	s.hadDeadline = ok
	if ok {
		s.remaining = time.Until(deadline)
	}
	return s.shutdownErr
}

func (s *shutdownRecorder) Close() error {
	s.closeCalls++
	return s.closeErr
}

func TestShutdownHTTPServerUsesProductionDeadline(t *testing.T) {
	t.Parallel()

	server := &shutdownRecorder{}
	timeout := 250 * time.Millisecond
	if err := shutdownHTTPServer(server, "production", timeout); err != nil {
		t.Fatalf("shutdownHTTPServer returned error: %v", err)
	}
	if !server.hadDeadline {
		t.Fatal("production shutdown context did not have a deadline")
	}
	if server.remaining <= 0 || server.remaining > timeout {
		t.Fatalf("production deadline outside expected bound: %s", server.remaining)
	}
	if server.closeCalls != 0 {
		t.Fatalf("successful graceful shutdown forced close %d times", server.closeCalls)
	}
}

func TestShutdownHTTPServerPreservesNonProductionBehavior(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("shutdown failed")
	server := &shutdownRecorder{shutdownErr: wantErr}
	err := shutdownHTTPServer(server, "development", 10*time.Millisecond)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected shutdown error %v, got %v", wantErr, err)
	}
	if server.hadDeadline {
		t.Fatal("non-production shutdown unexpectedly received a deadline")
	}
	if server.closeCalls != 1 {
		t.Fatalf("shutdown error forced close %d times, want 1", server.closeCalls)
	}
}

type blockingShutdowner struct {
	closeErr   error
	closeCalls int
}

func (s *blockingShutdowner) Shutdown(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func (s *blockingShutdowner) Close() error {
	s.closeCalls++
	return s.closeErr
}

func TestShutdownHTTPServerBoundsProductionDrain(t *testing.T) {
	t.Parallel()

	timeout := 20 * time.Millisecond
	startedAt := time.Now()
	server := &blockingShutdowner{}
	err := shutdownHTTPServer(server, "production", timeout)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
	if server.closeCalls != 1 {
		t.Fatalf("timed-out shutdown forced close %d times, want 1", server.closeCalls)
	}
	if elapsed := time.Since(startedAt); elapsed > 500*time.Millisecond {
		t.Fatalf("production shutdown exceeded deterministic bound: %s", elapsed)
	}
}

func TestShutdownHTTPServerReturnsShutdownAndCloseErrors(t *testing.T) {
	t.Parallel()

	shutdownErr := errors.New("shutdown failed")
	closeErr := errors.New("close failed")
	server := &shutdownRecorder{shutdownErr: shutdownErr, closeErr: closeErr}

	err := shutdownHTTPServer(server, "production", time.Second)
	if !errors.Is(err, shutdownErr) {
		t.Fatalf("combined error does not contain shutdown failure: %v", err)
	}
	if !errors.Is(err, closeErr) {
		t.Fatalf("combined error does not contain force-close failure: %v", err)
	}
	if server.closeCalls != 1 {
		t.Fatalf("failed shutdown forced close %d times, want 1", server.closeCalls)
	}
}

func TestShutdownHTTPServerDrainsInFlightRequest(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	requestStarted := make(chan struct{})
	releaseRequest := make(chan struct{})
	server := &http.Server{
		Handler: http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
			close(requestStarted)
			<-releaseRequest
			writer.WriteHeader(http.StatusNoContent)
		}),
	}
	defer server.Close()

	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.Serve(listener)
	}()

	responseStatus := make(chan int, 1)
	requestErr := make(chan error, 1)
	go func() {
		response, requestError := http.Get("http://" + listener.Addr().String()) //nolint:noctx // Test client is bounded by the shutdown deadline below.
		if requestError != nil {
			requestErr <- requestError
			return
		}
		defer response.Body.Close()
		responseStatus <- response.StatusCode
	}()

	select {
	case <-requestStarted:
	case <-time.After(time.Second):
		close(releaseRequest)
		t.Fatal("request did not reach the server")
	}

	shutdownDone := make(chan error, 1)
	go func() {
		shutdownDone <- shutdownHTTPServer(server, "production", time.Second)
	}()

	// Shutdown must wait for the active handler rather than force-closing it.
	select {
	case shutdownErr := <-shutdownDone:
		close(releaseRequest)
		t.Fatalf("shutdown returned before the request drained: %v", shutdownErr)
	case <-time.After(20 * time.Millisecond):
	}
	close(releaseRequest)

	if shutdownErr := <-shutdownDone; shutdownErr != nil {
		t.Fatalf("shutdownHTTPServer returned error: %v", shutdownErr)
	}
	select {
	case status := <-responseStatus:
		if status != http.StatusNoContent {
			t.Fatalf("request status = %d, want %d", status, http.StatusNoContent)
		}
	case requestError := <-requestErr:
		t.Fatalf("in-flight request failed: %v", requestError)
	case <-time.After(time.Second):
		t.Fatal("in-flight request did not complete")
	}
	if err = <-serveErr; !errors.Is(err, http.ErrServerClosed) {
		t.Fatalf("Serve returned %v, want http.ErrServerClosed", err)
	}
}

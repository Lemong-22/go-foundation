package playground

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"strings"
	"time"
)

const DefaultAddress = "127.0.0.1:8787"

//go:embed playground.html
var assets embed.FS

type Server struct {
	runner Runner
	index  *template.Template
}

type ServerOptions struct {
	Runner Runner
}

func NewServer(options ServerOptions) (*Server, error) {
	index, err := template.ParseFS(assets, "playground.html")
	if err != nil {
		return nil, fmt.Errorf("parse playground index: %w", err)
	}

	return &Server{
		runner: options.Runner,
		index:  index,
	}, nil
}

func (server *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handleIndex)
	mux.HandleFunc("/run", server.handleRun)
	mux.HandleFunc("/import/plan", server.handleImportPlan)
	mux.HandleFunc("/import/apply", server.handleImportApply)
	return mux
}

func (server *Server) ListenAndServe(ctx context.Context, address string) error {
	if strings.TrimSpace(address) == "" {
		address = DefaultAddress
	}
	if err := validateLoopbackAddress(address); err != nil {
		return err
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", address, err)
	}

	httpServer := &http.Server{
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- httpServer.Serve(listener)
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown playground server: %w", err)
		}
	}

	err = <-errCh
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func validateLoopbackAddress(address string) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("invalid playground address %q: %w", address, err)
	}
	if host == "localhost" {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("playground address must bind to loopback, got %q", address)
	}

	return nil
}

func (server *Server) handleIndex(response http.ResponseWriter, request *http.Request) {
	if request.URL.Path != "/" {
		http.NotFound(response, request)
		return
	}

	catalogJSON, err := json.Marshal(Catalog())
	if err != nil {
		http.Error(response, "catalog unavailable", http.StatusInternalServerError)
		return
	}

	response.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = server.index.Execute(response, struct {
		CatalogJSON template.JS
	}{
		CatalogJSON: template.JS(catalogJSON),
	})
	if err != nil {
		http.Error(response, "render playground index", http.StatusInternalServerError)
	}
}

func (server *Server) handleRun(response http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		response.Header().Set("Allow", http.MethodPost)
		http.Error(response, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	values, err := requestValues(request)
	if err != nil {
		http.Error(response, err.Error(), http.StatusBadRequest)
		return
	}

	action := values["action"]
	delete(values, "action")

	result := server.runner.Run(request.Context(), action, values)
	status := httpStatus(result.ExitCode)
	writeJSON(response, status, result)
}

func requestValues(request *http.Request) (map[string]string, error) {
	contentType := request.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		var values map[string]string
		if err := json.NewDecoder(request.Body).Decode(&values); err != nil {
			return nil, fmt.Errorf("decode request json: %w", err)
		}
		return values, nil
	}

	if err := request.ParseForm(); err != nil {
		return nil, fmt.Errorf("decode request form: %w", err)
	}

	values := make(map[string]string)
	for key, entries := range request.PostForm {
		if len(entries) > 0 {
			values[key] = entries[len(entries)-1]
		}
	}

	return values, nil
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(value)
}

func httpStatus(exitCode int) int {
	switch exitCode {
	case 0:
		return http.StatusOK
	case 2:
		return http.StatusNotFound
	case 5:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}

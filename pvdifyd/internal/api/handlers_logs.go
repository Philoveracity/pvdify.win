package api

import (
	"bufio"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// handleLogs streams logs for an app (SSE)
func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	app, err := s.db.GetApp(name)
	if err != nil || app == nil {
		s.error(w, http.StatusNotFound, "app not found")
		return
	}

	// Check for follow mode
	follow := r.URL.Query().Get("follow") == "true"

	// Get line count
	lines := 100
	if l := r.URL.Query().Get("lines"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			lines = parsed
		}
	}

	// Get process filter
	process := r.URL.Query().Get("process")
	if process == "" {
		process = "*"
	}

	// Build journalctl command
	unit := fmt.Sprintf("pvdify-%s-%s@*", name, process)
	args := []string{"-u", unit, "-n", strconv.Itoa(lines), "--no-pager", "-o", "short-iso"}
	if follow {
		args = append(args, "-f")
	}

	cmd := exec.CommandContext(r.Context(), "journalctl", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		s.logger.Error("create stdout pipe", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to start log stream")
		return
	}

	if err := cmd.Start(); err != nil {
		s.logger.Error("start journalctl", "error", err)
		s.error(w, http.StatusInternalServerError, "failed to start log stream")
		return
	}

	// Set up SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		s.error(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Fprintf(w, "data: %s\n\n", scanner.Text())
		flusher.Flush()

		// Check if client disconnected
		select {
		case <-r.Context().Done():
			cmd.Process.Kill()
			return
		default:
		}
	}

	cmd.Wait()
}

// Package middleware .
package middleware

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (res *responseRecorder) WriteHeader(code int) {
	res.statusCode = code
	res.ResponseWriter.WriteHeader(code)
}

func (res *responseRecorder) Write(p []byte) (int, error) {
	if res.statusCode == 0 {
		res.statusCode = http.StatusOK
	}
	n, err := res.ResponseWriter.Write(p)
	res.bytesWritten += n
	return n, err
}

type requestRecorder struct {
	io.ReadCloser
	bytesRead int
}

func (req *requestRecorder) Read(p []byte) (int, error) {
	n, err := req.ReadCloser.Read(p)
	req.bytesRead += n
	return n, err
}

func redactIP(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		host = addr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return host
	}
	if ip4 := ip.To4(); ip4 != nil {
		return fmt.Sprintf("%d.%d.%d.x", ip4[0], ip4[1], ip4[2])
	}
	return ip.String()
}

func LoggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqRec := &requestRecorder{ReadCloser: r.Body}
			r.Body = reqRec
			respRec := &responseRecorder{ResponseWriter: w}

			next.ServeHTTP(respRec, r)

			if r.URL.Path == "/health" {
				return
			}

			attrs := []any{
				slog.Duration("duration", time.Duration(time.Since(start).Milliseconds())),
				slog.Group("request",
					"method", r.Method,
					"uri", r.URL.RequestURI(),
					"ip", redactIP(r.RemoteAddr),
					// "request_body_bytes", reqRec.bytesRead, // not really useful atm - always 0 since no current endpoint needs a req body
				),
				slog.Group("response",
					"status", respRec.statusCode,
					"body_bytes", respRec.bytesWritten,
				),
			}
			logger.Info("Served request", attrs...)
		})
	}
}

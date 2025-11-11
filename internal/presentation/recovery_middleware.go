package presentation

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

// RecoveryMiddleware wraps an HTTP handler with panic recovery
func RecoveryMiddleware(logger interface{ Error(string) }) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic with stack trace
					stackTrace := debug.Stack()
					logger.Error(fmt.Sprintf("Panic recovered: %v\n%s", err, stackTrace))

					// Return 500 error to client
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// RecoveryHandlerFunc wraps a HandlerFunc with panic recovery
func RecoveryHandlerFunc(logger interface{ Error(string) }, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				stackTrace := debug.Stack()
				logger.Error(fmt.Sprintf("Panic recovered in %s %s: %v\n%s", r.Method, r.URL.Path, err, stackTrace))

				// Return 500 error to client
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		handler(w, r)
	}
}

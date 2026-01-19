package middleware

import (
	"log"
	"net/http"
)

func LogRequest(logger *log.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("%s - %s %s %s",
			r.RemoteAddr,
			r.Proto,
			r.Method,
			r.URL.RequestURI(),
		)
		next.ServeHTTP(w, r)
	})
}

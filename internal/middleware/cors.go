package middleware

import "net/http"

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// 🔑 Allow your frontend origin
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4000")

		// Required headers for fetch / axios
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Allowed HTTP methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Optional: allow cookies / auth headers
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

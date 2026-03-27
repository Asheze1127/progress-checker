package rest

import "net/http"

// CORSMiddleware returns an HTTP middleware that sets CORS headers for allowed origins.
func CORSMiddleware(allowedOrigins []string) func(http.Handler) http.Handler {
  return func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      w.Header().Add("Vary", "Origin")
      origin := r.Header.Get("Origin")
      if origin != "" {
        for _, allowed := range allowedOrigins {
          if origin == allowed {
            w.Header().Set("Access-Control-Allow-Origin", origin)
            w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
            w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
            w.Header().Set("Access-Control-Max-Age", "86400")
            break
          }
        }
      }
      next.ServeHTTP(w, r)
    })
  }
}

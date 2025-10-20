package middleware

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Middleware signature for httprouter handlers
type Middleware func(httprouter.Handle) httprouter.Handle

// Chain composes middlewares left-to-right
func Chain(middlewares ...Middleware) Middleware {
	return func(final httprouter.Handle) httprouter.Handle {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// ResponseWriterWithStatus wraps http.ResponseWriter to capture status codes
type ResponseWriterWithStatus struct {
	http.ResponseWriter
	status int
}

func (rw *ResponseWriterWithStatus) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// WrapResponseWriter ensures we can capture handler’s response status
func WrapResponseWriter(w http.ResponseWriter) *ResponseWriterWithStatus {
	return &ResponseWriterWithStatus{ResponseWriter: w, status: 200}
}

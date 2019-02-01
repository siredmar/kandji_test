package responselog

import (
	"encoding/json"
	"net/http"
	"runtime"
)

// LogResponseWriter is a wrapper around http.ResponseWriter to get additional
// information.
type LogResponseWriter struct {
	// HTTP
	http.ResponseWriter
	// Additional log fields
	fields map[string]interface{}
	// Tracking and tracing errors
	err   error
	stack []stackFrame
}

// NewLogResponseWriter creates a logResponseWriter.
func NewLogResponseWriter(w http.ResponseWriter, r *http.Request) *LogResponseWriter {
	return &LogResponseWriter{
		ResponseWriter: w,
		fields: map[string]interface{}{
			"remoteaddr":    r.RemoteAddr,
			"xforwardedfor": r.Header.Get("X-Forwarded-For"),
			"method":        r.Method,
			"path":          r.URL.Path,
			"query":         r.URL.RawQuery,
			"useragent":     r.UserAgent(),
			"transactionid": r.Header.Get("X-Amzn-Trace-Id"),
			"contentlength": 0,
		},
	}
}

// SetFields sets the log fields. It replaces any existing value associated
// with the key.
func (w *LogResponseWriter) SetFields(f map[string]interface{}) {
	for key, value := range f {
		w.fields[key] = value
	}
}

// Fields returns all log fields.
func (w *LogResponseWriter) Fields() map[string]interface{} {
	return w.fields
}

// WriteHeader sends an HTTP response header with status code.
func (w *LogResponseWriter) WriteHeader(code int) {
	w.fields["status"] = code
	w.ResponseWriter.WriteHeader(code)
}

// StatusCode returns the status code. e.g. 200.
func (w *LogResponseWriter) StatusCode() int {
	code, ok := w.fields["status"].(int)
	if ok {
		return code
	}
	return 0
}

// Write writes the data to the connetion as part of an HTTP reply.
func (w *LogResponseWriter) Write(b []byte) (int, error) {
	if w.StatusCode() == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(b)
	w.fields["contentlength"] = w.fields["contentlength"].(int) + n
	return n, err
}

// Flush sends any buffered data to the client.
func (w *LogResponseWriter) Flush() {
	flusher, ok := w.ResponseWriter.(http.Flusher)
	if ok {
		if w.StatusCode() == 0 {
			w.WriteHeader(http.StatusOK)
		}
		flusher.Flush()
	}
}

// WriteError writes an error to the connection as part of an HTTP reply.
// It also fills the stack in order to return a stack trace.
func (w *LogResponseWriter) WriteError(err error) (int, error) {
	// Limit stack trace depth.
	pc := make([]uintptr, 10)
	// Ignore first 2 calls because they're useless.
	n := runtime.Callers(2, pc)
	pc = pc[:n]
	if b, err := json.Marshal(retrieveStackFrames(pc)); err == nil {
		w.fields["stacktrace"] = string(b)
	}
	w.err = err
	if err, ok := err.(ResponseError); ok {
		w.WriteHeader(err.StatusCode())
		return w.Write(err.Bytes())
	}
	// Fallback because of unknown error type.
	w.WriteHeader(http.StatusInternalServerError)
	return w.Write([]byte(`something went wrong!`))
}

// Err may return an error.
func (w *LogResponseWriter) Err() error {
	return w.err
}

// ResponseError has
type ResponseError interface {
	StatusCode() int
	Bytes() []byte
}

// retrieveStackFrames returns a slice of stack frames for given program counters.
func retrieveStackFrames(pc []uintptr) []stackFrame {
	stackFrames := make([]stackFrame, 0, len(pc))
	frames := runtime.CallersFrames(pc)
	for {
		f, more := frames.Next()
		sf := stackFrame{
			Function: f.Func.Name(),
		}
		sf.File, sf.LineNumber = f.Func.FileLine(f.PC - 1)
		stackFrames = append(stackFrames, sf)
		if !more {
			break
		}
	}
	return stackFrames
}

type stackFrame struct {
	File       string `json:"file"`
	LineNumber int    `json:"line_number"`
	Function   string `json:"function"`
}

// ErrorResponseWriter interface extends the responseWriter interface by
// WriteError. Gets an error, returns fd and error
type ErrorResponseWriter interface {
	http.ResponseWriter
	WriteError(error) (int, error)
}

// ErrorHandler is an ServeHTTP interface
type ErrorHandler interface {
	ServeHTTP(ErrorResponseWriter, *http.Request)
}

// ErrorHandlerFunc does use ErrorResponseWriter to extend the ResponseWriter
// interface
type ErrorHandlerFunc func(ErrorResponseWriter, *http.Request)

package responselog

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// Middleware returns a middleware which emits a response log to the given
// logger
func Middleware(logger log.FieldLogger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := NewLogResponseWriter(w, r)
			s := time.Now()
			h.ServeHTTP(rw, r)
			rw.SetFields(log.Fields{
				"responsetime": float64(time.Since(s).Nanoseconds()) / 1000000.0, // in ms.
			})
			if rw.Err() != nil {
				l := logger.WithFields(rw.Fields())
				if rw.StatusCode() > 499 {
					l.Errorf("%v", rw.Err())
					return
				}
				l.Warnf("%v", rw.Err())
				return
			}
			logger.WithFields(rw.Fields()).Info("finished request")
		})
	}
}

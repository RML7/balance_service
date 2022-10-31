package middleware

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/avito-test/internal/config/logger"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	_ "github.com/google/uuid"
)

func ResponseHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.Clone(context.WithValue(r.Context(), "requestId", uuid.New().String())) //r.WithContext(context.WithValue(r.Context(), "requestId", uuid.New().String()))
		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
	body   []byte
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(body []byte) (int, error) {
	rw.body = body
	i, err := rw.ResponseWriter.Write(body)
	return i, err
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetLogger()

		bodyBuf, err := io.ReadAll(r.Body)

		if err != nil {
			log.Fatal(err.Error())
		}

		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBuf))

		requestId := r.Context().Value("requestId")

		rw := responseWriter{ResponseWriter: w}

		log.WithFields(logrus.Fields{
			"request_id": requestId,
			"method":     r.Method,
			"uri":        r.RequestURI,
			"body":       string(bodyBuf),
		}).Info(fmt.Sprintf("%s_REQUEST", requestId))

		next.ServeHTTP(&rw, r)

		log.WithFields(logrus.Fields{
			"request_id": requestId,
			"status":     rw.status,
			"body":       string(rw.body),
		}).Info(fmt.Sprintf("%s_RESPONSE", requestId))
	})
}

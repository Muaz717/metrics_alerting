package logger

import (
	"net/http"
	// "time"

	"go.uber.org/zap"
)

var Log *zap.Logger = zap.NewNop()

func Initialize(level string) error{
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil{
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	zl, err := cfg.Build()
	if err != nil{
		return err
	}

	Log = zl
	return nil
}

type responseData struct{
	size 	int
	status 	int
}

type loggingResponseWriter struct{
	http.ResponseWriter
	responseData *responseData
}

func (l *loggingResponseWriter) Write(b []byte) (int, error){
	size, err := l.ResponseWriter.Write(b)
	l.responseData.size += size
	return size, err
}

func (l *loggingResponseWriter) WriteHeader(statusCode int){
	l.ResponseWriter.WriteHeader(statusCode)
	l.responseData.status = statusCode
}

func WithLogging(next http.HandlerFunc) http.HandlerFunc{
	return func(w http.ResponseWriter, r *http.Request) {
		// start := time.Now()


		responseData := &responseData{
			size: 0,
			status: 0,
		}

		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData: responseData,
		}

		next(&lw, r)

		// duration := time.Since(start)
		// Log.Info(
		// 	"Got incoming HTTP request",
		// 	zap.String("url", r.RequestURI),
		// 	zap.String("method", r.Method),
		// 	zap.Duration("duration", duration),
		// 	zap.Int("size", responseData.size),
		// 	zap.Int("status", responseData.status),
		// )
	}
}

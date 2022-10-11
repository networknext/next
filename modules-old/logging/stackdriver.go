package logging

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/go-kit/kit/log/level"

	gcplogging "cloud.google.com/go/logging"
	"github.com/go-kit/kit/log"
	"github.com/go-logfmt/logfmt"
)

type logfmtEncoder struct {
	*logfmt.Encoder
	buf bytes.Buffer
}

func (l *logfmtEncoder) Reset() {
	l.Encoder.Reset()
	l.buf.Reset()
}

var logfmtEncoderPool = sync.Pool{
	New: func() interface{} {
		var enc logfmtEncoder
		enc.Encoder = logfmt.NewEncoder(&enc.buf)
		return &enc
	},
}

type stackdriverLogger struct {
	gcplogger    *gcplogging.Logger
	logfmtLogger log.Logger
}

// NewStackdriverLogger creates a Go kit log.Logger and adapts the log.Level to a GCP logging.Severity
func NewStackdriverLogger(client *gcplogging.Client, logID string) log.Logger {
	return &stackdriverLogger{
		gcplogger: client.Logger(logID),
	}
}

// Log satisfies the Go kit log.Logger interface
// It translates the keyvals into a GCP logging.Entry setting the
// correct logging.Severity from the keyvals and defaults to Default severity
func (s *stackdriverLogger) Log(keyvals ...interface{}) error {
	enc := logfmtEncoderPool.Get().(*logfmtEncoder)
	enc.Reset()
	defer logfmtEncoderPool.Put(enc)

	e := gcplogging.Entry{}

	if len(keyvals) >= 2 {
		lvl := fmt.Sprintf("%s", keyvals[1])

		switch lvl {
		case level.DebugValue().String():
			e.Severity = gcplogging.Debug
		case level.InfoValue().String():
			e.Severity = gcplogging.Info
		case level.WarnValue().String():
			e.Severity = gcplogging.Warning
		case level.ErrorValue().String():
			e.Severity = gcplogging.Error
		}
	}

	if e.Severity == 0 {
		if err := enc.EncodeKeyvals(keyvals...); err != nil {
			return err
		}
	} else {
		if err := enc.EncodeKeyvals(keyvals[2:]...); err != nil {
			return err
		}
	}

	if err := enc.EndRecord(); err != nil {
		return err
	}

	e.Payload = enc.buf.String()

	s.gcplogger.Log(e)

	return nil
}

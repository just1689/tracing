package tracing

import (
	"encoding/hex"
	"math/rand"
	"time"
)

func NewSpan(traceID string, serviceName, rpcName string, d time.Duration) Span {
	result := Span{
		ID:     newId(),
		Name:   rpcName,
		Shared: true,
		LocalEndpoint: map[string]string{
			"serviceName": serviceName,
		},
		TraceID: traceID,
	}
	result.SetDuration(d)
	result.SetTimestampNow()
	return result
}

func NewSpanChild(traceID string, parentID, serviceName, rpcName string, d time.Duration) Span {
	result := NewSpan(traceID, serviceName, rpcName, d)
	result.ParentID = parentID
	return result
}

type Span struct {
	Duration      int64             `json:"duration,omitempty"`
	ID            string            `json:"id"`
	Name          string            `json:"name,omitempty"`
	ParentID      string            `json:"parentId,omitempty"`
	Shared        bool              `json:"shared,omitempty"`
	LocalEndpoint map[string]string `json:"localEndpoint,omitempty"`
	Tags          Tags              `json:"tags,omitempty"`
	Timestamp     int64             `json:"timestamp,omitempty"`
	TraceID       string            `json:"traceId"`
}

func (s *Span) SetDuration(d time.Duration) {
	s.Duration = d.Nanoseconds() / 1000
}

func (s *Span) SetTimestamp(t time.Time) {
	s.Timestamp = t.UnixNano() / 1000
}

func (s *Span) SetTimestampNow() {
	s.SetTimestamp(time.Now())
}

type Tags map[string]string

const traceIDLen = 16

func NewId() string {
	return randString(traceIDLen)
}

var src = rand.New(rand.NewSource(time.Now().UnixNano()))

func randString(n int) string {
	b := make([]byte, (n+1)/2)
	if _, err := src.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)[:n]
}

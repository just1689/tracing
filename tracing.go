package tracing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

var GlobalPublisher *publisher

type Config struct {
	Url             string //http://localhost:9411/api/v2/spans
	CacheSize       int
	FlushTimeout    int
	FlushSize       int
	SleepBetweenErr int
	RetryErr        bool
}

func StartTracing(c Config) {
	p := &publisher{
		url:             c.Url,
		in:              make(chan Span, c.CacheSize),
		flushTimeout:    time.Duration(c.FlushTimeout) * time.Second,
		flushSize:       c.FlushSize,
		SleepBetweenErr: c.SleepBetweenErr,
		RetryErr:        c.RetryErr,
	}
	GlobalPublisher = p
	go p.run()
	return
}

type publisher struct {
	url             string
	in              chan Span
	flushTimeout    time.Duration
	flushSize       int
	SleepBetweenErr int
	RetryErr        bool
}

func (p *publisher) Enqueue(s Span) {
	p.in <- s
}
func (p *publisher) EnqueueAll(all []Span) {
	for _, s := range all {
		p.in <- s
	}
}
func (p *publisher) run() {
	var arr = make([]Span, 0)
	lastSend := time.Now()
	var readyToSend bool
	for {
		select {
		case s := <-p.in:
			arr = append(arr, s)
			readyToSend = len(arr) >= p.flushSize || time.Since(lastSend) >= p.flushTimeout
		case <-time.After(p.flushTimeout):
			readyToSend = len(arr) > 0
		}

		if readyToSend {
			p.sendNow(arr)
			arr = make([]Span, 0)
			lastSend = time.Now()
		}
	}
}

func (p *publisher) sendNow(s []Span) {
	b, err := json.Marshal(s)
	if err != nil {
		logrus.Errorln(err)
		return
	}
	r := bytes.NewReader(b)
	resp, err := http.Post(p.url, "application/json", r)
	if err != nil || resp.StatusCode < 200 || resp.StatusCode > 299 {
		logrus.Errorln(resp.StatusCode, err)
		b, err = ioutil.ReadAll(resp.Body)
		fmt.Println(string(b))
		if p.RetryErr {
			p.EnqueueAll(s)
		}
		if p.SleepBetweenErr != 0 {
			time.Sleep(time.Duration(p.SleepBetweenErr) * time.Second)
		}
		return
	}

}

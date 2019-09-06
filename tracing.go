package tracing

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"time"
)

var GlobalPublisher *publisher

func NewDefaultConfig() Config {
	return Config{
		Url:             "http://localhost:9411/api/v2/spans",
		CacheSize:       1024,
		FlushTimeout:    5,
		FlushSize:       25,
		SleepBetweenErr: 1,
	}
}

type Config struct {
	Url             string
	CacheSize       int
	FlushTimeout    int
	FlushSize       int
	SleepBetweenErr int
}

func StartTracing(c Config) {
	p := &publisher{
		url:             c.Url,
		in:              make(chan Span, c.CacheSize),
		collected:       make(chan []Span),
		marshaled:       make(chan []byte),
		flushTimeout:    time.Duration(c.FlushTimeout) * time.Second,
		flushSize:       c.FlushSize,
		SleepBetweenErr: c.SleepBetweenErr,
	}
	GlobalPublisher = p
	go p.collect()
	go p.marshal()
	go p.send()
	return
}

type publisher struct {
	url             string
	in              chan Span
	collected       chan []Span
	marshaled       chan []byte
	flushTimeout    time.Duration
	flushSize       int
	SleepBetweenErr int
}

func (p *publisher) Enqueue(s Span) {
	p.in <- s
}
func (p *publisher) EnqueueAll(all []Span) {
	for _, s := range all {
		p.in <- s
	}
}
func (p *publisher) collect() {
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
			p.collected <- arr
			arr = make([]Span, 0)
			lastSend = time.Now()
		}
	}
}

func (p *publisher) marshal() {
	var err error
	var b []byte
	for s := range p.collected {
		b, err = json.Marshal(s)
		if err != nil {
			logrus.Errorln(err)
			continue
		}
		p.marshaled <- b
	}
}

func (p *publisher) send() {
	var b []byte
	var err error
	for b = range p.marshaled {
		r := bytes.NewReader(b)
		err = httpPostJson(p.url, r)
		if err != nil {
			if p.SleepBetweenErr != 0 {
				time.Sleep(time.Duration(p.SleepBetweenErr) * time.Second)
			}
		}

	}
}

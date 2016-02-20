package main

import (
	"encoding/json"
	"net"
	"sync/atomic"

	log "github.com/Sirupsen/logrus"
	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/gravitational/trace"
	"github.com/jonboulle/clockwork"
)

const (
	// ELKType is ELK-required type field
	ELKType = "type"
	// ELKTimestamp is a timestamp required by ELK
	ELKTimestamp = "@timestamp"
	// ELKEntry is an object with events set by logrus
	ELKEntry = "entry"
	// ELKMessage contains user friendly message
	ELKMessage = "message"
	// ELKTrace is an object with trace data
	ELKTrace = "trace"
	// ELKBeatName is a beat name for ELKHook
	ELKBeatName = "trace"
	// ELKBeatVersion is a current version
	ELKBeatVersion = "0.0.1"
)

// ELKOptionSetter represents functional arguments passed to ELKHook
type ELKOptionSetter func(f *ELK)

// NewELK returns logrus-compatible hook that sends data to ELK
func NewELK(opts ...ELKOptionSetter) *ELK {
	f := &ELK{
		closeC: make(chan bool),
		SetupC: make(chan bool),
	}
	for _, o := range opts {
		o(f)
	}
	if f.Clock == nil {
		f.Clock = clockwork.NewRealClock()
	}
	return f
}

// ELK implements Elasticsearch-compatible logrus hook
// ELK's template is stored in template.json
// To initialize this template, use:
// curl -XPUT 'http://localhost:9200/_template/trace' -d@template.json
type ELK struct {
	Clock  clockwork.Clock
	beat   *beat.Beat
	closed uint32
	setup  uint32
	closeC chan bool
	SetupC chan bool
}

// Config configures ELKHook beat parameters
// read more about beats here: https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html#config-method
func (elk *ELK) Config(b *beat.Beat) error {
	return nil
}

// Setup initializes beat object
func (elk *ELK) Setup(b *beat.Beat) error {
	elk.beat = b
	if atomic.CompareAndSwapUint32(&elk.setup, 0, 1) {
		close(elk.SetupC)
	}
	return nil
}

// Run is called by ELK
func (elk *ELK) Run(b *beat.Beat) error {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:5000")
	l, err := net.ListenUDP(addr.Network(), addr)
	if err != nil {
		return trace.Wrap(err)
	}
	buf := make([]byte, 1024)
	var f trace.Frame
	for {
		count, _, err := l.ReadFrom(buf)
		if err != nil {
			e, ok := err.(net.Error)
			if ok && e.Timeout() {
				continue
			}
			return trace.Wrap(err)
		}
		err = json.Unmarshal(buf[:count], &f)
		if err != nil {
			log.Infof("failed to decode frame: %v", err)
		} else {
			elk.beat.Events.PublishEvent(common.MapStr{
				ELKTimestamp: common.Time(f.Time),
				ELKType:      f.Type,
				ELKEntry:     f.Entry,
				ELKMessage: map[string]string{
					ELKMessage:       f.Message,
					trace.LevelField: f.Level,
				},
			})
		}
		select {
		case <-elk.closeC:
			return nil
		default:
		}
	}
}

// Cleanup is a callback to cleanup any allocated resources
func (elk *ELK) Cleanup(b *beat.Beat) error {
	return nil
}

// Stop is a callback to stop everything
func (elk *ELK) Stop() {
	if atomic.CompareAndSwapUint32(&elk.closed, 0, 1) {
		close(elk.closeC)
	}
}

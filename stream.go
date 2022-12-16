package nevent

import (
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
)

type Stream struct {
	jet nats.JetStreamContext
}

func NewStream(nc *nats.Conn) (*Stream, error) {
	jet, err := nc.JetStream()
	if err != nil {
		return nil, err
	}
	return &Stream{jet: jet}, nil
}

// EnsureStream provide single subject for purperse
func (it *Stream) EnsureStream(s string, opts ...StreamOption) (*nats.StreamInfo, error) {
	o := it.GetOptions(opts...)
	subject := o.subjectTransformer(s)
	name := o.subjectNormalizer(subject)
	cfg := *o.config
	cfg.Name = name
	cfg.Subjects = []string{subject}

	si, err := it.jet.AddStream(&cfg)
	if err != nil && err.Error() == "stream name already in use" {
		si, err = it.jet.UpdateStream(&cfg)
		if err != nil {
			if strings.HasPrefix(err.Error(), "stream configuration update can not change") {
				if o.force {
					err = it.jet.DeleteStream(cfg.Name)
					if err != nil {
						return nil, fmt.Errorf("delete stream error when force update stream: %w", err)
					}
					si, err = it.jet.AddStream(&cfg)
					if err != nil {
						return nil, fmt.Errorf("add stream error when force update stream: %w", err)
					}
				} else {
					return nil, fmt.Errorf("try use force options(data loss) or do stream migration: %w", err)
				}
			} else {
				return nil, fmt.Errorf("update stream error: %w", err)
			}
		}
	}
	return si, err
}

func (it *Stream) GetOptions(opts ...StreamOption) *streamOptions {
	o := &streamOptions{}
	for _, opt := range opts {
		opt.apply(o)
	}
	if o.config == nil {
		o.config = new(nats.StreamConfig)
	}
	if o.subjectTransformer == nil {
		o.subjectTransformer = DefaultSubjectTransformer()
	}
	if o.subjectNormalizer == nil {
		o.subjectNormalizer = DefaultSubjectNormalize
	}
	return o
}

func StreamConfig(config *nats.StreamConfig) StreamOption {
	return newFuncStreamOption(func(o *streamOptions) {
		o.config = config
	})
}

// you will lost data in stream
func StreamForceUpdate() StreamOption {
	return newFuncStreamOption(func(o *streamOptions) {
		o.force = true
	})
}

func StreamSubjectTransformer(subjectTransformer SubjectTransformer) StreamOption {
	return newFuncStreamOption(func(o *streamOptions) {
		o.subjectTransformer = subjectTransformer
	})
}

func StreamSubjectNormalizer(subjectNormalizer SubjectNormalizer) StreamOption {
	return newFuncStreamOption(func(o *streamOptions) {
		o.subjectNormalizer = subjectNormalizer
	})
}

type streamOptions struct {
	config             *nats.StreamConfig
	subjectTransformer SubjectTransformer
	subjectNormalizer  SubjectNormalizer
	force              bool
}

type StreamOption interface {
	apply(*streamOptions)
}

type funcStreamOption struct {
	f func(*streamOptions)
}

func (it *funcStreamOption) apply(o *streamOptions) {
	it.f(o)
}

func newFuncStreamOption(f func(*streamOptions)) StreamOption {
	return &funcStreamOption{f: f}
}

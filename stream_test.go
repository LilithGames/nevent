package nevent

import (
	"fmt"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

func TestStream(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	assert.Nil(t, err)
	str, err := NewStream(nc)
	assert.Nil(t, err)
	si, err := str.EnsureStream("test1", StreamForceUpdate())
	assert.Nil(t, err)
	fmt.Printf("%+v\n", si)
	si, err = str.EnsureStream("test1", StreamConfig(&nats.StreamConfig{
		MaxConsumers: 10,
	}), StreamForceUpdate())
	assert.Nil(t, err)
	fmt.Printf("%+v\n", si)
}

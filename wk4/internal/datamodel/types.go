package datamodel

import (
	"fmt"
	"time"

	"github.com/google/wire"
	uuid "github.com/satori/go.uuid"
)

const (
	ProvisionStart = "start"
)

type ProvisionState struct {
	startTime time.Time
	duration  time.Duration
	state     string
}

type Cluster struct {
	name  string
	uuid  uuid.UUID
	state *ProvisionState
	event *Event
}

type Event struct {
	time    time.Time
	message string
}

func NewEvent(msg string) *Event {
	return &Event{
		time:    time.Now(),
		message: msg,
	}
}

func NewProvisionState() *ProvisionState {
	return &ProvisionState{
		startTime: time.Now(),
		duration:  time.Duration(0),
		state:     ProvisionStart,
	}
}

func NewCluster(name string, state *ProvisionState, event *Event) *Cluster {
	return &Cluster{
		name:  name,
		uuid:  uuid.NewV4(),
		state: state,
		event: event,
	}
}

func (c *Cluster) Start() {
	fmt.Printf("The cluster '%s' with id '%s' is started", c.name, c.uuid)
	return
}

var InitializeAll = wire.NewSet(NewEvent, NewProvisionState, NewCluster)

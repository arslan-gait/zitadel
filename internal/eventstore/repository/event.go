package repository

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/zitadel/zitadel/internal/eventstore"
)

var _ eventstore.Event = (*Event)(nil)

// Event represents all information about a manipulation of an aggregate
type Event struct {
	//ID is a generated uuid for this event
	ID string

	// Seq is the sequence of the event
	Seq uint64

	//PreviousAggregateSequence is the sequence of the previous sequence of the aggregate (e.g. org.250989)
	// if it's 0 then it's the first event of this aggregate
	PreviousAggregateSequence uint64

	//PreviousAggregateTypeSequence is the sequence of the previous sequence of the aggregate root (e.g. org)
	// the first event of the aggregate has previous aggregate root sequence 0
	PreviousAggregateTypeSequence uint64

	//CreationDate is the time the event is created
	// it's used for human readability.
	// Don't use it for event ordering,
	// time drifts in different services could cause integrity problems
	CreationDate time.Time

	// Typ describes the cause of the event (e.g. user.added)
	// it should always be in past-form
	Typ eventstore.EventType

	//Data describe the changed fields (e.g. userName = "hodor")
	// data must always a pointer to a struct, a struct or a byte array containing json bytes
	Data []byte

	//EditorService should be a unique identifier for the service which created the event
	// it's meant for maintainability
	EditorService string
	//EditorUser should be a unique identifier for the user which created the event
	// it's meant for maintainability.
	// It's recommend to use the aggregate id of the user
	EditorUser string

	//Version describes the definition of the aggregate at a certain point in time
	// it's used in read models to reduce the events in the correct definition
	Version eventstore.Version
	//AggregateID id is the unique identifier of the aggregate
	// the client must generate it by it's own
	AggregateID string
	//AggregateType describes the meaning of the aggregate for this event
	// it could an object like user
	AggregateType eventstore.AggregateType
	//ResourceOwner is the organisation which owns this aggregate
	// an aggregate can only be managed by one organisation
	// use the ID of the org
	ResourceOwner sql.NullString
	//InstanceID is the instance where this event belongs to
	// use the ID of the instance
	InstanceID string
}

// Aggregate implements [eventstore.Event]
func (e *Event) Aggregate() *eventstore.Aggregate {
	return &eventstore.Aggregate{
		ID:            e.AggregateID,
		Type:          e.AggregateType,
		ResourceOwner: e.ResourceOwner.String,
		InstanceID:    e.InstanceID,
		Version:       eventstore.Version(e.Version),
	}
}

// Creator implements [eventstore.Event]
func (e *Event) Creator() string {
	return e.EditorUser
}

// Type implements [eventstore.Event]
func (e *Event) Type() eventstore.EventType {
	return e.Typ
}

// Revision implements [eventstore.Event]
func (e *Event) Revision() uint16 {
	return 0
}

// Sequence implements [eventstore.Event]
func (e *Event) Sequence() uint64 {
	return e.Seq
}

// CreatedAt implements [eventstore.Event]
func (e *Event) CreatedAt() time.Time {
	return e.CreationDate
}

// Unmarshal implements [eventstore.Event]
func (e *Event) Unmarshal(ptr any) error {
	return json.Unmarshal(e.Data, ptr)
}

// DataAsBytes implements [eventstore.Event]
func (e *Event) DataAsBytes() []byte {
	return e.Data
}

func (e *Event) Payload() any {
	return e.Data
}

func (e *Event) UniqueConstraints() []*eventstore.UniqueConstraint { return nil }

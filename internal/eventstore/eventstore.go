package eventstore

import (
	"context"
	"sort"
	"sync"
	"time"
)

// Eventstore abstracts all functions needed to store valid events
// and filters the stored events
type Eventstore struct {
	interceptorMutex  sync.Mutex
	eventInterceptors map[EventType]eventTypeInterceptors
	eventTypes        []string
	aggregateTypes    []string
	PushTimeout       time.Duration

	pusher  Pusher
	querier Querier
}

type eventTypeInterceptors struct {
	eventMapper func(Event) (Event, error)
}

func NewEventstore(config *Config) *Eventstore {
	return &Eventstore{
		eventInterceptors: map[EventType]eventTypeInterceptors{},
		interceptorMutex:  sync.Mutex{},
		PushTimeout:       config.PushTimeout,

		pusher:  config.Pusher,
		querier: config.Querier,
	}
}

// Health checks if the eventstore can properly work
// It checks if the repository can serve load
func (es *Eventstore) Health(ctx context.Context) error {
	if err := es.pusher.Health(ctx); err != nil {
		return err
	}
	return es.querier.Health(ctx)
}

// Push pushes the events in a single transaction
// an event needs at least an aggregate
func (es *Eventstore) Push(ctx context.Context, cmds ...Command) ([]Event, error) {
	if es.PushTimeout > 0 {
		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, es.PushTimeout)
		defer cancel()
	}
	events, err := es.pusher.Push(ctx, cmds...)
	if err != nil {
		return nil, err
	}

	es.notify(events)
	return es.mapEvents(events)
}

// TODO(adlerhurst): still needed?
func (es *Eventstore) NewInstance(ctx context.Context, instanceID string) error {
	return es.querier.CreateInstance(ctx, instanceID)
}

func (es *Eventstore) EventTypes() []string {
	return es.eventTypes
}

func (es *Eventstore) AggregateTypes() []string {
	return es.aggregateTypes
}

// Filter filters the stored events based on the searchQuery
// and maps the events to the defined event structs
func (es *Eventstore) Filter(ctx context.Context, queryFactory *SearchQueryBuilder) ([]Event, error) {
	events, err := es.querier.Filter(ctx, queryFactory)
	if err != nil {
		return nil, err
	}

	return es.mapEvents(events)
}

func (es *Eventstore) mapEvents(events []Event) (mappedEvents []Event, err error) {
	mappedEvents = make([]Event, len(events))

	es.interceptorMutex.Lock()
	defer es.interceptorMutex.Unlock()

	for i, event := range events {
		mappedEvents[i], err = es.mapEvent(event)
		if err != nil {
			return nil, err
		}
	}

	return mappedEvents, nil
}

func (es *Eventstore) mapEvent(event Event) (Event, error) {
	interceptors, ok := es.eventInterceptors[event.Type()]
	if !ok || interceptors.eventMapper == nil {
		return BaseEventFromRepo(event), nil
	}
	return interceptors.eventMapper(event)
}

type reducer interface {
	//Reduce handles the events of the internal events list
	// it only appends the newly added events
	Reduce() error
	//AppendEvents appends the passed events to an internal list of events
	AppendEvents(...Event)
}

// FilterToReducer filters the events based on the search query, appends all events to the reducer and calls it's reduce function
func (es *Eventstore) FilterToReducer(ctx context.Context, searchQuery *SearchQueryBuilder, r reducer) error {
	events, err := es.Filter(ctx, searchQuery)
	if err != nil {
		return err
	}

	r.AppendEvents(events...)

	return r.Reduce()
}

// LatestSequence filters the latest sequence for the given search query
func (es *Eventstore) LatestSequence(ctx context.Context, queryFactory *SearchQueryBuilder) (time.Time, error) {
	return es.querier.LatestSequence(ctx, queryFactory)
}

// InstanceIDs returns the instance ids found by the search query
func (es *Eventstore) InstanceIDs(ctx context.Context, queryFactory *SearchQueryBuilder) ([]string, error) {
	return es.querier.InstanceIDs(ctx, queryFactory)
}

type QueryReducer interface {
	reducer
	//Query returns the SearchQueryFactory for the events needed in reducer
	Query() *SearchQueryBuilder
}

// FilterToQueryReducer filters the events based on the search query of the query function,
// appends all events to the reducer and calls it's reduce function
func (es *Eventstore) FilterToQueryReducer(ctx context.Context, r QueryReducer) error {
	events, err := es.Filter(ctx, r.Query())
	if err != nil {
		return err
	}
	r.AppendEvents(events...)

	return r.Reduce()
}

// RegisterFilterEventMapper registers a function for mapping an eventstore event to an event
func (es *Eventstore) RegisterFilterEventMapper(aggregateType AggregateType, eventType EventType, mapper func(Event) (Event, error)) *Eventstore {
	if mapper == nil || eventType == "" {
		return es
	}
	es.interceptorMutex.Lock()
	defer es.interceptorMutex.Unlock()

	es.appendEventType(eventType)
	es.appendAggregateType(aggregateType)

	interceptor := es.eventInterceptors[eventType]
	interceptor.eventMapper = mapper
	es.eventInterceptors[eventType] = interceptor

	return es
}

type Querier interface {
	// Health checks if the connection to the storage is available
	Health(ctx context.Context) error
	// Filter returns all events matching the given search query
	Filter(ctx context.Context, searchQuery *SearchQueryBuilder) (events []Event, err error)
	// LatestSequence returns the latest sequence found by the search query
	LatestSequence(ctx context.Context, queryFactory *SearchQueryBuilder) (time.Time, error)
	// InstanceIDs returns the instance ids found by the search query
	InstanceIDs(ctx context.Context, queryFactory *SearchQueryBuilder) ([]string, error)
	// CreateInstance creates a new sequence for the given instance
	CreateInstance(ctx context.Context, instanceID string) error
}

type Pusher interface {
	// Health checks if the connection to the storage is available
	Health(ctx context.Context) error
	// Push stores the actions
	Push(ctx context.Context, commands ...Command) (_ []Event, err error)
}

func (es *Eventstore) appendEventType(typ EventType) {
	i := sort.SearchStrings(es.eventTypes, string(typ))
	if i < len(es.eventTypes) && es.eventTypes[i] == string(typ) {
		return
	}
	es.eventTypes = append(es.eventTypes[:i], append([]string{string(typ)}, es.eventTypes[i:]...)...)
}

func (es *Eventstore) appendAggregateType(typ AggregateType) {
	i := sort.SearchStrings(es.aggregateTypes, string(typ))
	if len(es.aggregateTypes) > i && es.aggregateTypes[i] == string(typ) {
		return
	}
	es.aggregateTypes = append(es.aggregateTypes[:i], append([]string{string(typ)}, es.aggregateTypes[i:]...)...)
}

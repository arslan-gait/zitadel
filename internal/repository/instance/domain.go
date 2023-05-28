package instance

import (
	"context"

	"github.com/zitadel/zitadel/internal/errors"
	"github.com/zitadel/zitadel/internal/eventstore"
)

const (
	UniqueInstanceDomain              = "instance_domain"
	domainEventPrefix                 = instanceEventTypePrefix + "domain."
	InstanceDomainAddedEventType      = domainEventPrefix + "added"
	InstanceDomainPrimarySetEventType = domainEventPrefix + "primary.set"
	InstanceDomainRemovedEventType    = domainEventPrefix + "removed"
)

func NewAddInstanceDomainUniqueConstraint(domain string) *eventstore.UniqueConstraint {
	return eventstore.NewAddGlobalUniqueConstraint(
		UniqueInstanceDomain,
		domain,
		"Errors.Instance.Domain.AlreadyExists")
}

func NewRemoveInstanceDomainUniqueConstraint(domain string) *eventstore.UniqueConstraint {
	return eventstore.NewRemoveGlobalUniqueConstraint(
		UniqueInstanceDomain,
		domain)
}

type DomainAddedEvent struct {
	eventstore.BaseEvent `json:"-"`

	Domain    string `json:"domain,omitempty"`
	Generated bool   `json:"generated,omitempty"`
}

func (e *DomainAddedEvent) Payload() interface{} {
	return e
}

func (e *DomainAddedEvent) UniqueConstraints() []*eventstore.UniqueConstraint {
	return []*eventstore.UniqueConstraint{NewAddInstanceDomainUniqueConstraint(e.Domain)}
}

func NewDomainAddedEvent(ctx context.Context, aggregate *eventstore.Aggregate, domain string, generated bool) *DomainAddedEvent {
	return &DomainAddedEvent{
		BaseEvent: *eventstore.NewBaseEventForPush(
			ctx,
			aggregate,
			InstanceDomainAddedEventType,
		),
		Domain:    domain,
		Generated: generated,
	}
}

func DomainAddedEventMapper(event eventstore.Event) (eventstore.Event, error) {
	domainAdded := &DomainAddedEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}
	err := event.Unmarshal(domainAdded)
	if err != nil {
		return nil, errors.ThrowInternal(err, "INSTANCE-3noij", "unable to unmarshal instance domain added")
	}

	return domainAdded, nil
}

type DomainPrimarySetEvent struct {
	eventstore.BaseEvent `json:"-"`

	Domain string `json:"domain,omitempty"`
}

func (e *DomainPrimarySetEvent) Payload() interface{} {
	return e
}

func (e *DomainPrimarySetEvent) UniqueConstraints() []*eventstore.UniqueConstraint {
	return nil
}

func NewDomainPrimarySetEvent(ctx context.Context, aggregate *eventstore.Aggregate, domain string) *DomainPrimarySetEvent {
	return &DomainPrimarySetEvent{
		BaseEvent: *eventstore.NewBaseEventForPush(
			ctx,
			aggregate,
			InstanceDomainPrimarySetEventType,
		),
		Domain: domain,
	}
}

func DomainPrimarySetEventMapper(event eventstore.Event) (eventstore.Event, error) {
	domainAdded := &DomainPrimarySetEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}
	err := event.Unmarshal(domainAdded)
	if err != nil {
		return nil, errors.ThrowInternal(err, "INSTANCE-29jöF", "unable to unmarshal instance domain added")
	}

	return domainAdded, nil
}

type DomainRemovedEvent struct {
	eventstore.BaseEvent `json:"-"`

	Domain string `json:"domain,omitempty"`
}

func (e *DomainRemovedEvent) Payload() interface{} {
	return e
}

func (e *DomainRemovedEvent) UniqueConstraints() []*eventstore.UniqueConstraint {
	return []*eventstore.UniqueConstraint{NewRemoveInstanceDomainUniqueConstraint(e.Domain)}
}

func NewDomainRemovedEvent(ctx context.Context, aggregate *eventstore.Aggregate, domain string) *DomainRemovedEvent {
	return &DomainRemovedEvent{
		BaseEvent: *eventstore.NewBaseEventForPush(
			ctx,
			aggregate,
			InstanceDomainRemovedEventType,
		),
		Domain: domain,
	}
}

func DomainRemovedEventMapper(event eventstore.Event) (eventstore.Event, error) {
	domainRemoved := &DomainRemovedEvent{
		BaseEvent: *eventstore.BaseEventFromRepo(event),
	}
	err := event.Unmarshal(domainRemoved)
	if err != nil {
		return nil, errors.ThrowInternal(err, "INSTANCE-BngB2", "unable to unmarshal instance domain removed")
	}

	return domainRemoved, nil
}

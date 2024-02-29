package event

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"image"
	"time"
)

const (
	FinishedOK          FinishReason = "ok"
	FinishedDied        FinishReason = "death"
	FinishedChicken     FinishReason = "chicken"
	FinishedMercChicken FinishReason = "merc chicken"
	FinishedError       FinishReason = "error"
)

type FinishReason string

type Event interface {
	Message() string
	Image() image.Image
	OccurredAt() time.Time
}

type BaseEvent struct {
	message    string
	image      image.Image
	occurredAt time.Time
}

func (b BaseEvent) Message() string {
	return b.message
}

func (b BaseEvent) Image() image.Image {
	return b.image
}

func (b BaseEvent) OccurredAt() time.Time {
	return b.occurredAt
}

func WithScreenshot(message string, img image.Image) BaseEvent {
	return BaseEvent{
		message:    message,
		image:      img,
		occurredAt: time.Now(),
	}
}

func Text(message string) BaseEvent {
	return BaseEvent{
		message:    message,
		occurredAt: time.Now(),
	}
}

type UsedPotionEvent struct {
	BaseEvent
	PotionType data.PotionType
	OnMerc     bool
}

func UsedPotion(be BaseEvent, pt data.PotionType, onMerc bool) UsedPotionEvent {
	return UsedPotionEvent{
		BaseEvent:  be,
		PotionType: pt,
		OnMerc:     onMerc,
	}
}

type GameCreatedEvent struct {
	BaseEvent
	Name     string
	Password string
}

func GameCreated(be BaseEvent, name string, password string) GameCreatedEvent {
	return GameCreatedEvent{
		BaseEvent: be,
		Name:      name,
		Password:  password,
	}
}

type GameFinishedEvent struct {
	BaseEvent
	Reason FinishReason
}

func GameFinished(be BaseEvent, reason FinishReason) GameFinishedEvent {
	return GameFinishedEvent{
		BaseEvent: be,
		Reason:    reason,
	}
}

type RunFinishedEvent struct {
	BaseEvent
	RunName string
	Reason  FinishReason
}

func RunFinished(be BaseEvent, runName string, reason FinishReason) RunFinishedEvent {
	return RunFinishedEvent{
		BaseEvent: be,
		RunName:   runName,
		Reason:    reason,
	}
}

type ItemStashedEvent struct {
	BaseEvent
	Item data.Item
}

func ItemStashed(be BaseEvent, item data.Item) ItemStashedEvent {
	return ItemStashedEvent{
		BaseEvent: be,
		Item:      item,
	}
}

type RunStartedEvent struct {
	BaseEvent
	RunName string
}

func RunStarted(be BaseEvent, runName string) RunStartedEvent {
	return RunStartedEvent{
		BaseEvent: be,
		RunName:   runName,
	}
}

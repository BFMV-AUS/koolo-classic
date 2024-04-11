package step

import (
	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/game"
	"time"
)

const (
	StatusNotStarted Status = "NotStarted"
	StatusInProgress Status = "InProgress"
	StatusCompleted  Status = "Completed"
)

type Status string
type Step interface {
	Status(game.Data, container.Container) Status
	Run(game.Data, container.Container) error
	Reset()
	LastRun() time.Time
}

type basicStep struct {
	status  Status
	lastRun time.Time
}

func newBasicStep() basicStep {
	return basicStep{
		status: StatusNotStarted,
	}
}

func (bs *basicStep) LastRun() time.Time {
	return bs.lastRun
}

func (bs *basicStep) tryTransitionStatus(to Status) Status {
	if bs.status == StatusCompleted {
		return StatusCompleted
	}
	if bs.status == StatusInProgress && to != StatusCompleted {
		return StatusInProgress
	}

	bs.status = to
	return to
}

func (bs *basicStep) Reset() {
	bs.status = StatusNotStarted
	bs.lastRun = time.Time{}
}

package step

import (
	"github.com/hectorgimenez/koolo/internal/container"
	"github.com/hectorgimenez/koolo/internal/game"
)

type SyncActionStep struct {
	basicStep
	action        func(game.Data) error
	statusChecker func(game.Data) Status
}

func SyncStep(action func(game.Data) error) *SyncActionStep {
	return &SyncActionStep{
		basicStep: newBasicStep(),
		action:    action,
	}
}

func SyncStepWithCheck(action func(game.Data) error, statusChecker func(game.Data) Status) *SyncActionStep {
	return &SyncActionStep{
		basicStep:     newBasicStep(),
		action:        action,
		statusChecker: statusChecker,
	}
}

func (s *SyncActionStep) Status(d game.Data, container container.Container) Status {
	if s.status == StatusCompleted {
		return StatusCompleted
	}

	if s.statusChecker != nil {
		s.tryTransitionStatus(s.statusChecker(d))
	}

	return s.status
}

func (s *SyncActionStep) Run(d game.Data, container container.Container) error {
	if s.statusChecker == nil {
		s.tryTransitionStatus(StatusCompleted)
	}

	err := s.action(d)
	return err
}

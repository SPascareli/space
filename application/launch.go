package application

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

type LaunchStatus int

const (
	LaunchStatusSucceeded = iota + 1
	LaunchStatusFailed
)

type SearchParams struct {
	NETGreaterThanEqual time.Time
	NETLessThanEqual    time.Time
	Status              LaunchStatus
}

type errInvalid string

func (e errInvalid) Error() string { return string(e) }
func (e errInvalid) Invalid() bool { return true }

type LaunchProvider interface {
	Get(context.Context, SearchParams) ([]Launch, error)
}

type LaunchService struct {
	launch LaunchProvider
}

func NewLaunchService(launch LaunchProvider) *LaunchService {
	return &LaunchService{launch}
}

// Today get all the launches that happened or are expected to launch today
func (s *LaunchService) Today(ctx context.Context) ([]Launch, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Duration(24) * time.Hour)
	launches, err := s.launch.Get(ctx, SearchParams{
		NETGreaterThanEqual: start,
		NETLessThanEqual:    end,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get todays launches")
	}
	return launches, nil
}

// GetFailed gets all launches that failed in between the given time range
func (s *LaunchService) GetFailed(ctx context.Context, start time.Time, end time.Time) ([]Launch, error) {
	if end.Before(start) {
		return nil, errInvalid("end date cannot be before start date")
	}
	launches, err := s.launch.Get(ctx, SearchParams{
		NETGreaterThanEqual: start,
		NETLessThanEqual:    end,
		Status:              LaunchStatusFailed,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to get failed launches")
	}
	return launches, nil
}

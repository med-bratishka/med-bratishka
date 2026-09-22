package time_manager

import "time"

//go:generate  mockgen -source time_manager.go -destination ../../mocks/pkg/time_manager/time_manager.go

type TimeManager interface {
	Now() time.Time
	MillisecondsToTime(m int64) time.Time
}

type timeManager struct {
	locale *time.Location
}

func New(locale int64) TimeManager {
	return &timeManager{
		locale: time.FixedZone("MSC", int(locale)*3600),
	}
}

func (s timeManager) Now() time.Time {
	return time.Now().In(s.locale)
}

func (s timeManager) MillisecondsToTime(m int64) time.Time {
	return time.UnixMilli(m).In(s.locale)
}

package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bytekai/docker-auto-backup/internal/clock"
	"github.com/bytekai/docker-auto-backup/internal/logger"
	"github.com/bytekai/docker-auto-backup/internal/provider"
	"github.com/bytekai/docker-auto-backup/internal/storage"
)

type Frequency string

const (
	Daily   Frequency = "daily"
	Weekly  Frequency = "weekly"
	Monthly Frequency = "monthly"
	Yearly  Frequency = "yearly"
)

type Config struct {
	Enabled        bool
	Frequency      Frequency
	Time           string
	DayOfWeek      int
	DayOfMonth     int
	DayOfYear      int
	TimeZone       *time.Location
	Provider       string
	Location       string
	LocationPath   string
	StorageConfig  *storage.StorageConfig
	ProviderConfig *provider.ProviderConfig
}

type Scheduler interface {
	Start() error
	Stop() context.Context
}

type scheduler struct {
	config    Config
	task      func()
	clock     clock.Clock
	running   bool
	stop      chan struct{}
	runningMu sync.Mutex
	jobWaiter sync.WaitGroup
	logger    *logger.Logger
	session   *logger.Session
	location  *time.Location
}

type Option func(*scheduler)

func WithClock(clk clock.Clock) Option {
	return func(s *scheduler) {
		s.clock = clk
	}
}

func WithLogger(l *logger.Logger) Option {
	return func(s *scheduler) {
		s.logger = l
	}
}

func parseTime(timeStr string) (hour, minute int, err error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid time format, expected HH:MM, got %s", timeStr)
	}

	hour, err = strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, 0, fmt.Errorf("invalid hour: %s", parts[0])
	}

	minute, err = strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("invalid minute: %s", parts[1])
	}

	return hour, minute, nil
}

func isValidDayOfMonth(year int, month time.Month, day int) bool {
	if day < 1 {
		return false
	}
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	return day <= lastDay
}

func New(config Config, task func(), opts ...Option) (Scheduler, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}

	if _, _, err := parseTime(config.Time); err != nil {
		return nil, err
	}

	location := time.Local
	if config.TimeZone != nil {
		location = config.TimeZone
	}

	switch config.Frequency {
	case Daily:
	case Weekly:
		if config.DayOfWeek < 0 || config.DayOfWeek > 6 {
			return nil, fmt.Errorf("day of week must be between 0 and 6")
		}
	case Monthly:
		if config.DayOfMonth == -1 {
			break
		}
		validForAnyMonth := false
		for month := time.January; month <= time.December; month++ {
			if isValidDayOfMonth(time.Now().Year(), month, config.DayOfMonth) {
				validForAnyMonth = true
				break
			}
		}
		if !validForAnyMonth {
			return nil, fmt.Errorf("day of month %d is not valid for any month", config.DayOfMonth)
		}
	case Yearly:
		if config.DayOfYear < -1 || config.DayOfYear == 0 || config.DayOfYear > 366 {
			return nil, fmt.Errorf("day of year must be between 1 and 366, or -1")
		}
	default:
		return nil, fmt.Errorf("invalid frequency: %s", config.Frequency)
	}

	s := &scheduler{
		config:   config,
		clock:    clock.New(),
		stop:     make(chan struct{}),
		task:     task,
		logger:   logger.New(logger.INFO),
		location: location,
	}

	for _, opt := range opts {
		opt(s)
	}

	s.session = s.logger.NewSession("[scheduler] ")
	return s, nil
}

func (s *scheduler) nextRun(now time.Time) time.Time {
	hour, minute, _ := parseTime(s.config.Time)
	now = now.In(s.location)
	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, s.location)

	if next.Before(now) {
		next = next.AddDate(0, 0, 1)
	}

	switch s.config.Frequency {
	case Daily:
	case Weekly:
		for next.Weekday() != time.Weekday(s.config.DayOfWeek) {
			next = next.AddDate(0, 0, 1)
		}

	case Monthly:
		if s.config.DayOfMonth == -1 {
			next = time.Date(next.Year(), next.Month()+1, 0, hour, minute, 0, 0, s.location)
		} else {
			for {
				if isValidDayOfMonth(next.Year(), next.Month(), s.config.DayOfMonth) {
					next = time.Date(next.Year(), next.Month(), s.config.DayOfMonth, hour, minute, 0, 0, s.location)
					if !next.Before(now) {
						break
					}
				}
				next = time.Date(next.Year(), next.Month()+1, 1, hour, minute, 0, 0, s.location)
			}
		}
	case Yearly:
		if s.config.DayOfYear == -1 {
			next = time.Date(next.Year(), 12, 31, hour, minute, 0, 0, s.location)
			if next.Before(now) {
				next = time.Date(next.Year()+1, 12, 31, hour, minute, 0, 0, s.location)
			}
		} else {
			next = time.Date(next.Year(), 1, 1, hour, minute, 0, 0, s.location).
				AddDate(0, 0, s.config.DayOfYear-1)
			if next.Before(now) {
				next = time.Date(next.Year()+1, 1, 1, hour, minute, 0, 0, s.location).
					AddDate(0, 0, s.config.DayOfYear-1)
			}
		}
	}

	return next
}

func (s *scheduler) Start() error {
	s.runningMu.Lock()
	defer s.runningMu.Unlock()

	if s.running {
		return fmt.Errorf("scheduler already running")
	}

	s.running = true
	s.session.Info("Starting scheduler")
	go s.run()
	return nil
}

func (s *scheduler) Stop() context.Context {
	s.session.Info("Stopping scheduler")
	s.runningMu.Lock()
	defer s.runningMu.Unlock()

	if s.running {
		s.stop <- struct{}{}
		s.running = false
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		s.jobWaiter.Wait()
		cancel()
	}()
	return ctx
}

func (s *scheduler) run() {
	for {
		now := s.clock.Now()
		next := s.nextRun(now)
		wait := next.Sub(now)

		s.session.Debug("Next run scheduled at: %v (waiting %v)", next, wait)

		timer := time.NewTimer(wait)
		select {
		case <-timer.C:
			s.session.Info("Executing scheduled task")
			s.jobWaiter.Add(1)
			go func() {
				defer s.jobWaiter.Done()
				s.task()
			}()
		case <-s.stop:
			timer.Stop()
			return
		}
	}
}

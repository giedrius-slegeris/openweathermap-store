package cron

import (
	"os"
	"time"

	"github.com/go-co-op/gocron"
)

// StartTaskAsync initialises and runs a task from given cron intervals
func StartTaskAsync(task func()) error {
	loc, err := time.LoadLocation(os.Getenv("TIMEZONE"))
	if err != nil {
		return err
	}

	s := gocron.NewScheduler(loc)
	if _, err := s.Cron(os.Getenv("SCHEDULER_CRON")).Do(task); err != nil {
		return err
	}

	s.StartAsync()
	return nil
}

package cron

import (
	"testing"
)

func noop() {}

func TestStartTaskAsyncValidSchedule(t *testing.T) {
	t.Setenv("TIMEZONE", "Europe/London")
	t.Setenv("SCHEDULER_CRON", "*/5 * * * *")

	if err := StartTaskAsync(noop); err != nil {
		t.Fatalf("StartTaskAsync() returned error: %v", err)
	}
}

func TestStartTaskAsyncInvalidTimezone(t *testing.T) {
	t.Setenv("TIMEZONE", "Not/ARealZone")
	t.Setenv("SCHEDULER_CRON", "*/5 * * * *")

	if err := StartTaskAsync(noop); err == nil {
		t.Fatal("StartTaskAsync() returned nil error for an unknown timezone")
	}
}

func TestStartTaskAsyncInvalidCron(t *testing.T) {
	for _, expr := range []string{
		"",              // empty
		"not-a-cron",    // gibberish
		"*/5 * * *",     // four fields
		"*/5 * * * * *", // six fields, this scheduler expects five
		"99 * * * *",    // minute out of range
	} {
		t.Run(expr, func(t *testing.T) {
			t.Setenv("TIMEZONE", "UTC")
			t.Setenv("SCHEDULER_CRON", expr)

			if err := StartTaskAsync(noop); err == nil {
				t.Errorf("StartTaskAsync() returned nil error for cron %q", expr)
			}
		})
	}
}

// An unset TIMEZONE is not rejected: time.LoadLocation("") resolves to UTC.
// Worth pinning down, since it means a missing TIMEZONE silently shifts the
// schedule to UTC rather than failing loudly at startup.
func TestStartTaskAsyncEmptyTimezoneDefaultsToUTC(t *testing.T) {
	t.Setenv("TIMEZONE", "")
	t.Setenv("SCHEDULER_CRON", "*/5 * * * *")

	if err := StartTaskAsync(noop); err != nil {
		t.Fatalf("StartTaskAsync() returned error for empty TIMEZONE: %v", err)
	}
}

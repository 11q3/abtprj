package utils

import "time"

func ParseDateRange(dateStr string) (time.Time, time.Time, error) {
	now := time.Now()
	if dateStr == "" {
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return start, start.Add(24 * time.Hour), nil
	}
	parsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return parsed, parsed.Add(24 * time.Hour), nil
}

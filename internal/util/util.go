package util

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func plural(n int64, one, few, many string) string {
	if n%10 == 1 && n%100 != 11 {
		return fmt.Sprintf("%d %s", n, one)
	} else if n%10 >= 2 && n%10 <= 4 && (n%100 < 10 || n%100 >= 20) {
		return fmt.Sprintf("%d %s", n, few)
	}
	return fmt.Sprintf("%d %s", n, many)
}

func FormatDuration(d time.Duration) string {
	seconds := int64(d.Seconds())
	if seconds < 0 {
		seconds = -seconds
	}

	var parts []string

	weeks := seconds / (7 * 24 * 3600)
	if weeks > 0 {
		parts = append(parts, plural(weeks, "неделю", "недели", "недель"))
	}
	seconds %= (7 * 24 * 3600)

	days := seconds / (24 * 3600)
	if days > 0 {
		parts = append(parts, plural(days, "день", "дня", "дней"))
	}
	seconds %= (24 * 3600)

	hours := seconds / 3600
	if hours > 0 {
		parts = append(parts, plural(hours, "час", "часа", "часов"))
	}
	seconds %= 3600

	minutes := seconds / 60
	if minutes > 0 {
		parts = append(parts, plural(minutes, "минуту", "минуты", "минут"))
	}
	seconds %= 60

	if seconds > 0 || len(parts) == 0 {
		parts = append(parts, plural(seconds, "секунда", "секунды", "секунд"))
	}

	return strings.Join(parts, " ")
}

func ParseCustomDuration(s string) (time.Duration, error) {
	re := regexp.MustCompile(`(?i)(\d+)([wdhms])`)
	matches := re.FindAllStringSubmatch(s, -1)
	if matches == nil {
		return 0, fmt.Errorf("неверный формат продолжительности")
	}
	var d time.Duration
	for _, match := range matches {
		value, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, err
		}
		unit := strings.ToLower(match[2])
		switch unit {
		case "w":
			d += time.Duration(value) * 7 * 24 * time.Hour
		case "d":
			d += time.Duration(value) * 24 * time.Hour
		case "h":
			d += time.Duration(value) * time.Hour
		case "m":
			d += time.Duration(value) * time.Minute
		case "s":
			d += time.Duration(value) * time.Second
		default:
			return 0, fmt.Errorf("неизвестная единица %s", unit)
		}
	}
	return d, nil
}

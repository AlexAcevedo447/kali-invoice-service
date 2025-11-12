package helpers

import "time"

func OrTime(t *time.Time, defaultVal time.Time) time.Time {
    if t != nil {
        return *t
    }
    return defaultVal
}
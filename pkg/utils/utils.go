package utils

import (
	"strconv"
	"strings"
	"time"
)

func Contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
func TimeSectoTime(timeSec string) (int, error) {
	return strconv.Atoi(timeSec[0:strings.LastIndex(timeSec, ".")])
}
func LocalTime(unixTime int) time.Time {
	return time.Unix(int64(unixTime), 0)
}

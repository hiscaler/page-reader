package pagereader

import (
	"log"
	"time"
)

type RetryableFunc func() error

func Retry(retryableFunc RetryableFunc, maxTimes int, logger *log.Logger) {
	if maxTimes <= 0 {
		maxTimes = 1
	}
	currentTimes := 1
	for {
		logger.Printf("Retry %d time", currentTimes)
		err := retryableFunc()
		if err == nil || currentTimes > maxTimes {
			break
		}
		if err != nil {
			logger.Printf("Retry func execute error: %s", err.Error())
		}
		time.Sleep(1 * time.Second)
		currentTimes++
	}
}

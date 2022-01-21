package pagereader

type Config struct {
	Timeout       int // timeout second, default 10 seconds
	MaxTimeout    int
	RetryTimes    int
	MaxRetryTimes int
}

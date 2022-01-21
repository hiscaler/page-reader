package pagereader

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"log"
	"time"
)

type ChromeDP struct {
	Context struct {
		Value           context.Context
		CancelFunctions []context.CancelFunc
	}
	Headers              network.Headers
	ExecAllocatorOptions []chromedp.ExecAllocatorOption
}

// Start Start ChromeDP
// Flags
// headless: true
// blink-settings: imagesEnabled=false
func (c *ChromeDP) Start(timeout int, logger log.Logger) *ChromeDP {
	cancelFunctions := make([]context.CancelFunc, 0)
	options := c.ExecAllocatorOptions
	if len(options) == 0 {
		options = chromedp.DefaultExecAllocatorOptions[:]
	} else {
		options = append(chromedp.DefaultExecAllocatorOptions[:], options...)
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), options...)
	cancelFunctions = append(cancelFunctions, cancel)

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(logger.Printf))
	cancelFunctions = append(cancelFunctions, cancel)

	// create a timeout
	taskCtx, cancel = context.WithTimeout(taskCtx, timeoutDuration(timeout))
	cancelFunctions = append(cancelFunctions, cancel)

	c.Context.Value = taskCtx
	c.Context.CancelFunctions = cancelFunctions
	return c
}

func (c *ChromeDP) Cancel() *ChromeDP {
	for _, cancel := range c.Context.CancelFunctions {
		cancel()
	}
	return c
}

func timeoutDuration(timeout int) time.Duration {
	duration, _ := time.ParseDuration(fmt.Sprintf("%ds", timeout))
	return duration
}

func (c ChromeDP) Run() error {
	err := chromedp.Run(c.Context.Value)
	if err != nil {
		fmt.Println(fmt.Sprintf("chromedp.Run(taskCtx) error: %s", err.Error()))
	}
	return err
}

func (c ChromeDP) RunWithTimeOut(ctx *context.Context, timeout int, tasks chromedp.Tasks) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		timeoutContext, cancel := context.WithTimeout(ctx, timeoutDuration(timeout))
		defer cancel()
		return tasks.Do(timeoutContext)
	}
}

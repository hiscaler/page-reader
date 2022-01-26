package pagereader

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"log"
	"time"
)

type ChromeDP struct {
	Headers              network.Headers
	ExecAllocatorOptions []chromedp.ExecAllocatorOption
}

// NewContext New ChromeDP context
// Flags
// headless: true
// blink-settings: imagesEnabled=false
func (c *ChromeDP) NewContext(timeout int, logger log.Logger) (context.Context, []context.CancelFunc) {
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

	return taskCtx, cancelFunctions
}

func timeoutDuration(timeout int) time.Duration {
	duration, _ := time.ParseDuration(fmt.Sprintf("%ds", timeout))
	return duration
}

func (c ChromeDP) RunWithTimeOut(ctx *context.Context, timeout int, tasks chromedp.Tasks) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		timeoutContext, cancel := context.WithTimeout(ctx, timeoutDuration(timeout))
		defer cancel()
		return tasks.Do(timeoutContext)
	}
}

func (c ChromeDP) Click(sel interface{}, opts ...chromedp.QueryOption) chromedp.QueryAction {
	return chromedp.QueryAfter(sel, func(ctx context.Context, execCtx runtime.ExecutionContextID, nodes ...*cdp.Node) error {
		if len(nodes) > 0 {
			return chromedp.MouseClickNode(nodes[0]).Do(ctx)
		}
		return nil
	}, opts...)
}

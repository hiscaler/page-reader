package pagereader

import (
	"context"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"log"
	"strings"
)

type PageReader struct {
	Config
	Logger     *log.Logger
	ChromeDP   ChromeDP
	URL        string
	htmlSource string
	Doc        *goquery.Document
}

func NewPageReader(timeout int, logger *log.Logger) *PageReader {
	if logger.Prefix() == "" {
		logger.SetPrefix("[ Page Reader ] ")
	}
	return &PageReader{
		Config: Config{
			Timeout:       timeout,
			MaxTimeout:    timeout * 2,
			RetryTimes:    1,
			MaxRetryTimes: 3,
		},
		Logger: logger,
	}
}

func (pr *PageReader) SetMaxTryTimes(times int) *PageReader {
	if times <= 0 {
		times = 1
	}
	pr.Config.MaxRetryTimes = times
	return pr
}

func (pr *PageReader) SetUrl(url string) *PageReader {
	pr.URL = url
	return pr
}

func (pr PageReader) HtmlSource() string {
	return pr.htmlSource
}

func (pr PageReader) Headers() network.Headers {
	headers := network.Headers{
		"accept-encoding":           "gzip, deflate, br",
		"accept-language":           "zh-CN,zh;q=0.9",
		"upgrade-insecure-requests": "1",
	}
	return headers
}

func (pr PageReader) PageSource(timeout int) (htmlSource string, err error) {
	pr.Logger.Printf("Time %d：%s", pr.RetryTimes, pr.URL)
	if timeout <= 0 || timeout > pr.Config.Timeout {
		timeout = pr.Config.Timeout
	}
	pr.ChromeDP.Start(timeout, *pr.Logger)
	defer func() {
		pr.ChromeDP.Cancel()
	}()
	ctx := pr.ChromeDP.Context.Value
	err = chromedp.Run(ctx, pr.ChromeDP.RunWithTimeOut(&ctx, timeout, chromedp.Tasks{
		network.Enable(),
		network.SetExtraHTTPHeaders(pr.Headers()),
		chromedp.Navigate(pr.URL),
		chromedp.ActionFunc(func(ctx context.Context) error {
			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			htmlSource, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
			if err != nil {
				return err
			}
			if htmlSource != "" {
				htmlSource = strings.TrimSpace(htmlSource)
			}
			pr.htmlSource = htmlSource
			if doc, e := goquery.NewDocumentFromReader(strings.NewReader(htmlSource)); e == nil {
				pr.Doc = doc
			}
			return err
		}),
	}))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			timeout += 10
			pr.Config.RetryTimes += 1
			if timeout <= pr.Config.MaxTimeout && pr.Config.RetryTimes <= pr.Config.MaxRetryTimes {
				pr.PageSource(timeout)
			}
		}
	}

	return
}

func (pr PageReader) Contains(s string) bool {
	html := pr.htmlSource
	if html == "" {
		return strings.Contains(html, s)
	}
	return false
}

func findBySelector(doc *goquery.Document, selector string) *goquery.Selection {
	var s *goquery.Selection
	if doc != nil {
		s = doc.Find(selector)
	}

	return s
}

func (pr PageReader) Text(selector string) string {
	return findBySelector(pr.Doc, selector).Text()
}

func (pr PageReader) Attr(selector, attrName string) (string, bool) {
	return findBySelector(pr.Doc, selector).Attr(attrName)
}

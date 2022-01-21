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
	Debug bool
	Config
	Logger     *log.Logger
	ChromeDP   ChromeDP
	URL        string
	Title      string
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

func (pr PageReader) Headers() network.Headers {
	headers := network.Headers{
		"accept-encoding":           "gzip, deflate, br",
		"accept-language":           "zh-CN,zh;q=0.9",
		"upgrade-insecure-requests": "1",
	}
	return headers
}

func (pr *PageReader) Reset() *PageReader {
	pr.htmlSource = ""
	pr.Title = ""
	pr.Doc = nil
	return pr
}

func (pr *PageReader) Open(url string, timeout int, extraTasks ...chromedp.Action) (htmlSource string, err error) {
	pr.URL = url
	pr.Logger.Printf("Time %d：%s", pr.RetryTimes, pr.URL)
	if timeout <= 0 || timeout > pr.Config.Timeout {
		timeout = pr.Config.Timeout
	}
	pr.ChromeDP.Start(pr.Config.Timeout, *pr.Logger)
	defer func() {
		pr.ChromeDP.Cancel()
	}()
	ctx := pr.ChromeDP.Context.Value
	var title string
	tasks := []chromedp.Action{
		network.Enable(),
		network.SetExtraHTTPHeaders(pr.Headers()),
		chromedp.Navigate(pr.URL),
	}
	if len(extraTasks) > 0 {
		tasks = append(tasks, extraTasks...)
	}
	tasks = append(tasks, []chromedp.Action{
		chromedp.Title(&title),
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
			return err
		}),
	}...)
	err = chromedp.Run(ctx, pr.ChromeDP.RunWithTimeOut(&ctx, timeout, tasks))
	if err != nil {
		pr.Reset()
		pr.Logger.Printf("Error: %s", err.Error())
		if errors.Is(err, context.DeadlineExceeded) {
			timeout += 10
			pr.Config.RetryTimes += 1
			if timeout <= pr.Config.MaxTimeout && pr.Config.RetryTimes <= pr.Config.MaxRetryTimes {
				pr.Open(url, timeout)
			}
		}
	} else {
		if title != "" {
			title = strings.TrimSpace(title)
		}
		pr.Title = title
		pr.Doc = nil
		if pr.htmlSource != "" {
			if doc, e := goquery.NewDocumentFromReader(strings.NewReader(htmlSource)); e == nil {
				pr.Doc = doc
			} else {
				pr.Logger.Printf("goQuery create document Error: %s", e.Error())
			}
		}
	}

	return
}

func (pr PageReader) HtmlSource() string {
	return pr.htmlSource
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

func (pr PageReader) Text(selector string, selectors ...string) (value string) {
	selectorValues := append([]string{selector}, selectors...)
	for _, selector := range selectorValues {
		if s := findBySelector(pr.Doc, selector); s != nil {
			value = s.Text()
			if value != "" {
				value = strings.TrimSpace(value)
			}
		}
		if pr.Debug {
			pr.Logger.Printf("选择器 %s 文本查询结果：%s", selector, value)
		}
		if value != "" {
			break
		}
	}

	return
}

func (pr PageReader) Attr(selector, attrName string) (value string, exists bool) {
	if s := findBySelector(pr.Doc, selector); s != nil {
		value, exists = s.Attr(attrName)
		if exists && value != "" {
			value = strings.TrimSpace(value)
		}
	}
	if pr.Debug {
		pr.Logger.Printf("选择器 %s 属性 %s 查询结果为：%s，属性 %s %v", selector, attrName, value, attrName, exists)
	}
	return
}

package pagereader

import (
	"context"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"time"
	// "github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"log"
	"strings"
)

type PageReader struct {
	Debug bool
	Config
	Logger   *log.Logger
	ChromeDP *ChromeDP
	URL      string
	Title    string
	html     string
	Doc      *goquery.Document
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
		Logger:   logger,
		ChromeDP: &ChromeDP{},
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
	pr.html = ""
	pr.Title = ""
	pr.Doc = nil
	return pr
}

func (pr PageReader) RunTasks(ctx context.Context, name string, timeout int, tasks []chromedp.Action) error {
	var err error
	if name == "" {
		name = "Unknown"
	}
	notify := NewNotify("RunTasks", name)
	if timeout == 0 {
		err = chromedp.Run(ctx, tasks...)
	} else {
		err = chromedp.Run(ctx, pr.ChromeDP.RunWithTimeOut(&ctx, timeout, tasks))
	}

	notify.Error = err
	pr.Logger.Print(notify.String())

	return err
}

// JQueryIsLoaded Check JQuery is loaded
func (pr PageReader) JQueryIsLoaded(ctx context.Context) (loaded bool) {
	checkJQueryLoadTasks := chromedp.Tasks{
		chromedp.EvaluateAsDevTools(`typeof jQuery === "function";`, &loaded),
	}
	pr.RunTasks(ctx, "JQueryIsLoaded", 1, checkJQueryLoadTasks)
	pr.Logger.Printf("JQuery loaded: %v", loaded)
	return
}

func (pr *PageReader) AddJQuery(ctx context.Context, timeout int) (loaded bool, err error) {
	ts := time.Now().Unix()
	if pr.JQueryIsLoaded(ctx) {
		return true, nil
	}

	if timeout < 6 {
		timeout = 6
	}
	err = pr.RunTasks(ctx, "AddJQuery", timeout, chromedp.Tasks{
		chromedp.EvaluateAsDevTools(`
var JQ = document.createElement('script');
JQ.src = "https://cdn.bootcss.com/jquery/1.4.2/jquery.js";
document.getElementsByTagName('head')[0].appendChild(JQ);
function sleep(delay) {
    const start = (new Date()).getTime();
    while ((new Date()).getTime() < (start + delay)) {
    }
    let seconds = ((new Date()).getTime() - start) / 1000;
    console.info("Sleep " + seconds + " Seconds");
}
`, nil),
		chromedp.Sleep(2 * time.Second),
	})
	if err != nil {
		pr.Logger.Printf("AddJQuery error: %s", err.Error())
	} else {
		for {
			loaded = pr.JQueryIsLoaded(ctx)
			if loaded || time.Now().Unix()-ts >= int64(timeout) {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
	}

	return
}

func (pr *PageReader) Open(ctx context.Context, url string, timeout int, extraTasks ...chromedp.Action) (html string, err error) {
	notify := NewNotify("Open", url)
	pr.Reset()
	pr.URL = url
	notify.AddLogf("#%d Open %s", pr.RetryTimes, pr.URL)
	if timeout <= 0 || timeout > pr.Config.Timeout {
		timeout = pr.Config.Timeout
	}

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
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
		// chromedp.ActionFunc(func(ctx context.Context) error {
		// 	node, err := dom.GetDocument().Do(ctx)
		// 	if err != nil {
		// 		pr.Logger.Printf("err1=====%s", err.Error())
		// 		return err
		// 	}
		//
		// 	htmlSource, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
		// 	if err != nil {
		// 		pr.Logger.Printf("err2=====%s", err.Error())
		// 		return err
		// 	}
		// 	if htmlSource != "" {
		// 		htmlSource = strings.TrimSpace(htmlSource)
		// 	}
		// 	pr.htmlSource = htmlSource
		// 	return err
		// }),
	}...)
	err = chromedp.Run(ctx, pr.ChromeDP.RunWithTimeOut(&ctx, timeout, tasks))
	pr.SetHtml(html)
	if err != nil {
		notify.AddLogf("Open faild, error: %s", err.Error())
		if errors.Is(err, context.DeadlineExceeded) {
			timeout += 10
			pr.Config.RetryTimes += 1
			if timeout <= pr.Config.MaxTimeout && pr.Config.RetryTimes <= pr.Config.MaxRetryTimes {
				pr.Open(ctx, url, timeout)
			}
		}
	} else {
		notify.AddLog("Open success")
		if title != "" {
			title = strings.TrimSpace(title)
		}
		notify.AddLogf("Title: %s", title)
		pr.Title = title
	}
	notify.Error = err
	pr.Logger.Print(notify.String())

	return
}

func (pr *PageReader) WaitReady(ctx context.Context, sel interface{}, opts ...chromedp.QueryOption) *PageReader {
	var html string
	tasks := chromedp.Tasks{
		chromedp.WaitReady(sel, opts...),
		chromedp.OuterHTML("html", &html, chromedp.ByQuery),
	}
	pr.RunTasks(ctx, "WaitReady", 0, tasks)
	pr.SetHtml(html)
	return pr
}

func (pr *PageReader) ObtainHtml(ctx context.Context) *PageReader {
	var html string
	tasks := chromedp.Tasks{chromedp.Sleep(2 * time.Second)}
	if pr.JQueryIsLoaded(ctx) {
		tasks = append(tasks, chromedp.EvaluateAsDevTools(`$("html").html();`, &html))
	} else {
		tasks = append(tasks, chromedp.OuterHTML("html", &html, chromedp.ByQuery))
	}
	err := pr.RunTasks(ctx, "ObtainHtml", 6, tasks)
	if err != nil {
		pr.Logger.Printf("ObtainHtml error: %s", err.Error())
	}
	pr.SetHtml(html)
	return pr
}

func (pr *PageReader) SetHtml(html string) *PageReader {
	if html == "" {
		pr.Logger.Printf("HTML is empty")
	}
	pr.Doc = nil
	pr.html = strings.TrimSpace(html)
	if pr.html != "" {
		if doc, e := goquery.NewDocumentFromReader(strings.NewReader(pr.html)); e == nil {
			pr.Doc = doc
		} else {
			pr.Logger.Printf("goQuery create document Error: %s", e.Error())
		}
	}
	return pr
}

func (pr PageReader) Html() string {
	return pr.html
}

func (pr PageReader) Contains(s string) bool {
	if pr.html != "" {
		return strings.Contains(strings.ToLower(pr.html), strings.ToLower(s))
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
	for _, sel := range selectorValues {
		if s := findBySelector(pr.Doc, sel); s != nil {
			value = s.Text()
			if value != "" {
				value = strings.TrimSpace(value)
			}
		}
		if pr.Debug {
			pr.Logger.Printf(`
  Selector: %s
Query Text: %s`, sel, value)
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
		pr.Logger.Printf(` 
    Selector: %s [ Attr: %s ]
Query Result: [ Exists: %v ] [ Value: %s ]"`, selector, attrName, exists, value)
	}
	return
}

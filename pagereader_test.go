package pagereader

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"testing"
)

var pageReader *PageReader
var ctx context.Context
var ctxCancelFunctions []context.CancelFunc

func init() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	pageReader = NewPageReader(40, logger)
	pageReader.Debug = true
	pageReader.ChromeDP.ExecAllocatorOptions = []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", false),
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
	}
	ctx, ctxCancelFunctions = pageReader.ChromeDP.NewContext(30, *logger)
}

func TestPageReader_PageSource(t *testing.T) {
	defer func() {
		for _, cancelFunc := range ctxCancelFunctions {
			cancelFunc()
		}
	}()
	_, err := pageReader.Open(ctx, "https://www.amazon.com/dp/B092M62439", 20)
	if err != nil {
		t.Errorf("error: %s", err.Error())
	} else {
		brandUrl, _ := pageReader.Attr("#bylineInfo", "href")
		fmt.Println(fmt.Sprintf(`
Title: %s
Product Name: %s
Brand URL: %s
`,
			pageReader.Title,
			pageReader.Text("#a", "#b", "#productTitle"),
			brandUrl,
		))
	}
}

func TestPageReader_Text(t *testing.T) {
	defer func() {
		for _, cancelFunc := range ctxCancelFunctions {
			cancelFunc()
		}
	}()
	_, err := pageReader.Open(ctx, "https://www.amazon.com/s?me=A21ML91ENNQT46&marketplaceID=ATVPDKIKX0DER", 20)
	if err != nil {
		t.Errorf("error: %s", err.Error())
	} else {
		text := pageReader.Text("#search > span > div > h1 > div > div.sg-col-14-of-20.sg-col.s-breadcrumb.sg-col-10-of-16.sg-col-6-of-12 > div > div > span", "#search > span")
		fmt.Println(fmt.Sprintf("Text: %s", text))
	}
}

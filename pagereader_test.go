package pagereader

import (
	"fmt"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"testing"
)

func TestAmazon_PageSource(t *testing.T) {
	logger := log.New(os.Stdout, "[ Amazon Fetcher ] ", log.LstdFlags)
	pageReader := NewPageReader(20, logger)
	pageReader.ChromeDP.ExecAllocatorOptions = []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", true),
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
	}
	pageReader.SetUrl("https://www.pageReader.com/dp/B092M62439")
	_, err := pageReader.PageSource(20)
	if err != nil {
		t.Errorf("error: %s", err.Error())
	} else {
		fmt.Println(pageReader.Text("title"))
	}
}

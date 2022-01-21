package pagereader

import (
	"fmt"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"testing"
)

func TestPageReader_PageSource(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	pageReader := NewPageReader(40, logger)
	pageReader.Debug = true
	pageReader.ChromeDP.ExecAllocatorOptions = []chromedp.ExecAllocatorOption{
		chromedp.Flag("headless", false),
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
	}
	_, err := pageReader.Open("https://www.amazon.com/dp/B092M62439", 20)
	if err != nil {
		t.Errorf("error: %s", err.Error())
	} else {
		fmt.Println(fmt.Sprintf(`
Title: %s
Product Name: %s
`, pageReader.Title, pageReader.Text("#productTitle")))
	}
}

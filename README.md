# Page Reader

Use ChromeDP read page content

## 使用方法
```go
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
```

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

var (
	uri           string
	width, height int
	userAgent     string
	output        string
	timeout       time.Duration
	wait          time.Duration
)

func main() {
	parseFlags()
	browser := rod.New().MustConnect()
	if err := getScreenshot(getUrl(uri, "https://"), browser); err != nil {
		log.Println("Screenshot for https failed, proceeding with http", err)
		if err := getScreenshot(getUrl(uri, "http://"), browser); err != nil {
			panic(fmt.Errorf("Screenshot capture failed %w", err))
		}
	}
	log.Println("Screenshot capture successful")
}

func parseFlags() {
	uriFlag := flag.String("uri", "undefined", "The URI")
	widthFlag := flag.Int("width", 1920, "Screenshot width")
	heightFlag := flag.Int("height", 1080, "Screenshot height")
	userAgentFlag := flag.String("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.111 Safari/537.36", "Fake User Agent for page load")
	timeoutFlag := flag.Duration("timeout", 2*time.Minute, "Timeout to load page and take screenshot")
	waitFlag := flag.Duration("wait", 0, "Extra wait time after page load")
	outputFlag := flag.String("output", "example.png", "Path to the output file")
	flag.Parse()
	if *uriFlag == "undefined" {
		flag.Usage()
		fmt.Println("\nERROR: URI must be defined")
		os.Exit(1)
	}
	uri = *uriFlag
	width, height = *widthFlag, *heightFlag
	userAgent = *userAgentFlag
	timeout = *timeoutFlag
	wait = *waitFlag
	output = *outputFlag
}

func getUrl(uri, prefix string) string {
	r := regexp.MustCompile("^http(s|)://")
	return prefix + r.ReplaceAllString(uri, "")
}

func getScreenshot(url string, browser *rod.Browser) (err error) {
	var page *rod.Page
	// recover error and close page
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(x)
			case error:
				err = x
			default:
				err = errors.New("Unknown panic")
			}
		}
		if page != nil {
			page.Close()
		}
	}()
	// Take Screenshot
	page = browser.MustPage("")
	withTimeout := page.Timeout(timeout)
	withTimeout.
		MustSetViewport(width, height, 1, false).
		MustSetUserAgent(&proto.NetworkSetUserAgentOverride{
			UserAgent: userAgent,
		}).
		MustNavigate(url).
		MustWaitLoad()
	if wait != 0 {
		page.MustEvaluate(createWaitFunc(wait))
	}
	page.MustScreenshot(output)
	return
}

func createWaitFunc(d time.Duration) *rod.EvalOptions {
	millis := d / time.Millisecond
	return &rod.EvalOptions{
		ByValue:      true,
		JS:           "new Promise(r => setTimeout(r, " + strconv.Itoa(int(millis)) + "));",
		AwaitPromise: true,
	}
}

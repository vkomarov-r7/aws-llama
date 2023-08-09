package browser

import (
	"aws-llama/config"
	"fmt"
	"net/url"
	"time"

	"github.com/playwright-community/playwright-go"
)

func ensurePage(context playwright.BrowserContext) (playwright.Page, error) {
	pages := context.Pages()
	if len(pages) > 0 {
		return pages[0], nil
	}

	page, err := context.NewPage()
	return page, err
}

func Authenticate(headless bool) error {
	runOptions := playwright.RunOptions{
		Browsers: []string{"chromium"},
	}
	err := playwright.Install(&runOptions)
	if err != nil {
		return err
	}

	pw, err := playwright.Run()
	if err != nil {
		return err
	}
	defer pw.Stop()

	loginPath, err := url.JoinPath(config.ROOT_URL_RAW, "/login")
	if err != nil {
		return err
	}

	opts := playwright.BrowserTypeLaunchPersistentContextOptions{
		Headless: &headless,
	}
	browserCtx, err := pw.Chromium.LaunchPersistentContext("./chromium-context", opts)
	if err != nil {
		return err
	}

	page, err := ensurePage(browserCtx)
	if err != nil {
		return err
	}
	response, err := page.Goto(loginPath)
	if err != nil {
		return err
	}
	fmt.Printf("Browser Response: %+v\n", response)

	time.Sleep(600 * time.Second)
	return nil
}

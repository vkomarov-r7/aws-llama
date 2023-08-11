package browser

import (
	"aws-llama/config"
	"fmt"
	"net/url"

	"github.com/playwright-community/playwright-go"
)

func ensurePage(context playwright.BrowserContext) (playwright.Page, error) {
	pages := context.Pages()
	if len(pages) > 0 {
		return pages[0], nil
	}
	fmt.Printf("Current number of pages: %d\n", len(pages))

	page, err := context.NewPage()
	return page, err
}

func Authenticate(headless bool) error {
	runOptions := playwright.RunOptions{
		SkipInstallBrowsers: true,
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

	loginPath, err := url.JoinPath(config.CurrentConfig.RootUrl.String(), "/login")
	if err != nil {
		return err
	}

	opts := playwright.BrowserTypeLaunchPersistentContextOptions{
		Headless: &headless,
		Channel:  playwright.String("chrome"),
	}

	browserCtx, err := pw.Chromium.LaunchPersistentContext(config.CurrentConfig.ChromeUserDataDir, opts)
	if err != nil {
		return err
	}

	page, err := ensurePage(browserCtx)
	if err != nil {
		return err
	}
	_, err = page.Goto(loginPath, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		return err
	}

	// page.WaitForTimeout(15 * 1000)
	fmt.Printf("Current URL: %s\n", page.URL())
	fmt.Printf("Waiting for response..\n")
	err = page.WaitForURL("http://localhost:2600/")
	if err != nil {
		return err
	}
	fmt.Printf("Finished waiting for response, closing stuff out..\n")
	page.Close()
	browserCtx.Close()

	// time.Sleep(600 * time.Second)
	fmt.Printf("Finished authing. Toodles!\n")
	return nil
}

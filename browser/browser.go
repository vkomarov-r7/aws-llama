package browser

import (
	"aws-llama/config"
	"aws-llama/credentials"
	"aws-llama/log"
	"fmt"
	"net/url"
	"time"

	"github.com/playwright-community/playwright-go"
)

func ensurePage(context playwright.BrowserContext) (playwright.Page, error) {
	pages := context.Pages()
	fmt.Printf("Current number of pages: %d\n", len(pages))

	// Close any other pages that might exist in the current BrowserContext.
	// We do this to avoid confusion in case there are other tabs that had
	// a login page or something.
	if len(pages) > 1 {
		for _, page := range pages[1:] {
			page.Close()
		}
	}

	if len(pages) > 0 {
		return pages[0], nil
	}

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

	waitOpts := playwright.FrameWaitForURLOptions{Timeout: playwright.Float(5 * 60 * 1000)}
	err = page.WaitForURL("http://localhost:2600/", waitOpts)
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

func AuthenticationLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	for {
		<-ticker.C

		metadataURL := credentials.NextMetadataURLForRefresh()
		if metadataURL != "" {
			log.Logger.Info("Eligible to authenticate a metadata url. Opening browser")
			err := Authenticate(false)
			if err != nil {
				log.Logger.Info("There was an error during authentication", "error", err.Error())
			}
		}
	}
}

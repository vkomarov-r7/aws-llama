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

type Browser struct {
	playwright     *playwright.Playwright
	browserContext playwright.BrowserContext
	loginPath      string
	// page           *playwright.Page
}

func NewBrowser() (*Browser, error) {
	pw, err := createPlaywright()
	if err != nil {
		return nil, err
	}

	loginPath, err := url.JoinPath(config.CurrentConfig.RootUrl.String(), "/login")
	if err != nil {
		return nil, err
	}

	browser := Browser{
		playwright: pw,
		loginPath:  loginPath,
	}
	return &browser, nil
}

func (b *Browser) Authenticate() error {
	authenticated, err := b.authenticate(true)
	if err != nil {
		return err
	}

	if authenticated {
		log.Logger.Debug("Authentication via headless mode successful.")
		return nil
	}

	log.Logger.Debug("Headless auth didn't work. Attemping Browser auth..")
	authenticated, err = b.authenticate(false)
	if err != nil {
		return err
	}

	if authenticated {
		log.Logger.Debug("Authentication mode via browser successful.")
		return nil
	}

	return fmt.Errorf("failed to authenticate via both headed and headless browsers")
}

func (b *Browser) Close() error {
	err := b.browserContext.Close()
	if err != nil {
		return err
	}
	err = b.playwright.Stop()
	if err != nil {
		return err
	}

	return nil
}

func (b *Browser) authenticate(headless bool) (bool, error) {
	err := b.ensureBrowserContext(headless)
	if err != nil {
		return false, err
	}

	page, err := b.ensurePage()
	if err != nil {
		return false, err
	}

	// TODO: Error checking here..?
	defer page.Close()

	_, err = page.Goto(b.loginPath, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded})
	if err != nil {
		return false, err
	}

	err = page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{State: playwright.LoadStateNetworkidle})
	if err != nil {
		return false, err
	}

	url := page.URL()
	log.Logger.Debug("Current browser URL:", page.URL())
	// If we got redirected back to the home page, auth is completed.
	if url == "http://localhost:2600/" {
		return true, nil
	}

	log.Logger.Info("Authentication ended up at url (likely needs user input):", url)

	// Wait for 5 minutes for user input if a window is displayed.
	if !headless {
		log.Logger.Info("Waiting for 5 minutes for user input (headed browser mode)...")
		waitOpts := playwright.FrameWaitForURLOptions{Timeout: playwright.Float(5 * 60 * 1000)}
		err := page.WaitForURL("http://localhost:2600/", waitOpts)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func createPlaywright() (*playwright.Playwright, error) {
	runOptions := playwright.RunOptions{
		SkipInstallBrowsers: true,
	}
	err := playwright.Install(&runOptions)
	if err != nil {
		return nil, err
	}

	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}

	return pw, nil
}

func (b *Browser) ensureBrowserContext(headless bool) error {
	if b.browserContext != nil {
		err := b.browserContext.Close()
		if err != nil {
			return err
		}
	}

	opts := playwright.BrowserTypeLaunchPersistentContextOptions{
		Headless: &headless,
		Channel:  playwright.String("chrome"),
	}

	browserCtx, err := b.playwright.Chromium.LaunchPersistentContext(config.CurrentConfig.ChromeUserDataDir, opts)
	if err != nil {
		return err
	}

	b.browserContext = browserCtx
	return nil
}

func (b *Browser) ensurePage() (playwright.Page, error) {
	pages := b.browserContext.Pages()
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

	page, err := b.browserContext.NewPage()
	return page, err
}

func AuthenticationLoop() {
	log.Logger.Debug("Starting browser auth loop.")
	ticker := time.NewTicker(5 * time.Minute)

	// Perform the initial tick.
	AttemptAuthentication()
	for {
		<-ticker.C
		AttemptAuthentication()
	}
}

func AttemptAuthentication() {
	metadataURL := credentials.NextMetadataURLForRefresh()
	if metadataURL != "" {
		log.Logger.Info("Eligible to authenticate a metadata url. Opening browser")
		b, err := NewBrowser()
		if err != nil {
			log.Logger.Errorf("Error in new browser creation:", err.Error())
		}

		err = b.Authenticate()
		if err != nil {
			log.Logger.Errorf("Error during authentication:", err.Error())
		}

		err = b.Close()
		if err != nil {
			log.Logger.Errorf("Error during browser closure", err.Error())
		}
	} else {
		log.Logger.Debug("No credentials need refreshing at this time.")
	}
}

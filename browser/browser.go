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
	browser        playwright.Browser
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
		return fmt.Errorf("failed to auth via headless mode: %w", err)
	}

	if authenticated {
		b.browserContext.StorageState(config.CurrentConfig.StorageStatePath)
		log.Logger.Debug("Authentication via headless mode successful.")
		return nil
	}

	log.Logger.Debug("Headless auth didn't work. Attemping Browser auth..")
	authenticated, err = b.authenticate(false)
	if err != nil {
		return fmt.Errorf("failed to auth via headed mode: %w", err)
	}

	if authenticated {
		b.browserContext.StorageState(config.CurrentConfig.StorageStatePath)
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
	err = b.browser.Close()
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

	_, err = page.Goto(b.loginPath, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateNetworkidle})
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

	log.Logger.Info("Authentication ended up at url (likely needs user input): ", url)

	// Wait for 5 minutes for user input if a window is displayed.
	if !headless {
		err = attemptAuth(page)
		if err != nil {
			log.Logger.Error("failed to automatically authenticate, reverting to manual mode: %w", err)
		}

		log.Logger.Info("Waiting for 5 minutes for user input (headed browser mode)...")
		waitOpts := playwright.PageWaitForURLOptions{Timeout: playwright.Float(5 * 60 * 1000)}
		err := page.WaitForURL("http://localhost:2600/", waitOpts)
		if err != nil {
			return false, fmt.Errorf("error while waiting for localhost url: %w", err)
		}
		return true, nil
	}
	return false, nil
}

func attemptAuth(page playwright.Page) error {
	if !config.CurrentConfig.HasLogin() {
		log.Logger.Info("Skipping auth attempt. Credentials not configured.")
		return nil
	}

	submitBtnLocator := page.Locator("input[value=\"Sign in\"]")
	submitButtonCount, err := submitBtnLocator.Count()
	if err != nil {
		return err
	}

	if submitButtonCount != 1 {
		log.Logger.Info("Didn't detect a login page. Skipping automated auth attempt.")
		return nil
	}

	// Fill in Username and Password
	usernameField := page.Locator("input[name=\"identifier\"]")
	err = usernameField.Fill(config.CurrentConfig.Username)
	if err != nil {
		return fmt.Errorf("failed to fill in username field: %w", err)
	}
	passwordField := page.Locator("input[name=\"credentials.passcode\"]")
	err = passwordField.Fill(config.CurrentConfig.Password)
	if err != nil {
		return fmt.Errorf("failed to fill in password field: %w", err)
	}

	// Check the "Keep me signed in" box for less refreshing later.
	rememberMeField := page.Locator("[name=\"rememberMe\"]")
	n, err := rememberMeField.Count()
	log.Logger.Info("Number of checkboxes found: %d %w", n, err)
	err = rememberMeField.Check(playwright.LocatorCheckOptions{Timeout: playwright.Float(5 * 1000), Force: playwright.Bool(true)})
	if err != nil {
		return fmt.Errorf("failed to check the 'Keep me signed in' checkbox: %w", err)
	}

	// Click the submit button
	err = submitBtnLocator.Click()
	if err != nil {
		return fmt.Errorf("failed to click the submit button: %w", err)
	}

	// Check if there's a "Verify" form and automatically click it.
	verifyBtnLocator := page.Locator("input[value=\"Verify\"]")
	err = verifyBtnLocator.Click()
	if err != nil {
		return fmt.Errorf("failed to click the 'Verify' button: %w", err)
	}

	return nil
}

func createPlaywright() (*playwright.Playwright, error) {
	runOptions := playwright.RunOptions{
		Browsers: []string{"chromium"},
	}

	log.Logger.Info("Installing browser if necessary...")
	err := playwright.Install(&runOptions)
	if err != nil {
		return nil, err
	}
	log.Logger.Info("Installation completed successfully.")

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

	browserLaunchOpts := playwright.BrowserTypeLaunchOptions{
		Headless: &headless,
	}
	browser, err := b.playwright.Chromium.Launch(browserLaunchOpts)
	if err != nil {
		return err
	}
	b.browser = browser

	contextOpts := playwright.BrowserNewContextOptions{
		StorageStatePath: playwright.String(config.CurrentConfig.StorageStatePath),
	}
	browserCtx, err := browser.NewContext(contextOpts)
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

		// log.Logger.Info("Skipping browser closure!")
		err = b.Close()
		if err != nil {
			log.Logger.Errorf("Error during browser closure", err.Error())
		}
	} else {
		log.Logger.Debug("No credentials need refreshing at this time.")
	}
}

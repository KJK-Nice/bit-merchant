package dsl

import (
	"fmt"
	"time"
)

// WaitForDOMUpdate waits for DOM to update via Datastar SSE
func (app *TestApplication) WaitForDOMUpdate(selector string, expectedCount int, timeout time.Duration) bool {
	page := app.GetPage()
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		count := page.MustEval(`() => {
			const el = document.querySelector('` + selector + `');
			return el ? el.children.length : 0;
		}`).Int()

		if count == expectedCount {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// WaitForSSEEvent waits for SSE event to be received and processed
func (app *TestApplication) WaitForSSEEvent(timeout time.Duration) bool {
	page := app.GetPage()
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		// Check if DOM was updated (via MutationObserver)
		domChanged := page.MustEval(`() => window.__domChanges && window.__domChanges.length > 0`).Bool()
		if domChanged {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// SetupDOMObserver sets up a MutationObserver to track DOM changes
// Useful for detecting Datastar updates
func (app *TestApplication) SetupDOMObserver(selector string) {
	page := app.GetPage()
	page.MustEval(`
		() => {
			window.__domChanges = [];
			window.__orderCount = 0;
			
			const observer = new MutationObserver((mutations) => {
				const ordersList = document.getElementById('` + selector + `');
				if (ordersList) {
					const newCount = ordersList.children.length;
					if (newCount !== window.__orderCount) {
						window.__orderCount = newCount;
						window.__domChanges.push({
							count: newCount,
							timestamp: Date.now()
						});
					}
				}
			});
			
			const ordersList = document.getElementById('` + selector + `');
			if (ordersList) {
				window.__orderCount = ordersList.children.length;
				observer.observe(ordersList, { childList: true, subtree: true });
			}
		}
	`)
}

// WaitForPageStable waits for the page to be stable (no network requests)
func (app *TestApplication) WaitForPageStable(timeout time.Duration) {
	page := app.GetPage()
	page.Timeout(timeout).MustWaitStable()
	time.Sleep(300 * time.Millisecond) // Give time for any async updates
}

// SetCookie sets a cookie on the current page
func (app *TestApplication) SetCookie(name, value string) {
	page := app.GetPage()
	// Rod uses a different API - set cookies via JavaScript or use MustSetCookies with proper type
	// For now, we'll use JavaScript to set cookies
	page.MustEval(fmt.Sprintf(`() => {
		document.cookie = "%s=%s; path=/";
	}`, name, value))
}

// ReloadPage reloads the current page
func (app *TestApplication) ReloadPage() {
	page := app.GetPage()
	page.MustReload()
	page.Timeout(10 * time.Second).MustWaitLoad()
	time.Sleep(200 * time.Millisecond)
}

// GetCurrentURL returns the current page URL
func (app *TestApplication) GetCurrentURL() string {
	page := app.GetPage()
	return page.MustEval(`() => window.location.pathname`).String()
}

// ElementExists checks if an element exists on the page
func (app *TestApplication) ElementExists(selector string) bool {
	page := app.GetPage()
	_, err := page.Element(selector)
	return err == nil
}

// GetElementText returns the text content of an element
func (app *TestApplication) GetElementText(selector string) string {
	page := app.GetPage()
	el, err := page.Element(selector)
	if err != nil {
		return ""
	}
	return el.MustEval(`el => el.textContent`).String()
}

// GetElementCount returns the number of child elements
func (app *TestApplication) GetElementCount(selector string) int {
	page := app.GetPage()
	el, err := page.Element(selector)
	if err != nil {
		return 0
	}
	return el.MustEval(`el => el ? el.children.length : 0`).Int()
}

// WaitForElement waits for an element to appear on the page
func (app *TestApplication) WaitForElement(selector string, timeout time.Duration) error {
	page := app.GetPage()
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		_, err := page.Element(selector)
		if err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("element %s not found within %v", selector, timeout)
}

// ClickElement clicks an element on the page
func (app *TestApplication) ClickElement(selector string) error {
	page := app.GetPage()
	el, err := page.Element(selector)
	if err != nil {
		return fmt.Errorf("element %s not found: %w", selector, err)
	}
	el.MustClick()
	time.Sleep(200 * time.Millisecond) // Give time for click handlers
	return nil
}

// FillInput fills an input field with the given value
func (app *TestApplication) FillInput(selector, value string) error {
	page := app.GetPage()
	el, err := page.Element(selector)
	if err != nil {
		return fmt.Errorf("input %s not found: %w", selector, err)
	}
	el.MustInput(value)
	time.Sleep(100 * time.Millisecond) // Give time for input handlers
	return nil
}

// GetElementAttribute returns an element's attribute value
func (app *TestApplication) GetElementAttribute(selector, attr string) string {
	page := app.GetPage()
	el, err := page.Element(selector)
	if err != nil {
		return ""
	}
	return el.MustEval(fmt.Sprintf(`el => el.getAttribute('%s')`, attr)).String()
}

// Screenshot takes a screenshot of the current page
// Useful for debugging test failures
// Returns the screenshot as bytes - caller should save to file if needed
func (app *TestApplication) Screenshot() ([]byte, error) {
	page := app.GetPage()
	return page.Screenshot(false, nil)
}

// ExecuteJavaScript executes JavaScript in the page context and returns the result
func (app *TestApplication) ExecuteJavaScript(script string) interface{} {
	page := app.GetPage()
	return page.MustEval(script)
}

// WaitForNetworkIdle waits for network requests to complete
func (app *TestApplication) WaitForNetworkIdle(timeout time.Duration) error {
	page := app.GetPage()
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		// Check if network is idle by evaluating network state
		// Rod doesn't have a direct network idle API, so we check for pending requests
		idle := page.MustEval(`() => {
			// Check if there are any pending fetch/XHR requests
			// This is a simplified check - in practice, you might want to track requests
			return performance.getEntriesByType('resource').length > 0;
		}`).Bool()
		
		if idle {
			// Give a bit more time to ensure all handlers are done
			time.Sleep(200 * time.Millisecond)
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("network did not become idle within %v", timeout)
}

// HasClass checks if an element has a specific CSS class
func (app *TestApplication) HasClass(selector, className string) bool {
	page := app.GetPage()
	el, err := page.Element(selector)
	if err != nil {
		return false
	}
	return el.MustEval(fmt.Sprintf(`el => el.classList.contains('%s')`, className)).Bool()
}

// GetComputedStyle returns a computed CSS property value
func (app *TestApplication) GetComputedStyle(selector, property string) string {
	page := app.GetPage()
	el, err := page.Element(selector)
	if err != nil {
		return ""
	}
	return el.MustEval(fmt.Sprintf(`el => window.getComputedStyle(el).getPropertyValue('%s')`, property)).String()
}

// SelectOption selects an option in a select/dropdown element
func (app *TestApplication) SelectOption(selector, value string) error {
	page := app.GetPage()
	el, err := page.Element(selector)
	if err != nil {
		return fmt.Errorf("select element %s not found: %w", selector, err)
	}
	// Rod's MustSelect expects a single value string, not a slice
	el.MustSelect(value)
	time.Sleep(100 * time.Millisecond) // Give time for change handlers
	return nil
}

// ScrollToElement scrolls the page to make an element visible
func (app *TestApplication) ScrollToElement(selector string) error {
	page := app.GetPage()
	el, err := page.Element(selector)
	if err != nil {
		return fmt.Errorf("element %s not found: %w", selector, err)
	}
	el.MustScrollIntoView()
	time.Sleep(200 * time.Millisecond) // Give time for scroll
	return nil
}


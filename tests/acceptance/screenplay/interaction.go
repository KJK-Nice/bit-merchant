package screenplay

import (
	"fmt"
	"time"

	"github.com/go-rod/rod/lib/proto"
)

// NavigateTo navigates to a specific URL.
type NavigateTo struct {
	URL string
}

func (n NavigateTo) PerformAs(actor *Actor) error {
	page, err := BrowseTheWebWith(actor)
	if err != nil {
		return err
	}
	return page.Navigate(n.URL)
}

func (n NavigateTo) Description() string {
	return fmt.Sprintf("navigate to %s", n.URL)
}

// ClickOn clicks on an element matching the selector.
type ClickOn struct {
	Selector string
}

func (c ClickOn) PerformAs(actor *Actor) error {
	page, err := BrowseTheWebWith(actor)
	if err != nil {
		return err
	}

	// Wait for element to be visible and clickable
	el, err := page.Element(c.Selector)
	if err != nil {
		return fmt.Errorf("failed to find element %s: %w", c.Selector, err)
	}

	return el.Click(proto.InputMouseButtonLeft, 1)
}

func (c ClickOn) Description() string {
	return fmt.Sprintf("click on %s", c.Selector)
}

// EnterText enters text into an input field.
type EnterText struct {
	Text     string
	Selector string
}

func (e EnterText) PerformAs(actor *Actor) error {
	page, err := BrowseTheWebWith(actor)
	if err != nil {
		return err
	}

	el, err := page.Element(e.Selector)
	if err != nil {
		return fmt.Errorf("failed to find element %s: %w", e.Selector, err)
	}

	if err := el.SelectAllText(); err != nil {
		return err
	}
	return el.Input(e.Text)
}

func (e EnterText) Description() string {
	return fmt.Sprintf("enter text '%s' into %s", e.Text, e.Selector)
}

// SeeText asserts that an element contains specific text.
type SeeText struct {
	Text     string
	Selector string
}

func (s SeeText) PerformAs(actor *Actor) error {
	page, err := BrowseTheWebWith(actor)
	if err != nil {
		return err
	}

	// Retry a few times for eventual consistency (hypermedia updates)
	// Rod's Element() waits for existence, but text might update dynamically.
	// For strictness we might want a custom waiter, but let's try simple verification first.
	// Or better, use Rod's Wait for text logic?
	// Let's stick to simple check, maybe with a small wait if needed,
	// but generally PerformAs is synchronous action.
	// Ideally, we should use a "Question" pattern for assertions, but Task is fine for now.

	// Using Rod's built-in wait for text is safer for acceptance tests.
	// But Rod doesn't have a direct "WaitText" on Page without selector.

	// Let's find the element first.
	el, err := page.Element(s.Selector)
	if err != nil {
		return fmt.Errorf("element %s not found: %w", s.Selector, err)
	}

	text, err := el.Text()
	if err != nil {
		return err
	}

	if text != s.Text {
		// Naive wait logic for now if immediate check fails
		time.Sleep(500 * time.Millisecond)
		text, _ = el.Text()
		if text != s.Text {
			return fmt.Errorf("expected text '%s' in %s, but found '%s'", s.Text, s.Selector, text)
		}
	}

	return nil
}

func (s SeeText) Description() string {
	return fmt.Sprintf("see text '%s' in %s", s.Text, s.Selector)
}

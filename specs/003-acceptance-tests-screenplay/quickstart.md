# Quickstart: Implementing Acceptance Tests

This guide shows you how to get started implementing the acceptance tests with the Screenplay pattern and go-rod.

## Step 1: Install go-rod

```bash
go get github.com/go-rod/rod
```

## Step 2: Create the Base Structure

Create the initial directory structure:

```bash
mkdir -p tests/acceptance/screenplay/{actors,abilities,tasks/{customer,kitchen,owner},interactions,questions}
mkdir -p tests/acceptance/{scenarios,fixtures,support}
```

## Step 3: Implement Core Interfaces

### Actor (tests/acceptance/screenplay/actors/actor.go)

```go
package actors

import (
	"fmt"

	"bitmerchant/tests/acceptance/screenplay/abilities"
)

// Task represents a high-level business action
type Task interface {
	Name() string
	PerformAs(actor *Actor) error
}

// Question represents a way to query system state
type Question interface {
	AnsweredBy(actor *Actor) (interface{}, error)
}

// Actor represents a user interacting with the system
type Actor struct {
	name         string
	browseTheWeb *abilities.BrowseTheWeb
}

// NewActor creates a new actor with the given name
func NewActor(name string) *Actor {
	return &Actor{
		name: name,
	}
}

// Name returns the actor's name
func (a *Actor) Name() string {
	return a.name
}

// Can grants an ability to the actor
func (a *Actor) Can(ability interface{}) *Actor {
	switch ab := ability.(type) {
	case *abilities.BrowseTheWeb:
		a.browseTheWeb = ab
	}
	return a
}

// BrowseTheWeb returns the actor's web browsing ability
func (a *Actor) BrowseTheWeb() *abilities.BrowseTheWeb {
	if a.browseTheWeb == nil {
		panic(fmt.Sprintf("%s does not have the ability to browse the web", a.name))
	}
	return a.browseTheWeb
}

// AttemptsTo performs a series of tasks
func (a *Actor) AttemptsTo(tasks ...Task) error {
	for _, task := range tasks {
		if err := task.PerformAs(a); err != nil {
			return fmt.Errorf("%s failed to %s: %w", a.name, task.Name(), err)
		}
	}
	return nil
}

// AsksFor queries the system state
func (a *Actor) AsksFor(question Question) (interface{}, error) {
	return question.AnsweredBy(a)
}

// AsksForString is a helper that returns the answer as a string
func (a *Actor) AsksForString(question Question) (string, error) {
	result, err := question.AnsweredBy(a)
	if err != nil {
		return "", err
	}
	return result.(string), nil
}

// AsksForFloat is a helper that returns the answer as a float64
func (a *Actor) AsksForFloat(question Question) (float64, error) {
	result, err := question.AnsweredBy(a)
	if err != nil {
		return 0, err
	}
	return result.(float64), nil
}

// AsksForInt is a helper that returns the answer as an int
func (a *Actor) AsksForInt(question Question) (int, error) {
	result, err := question.AnsweredBy(a)
	if err != nil {
		return 0, err
	}
	return result.(int), nil
}
```

### BrowseTheWeb Ability (tests/acceptance/screenplay/abilities/browse_web.go)

```go
package abilities

import (
	"time"

	"github.com/go-rod/rod"
)

// BrowseTheWeb represents the ability to interact with a web browser
type BrowseTheWeb struct {
	browser *rod.Browser
	page    *rod.Page
	baseURL string
	timeout time.Duration
}

// BrowseTheWebUsing creates a new browse ability with the given browser and base URL
func BrowseTheWebUsing(browser *rod.Browser, baseURL string) *BrowseTheWeb {
	return &BrowseTheWeb{
		browser: browser,
		baseURL: baseURL,
		timeout: 30 * time.Second,
	}
}

// WithTimeout sets the default timeout for page operations
func (b *BrowseTheWeb) WithTimeout(timeout time.Duration) *BrowseTheWeb {
	b.timeout = timeout
	return b
}

// OpenPage opens a new browser page
func (b *BrowseTheWeb) OpenPage() *rod.Page {
	if b.page == nil {
		b.page = b.browser.MustPage()
		b.page = b.page.Timeout(b.timeout)
	}
	return b.page
}

// NavigateTo navigates to a path relative to the base URL
func (b *BrowseTheWeb) NavigateTo(path string) error {
	page := b.OpenPage()
	return page.Navigate(b.baseURL + path)
}

// Page returns the current page
func (b *BrowseTheWeb) Page() *rod.Page {
	return b.OpenPage()
}

// BaseURL returns the base URL
func (b *BrowseTheWeb) BaseURL() string {
	return b.baseURL
}

// WaitForIdle waits for the page to be idle (no pending network requests)
func (b *BrowseTheWeb) WaitForIdle() error {
	return b.page.WaitIdle(b.timeout)
}

// WaitForElement waits for an element to appear
func (b *BrowseTheWeb) WaitForElement(selector string) (*rod.Element, error) {
	return b.page.Timeout(b.timeout).Element(selector)
}

// Close closes the page
func (b *BrowseTheWeb) Close() {
	if b.page != nil {
		b.page.Close()
		b.page = nil
	}
}
```

### Test Server Support (tests/acceptance/support/server.go)

```go
package support

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"bitmerchant/internal/application/cart"
	"bitmerchant/internal/application/kitchen"
	"bitmerchant/internal/application/menu"
	"bitmerchant/internal/application/order"
	"bitmerchant/internal/domain"
	"bitmerchant/internal/infrastructure/events"
	"bitmerchant/internal/infrastructure/logging"
	"bitmerchant/internal/infrastructure/payment/cash"
	"bitmerchant/internal/infrastructure/repositories/memory"
	handler "bitmerchant/internal/interfaces/http"
	"bitmerchant/internal/interfaces/http/middleware"

	"github.com/labstack/echo/v4"
)

// TestServer wraps an httptest.Server with test utilities
type TestServer struct {
	server   *httptest.Server
	echo     *echo.Echo
	eventBus *events.EventBus
	
	// Repositories (exposed for test data setup)
	RestRepo     *memory.MemoryRestaurantRepository
	MenuCatRepo  *memory.MemoryMenuCategoryRepository
	MenuItemRepo *memory.MemoryMenuItemRepository
	OrderRepo    *memory.MemoryOrderRepository
}

// StartTestServer creates and starts a test server
func StartTestServer(t *testing.T) *TestServer {
	t.Helper()

	// Infrastructure
	logger := logging.NewLogger()
	eventBus := events.NewEventBus()

	// Repositories
	restRepo := memory.NewMemoryRestaurantRepository()
	menuCatRepo := memory.NewMemoryMenuCategoryRepository()
	menuItemRepo := memory.NewMemoryMenuItemRepository()
	orderRepo := memory.NewMemoryOrderRepository()
	paymentRepo := memory.NewMemoryPaymentRepository()

	// Services
	cartService := cart.NewCartService()
	paymentMethod := cash.NewCashPaymentMethod()
	sseHandler := handler.NewSSEHandler()

	// Seed test data
	seedTestData(restRepo, menuCatRepo, menuItemRepo)

	// Use Cases
	getMenuUC := menu.NewGetMenuUseCase(menuCatRepo, menuItemRepo, restRepo)
	createOrderUC := order.NewCreateOrderUseCase(orderRepo, paymentRepo, restRepo, eventBus, paymentMethod, logger)
	getOrderUC := order.NewGetOrderByNumberUseCase(orderRepo)
	getCustomerOrdersUC := order.NewGetCustomerOrdersUseCase(orderRepo)
	getKitchenOrdersUC := kitchen.NewGetKitchenOrdersUseCase(orderRepo)
	markPaidUC := kitchen.NewMarkOrderPaidUseCase(orderRepo, eventBus)
	markPreparingUC := kitchen.NewMarkOrderPreparingUseCase(orderRepo, eventBus)
	markReadyUC := kitchen.NewMarkOrderReadyUseCase(orderRepo, eventBus)

	// Handlers
	menuHandler := handler.NewMenuHandler(getMenuUC, cartService)
	cartHandler := handler.NewCartHandler(cartService, menuItemRepo)
	orderHandler := handler.NewOrderHandler(createOrderUC, getOrderUC, getCustomerOrdersUC, cartService)
	kitchenHandler := handler.NewKitchenHandler(getKitchenOrdersUC, markPaidUC, markPreparingUC, markReadyUC)

	// Echo setup
	e := echo.New()
	e.Use(middleware.SessionMiddleware())

	// Routes
	e.GET("/menu", menuHandler.GetMenu)
	e.GET("/cart", cartHandler.GetCart)
	e.POST("/cart/add", cartHandler.AddToCart)
	e.POST("/cart/remove", cartHandler.RemoveFromCart)
	e.GET("/order/confirm", orderHandler.GetConfirmOrder)
	e.POST("/order/create", orderHandler.CreateOrder)
	e.GET("/order/:orderNumber", orderHandler.GetOrder)
	e.GET("/order/:orderNumber/stream", sseHandler.OrderStatusStream)
	e.GET("/kitchen", kitchenHandler.GetKitchen)
	e.GET("/kitchen/stream", sseHandler.KitchenStream)
	e.POST("/kitchen/order/:id/mark-paid", kitchenHandler.MarkPaid)
	e.POST("/kitchen/order/:id/mark-preparing", kitchenHandler.MarkPreparing)
	e.POST("/kitchen/order/:id/mark-ready", kitchenHandler.MarkReady)

	// Static files
	e.Static("/assets", "assets")

	server := httptest.NewServer(e)

	ts := &TestServer{
		server:       server,
		echo:         e,
		eventBus:     eventBus,
		RestRepo:     restRepo,
		MenuCatRepo:  menuCatRepo,
		MenuItemRepo: menuItemRepo,
		OrderRepo:    orderRepo,
	}

	t.Cleanup(func() {
		ts.Stop()
	})

	return ts
}

// URL returns the test server URL
func (s *TestServer) URL() string {
	return s.server.URL
}

// Stop stops the test server
func (s *TestServer) Stop() {
	s.eventBus.Close()
	s.server.Close()
}

func seedTestData(restRepo *memory.MemoryRestaurantRepository, catRepo *memory.MemoryMenuCategoryRepository, itemRepo *memory.MemoryMenuItemRepository) {
	restaurantID := domain.RestaurantID("restaurant_1")
	rest, _ := domain.NewRestaurant(restaurantID, "Test Restaurant")
	restRepo.Save(rest)

	cat1, _ := domain.NewMenuCategory("cat_1", restaurantID, "Appetizers", 1)
	cat2, _ := domain.NewMenuCategory("cat_2", restaurantID, "Mains", 2)
	cat3, _ := domain.NewMenuCategory("cat_3", restaurantID, "Drinks", 3)
	catRepo.Save(cat1)
	catRepo.Save(cat2)
	catRepo.Save(cat3)

	item1, _ := domain.NewMenuItem("item_1", "cat_1", restaurantID, "Bruschetta", 8.50)
	item1.SetDescription("Toasted bread with tomatoes and basil")
	itemRepo.Save(item1)

	item2, _ := domain.NewMenuItem("item_2", "cat_2", restaurantID, "Bitcoin Burger", 15.00)
	item2.SetDescription("Premium beef patty with cheese")
	itemRepo.Save(item2)

	item3, _ := domain.NewMenuItem("item_3", "cat_3", restaurantID, "Satoshi Soda", 3.00)
	itemRepo.Save(item3)
}
```

### Browser Support (tests/acceptance/support/browser.go)

```go
package support

import (
	"os"
	"testing"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

// NewBrowser creates a new browser instance for testing
func NewBrowser(t *testing.T) *rod.Browser {
	t.Helper()

	headless := os.Getenv("HEADLESS") != "false"

	l := launcher.New().
		Headless(headless)

	if !headless {
		l = l.Devtools(true)
	}

	url := l.MustLaunch()

	browser := rod.New().
		ControlURL(url).
		MustConnect()

	t.Cleanup(func() {
		browser.MustClose()
	})

	return browser
}
```

## Step 4: Implement Example Tasks

### Navigate to Menu (tests/acceptance/screenplay/tasks/customer/navigate_to_menu.go)

```go
package customer

import "bitmerchant/tests/acceptance/screenplay/actors"

type navigateToMenu struct{}

// NavigateToMenu returns a task to navigate to the menu page
func NavigateToMenu() *navigateToMenu {
	return &navigateToMenu{}
}

func (t *navigateToMenu) Name() string {
	return "navigate to menu"
}

func (t *navigateToMenu) PerformAs(actor *actors.Actor) error {
	return actor.BrowseTheWeb().NavigateTo("/menu")
}
```

### Add to Cart (tests/acceptance/screenplay/tasks/customer/add_to_cart.go)

```go
package customer

import (
	"fmt"

	"bitmerchant/tests/acceptance/screenplay/actors"
)

type addToCart struct {
	itemName string
}

// AddToCart returns a task to add an item to the cart
func AddToCart(itemName string) *addToCart {
	return &addToCart{itemName: itemName}
}

func (t *addToCart) Name() string {
	return fmt.Sprintf("add %s to cart", t.itemName)
}

func (t *addToCart) PerformAs(actor *actors.Actor) error {
	page := actor.BrowseTheWeb().Page()

	// Find the card containing the item name and click its "Add to Cart" button
	// Using XPath to find button within card containing the item name
	selector := fmt.Sprintf(`//div[contains(@class, 'card') and .//text()[contains(., '%s')]]//button[contains(text(), 'Add to Cart')]`, t.itemName)
	
	button, err := page.ElementX(selector)
	if err != nil {
		return fmt.Errorf("could not find Add to Cart button for %s: %w", t.itemName, err)
	}

	if err := button.Click("left", 1); err != nil {
		return fmt.Errorf("could not click Add to Cart: %w", err)
	}

	// Wait for Datastar to update
	return page.WaitIdle(5 * time.Second)
}
```

## Step 5: Implement Example Questions

### Cart Item Count (tests/acceptance/screenplay/questions/cart_item_count.go)

```go
package questions

import (
	"strconv"
	"strings"

	"bitmerchant/tests/acceptance/screenplay/actors"
)

type cartItemCount struct{}

// TheCartItemCount returns a question about the number of items in cart
func TheCartItemCount() *cartItemCount {
	return &cartItemCount{}
}

func (q *cartItemCount) AnsweredBy(actor *actors.Actor) (interface{}, error) {
	page := actor.BrowseTheWeb().Page()

	// Look for cart badge or item count element
	el, err := page.Element(`[data-testid="cart-item-count"]`)
	if err != nil {
		// Cart might be empty
		return 0, nil
	}

	text, err := el.Text()
	if err != nil {
		return 0, err
	}

	text = strings.TrimSpace(text)
	if text == "" {
		return 0, nil
	}

	return strconv.Atoi(text)
}
```

## Step 6: Write Your First Test

### tests/acceptance/scenarios/customer_ordering_test.go

```go
//go:build acceptance

package scenarios

import (
	"testing"

	"bitmerchant/tests/acceptance/screenplay/abilities"
	"bitmerchant/tests/acceptance/screenplay/actors"
	"bitmerchant/tests/acceptance/screenplay/questions"
	"bitmerchant/tests/acceptance/screenplay/tasks/customer"
	"bitmerchant/tests/acceptance/support"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomerCanBrowseMenuAndAddToCart(t *testing.T) {
	// Setup
	server := support.StartTestServer(t)
	browser := support.NewBrowser(t)

	// Create actor
	sarah := actors.NewActor("Sarah").
		Can(abilities.BrowseTheWebUsing(browser, server.URL()))
	defer sarah.BrowseTheWeb().Close()

	// Given Sarah navigates to the menu
	err := sarah.AttemptsTo(
		customer.NavigateToMenu(),
	)
	require.NoError(t, err)

	// When she adds a Bitcoin Burger to her cart
	err = sarah.AttemptsTo(
		customer.AddToCart("Bitcoin Burger"),
	)
	require.NoError(t, err)

	// Then the cart should have 1 item
	count, err := sarah.AsksForInt(questions.TheCartItemCount())
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
```

## Step 7: Run Tests

```bash
# Run with headless browser (CI)
go test -v ./tests/acceptance/... -tags=acceptance

# Run with visible browser (debugging)
HEADLESS=false go test -v ./tests/acceptance/... -tags=acceptance
```

## Next Steps

1. Add `data-testid` attributes to your templates for reliable element selection
2. Implement remaining tasks and questions as per the tasks.md file
3. Add more test scenarios for each user story
4. Configure CI pipeline to run acceptance tests

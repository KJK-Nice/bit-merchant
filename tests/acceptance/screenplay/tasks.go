package screenplay

import (
	"fmt"
)

// --- Tasks for Customer ---

// PlaceCashOrder is a task to place an order using cash.
type PlaceCashOrder struct {
	ItemName string
}

func (p PlaceCashOrder) PerformAs(actor *Actor) error {
	// 1. Add item to cart
	// Assuming the menu has buttons like "Add to Cart" near items.
	// Spec says: "When customer taps food items, Then items are added to cart"
	// This might mean clicking the item card or an add button.
	// Let's assume there's an "Add" button for the specific item.
	// Selector strategy: Find element containing text ItemName, then find 'Add' button within it?
	// Or maybe a specific ID if we control the HTML.
	// For now, let's assume we click the item name to add it (simple implementation) or a specific button.
	// "Add" button for the specific item
	addToCartSelector := fmt.Sprintf(`button[data-on*="/cart/add?itemID=%s"]`, p.ItemName)

	err := actor.AttemptsTo(
		ClickOn{Selector: addToCartSelector},
	)
	if err != nil {
		return fmt.Errorf("failed to add item: %w", err)
	}

	// 2. Go to Cart/Checkout
	// Assume a cart link/button exists
	// "When customer taps 'Place Order'" (implied from cart view)
	// First navigate to cart or open cart modal?
	// Spec: "System MUST provide shopping cart where customers can... see running total"
	// Spec: "When customer taps 'Place Order', Then order confirmation page appears"

	cartSelector := `a[href="/cart"]`
	err = actor.AttemptsTo(
		ClickOn{Selector: cartSelector},
	)
	if err != nil {
		return fmt.Errorf("failed to go to cart: %w", err)
	}

	// 3. Place Order
	// "Place Order" button on Cart page
	placeOrderSelector := `button[type="submit"]` // Assuming form submission
	err = actor.AttemptsTo(
		ClickOn{Selector: placeOrderSelector},
	)
	if err != nil {
		return fmt.Errorf("failed to click place order: %w", err)
	}

	// 4. Confirm Cash Payment
	// Spec: "When customer confirms 'I will pay with cash'"
	// Assume a button "Pay with Cash" on the confirmation page
	payCashSelector := `button[data-testid="pay-cash"]`
	err = actor.AttemptsTo(
		ClickOn{Selector: payCashSelector},
	)
	if err != nil {
		return fmt.Errorf("failed to select cash payment: %w", err)
	}

	return nil
}

func (p PlaceCashOrder) Description() string {
	return fmt.Sprintf("place a cash order for %s", p.ItemName)
}

// --- Tasks for Kitchen Staff ---

// ProcessCashOrder is a task for kitchen staff to move an order through the workflow.
type ProcessCashOrder struct {
	OrderNumber string
}

func (p ProcessCashOrder) PerformAs(actor *Actor) error {
	// The kitchen dashboard shows orders.
	// Actions: Mark Paid -> Mark Preparing -> Mark Ready

	// Selectors might need to be specific to the order card.
	// e.g. #order-123 .mark-paid-btn

	// We don't know the exact DOM structure yet, so we'll make reasonable assumptions
	// and these might fail initially, requiring updates to the HTML templates or these tests.
	// This is normal for TDD/BDD.

	// NOTE: The Spec says "When he marks the order as 'Paid'".
	// We need separate tasks for granular steps if the feature file splits them.
	// The feature file has:
	// And he marks the order as "Paid"
	// When he marks the order as "Preparing"
	// When he marks the order as "Ready"

	// So we should probably expose smaller tasks or interactions,
	// or a flexible task like "MarkOrder{Status: ...}"
	return nil // This task is a placeholder if we use granular tasks instead.
}
func (p ProcessCashOrder) Description() string { return "process order" }

type MarkOrder struct {
	Status string // "Paid", "Preparing", "Ready"
	// We might need OrderID or we assume the most recent one?
	// Feature file says "new order is placed... And he marks THE order..."
	// Ideally we pass the order identifier.
	// But in the test we might not know the generated ID easily unless we capture it.
	// For simplicity in acceptance test, maybe we just pick the first/latest order?
}

func (m MarkOrder) PerformAs(actor *Actor) error {
	// Strategy: Find the button that transitions to the target status.
	// e.g. "Mark Paid" button.

	var selector string
	switch m.Status {
	case "Paid":
		selector = `button[data-action="mark-paid"]`
	case "Preparing":
		selector = `button[data-action="mark-preparing"]`
	case "Ready":
		selector = `button[data-action="mark-ready"]`
	default:
		return fmt.Errorf("unknown status transition: %s", m.Status)
	}

	return actor.AttemptsTo(
		ClickOn{Selector: selector},
	)
}

func (m MarkOrder) Description() string {
	return fmt.Sprintf("mark order as %s", m.Status)
}

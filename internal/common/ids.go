package common

// RestaurantID represents a unique restaurant identifier.
type RestaurantID string

// CategoryID represents a unique menu category identifier.
type CategoryID string

// ItemID represents a unique menu item identifier.
type ItemID string

// OrderID represents a unique order identifier.
type OrderID string

// OrderNumber represents a human-readable order number.
type OrderNumber string

// OrderItemID represents a unique order item identifier.
type OrderItemID string

// PaymentID represents a unique payment identifier.
type PaymentID string

// UserID represents a unique user identifier.
type UserID string

// MembershipID represents a unique membership identifier.
type MembershipID string

// InvitationID represents a unique invitation identifier.
type InvitationID string

// MemberRole controls permissions inside a restaurant organization.
type MemberRole string

const (
	RoleOwner        MemberRole = "owner"
	RoleKitchenStaff MemberRole = "kitchen_staff"
	RoleCustomer     MemberRole = "customer"
)

// PaymentMethodType represents payment method type.
type PaymentMethodType string

const (
	PaymentMethodTypeCash      PaymentMethodType = "cash"
	PaymentMethodTypeLightning PaymentMethodType = "lightning"
)

// PaymentStatus represents payment status.
type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusFailed  PaymentStatus = "failed"
	PaymentStatusExpired PaymentStatus = "expired"
)

// FulfillmentStatus represents order fulfillment status.
type FulfillmentStatus string

const (
	FulfillmentStatusPaid      FulfillmentStatus = "paid"
	FulfillmentStatusPreparing FulfillmentStatus = "preparing"
	FulfillmentStatusReady     FulfillmentStatus = "ready"
	FulfillmentStatusCompleted FulfillmentStatus = "completed"
)

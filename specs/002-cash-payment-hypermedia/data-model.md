# Data Model: Cash Payment with Hypermedia UI

**Date**: 2025-01-27  
**Feature**: Cash Payment with Hypermedia UI

## Entities

### Restaurant

**Purpose**: Represents a single restaurant tenant with menu structure and owner credentials.

**Fields**:
- `ID` (RestaurantID): Unique identifier (UUID or string)
- `Name` (string): Restaurant name (required)
- `IsOpen` (bool): Restaurant open/closed status (default: true)
- `ClosedMessage` (string, optional): Custom message displayed when closed
- `ReopeningHours` (string, optional): Expected reopening time when closed
- `CreatedAt` (time.Time): Account creation timestamp
- `UpdatedAt` (time.Time): Last update timestamp

**Validation Rules**:
- Name: 1-100 characters, required
- CreatedAt: Set on creation, immutable
- UpdatedAt: Updated on every modification

**State Transitions**:
- `IsOpen: false → true`: Restaurant opens, menu becomes orderable
- `IsOpen: true → false`: Restaurant closes, ordering disabled, menu visible

**Relationships**:
- Has many MenuCategories
- Has many Orders
- Has many Payments

---

### MenuCategory

**Purpose**: Logical grouping of menu items (e.g., Appetizers, Mains, Desserts, Drinks).

**Fields**:
- `ID` (CategoryID): Unique identifier (UUID or string)
- `RestaurantID` (RestaurantID): Foreign key to Restaurant
- `Name` (string): Category name (required)
- `DisplayOrder` (int): Order for display in menu (lower = first)
- `IsActive` (bool): Category visibility (default: true)
- `CreatedAt` (time.Time): Creation timestamp
- `UpdatedAt` (time.Time): Last update timestamp

**Validation Rules**:
- Name: 1-50 characters, required
- DisplayOrder: >= 0, unique within restaurant
- RestaurantID: Must reference existing Restaurant

**State Transitions**:
- `IsActive: true → false`: Category hidden from customer menu, items retained
- `IsActive: false → true`: Category visible in customer menu

**Relationships**:
- Belongs to Restaurant
- Has many MenuItems

---

### MenuItem

**Purpose**: Food/drink item with name, description, price, photo, and availability status.

**Fields**:
- `ID` (ItemID): Unique identifier (UUID or string)
- `CategoryID` (CategoryID): Foreign key to MenuCategory
- `RestaurantID` (RestaurantID): Foreign key to Restaurant (denormalized for queries)
- `Name` (string): Item name (required)
- `Description` (string, optional): Item description
- `Price` (decimal): Price in local fiat currency (required, > 0)
- `PhotoURL` (string, optional): URL to optimized photo (300KB display version)
- `PhotoOriginalURL` (string, optional): URL to original photo (up to 2MB)
- `IsAvailable` (bool): In stock / out of stock status (default: true)
- `CreatedAt` (time.Time): Creation timestamp
- `UpdatedAt` (time.Time): Last update timestamp

**Validation Rules**:
- Name: 1-100 characters, required
- Description: 0-500 characters, optional
- Price: > 0, required, precision 2 decimal places
- PhotoURL: Valid URL format if provided
- CategoryID: Must reference existing MenuCategory
- RestaurantID: Must reference existing Restaurant

**State Transitions**:
- `IsAvailable: true → false`: Item hidden from customer menu, retained in system
- `IsAvailable: false → true`: Item visible in customer menu

**Relationships**:
- Belongs to MenuCategory
- Belongs to Restaurant (denormalized)
- Has many OrderItems (through Order)

---

### Order

**Purpose**: Customer purchase record with items, payment method, payment status, and fulfillment status.

**Fields**:
- `ID` (OrderID): Unique identifier (UUID or string)
- `OrderNumber` (string): Human-readable order number (unique, e.g., "ORD-001")
- `RestaurantID` (RestaurantID): Foreign key to Restaurant
- `PaymentMethodType` (string): Payment method type ("cash", future: "lightning") (required)
- `PaymentStatus` (PaymentStatus): Payment status enum (pending_payment, paid, not_paid)
- `FulfillmentStatus` (FulfillmentStatus): Fulfillment status enum (pending_payment, paid, preparing, ready, completed)
- `TotalAmount` (decimal): Total order amount in local currency (required, > 0)
- `Items` ([]OrderItem): Order items with quantities
- `CreatedAt` (time.Time): Order creation timestamp
- `PaidAt` (time.Time, optional): Payment confirmation timestamp
- `PreparingAt` (time.Time, optional): Kitchen started preparing timestamp
- `ReadyAt` (time.Time, optional): Order ready timestamp
- `CompletedAt` (time.Time, optional): Order completed timestamp

**Validation Rules**:
- OrderNumber: Unique, required, format "ORD-{number}"
- PaymentMethodType: Must be valid payment method type ("cash" initially, future: "lightning")
- TotalAmount: > 0, required, precision 2 decimal places
- RestaurantID: Must reference existing Restaurant
- Items: At least one item required
- PaymentStatus transitions: pending_payment → paid → (can proceed to preparing)
- FulfillmentStatus transitions: pending_payment → paid → preparing → ready → completed

**State Transitions**:
- Payment Status:
  - `pending_payment → paid`: Customer confirms cash payment, staff marks as paid
  - `pending_payment → not_paid`: Order cancelled before payment
- Fulfillment Status:
  - `pending_payment → paid`: Payment confirmed (PaymentStatus must be "paid")
  - `paid → preparing`: Kitchen staff starts preparing order
  - `preparing → ready`: Kitchen staff marks order as ready
  - `ready → completed`: Order picked up by customer (after 1 hour auto-archive)

**Relationships**:
- Belongs to Restaurant
- Has many OrderItems
- Has one Payment (optional, for tracking payment details)

---

### OrderItem

**Purpose**: Individual item within an order with quantity.

**Fields**:
- `ID` (OrderItemID): Unique identifier (UUID or string)
- `OrderID` (OrderID): Foreign key to Order
- `MenuItemID` (ItemID): Foreign key to MenuItem (snapshot at time of order)
- `Name` (string): Item name snapshot (required)
- `Price` (decimal): Item price snapshot at time of order (required, > 0)
- `Quantity` (int): Quantity ordered (required, > 0)
- `Subtotal` (decimal): Quantity * Price (required, > 0)

**Validation Rules**:
- Quantity: > 0, required
- Price: > 0, required, precision 2 decimal places
- Subtotal: Must equal Quantity * Price
- OrderID: Must reference existing Order
- MenuItemID: Must reference existing MenuItem

**Relationships**:
- Belongs to Order
- References MenuItem (snapshot)

---

### Payment

**Purpose**: Payment record tracking payment method, status, and details.

**Fields**:
- `ID` (PaymentID): Unique identifier (UUID or string)
- `OrderID` (OrderID): Foreign key to Order (optional, may be created before order)
- `RestaurantID` (RestaurantID): Foreign key to Restaurant
- `PaymentMethodType` (string): Payment method type ("cash", future: "lightning") (required)
- `Amount` (decimal): Payment amount in local currency (required, > 0)
- `Status` (PaymentStatus): Payment status enum (pending_payment, paid, not_paid)
- `CreatedAt` (time.Time): Payment creation timestamp
- `PaidAt` (time.Time, optional): Payment confirmation timestamp
- `FailedAt` (time.Time, optional): Payment failure timestamp
- `FailureReason` (string, optional): Reason for payment failure

**Validation Rules**:
- PaymentMethodType: Must be valid payment method type ("cash" initially)
- Amount: > 0, required, precision 2 decimal places
- RestaurantID: Must reference existing Restaurant
- Status transitions: pending_payment → paid or not_paid

**State Transitions**:
- `pending_payment → paid`: Payment confirmed (cash: staff marks as paid, future lightning: payment confirmed)
- `pending_payment → not_paid`: Payment failed or cancelled

**Relationships**:
- Belongs to Restaurant
- Belongs to Order (optional)

---

### Cart (Ephemeral)

**Purpose**: Session-only shopping cart containing selected items and quantities. Not persisted to database.

**Fields** (in-memory only):
- `SessionID` (string): Session identifier
- `RestaurantID` (RestaurantID): Restaurant context
- `Items` (map[ItemID]CartItem): Cart items keyed by item ID
- `TotalAmount` (decimal): Running total
- `CreatedAt` (time.Time): Cart creation timestamp
- `UpdatedAt` (time.Time): Last update timestamp

**Validation Rules**:
- SessionID: Required, unique per session
- Items: Can be empty (cart can be empty)
- TotalAmount: Calculated from items, >= 0

**State Transitions**:
- Cart is ephemeral - no persistence, cleared on session expiry
- Items can be added, removed, quantities adjusted
- Cart cleared when order is placed

**Relationships**:
- References Restaurant
- References MenuItems (for validation)

---

## Value Objects

### PaymentStatus

**Enum Values**:
- `pending_payment`: Payment not yet confirmed
- `paid`: Payment confirmed
- `not_paid`: Payment failed or cancelled

### FulfillmentStatus

**Enum Values**:
- `pending_payment`: Waiting for payment confirmation
- `paid`: Payment confirmed, waiting for kitchen
- `preparing`: Kitchen is preparing order
- `ready`: Order ready for pickup
- `completed`: Order completed (picked up)

### PaymentMethodType

**String Values**:
- `"cash"`: Cash payment (v1.0)
- `"lightning"`: Lightning Network payment (future)

**Validation**: Must be one of the supported payment method types. New payment methods can be added without changing Order/Payment entity structure.

---

## Domain Events

### OrderCreated
- **Trigger**: Order created with "Pending Payment" status
- **Payload**: OrderID, OrderNumber, RestaurantID, TotalAmount, PaymentMethodType
- **Subscribers**: Kitchen display (show new order), Analytics (track order creation)

### OrderPaid
- **Trigger**: Payment status changes to "paid"
- **Payload**: OrderID, PaymentID, PaidAt
- **Subscribers**: Kitchen display (enable "Preparing" action), Customer SSE (update order status), Analytics (track payment)

### OrderPreparing
- **Trigger**: Fulfillment status changes to "preparing"
- **Payload**: OrderID, PreparingAt
- **Subscribers**: Customer SSE (update order status), Analytics (track preparation start)

### OrderReady
- **Trigger**: Fulfillment status changes to "ready"
- **Payload**: OrderID, ReadyAt
- **Subscribers**: Customer SSE (notify customer), Kitchen display (move to ready queue), Analytics (track ready time)

### OrderCompleted
- **Trigger**: Fulfillment status changes to "completed"
- **Payload**: OrderID, CompletedAt
- **Subscribers**: Analytics (track completion, calculate fulfillment time)

---

## Repository Interfaces

### RestaurantRepository
```go
type RestaurantRepository interface {
    GetByID(ctx context.Context, id RestaurantID) (*Restaurant, error)
    Create(ctx context.Context, restaurant *Restaurant) error
    Update(ctx context.Context, restaurant *Restaurant) error
}
```

### MenuCategoryRepository
```go
type MenuCategoryRepository interface {
    GetByRestaurantID(ctx context.Context, restaurantID RestaurantID) ([]*MenuCategory, error)
    GetByID(ctx context.Context, id CategoryID) (*MenuCategory, error)
    Create(ctx context.Context, category *MenuCategory) error
    Update(ctx context.Context, category *MenuCategory) error
}
```

### MenuItemRepository
```go
type MenuItemRepository interface {
    GetByCategoryID(ctx context.Context, categoryID CategoryID) ([]*MenuItem, error)
    GetByRestaurantID(ctx context.Context, restaurantID RestaurantID) ([]*MenuItem, error)
    GetByID(ctx context.Context, id ItemID) (*MenuItem, error)
    Create(ctx context.Context, item *MenuItem) error
    Update(ctx context.Context, item *MenuItem) error
}
```

### OrderRepository
```go
type OrderRepository interface {
    GetByID(ctx context.Context, id OrderID) (*Order, error)
    GetByOrderNumber(ctx context.Context, orderNumber string) (*Order, error)
    GetByRestaurantID(ctx context.Context, restaurantID RestaurantID) ([]*Order, error)
    GetPendingByRestaurantID(ctx context.Context, restaurantID RestaurantID) ([]*Order, error)
    Create(ctx context.Context, order *Order) error
    Update(ctx context.Context, order *Order) error
}
```

### PaymentRepository
```go
type PaymentRepository interface {
    GetByID(ctx context.Context, id PaymentID) (*Payment, error)
    GetByOrderID(ctx context.Context, orderID OrderID) (*Payment, error)
    Create(ctx context.Context, payment *Payment) error
    Update(ctx context.Context, payment *Payment) error
}
```

---

## Payment Method Abstraction

### PaymentMethod Interface
```go
type PaymentMethod interface {
    // ProcessPayment initiates payment processing
    ProcessPayment(ctx context.Context, order *Order) (*Payment, error)
    
    // ValidatePayment validates payment status
    ValidatePayment(ctx context.Context, paymentID PaymentID) (*Payment, error)
    
    // GetPaymentMethodType returns the payment method type identifier
    GetPaymentMethodType() string
}
```

### CashPaymentMethod Implementation
```go
type CashPaymentMethod struct {
    // Cash-specific implementation
    // No external API calls, immediate confirmation
}

func (c *CashPaymentMethod) ProcessPayment(ctx context.Context, order *Order) (*Payment, error) {
    // Create payment with "pending_payment" status
    // Customer confirms cash payment, staff marks as paid later
}

func (c *CashPaymentMethod) ValidatePayment(ctx context.Context, paymentID PaymentID) (*Payment, error) {
    // Return current payment status
    // For cash, validation is manual (staff marks as paid)
}

func (c *CashPaymentMethod) GetPaymentMethodType() string {
    return "cash"
}
```

### Future: LightningPaymentMethod Implementation
```go
type LightningPaymentMethod struct {
    // Lightning-specific implementation
    // Integrates with Lightning Network API
}

func (l *LightningPaymentMethod) ProcessPayment(ctx context.Context, order *Order) (*Payment, error) {
    // Generate Lightning invoice
    // Create payment with "pending_payment" status
    // Poll for payment confirmation
}

func (l *LightningPaymentMethod) ValidatePayment(ctx context.Context, paymentID PaymentID) (*Payment, error) {
    // Check Lightning invoice status via API
    // Update payment status based on invoice state
}

func (l *LightningPaymentMethod) GetPaymentMethodType() string {
    return "lightning"
}
```

---

## Data Constraints

### Uniqueness Constraints
- Restaurant.Name: Unique per deployment (v1.0 single tenant)
- Order.OrderNumber: Globally unique
- MenuCategory.DisplayOrder: Unique within Restaurant
- Payment.ID: Globally unique

### Referential Integrity
- Order.RestaurantID → Restaurant.ID
- OrderItem.OrderID → Order.ID
- OrderItem.MenuItemID → MenuItem.ID
- MenuItem.CategoryID → MenuCategory.ID
- MenuCategory.RestaurantID → Restaurant.ID
- Payment.RestaurantID → Restaurant.ID
- Payment.OrderID → Order.ID (optional)

### Business Rules
- Order cannot proceed to "preparing" until PaymentStatus is "paid"
- Order.TotalAmount must equal sum of OrderItem.Subtotal
- Payment.Amount must equal Order.TotalAmount
- PaymentMethodType must match between Order and Payment
- MenuItem.Price changes do not affect existing OrderItems (snapshot)

---

## Indexes (Future PostgreSQL)

- Restaurant.ID (primary key)
- Order.OrderNumber (unique index)
- Order.RestaurantID (index for restaurant queries)
- Order.PaymentStatus (index for pending payment queries)
- Order.FulfillmentStatus (index for kitchen display queries)
- MenuItem.CategoryID (index for category queries)
- MenuItem.RestaurantID (index for restaurant queries)
- Payment.OrderID (index for order payment lookup)


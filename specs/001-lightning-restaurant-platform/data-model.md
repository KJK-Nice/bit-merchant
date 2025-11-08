# Data Model: BitMerchant v1.0

**Date**: 2025-11-08  
**Feature**: BitMerchant v1.0 - Lightning Payment Platform for Restaurants

## Entities

### Restaurant

**Purpose**: Represents a single restaurant tenant with menu structure and owner credentials.

**Fields**:
- `ID` (RestaurantID): Unique identifier (UUID or string)
- `Name` (string): Restaurant name (required)
- `LightningAddress` (string): Lightning Network address for receiving payments (required)
- `IsOpen` (bool): Restaurant open/closed status (default: true)
- `ClosedMessage` (string, optional): Custom message displayed when closed
- `ReopeningHours` (string, optional): Expected reopening time when closed
- `CreatedAt` (time.Time): Account creation timestamp
- `UpdatedAt` (time.Time): Last update timestamp

**Validation Rules**:
- Name: 1-100 characters, required
- LightningAddress: Valid Lightning address format, required
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

**Photo Storage Constraints** (FR-020, FR-042):
- Maximum 2MB per photo upload
- Maximum 100 photos per restaurant
- Automatic compression to 300KB optimized version for display
- If restaurant reaches 100 photo limit, must delete existing photos before uploading new ones

---

### Order

**Purpose**: Customer purchase record with items, payment status, and fulfillment status.

**Fields**:
- `ID` (OrderID): Unique identifier (UUID or string)
- `OrderNumber` (string): Human-readable order number (e.g., "ORD-001", unique)
- `RestaurantID` (RestaurantID): Foreign key to Restaurant
- `Items` ([]OrderItem): Ordered items with quantities
- `TotalAmount` (int64): Total amount in satoshis (calculated from items)
- `FiatAmount` (decimal): Total amount in local fiat currency (for display)
- `PaymentStatus` (PaymentStatus): pending | paid | failed
- `FulfillmentStatus` (FulfillmentStatus): paid | preparing | ready | completed
- `LightningInvoiceID` (string, optional): Strike API invoice ID
- `LightningInvoice` (string, optional): Lightning invoice string (for QR code)
- `CreatedAt` (time.Time): Order creation timestamp (when payment confirmed)
- `UpdatedAt` (time.Time): Last update timestamp
- `CompletedAt` (time.Time, optional): Order completion timestamp (when picked up)

**Validation Rules**:
- OrderNumber: Unique within restaurant, auto-generated format "ORD-{sequence}"
- Items: At least 1 item required
- TotalAmount: > 0, calculated from items + exchange rate
- PaymentStatus: Must be "paid" before order can be fulfilled
- FulfillmentStatus: Valid transition: paid → preparing → ready → completed
- LightningInvoiceID: Required when PaymentStatus is "paid" or "pending"

**State Transitions**:
- `PaymentStatus: pending → paid`: Order created, appears in kitchen queue
- `PaymentStatus: pending → failed`: Order not created, customer can retry
- `FulfillmentStatus: paid → preparing`: Kitchen staff starts preparing
- `FulfillmentStatus: preparing → ready`: Food ready for pickup
- `FulfillmentStatus: ready → completed`: Customer picked up, order archived after 1 hour

**Relationships**:
- Belongs to Restaurant
- Has many OrderItems
- Has one Payment (when paid)

**Business Rules**:
- Order only created when PaymentStatus is "paid" (FR-029)
- Order sequence integrity based on CreatedAt timestamp (FR-035)
- Completed orders archived after 1 hour (FR-015)

---

### OrderItem

**Purpose**: Individual item within an order with quantity.

**Fields**:
- `ID` (OrderItemID): Unique identifier (UUID or string)
- `OrderID` (OrderID): Foreign key to Order
- `MenuItemID` (ItemID): Foreign key to MenuItem
- `Quantity` (int): Quantity ordered (required, > 0)
- `UnitPrice` (decimal): Price per unit in fiat currency (snapshot at order time)
- `UnitPriceSatoshis` (int64): Price per unit in satoshis (snapshot at order time)
- `Subtotal` (int64): Quantity × UnitPriceSatoshis (calculated)

**Validation Rules**:
- Quantity: > 0, required
- UnitPrice: > 0, snapshot from MenuItem.Price at order time
- UnitPriceSatoshis: > 0, calculated from UnitPrice + exchange rate at order time
- MenuItemID: Must reference existing MenuItem

**Relationships**:
- Belongs to Order
- References MenuItem (snapshot, not foreign key - item may be deleted)

---

### Payment

**Purpose**: Lightning Network transaction record with invoice, amount, status, and settlement.

**Fields**:
- `ID` (PaymentID): Unique identifier (UUID or string)
- `RestaurantID` (RestaurantID): Foreign key to Restaurant
- `OrderID` (OrderID, optional): Foreign key to Order (when payment succeeds)
- `InvoiceID` (string): Strike API invoice ID (required)
- `Invoice` (string): Lightning invoice string (required)
- `AmountSatoshis` (int64): Amount in satoshis (required, > 0)
- `AmountFiat` (decimal): Amount in local fiat currency (required, > 0)
- `ExchangeRate` (decimal): Exchange rate used at invoice generation (required)
- `Status` (PaymentStatus): pending | paid | failed | expired
- `SettlementStatus` (SettlementStatus): pending | settled
- `SettledAt` (time.Time, optional): Settlement timestamp (when paid to restaurant)
- `SettlementTransactionID` (string, optional): Settlement transaction ID
- `CreatedAt` (time.Time): Payment creation timestamp
- `PaidAt` (time.Time, optional): Payment completion timestamp
- `FailedAt` (time.Time, optional): Payment failure timestamp
- `FailureReason` (string, optional): Failure reason if Status is "failed"

**Validation Rules**:
- InvoiceID: Unique, required, from Strike API
- AmountSatoshis: > 0, required
- AmountFiat: > 0, required
- ExchangeRate: > 0, required, snapshot at invoice generation
- Status: Valid values: pending, paid, failed, expired
- SettlementStatus: Valid values: pending, settled

**State Transitions**:
- `Status: pending → paid`: Payment confirmed, Order created, PaidAt set
- `Status: pending → failed`: Payment failed, FailureReason set, FailedAt set
- `Status: pending → expired`: Invoice expired (typically 15 minutes)
- `SettlementStatus: pending → settled`: Daily settlement completed, SettledAt set

**Relationships**:
- Belongs to Restaurant
- Has one Order (when Status is "paid")

**Business Rules**:
- Payment must be validated before Order creation (FR-029)
- Payment failures detected within 30 seconds (FR-030)
- Daily settlement happens at end of business day (11:59 PM local time, FR-031)
- Exchange rate snapshot at invoice generation (FR-033)

---

## Value Objects

### PaymentStatus

**Type**: Enum/string  
**Values**: `pending`, `paid`, `failed`, `expired`

**Purpose**: Tracks Lightning payment lifecycle.

---

### FulfillmentStatus

**Type**: Enum/string  
**Values**: `paid`, `preparing`, `ready`, `completed`

**Purpose**: Tracks order fulfillment lifecycle from payment to pickup.

---

### SettlementStatus

**Type**: Enum/string  
**Values**: `pending`, `settled`

**Purpose**: Tracks daily Bitcoin settlement to restaurant Lightning address.

---

## Domain Events

### OrderPaid

**Triggered**: When Payment.Status changes from `pending` to `paid`  
**Payload**:
- OrderID
- RestaurantID
- OrderNumber
- TotalAmount (satoshis)
- CreatedAt

**Handlers**:
- Create Order entity
- Publish to kitchen display (SSE)
- Send push notification to customer

---

### OrderStatusChanged

**Triggered**: When Order.FulfillmentStatus changes  
**Payload**:
- OrderID
- OrderNumber
- PreviousStatus
- NewStatus
- UpdatedAt

**Handlers**:
- Update customer view (SSE)
- Send push notification if status is "ready"

---

### PaymentFailed

**Triggered**: When Payment.Status changes to `failed`  
**Payload**:
- PaymentID
- InvoiceID
- FailureReason
- FailedAt

**Handlers**:
- Show error message to customer
- Allow retry (generate new invoice)

---

## Repository Interfaces

### RestaurantRepository

```go
type RestaurantRepository interface {
    Save(restaurant *Restaurant) error
    FindByID(id RestaurantID) (*Restaurant, error)
    Update(restaurant *Restaurant) error
}
```

---

### MenuCategoryRepository

```go
type MenuCategoryRepository interface {
    Save(category *MenuCategory) error
    FindByID(id CategoryID) (*MenuCategory, error)
    FindByRestaurantID(restaurantID RestaurantID) ([]*MenuCategory, error)
    Update(category *MenuCategory) error
    Delete(id CategoryID) error
}
```

---

### MenuItemRepository

```go
type MenuItemRepository interface {
    Save(item *MenuItem) error
    FindByID(id ItemID) (*MenuItem, error)
    FindByCategoryID(categoryID CategoryID) ([]*MenuItem, error)
    FindByRestaurantID(restaurantID RestaurantID) ([]*MenuItem, error)
    FindAvailableByRestaurantID(restaurantID RestaurantID) ([]*MenuItem, error)
    Update(item *MenuItem) error
    Delete(id ItemID) error
    CountByRestaurantID(restaurantID RestaurantID) (int, error) // For 100 photo limit check
}
```

---

### OrderRepository

```go
type OrderRepository interface {
    Save(order *Order) error
    FindByID(id OrderID) (*Order, error)
    FindByOrderNumber(restaurantID RestaurantID, orderNumber string) (*Order, error)
    FindByRestaurantID(restaurantID RestaurantID) ([]*Order, error)
    FindActiveByRestaurantID(restaurantID RestaurantID) ([]*Order, error) // paid, preparing, ready
    Update(order *Order) error
}
```

---

### PaymentRepository

```go
type PaymentRepository interface {
    Save(payment *Payment) error
    FindByID(id PaymentID) (*Payment, error)
    FindByInvoiceID(invoiceID string) (*Payment, error)
    FindByRestaurantID(restaurantID RestaurantID) ([]*Payment, error)
    FindPendingSettlements(restaurantID RestaurantID) ([]*Payment, error)
    Update(payment *Payment) error
}
```

---

## Data Constraints Summary

- **Restaurant**: 1 per deployment (v1.0 single tenant)
- **MenuCategory**: Unlimited per restaurant
- **MenuItem**: Maximum 100 photos per restaurant (FR-042), 2MB max upload per photo (FR-020)
- **Order**: Unlimited per restaurant, archived after 1 hour when completed
- **Payment**: Unlimited per restaurant, daily settlement batch

## Notes

- All timestamps use restaurant's local timezone (Assumption)
- Exchange rates snapshot at invoice generation (FR-033)
- Order items snapshot MenuItem prices (denormalized for historical accuracy)
- Photo optimization: 2MB upload → 300KB display version (FR-020)


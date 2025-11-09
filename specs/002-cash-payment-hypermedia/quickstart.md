# Quickstart: Cash Payment with Hypermedia UI Development Setup

**Date**: 2025-01-27  
**Feature**: Cash Payment with Hypermedia UI

## Prerequisites

- **Go 1.25+** (or Go 1.21+ minimum): [Install Go](https://go.dev/doc/install)
- **Git**: For version control
- **Photo Storage**: AWS S3 account or similar (for production) or local storage (for development)

## Initial Setup

### 1. Clone Repository

```bash
git clone <repository-url>
cd bit-merchant
git checkout 002-cash-payment-hypermedia
```

### 2. Install Dependencies

```bash
go mod download
```

**Key Dependencies**:
- `github.com/labstack/echo/v4` - Web framework with HTTP/2 SSE support
- `github.com/a-h/templ` - Type-safe server-rendered HTML templates
- `github.com/delaneyj/datastar` - Hypermedia-driven UI with SSE for partial page updates and real-time DOM updates (REQUIRED - NOT HTMX)
- `github.com/ThreeDotsLabs/watermill` - Event bus for domain events
- `github.com/stretchr/testify` - Testing assertions

### 3. Install Development Tools

```bash
# Templ template compiler
go install github.com/a-h/templ/cmd/templ@latest

# Linting
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Complexity checker
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
```

### 4. Configure Environment

Create `.env` file in project root:

```bash
# Server
PORT=8080
ENV=development

# Restaurant (v1.0 single tenant)
RESTAURANT_ID=rest_001
RESTAURANT_NAME="My Restaurant"

# Photo Storage (development - use local storage)
PHOTO_STORAGE_TYPE=local
PHOTO_STORAGE_PATH=./static/photos

# Photo Storage (production - use S3)
# PHOTO_STORAGE_TYPE=s3
# AWS_REGION=us-east-1
# AWS_BUCKET=bitmerchant-photos
# AWS_ACCESS_KEY_ID=your_access_key
# AWS_SECRET_ACCESS_KEY=your_secret_key
```

### 5. Generate Templates

```bash
# Generate Go code from Templ templates
templ generate
```

**Note**: Run `templ generate` after every template change. Use `templ generate --watch` for development with auto-reload.

### 6. Run Development Server

```bash
# With live reload (recommended for development)
templ generate --watch --proxy="http://localhost:8080" --cmd="go run cmd/server/main.go"

# Or manually
go run cmd/server/main.go
```

Server starts on `http://localhost:8080` (or port specified in `.env`).

## Development Workflow

### Template Development

1. **Create/Edit Templates**: Edit `.templ` files in `internal/interfaces/templates/`
2. **Generate Go Code**: Run `templ generate` (or use `--watch` mode)
3. **Use Templ UI Components**: Import components from templui.io patterns
4. **Reference Templ Guide**: Follow patterns from https://templ.guide/llms.md for AI-assisted development

**Example Template**:
```templ
package templates

import "bitmerchant/internal/domain"

templ MenuPage(restaurant *domain.Restaurant, categories []*domain.MenuCategory) {
    <!DOCTYPE html>
    <html>
        <head>
            <title>{ restaurant.Name } - Menu</title>
        </head>
        <body>
            <main>
                <h1>{ restaurant.Name }</h1>
                for _, category := range categories {
                    <section class="category">
                        <h2>{ category.Name }</h2>
                        <!-- Menu items rendered here -->
                    </section>
                }
            </main>
        </body>
    </html>
}
```

### Payment Method Development

1. **Payment Method Interface**: Implement `PaymentMethod` interface in `internal/domain/payment.go`
2. **Cash Implementation**: Create `CashPaymentMethod` in `internal/infrastructure/payment/cash/`
3. **Future Lightning**: Create `LightningPaymentMethod` in `internal/infrastructure/payment/lightning/` (future)

**Example Payment Method**:
```go
package cash

import "bitmerchant/internal/domain"

type CashPaymentMethod struct {
    // Cash-specific fields
}

func (c *CashPaymentMethod) ProcessPayment(ctx context.Context, order *domain.Order) (*domain.Payment, error) {
    // Create payment with "pending_payment" status
    // Customer confirms cash, staff marks as paid later
}

func (c *CashPaymentMethod) GetPaymentMethodType() string {
    return "cash"
}
```

### Hypermedia Interactions

1. **Form Submissions**: Use standard HTML forms with Datastar attributes (ds-post, ds-target) for partial page updates
2. **Real-Time Updates**: Implement SSE endpoints for order status updates via Datastar (ds-sse-connect)
3. **No JavaScript Required**: Datastar handles DOM updates automatically - basic functionality works without JavaScript

**Example Form with Datastar**:
```html
<form method="POST" action="/cart/add" ds-post="/cart/add" ds-target="#cart-summary">
    <input type="hidden" name="itemId" value="{ItemID}" />
    <input type="number" name="quantity" value="1" min="1" />
    <button type="submit">Add to Cart</button>
</form>
```

## Testing

### Run All Tests

```bash
go test ./...
```

### Run Tests with Coverage

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Coverage Requirements

- **Payment processing**: 95% coverage (critical path)
- **Order management**: 95% coverage (critical path)
- **Menu management**: 80% coverage (standard)

### Test Types

- **Unit Tests**: Domain logic, use cases, repositories
- **Integration Tests**: SSE via Datastar, photo storage, hypermedia interactions
- **Contract Tests**: HTML endpoint contracts

**Example Test**:
```go
package order_test

import (
    "testing"
    "bitmerchant/internal/application/order"
)

func TestCreateOrder(t *testing.T) {
    // Test order creation with cash payment
    // Verify payment method abstraction works
}
```

## Linting

### Run Linter

```bash
golangci-lint run
```

### Check Complexity

```bash
gocyclo -over 10 .
```

**Requirements**:
- Functions max 50 lines
- Classes max 300 lines
- Cyclomatic complexity max 10 per function

## Project Structure

```
bit-merchant/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── domain/               # Business logic (zero dependencies)
│   ├── application/          # Use cases
│   ├── infrastructure/       # External adapters
│   │   ├── payment/          # Payment method implementations
│   │   │   └── cash/         # Cash payment implementation
│   │   └── repositories/     # Data access
│   └── interfaces/           # HTTP handlers, templates
│       └── templates/        # Templ templates
├── tests/
│   ├── contract/             # HTML contract tests
│   ├── integration/          # Integration tests
│   └── unit/                 # Unit tests
└── static/                   # Static assets, PWA
```

## Key References

- **Templ Documentation**: https://templ.guide/
- **Templ LLM Guide**: https://templ.guide/llms.md (use for AI-assisted Templ development)
- **Templ UI Components**: https://templui.io/ (pre-built UI component library)
- **Datastar**: https://github.com/delaneyj/datastar - Hypermedia-driven UI with SSE for Go/Templ (REQUIRED - NOT HTMX)
- **Echo Framework**: https://echo.labstack.com/
- **Watermill Event Bus**: https://watermill.io/

## Common Tasks

### Add New Menu Item

1. Create menu item via owner dashboard (`/dashboard/menu`)
2. Item appears in customer menu within 5 seconds (FR-024)

### Mark Order as Paid

1. Customer confirms cash payment
2. Kitchen staff marks order as "Paid" via kitchen display
3. Order status updates in real-time via SSE

### Add New Payment Method (Future)

1. Implement `PaymentMethod` interface
2. Create implementation in `internal/infrastructure/payment/{method}/`
3. Register payment method in application setup
4. No changes to Order/Payment entities required (payment method abstraction)

## Troubleshooting

### Templates Not Updating

```bash
# Regenerate templates
templ generate

# Check for syntax errors
templ generate --help
```

### SSE Not Working

- Verify HTTP/2 is enabled in Echo
- Check browser console for SSE connection errors
- Ensure `text/event-stream` content type is set

### Payment Method Not Found

- Verify payment method is registered in `cmd/server/main.go`
- Check payment method type matches Order.PaymentMethodType

## Next Steps

1. **Read Specification**: Review `spec.md` for complete requirements
2. **Review Data Model**: See `data-model.md` for entity structure
3. **Check Contracts**: See `contracts/html-contracts.md` for HTML endpoint contracts
4. **Start Development**: Begin with User Story 1 (Customer Ordering)


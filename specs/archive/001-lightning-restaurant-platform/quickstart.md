# Quickstart: BitMerchant v1.0 Development Setup

**Date**: 2025-11-08  
**Feature**: BitMerchant v1.0 - Lightning Payment Platform for Restaurants

## Prerequisites

- **Go 1.25+** (or Go 1.21+ minimum): [Install Go](https://go.dev/doc/install)
- **Git**: For version control
- **Strike API Account**: For Lightning Network payment processing (test mode available)
- **Photo Storage**: AWS S3 account or similar (for production) or local storage (for development)

## Initial Setup

### 1. Clone Repository

```bash
git clone <repository-url>
cd bit-merchant
git checkout 001-lightning-restaurant-platform
```

### 2. Install Dependencies

```bash
go mod download
```

**Key Dependencies**:
- `github.com/labstack/echo/v4` - Web framework
- `github.com/a-h/templ` - Type-safe templates
- `github.com/delaneyj/datastar` - Real-time updates
- `github.com/ThreeDotsLabs/watermill` - Event bus
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

# Strike API
STRIKE_API_KEY=your_strike_api_key
STRIKE_API_URL=https://api.strike.me/v1

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

Run this whenever you modify `.templ` files.

## Project Structure

```
bit-merchant/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── domain/                  # Business logic (zero dependencies)
│   │   ├── restaurant.go
│   │   ├── menu.go
│   │   ├── order.go
│   │   ├── payment.go
│   │   └── events.go
│   ├── application/             # Use cases
│   │   ├── menu/
│   │   ├── order/
│   │   ├── payment/
│   │   └── kitchen/
│   ├── infrastructure/          # External adapters
│   │   ├── repositories/
│   │   │   └── memory/          # In-memory repos (v1.0)
│   │   ├── strike/              # Strike API client
│   │   ├── storage/             # Photo storage
│   │   └── events/              # Watermill event bus
│   └── interfaces/              # HTTP handlers, templates
│       ├── http/
│       └── templates/
├── tests/
│   ├── contract/                # Contract tests
│   ├── integration/             # Integration tests
│   └── unit/                    # Unit tests
├── static/
│   └── pwa/                     # PWA manifest, service worker
└── go.mod
```

## Running the Application

### Development Mode

```bash
# Watch templates and regenerate
templ generate --watch &

# Run server
go run cmd/server/main.go
```

Server starts at `http://localhost:8080`

### Production Build

```bash
# Build binary
go build -o bin/bitmerchant cmd/server/main.go

# Run binary
./bin/bitmerchant
```

## Testing

### Run All Tests

```bash
go test ./...
```

### Run with Coverage

```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Coverage Requirements

- **Payment processing**: 95% coverage (critical path)
- **Order management, kitchen display**: 95% coverage (critical path)
- **Menu management, analytics**: 80% coverage (standard)
- **Overall**: Minimum 80% coverage

### Run Specific Test Types

```bash
# Unit tests
go test ./internal/domain/... ./internal/application/...

# Integration tests
go test ./tests/integration/...

# Contract tests
go test ./tests/contract/...
```

### Test Execution Time

All unit tests must complete in under 5 minutes (Constitution requirement).

## Linting and Code Quality

### Run Linter

```bash
golangci-lint run
```

### Check Complexity

```bash
gocyclo -over 10 ./internal
```

**Requirements**:
- Functions max 50 lines
- Classes max 300 lines
- Cyclomatic complexity max 10 per function

### Format Code

```bash
go fmt ./...
```

## Development Workflow

### 1. Create Feature Branch

```bash
git checkout -b feature/my-feature
```

### 2. Write Tests First (TDD)

```bash
# Write test
# Run test (should fail)
go test ./internal/application/order/...

# Implement feature
# Run test (should pass)
go test ./internal/application/order/...
```

### 3. Generate Templates

After modifying `.templ` files:

```bash
templ generate
```

### 4. Run Linter

```bash
golangci-lint run
```

### 5. Run Tests

```bash
go test ./...
```

### 6. Commit Changes

```bash
git add .
git commit -m "feat: add order creation use case"
```

## Common Tasks

### Add New Menu Item

1. Create domain entity in `internal/domain/menu.go`
2. Create repository interface in `internal/domain/`
3. Implement in-memory repository in `internal/infrastructure/repositories/memory/`
4. Create use case in `internal/application/menu/`
5. Create HTTP handler in `internal/interfaces/http/`
6. Create Templ template in `internal/interfaces/templates/`
7. Write tests (unit, integration, contract)

### Add New Domain Event

1. Define event in `internal/domain/events.go`
2. Publish event in use case
3. Create Watermill handler in `internal/infrastructure/events/`
4. Stream via SSE in `internal/interfaces/http/sse.go`
5. Write tests

### Add New API Endpoint

1. Define contract in `contracts/api-contracts.md`
2. Create HTTP handler in `internal/interfaces/http/`
3. Register route in `cmd/server/main.go`
4. Write contract test in `tests/contract/`
5. Write integration test in `tests/integration/`

## Strike API Integration

### Test Mode

Strike API provides test mode for development:

```bash
STRIKE_API_KEY=test_key_123
STRIKE_API_URL=https://api.strike.me/v1
```

### Create Invoice

```go
// Example: Create Lightning invoice
invoice, err := strikeClient.CreateInvoice(ctx, CreateInvoiceRequest{
    Amount: Amount{
        Currency: "USD",
        Amount:   "11.98",
    },
    Description: "Order ORD-001",
})
```

### Poll Payment Status

```go
// Poll every 2-3 seconds
status, err := strikeClient.GetInvoiceStatus(ctx, invoiceID)
if status == "PAID" {
    // Create order
}
```

## Photo Storage

### Local Storage (Development)

```bash
PHOTO_STORAGE_TYPE=local
PHOTO_STORAGE_PATH=./static/photos
```

Photos stored in `./static/photos/` directory.

### S3 Storage (Production)

```bash
PHOTO_STORAGE_TYPE=s3
AWS_REGION=us-east-1
AWS_BUCKET=bitmerchant-photos
```

Photos uploaded to S3, CDN URL returned.

### Photo Optimization

Photos automatically optimized:
- Upload: Max 2MB
- Display: 300KB optimized version
- Limit: 100 photos per restaurant

## Real-Time Updates (SSE)

### Customer Order Status

```javascript
// Client-side (Datastar or EventSource)
const eventSource = new EventSource('/order/ORD-001/stream');
eventSource.addEventListener('order-status', (e) => {
    const data = JSON.parse(e.data);
    // Update DOM: data.status, data.updatedAt
});
```

### Kitchen Display

```javascript
// Kitchen display SSE stream
const eventSource = new EventSource('/kitchen/rest_123/stream');
eventSource.addEventListener('new-order', (e) => {
    const order = JSON.parse(e.data);
    // Display new order
});
```

## Performance Testing

### Menu Load Time

```bash
# Test menu load <2 seconds on 3G
curl -w "@curl-format.txt" http://localhost:8080/menu
```

### Payment Flow

```bash
# Test payment completion <10 seconds
# Create invoice, poll status until paid
```

### Order Status Updates

```bash
# Test status update propagation <5 seconds
# Update order status, verify SSE event received
```

## Troubleshooting

### Templates Not Updating

```bash
# Regenerate templates
templ generate
```

### Tests Failing

```bash
# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run TestOrderCreation ./internal/application/order/...
```

### Linter Errors

```bash
# Auto-fix where possible
golangci-lint run --fix
```

### SSE Not Working

- Check HTTP/2 support in Echo configuration
- Verify Watermill event bus is running
- Check browser console for EventSource errors

## Next Steps

1. **Read Specification**: Review `spec.md` for feature requirements
2. **Review Data Model**: Understand entities in `data-model.md`
3. **Review Contracts**: Understand API endpoints in `contracts/api-contracts.md`
4. **Start Development**: Begin with User Story 1 (Customer Ordering)

## Resources

- **Go Documentation**: https://go.dev/doc/
- **Echo Framework**: https://echo.labstack.com/
- **Templ Templates**: https://templ.guide/
- **Watermill Events**: https://watermill.io/
- **Datastar**: https://github.com/delaneyj/datastar
- **Strike API**: https://strike.me/developers

## Support

For questions or issues:
1. Check specification: `spec.md`
2. Review architecture: `plan.md`
3. Check contracts: `contracts/api-contracts.md`


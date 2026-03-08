# BitMerchant

BitMerchant is a lightning-fast restaurant ordering platform designed for cash-first payments with a hypermedia-driven UI.

## Features

- **Customer Ordering**: Scan QR code, browse menu, order, and pay with cash. No account required.
- **Kitchen Display**: Real-time order management for kitchen staff.
- **Owner Dashboard**: Menu management and sales analytics.
- **PWA**: Installable as an app, works offline (menu browsing).
- **Performance**: Server-rendered HTML for speed, SSE for real-time updates.

## Tech Stack

- **Language**: Go 1.25+
- **Web Framework**: Echo v4
- **Templating**: Templ (Type-safe Go templates)
- **UI Library**: Datastar (Hypermedia) + TemplUI
- **Database**: In-memory (MVP), Interface-ready for PostgreSQL

## Getting Started

### Prerequisites

- Go 1.25+
- Make (optional)

### Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Install templ:
   ```bash
   go install github.com/a-h/templ/cmd/templ@latest
   ```

### Running the App

1. Generate templates:
   ```bash
   templ generate
   ```
2. Run the server:
   ```bash
   go run cmd/server/main.go
   ```
3. Open http://localhost:8080

### Development

- Run tests:
  ```bash
  go test ./...
  ```

## Current Status

- Dashboard analytics now support `today`, `week`, and `month` date ranges with unit/integration coverage.
- CSRF protection is enabled in the Echo middleware stack, with token handling for both form submissions and Datastar-driven POST requests.
- Dashboard and mobile footer navigation were revalidated, and menu closed-state rendering was fixed (`circle-alert` icon usage).

## License

Proprietary.


# HTML Contracts: Cash Payment with Hypermedia UI

**Date**: 2025-01-27  
**Feature**: Cash Payment with Hypermedia UI

## Overview

All HTTP endpoints return HTML pages instead of JSON responses (FR-038, FR-039, FR-040). Hypermedia-driven interactions use Datastar for form submissions, links, and Server-Sent Events (SSE) for real-time updates (FR-041, FR-042, FR-043). **Note: We are using Datastar, NOT HTMX, for the hypermedia UI.**

## Customer-Facing Routes

### GET /menu

**Purpose**: Display restaurant menu with categories, items, prices, and photos.

**Response**: HTML page (200 OK)

**HTML Structure**:
```html
<!DOCTYPE html>
<html>
<head>
    <title>{Restaurant Name} - Menu</title>
    <link rel="manifest" href="/static/pwa/manifest.json">
    <!-- PWA and styling -->
</head>
<body>
    <main>
        <h1>{Restaurant Name}</h1>
        {# If restaurant closed #}
        <div class="closed-banner">
            <p>{ClosedMessage}</p>
            <p>Reopening: {ReopeningHours}</p>
        </div>
        {# End if #}
        
        {# For each category #}
        <section class="category">
            <h2>{Category Name}</h2>
            {# For each item in category #}
            <article class="menu-item">
                <img src="{PhotoURL}" alt="{Name}" />
                <h3>{Name}</h3>
                <p>{Description}</p>
                <p class="price">{Price}</p>
                <form method="POST" action="/cart/add" ds-post="/cart/add" ds-target="#cart-summary">
                    <input type="hidden" name="itemId" value="{ItemID}" />
                    <input type="number" name="quantity" value="1" min="1" />
                    <button type="submit">Add to Cart</button>
                </form>
            </article>
            {# End for #}
        </section>
        {# End for #}
        
        <div id="cart-summary">
            <!-- Cart summary updated via hypermedia -->
        </div>
    </main>
</body>
</html>
```

**Performance**: <2 seconds on 3G (SC-003)

**Error Responses**:
- 404: Restaurant not found
- 503: Restaurant is closed (menu still visible, ordering disabled)

---

### POST /cart/add

**Purpose**: Add item to cart with hypermedia update (no full page reload).

**Request**: Form data
- `itemId` (string, required): Menu item ID
- `quantity` (int, required): Quantity to add

**Response**: HTML fragment (200 OK) - cart summary section

**HTML Fragment**:
```html
<div id="cart-summary">
    <h2>Cart</h2>
    {# For each item in cart #}
    <div class="cart-item">
        <span>{Name}</span>
        <span>{Quantity}</span>
        <span>{Subtotal}</span>
        <form method="POST" action="/cart/remove" ds-post="/cart/remove" ds-target="#cart-summary">
            <input type="hidden" name="itemId" value="{ItemID}" />
            <button type="submit">Remove</button>
        </form>
    </div>
    {# End for #}
    <p>Total: {TotalAmount}</p>
    <form method="GET" action="/order/confirm">
        <button type="submit">Place Order</button>
    </form>
</div>
```

**Error Responses**:
- 400: Invalid request (missing itemId or quantity)
- 404: Item not found or not available
- 503: Restaurant is closed

---

### POST /cart/remove

**Purpose**: Remove item from cart with hypermedia update.

**Request**: Form data
- `itemId` (string, required): Menu item ID
- `quantity` (int, optional): Quantity to remove (default: all)

**Response**: HTML fragment (200 OK) - cart summary section

**Error Responses**:
- 400: Invalid request

---

### GET /cart

**Purpose**: Display current cart contents.

**Response**: HTML page (200 OK) - same structure as cart summary fragment

---

### GET /order/confirm

**Purpose**: Display order confirmation page with order summary.

**Response**: HTML page (200 OK)

**HTML Structure**:
```html
<!DOCTYPE html>
<html>
<head>
    <title>Confirm Order</title>
</head>
<body>
    <main>
        <h1>Confirm Your Order</h1>
        <div class="order-summary">
            {# For each item #}
            <div class="order-item">
                <span>{Name}</span>
                <span>{Quantity}</span>
                <span>{Subtotal}</span>
            </div>
            {# End for #}
            <p>Total: {TotalAmount}</p>
        </div>
        
        <form method="POST" action="/order/create">
            <input type="hidden" name="paymentMethod" value="cash" />
            <button type="submit">Confirm Cash Payment</button>
        </form>
    </main>
</body>
</html>
```

---

### POST /order/create

**Purpose**: Create order with cash payment confirmation.

**Request**: Form data
- `paymentMethod` (string, required): "cash"

**Response**: HTML page (200 OK) - order confirmation with order number

**HTML Structure**:
```html
<!DOCTYPE html>
<html>
<head>
    <title>Order Confirmed</title>
</head>
<body>
    <main>
        <h1>Order Confirmed</h1>
        <p>Order Number: <strong>{OrderNumber}</strong></p>
        <p>Status: <span id="order-status">Pending Payment</span></p>
        
        <!-- SSE connection for real-time updates via Datastar -->
        <div id="order-updates" ds-sse-connect="/order/{OrderNumber}/stream">
            <!-- Order status updates appear here -->
        </div>
        
        <p>Please pay cash to staff. Your order will be prepared once payment is confirmed.</p>
    </main>
</body>
</html>
```

**Error Responses**:
- 400: Invalid request or empty cart
- 503: Restaurant is closed

---

### GET /order/:orderNumber

**Purpose**: Lookup order by order number (for customers who lost connection).

**Response**: HTML page (200 OK) - same structure as order confirmation page

**Error Responses**:
- 404: Order not found

---

### GET /order/:orderNumber/stream

**Purpose**: Server-Sent Events stream for real-time order status updates.

**Response**: text/event-stream (200 OK)

**Event Format**:
```
event: order-status
data: {"status": "paid", "timestamp": "2025-01-27T10:30:00Z"}

event: order-status
data: {"status": "preparing", "timestamp": "2025-01-27T10:35:00Z"}

event: order-status
data: {"status": "ready", "timestamp": "2025-01-27T10:45:00Z"}
```

**Performance**: Updates appear within 5 seconds of status change (SC-004)

---

## Kitchen Display Routes

### GET /kitchen

**Purpose**: Kitchen display showing all orders in chronological order.

**Response**: HTML page (200 OK)

**HTML Structure**:
```html
<!DOCTYPE html>
<html>
<head>
    <title>Kitchen Display</title>
</head>
<body>
    <main>
        <h1>Kitchen Orders</h1>
        <div id="orders-list" ds-sse-connect="/kitchen/stream">
            {# For each order #}
            <article class="order-card" data-order-id="{OrderID}">
                <h2>Order #{OrderNumber}</h2>
                <p>Status: {FulfillmentStatus}</p>
                <p>Payment: {PaymentStatus}</p>
                <p>Time: {CreatedAt}</p>
                
                <ul>
                    {# For each item #}
                    <li>{Quantity}x {Name}</li>
                    {# End for #}
                </ul>
                
                <p>Total: {TotalAmount}</p>
                
                {# If payment status is pending_payment #}
                <form method="POST" action="/kitchen/order/{OrderID}/mark-paid" ds-post="/kitchen/order/{OrderID}/mark-paid" ds-target="closest article">
                    <button type="submit">Mark as Paid</button>
                </form>
                {# End if #}
                
                {# If payment status is paid and fulfillment status is paid #}
                <form method="POST" action="/kitchen/order/{OrderID}/mark-preparing" ds-post="/kitchen/order/{OrderID}/mark-preparing" ds-target="closest article">
                    <button type="submit">Start Preparing</button>
                </form>
                {# End if #}
                
                {# If fulfillment status is preparing #}
                <form method="POST" action="/kitchen/order/{OrderID}/mark-ready" ds-post="/kitchen/order/{OrderID}/mark-ready" ds-target="closest article">
                    <button type="submit">Mark as Ready</button>
                </form>
                {# End if #}
            </article>
            {# End for #}
        </div>
    </main>
</body>
</html>
```

**Performance**: New orders appear within 5 seconds (SC-005)

---

### GET /kitchen/stream

**Purpose**: Server-Sent Events stream for real-time kitchen order updates.

**Response**: text/event-stream (200 OK)

**Event Format**:
```
event: new-order
data: {"orderId": "order-123", "orderNumber": "ORD-001", "items": [...], "totalAmount": 25.50}

event: order-updated
data: {"orderId": "order-123", "paymentStatus": "paid", "fulfillmentStatus": "preparing"}
```

---

### POST /kitchen/order/:orderID/mark-paid

**Purpose**: Mark order as paid when cash is received.

**Response**: HTML fragment (200 OK) - updated order card

**Error Responses**:
- 404: Order not found
- 400: Invalid state transition

---

### POST /kitchen/order/:orderID/mark-preparing

**Purpose**: Mark order as preparing.

**Response**: HTML fragment (200 OK) - updated order card

**Error Responses**:
- 404: Order not found
- 400: Invalid state transition (payment must be confirmed)

---

### POST /kitchen/order/:orderID/mark-ready

**Purpose**: Mark order as ready for pickup.

**Response**: HTML fragment (200 OK) - updated order card

**Error Responses**:
- 404: Order not found
- 400: Invalid state transition

---

## Owner Dashboard Routes

### GET /dashboard

**Purpose**: Owner dashboard showing sales, orders, and analytics.

**Response**: HTML page (200 OK)

**HTML Structure**:
```html
<!DOCTYPE html>
<html>
<head>
    <title>Dashboard</title>
</head>
<body>
    <main>
        <h1>Dashboard</h1>
        
        <section class="stats">
            <div>
                <h2>Orders Today</h2>
                <p>{OrdersToday}</p>
            </div>
            <div>
                <h2>Total Sales</h2>
                <p>{TotalSales}</p>
            </div>
            <div>
                <h2>Average Order Value</h2>
                <p>{AverageOrderValue}</p>
            </div>
        </section>
        
        <section class="orders">
            <h2>Recent Orders</h2>
            <table>
                <thead>
                    <tr>
                        <th>Time</th>
                        <th>Order #</th>
                        <th>Items</th>
                        <th>Amount</th>
                        <th>Payment</th>
                        <th>Status</th>
                    </tr>
                </thead>
                <tbody>
                    {# For each order #}
                    <tr>
                        <td>{CreatedAt}</td>
                        <td>{OrderNumber}</td>
                        <td>{ItemsSummary}</td>
                        <td>{TotalAmount}</td>
                        <td>{PaymentStatus}</td>
                        <td>{FulfillmentStatus}</td>
                    </tr>
                    {# End for #}
                </tbody>
            </table>
        </section>
        
        <section class="top-items">
            <h2>Top Selling Items</h2>
            <ol>
                {# For each top item #}
                <li>{Name} - {QuantitySold} sold - {Revenue}</li>
                {# End for #}
            </ol>
        </section>
        
        <section class="restaurant-controls">
            <form method="POST" action="/dashboard/toggle-open">
                <button type="submit">
                    {# If restaurant is open #}
                    Close Restaurant
                    {# Else #}
                    Open Restaurant
                    {# End if #}
                </button>
            </form>
        </section>
    </main>
</body>
</html>
```

---

### POST /dashboard/toggle-open

**Purpose**: Toggle restaurant open/closed status.

**Response**: HTML page (200 OK) - updated dashboard

---

### GET /dashboard/menu

**Purpose**: Menu management interface for owners.

**Response**: HTML page (200 OK) - menu editing interface

---

### POST /dashboard/menu/category

**Purpose**: Create menu category.

**Request**: Form data
- `name` (string, required): Category name
- `displayOrder` (int, optional): Display order

**Response**: HTML fragment (200 OK) - updated menu structure

---

### POST /dashboard/menu/item

**Purpose**: Create menu item.

**Request**: Form data (multipart/form-data)
- `categoryId` (string, required): Category ID
- `name` (string, required): Item name
- `description` (string, optional): Item description
- `price` (decimal, required): Item price
- `photo` (file, optional): Photo file (max 2MB)

**Response**: HTML fragment (200 OK) - updated menu structure

**Performance**: Item appears within 30 seconds, photo within 10 seconds

---

## Error Pages

### 404 Not Found

**Response**: HTML page (404)

```html
<!DOCTYPE html>
<html>
<head>
    <title>Not Found</title>
</head>
<body>
    <main>
        <h1>404 - Not Found</h1>
        <p>The requested resource was not found.</p>
        <a href="/menu">Return to Menu</a>
    </main>
</body>
</html>
```

### 400 Bad Request

**Response**: HTML page (400)

```html
<!DOCTYPE html>
<html>
<head>
    <title>Bad Request</title>
</head>
<body>
    <main>
        <h1>400 - Bad Request</h1>
        <p>{ErrorMessage}</p>
        <a href="javascript:history.back()">Go Back</a>
    </main>
</body>
</html>
```

### 503 Service Unavailable

**Response**: HTML page (503)

```html
<!DOCTYPE html>
<html>
<head>
    <title>Service Unavailable</title>
</head>
<body>
    <main>
        <h1>503 - Service Unavailable</h1>
        <p>{ErrorMessage}</p>
        {# If restaurant closed #}
        <p>Reopening: {ReopeningHours}</p>
        {# End if #}
    </main>
</body>
</html>
```

---

## Hypermedia Interactions

### Form Submissions
- All forms use standard HTML form submission
- Datastar attributes (ds-post, ds-target) enable partial page updates
- No JavaScript required for basic functionality - Datastar handles DOM updates automatically

### Links
- Standard HTML anchor tags for navigation
- Relative URLs preferred for portability

### Real-Time Updates
- Server-Sent Events (SSE) for order status updates via Datastar
- SSE connection established via Datastar `ds-sse-connect` attribute pointing to `/order/:orderNumber/stream` or `/kitchen/stream`
- Datastar automatically receives SSE events and updates DOM without page refresh
- No JavaScript required - Datastar handles SSE connection and DOM updates

---

## Performance Requirements

- **Menu page load**: <2 seconds on 3G (SC-003)
- **Order status updates**: <5 seconds propagation (SC-004, SC-005)
- **API endpoints**: <200ms p95 latency (Constitution)
- **Critical flow**: <2 minutes total (SC-001)

---

## Accessibility

- Semantic HTML5 elements
- WCAG 2.1 AA compliance
- Keyboard navigation support
- Screen reader friendly
- Alt text for images

---

## Browser Support

- Chrome (latest)
- Safari (latest)
- Firefox (latest)
- Edge (latest)
- Mobile browsers (iOS Safari, Chrome Mobile)

---

## References

- **Templ Documentation**: https://templ.guide/
- **Templ LLM Guide**: https://templ.guide/llms.md
- **Templ UI Components**: https://templui.io/
- **Datastar**: https://github.com/delaneyj/datastar - Hypermedia-driven UI with SSE for Go/Templ (REQUIRED - NOT HTMX)
- **Server-Sent Events**: https://html.spec.whatwg.org/multipage/server-sent-events.html


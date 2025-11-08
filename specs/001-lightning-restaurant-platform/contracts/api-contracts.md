# API Contracts: BitMerchant v1.0

**Date**: 2025-11-08  
**Feature**: BitMerchant v1.0 - Lightning Payment Platform for Restaurants  
**Format**: OpenAPI 3.0 (simplified for documentation)

## Base URL

- **Customer Menu**: `https://{restaurant-slug}.bitmerchant.app` or `https://bitmerchant.app/{restaurant-id}`
- **Kitchen Display**: `https://bitmerchant.app/kitchen/{restaurant-id}`
- **Owner Dashboard**: `https://bitmerchant.app/dashboard/{restaurant-id}`

## Authentication

- **Customer Menu**: No authentication required (public)
- **Kitchen Display**: Session-based (simple token or cookie)
- **Owner Dashboard**: Session-based (simple token or cookie)

---

## Customer Menu Endpoints

### GET /menu

**Purpose**: Display restaurant menu with categories and items (FR-001, FR-002)

**Response**: HTML (Templ template) or JSON

**JSON Response**:
```json
{
  "restaurant": {
    "id": "rest_123",
    "name": "Cafe Lightning",
    "isOpen": true,
    "closedMessage": null,
    "reopeningHours": null
  },
  "categories": [
    {
      "id": "cat_001",
      "name": "Appetizers",
      "displayOrder": 1,
      "items": [
        {
          "id": "item_001",
          "name": "Garlic Bread",
          "description": "Fresh baked with garlic butter",
          "price": 5.99,
          "photoUrl": "https://cdn.example.com/photo_001_300kb.jpg",
          "isAvailable": true
        }
      ]
    }
  ]
}
```

**Status Codes**:
- `200 OK`: Menu loaded successfully
- `404 Not Found`: Restaurant not found
- `503 Service Unavailable`: Restaurant closed (if IsOpen is false)

**Performance**: <2 seconds on 3G (SC-003)

---

### POST /cart/add

**Purpose**: Add item to shopping cart (FR-003)

**Request Body**:
```json
{
  "itemId": "item_001",
  "quantity": 2
}
```

**Response**:
```json
{
  "cart": {
    "items": [
      {
        "itemId": "item_001",
        "name": "Garlic Bread",
        "quantity": 2,
        "unitPrice": 5.99,
        "subtotal": 11.98
      }
    ],
    "total": 11.98
  }
}
```

**Status Codes**:
- `200 OK`: Item added to cart
- `400 Bad Request`: Invalid itemId or quantity
- `404 Not Found`: Item not found or not available
- `503 Service Unavailable`: Restaurant closed

---

### POST /cart/remove

**Purpose**: Remove item from cart or adjust quantity (FR-003)

**Request Body**:
```json
{
  "itemId": "item_001",
  "quantity": 0  // 0 removes item, >0 updates quantity
}
```

**Response**: Same as `/cart/add`

---

### GET /cart

**Purpose**: Get current cart contents (FR-003)

**Response**: Same as `/cart/add` response

---

### POST /payment/create-invoice

**Purpose**: Generate Lightning Network invoice for cart (FR-004, FR-033)

**Request Body**:
```json
{
  "items": [
    {
      "itemId": "item_001",
      "quantity": 2
    }
  ]
}
```

**Response**:
```json
{
  "invoiceId": "inv_strike_abc123",
  "invoice": "lnbc1234567890...",
  "qrCode": "data:image/png;base64,iVBORw0KG...",
  "amountSatoshis": 12345,
  "amountFiat": 11.98,
  "exchangeRate": 0.0001031,
  "expiresAt": "2025-11-08T12:15:00Z"
}
```

**Status Codes**:
- `200 OK`: Invoice created
- `400 Bad Request`: Empty cart or invalid items
- `503 Service Unavailable`: Restaurant closed

**Performance**: Invoice generation <200ms (Constitution requirement)

---

### GET /payment/status/{invoiceId}

**Purpose**: Poll payment status (FR-005, FR-030)

**Response**:
```json
{
  "invoiceId": "inv_strike_abc123",
  "status": "paid",  // pending | paid | failed | expired
  "orderId": "ord_001",
  "orderNumber": "ORD-001",
  "paidAt": "2025-11-08T12:10:00Z"
}
```

**Status Codes**:
- `200 OK`: Payment status retrieved
- `404 Not Found`: Invoice not found

**Polling**: Client polls every 2-3 seconds until status is "paid" or "failed" (detect within 10s per SC-002)

---

### GET /order/{orderNumber}

**Purpose**: Look up order by order number (FR-037)

**Response**:
```json
{
  "orderId": "ord_001",
  "orderNumber": "ORD-001",
  "status": "preparing",  // paid | preparing | ready | completed
  "items": [
    {
      "itemId": "item_001",
      "name": "Garlic Bread",
      "quantity": 2,
      "unitPrice": 5.99
    }
  ],
  "total": 11.98,
  "createdAt": "2025-11-08T12:10:00Z",
  "updatedAt": "2025-11-08T12:12:00Z"
}
```

**Status Codes**:
- `200 OK`: Order found
- `404 Not Found`: Order not found

---

### GET /order/{orderNumber}/stream

**Purpose**: Server-Sent Events (SSE) stream for real-time order status updates (FR-007, FR-014)

**Response**: `text/event-stream`

**Event Format**:
```
event: order-status
data: {"orderNumber":"ORD-001","status":"ready","updatedAt":"2025-11-08T12:15:00Z"}

event: order-status
data: {"orderNumber":"ORD-001","status":"completed","updatedAt":"2025-11-08T12:20:00Z"}
```

**Status Codes**:
- `200 OK`: SSE stream established
- `404 Not Found`: Order not found

**Performance**: Status updates appear within 5 seconds (SC-004)

---

## Kitchen Display Endpoints

### GET /kitchen/{restaurantId}

**Purpose**: Kitchen display view with active orders (FR-010, FR-011)

**Response**: HTML (Templ template) or JSON

**JSON Response**:
```json
{
  "restaurant": {
    "id": "rest_123",
    "name": "Cafe Lightning"
  },
  "orders": [
    {
      "orderId": "ord_001",
      "orderNumber": "ORD-001",
      "status": "paid",  // paid | preparing | ready
      "items": [
        {
          "itemId": "item_001",
          "name": "Garlic Bread",
          "quantity": 2
        }
      ],
      "total": 11.98,
      "createdAt": "2025-11-08T12:10:00Z"
    }
  ]
}
```

**Status Codes**:
- `200 OK`: Kitchen display loaded
- `404 Not Found`: Restaurant not found

**Performance**: New orders appear within 5 seconds (SC-005)

---

### GET /kitchen/{restaurantId}/stream

**Purpose**: Server-Sent Events (SSE) stream for real-time new orders (FR-012, FR-014)

**Response**: `text/event-stream`

**Event Format**:
```
event: new-order
data: {"orderId":"ord_002","orderNumber":"ORD-002","items":[...],"createdAt":"2025-11-08T12:20:00Z"}

event: order-status-changed
data: {"orderId":"ord_001","orderNumber":"ORD-001","status":"preparing","updatedAt":"2025-11-08T12:25:00Z"}
```

**Status Codes**:
- `200 OK`: SSE stream established
- `404 Not Found`: Restaurant not found

---

### POST /kitchen/order/{orderId}/status

**Purpose**: Update order fulfillment status (FR-013)

**Request Body**:
```json
{
  "status": "preparing"  // preparing | ready
}
```

**Response**:
```json
{
  "orderId": "ord_001",
  "orderNumber": "ORD-001",
  "status": "preparing",
  "updatedAt": "2025-11-08T12:25:00Z"
}
```

**Status Codes**:
- `200 OK`: Status updated
- `400 Bad Request`: Invalid status transition
- `404 Not Found`: Order not found

**Business Rules**:
- Valid transitions: `paid → preparing → ready`
- Status change triggers SSE event to customer (FR-014)

---

## Owner Dashboard Endpoints

### POST /dashboard/restaurant

**Purpose**: Create restaurant account (FR-017)

**Request Body**:
```json
{
  "name": "Cafe Lightning",
  "lightningAddress": "cafe@strike.me"
}
```

**Response**:
```json
{
  "restaurantId": "rest_123",
  "name": "Cafe Lightning",
  "lightningAddress": "cafe@strike.me",
  "createdAt": "2025-11-08T10:00:00Z"
}
```

**Status Codes**:
- `201 Created`: Restaurant created
- `400 Bad Request`: Invalid name or Lightning address

---

### POST /dashboard/restaurant/{restaurantId}/status

**Purpose**: Toggle restaurant open/closed status (FR-039)

**Request Body**:
```json
{
  "isOpen": false,
  "closedMessage": "Closed for holiday",
  "reopeningHours": "Tomorrow 9 AM"
}
```

**Response**:
```json
{
  "restaurantId": "rest_123",
  "isOpen": false,
  "closedMessage": "Closed for holiday",
  "reopeningHours": "Tomorrow 9 AM",
  "updatedAt": "2025-11-08T18:00:00Z"
}
```

**Status Codes**:
- `200 OK`: Status updated
- `404 Not Found`: Restaurant not found

---

### POST /dashboard/menu/category

**Purpose**: Create menu category (FR-018)

**Request Body**:
```json
{
  "restaurantId": "rest_123",
  "name": "Appetizers",
  "displayOrder": 1
}
```

**Response**:
```json
{
  "categoryId": "cat_001",
  "name": "Appetizers",
  "displayOrder": 1,
  "createdAt": "2025-11-08T10:05:00Z"
}
```

---

### POST /dashboard/menu/item

**Purpose**: Create menu item (FR-019, FR-020)

**Request Body** (multipart/form-data):
```
categoryId: cat_001
name: Garlic Bread
description: Fresh baked with garlic butter
price: 5.99
photo: [file, max 2MB]
```

**Response**:
```json
{
  "itemId": "item_001",
  "name": "Garlic Bread",
  "description": "Fresh baked with garlic butter",
  "price": 5.99,
  "photoUrl": "https://cdn.example.com/photo_001_300kb.jpg",
  "createdAt": "2025-11-08T10:10:00Z"
}
```

**Status Codes**:
- `201 Created`: Item created
- `400 Bad Request`: Invalid data or photo >2MB
- `403 Forbidden`: Restaurant reached 100 photo limit (FR-042)

**Validation**:
- Photo max 2MB (FR-020)
- Restaurant max 100 photos (FR-042)
- Photo automatically compressed to 300KB (FR-020)

---

### PUT /dashboard/menu/item/{itemId}

**Purpose**: Update menu item (FR-021)

**Request Body**: Same as POST, all fields optional

**Response**: Updated item JSON

**Status Codes**:
- `200 OK`: Item updated
- `404 Not Found`: Item not found

---

### GET /dashboard/restaurant/{restaurantId}/qr

**Purpose**: Generate customer-facing QR code (FR-022)

**Response**: 
- HTML: QR code image + shareable link
- JSON: `{"qrCode": "data:image/png;base64,...", "url": "https://..."}`

---

### GET /dashboard/restaurant/{restaurantId}/analytics

**Purpose**: Get sales dashboard (FR-024, FR-025, FR-026, FR-027)

**Query Parameters**:
- `startDate` (optional): Start date for date range
- `endDate` (optional): End date for date range

**Response**:
```json
{
  "ordersToday": 25,
  "totalSalesBTC": 0.001234,
  "totalSalesFiat": 299.75,
  "averageOrderValue": 11.99,
  "topItems": [
    {
      "itemId": "item_001",
      "name": "Garlic Bread",
      "quantitySold": 50,
      "revenue": 299.50
    }
  ],
  "settlementStatus": {
    "status": "pending",  // pending | settled
    "amount": 0.001234,
    "destinationAddress": "cafe@strike.me",
    "settledAt": null
  }
}
```

---

## Error Responses

All endpoints return consistent error format:

```json
{
  "error": {
    "code": "PAYMENT_FAILED",
    "message": "Lightning payment failed. Please try again.",
    "details": "Invoice expired after 15 minutes"
  }
}
```

**Error Codes**:
- `RESTAURANT_NOT_FOUND`: Restaurant ID invalid
- `ITEM_NOT_FOUND`: Menu item not found
- `ITEM_UNAVAILABLE`: Item out of stock
- `RESTAURANT_CLOSED`: Restaurant is closed
- `PAYMENT_FAILED`: Lightning payment failed
- `INVOICE_EXPIRED`: Lightning invoice expired
- `PHOTO_LIMIT_REACHED`: Restaurant reached 100 photo limit
- `INVALID_REQUEST`: Request validation failed

---

## Performance Requirements

- **API Endpoints**: <200ms p95 latency (Constitution)
- **Menu Load**: <2 seconds on 3G (SC-003)
- **Payment Status**: Detect completion within 10 seconds (SC-002)
- **Order Status Updates**: <5 seconds propagation (SC-004, SC-005)

---

## Real-Time Updates (SSE)

All SSE endpoints:
- Use HTTP/2 for better performance
- Send `keep-alive` events every 30 seconds
- Close connection on error or client disconnect
- Support reconnection with last event ID

**Client Implementation**:
- Use EventSource API or Datastar client
- Handle reconnection automatically
- Update DOM based on event data

---

## Notes

- All prices displayed in local fiat currency, payments processed in Bitcoin (FR-032)
- Exchange rates snapshot at invoice generation (FR-033)
- Orders only created when payment status is "paid" (FR-029)
- Menu changes reflect in customer view within 5 seconds (FR-023)


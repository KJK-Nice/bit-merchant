# Feature Specification: Cash Payment with Hypermedia UI

**Feature Branch**: `002-cash-payment-hypermedia`  
**Created**: 2025-01-27  
**Status**: Draft  
**Input**: User description: "I want to change payment method to focus on accepting cash first. you can remove everything about Strike API. Also on about our http routes do we need to implement JSON rest? Can we change to build hypermedia UI directly with HTML-returning handlers?"

## Clarifications

### Session 2025-01-27

- Q: How should the architecture support future Lightning payment integration? → A: Design payment method abstraction now (cash-only initially, Lightning added later as new payment type) - architecture supports multiple payment methods from start

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Customer Orders Food with Cash Payment (Priority: P1) 🎯

Sarah walks into a cafe, sees a QR code on the table, and wants to order food. She browses the menu, adds items to her cart, and completes her order by confirming cash payment. She receives an order number and can track her order status in real-time.

**Why this priority**: This is the core value proposition - customers can order food and pay with cash without creating an account or sharing personal information. Without this working flawlessly, the entire platform has no purpose. This delivers immediate value to both customers (fast ordering, simple payment) and restaurants (streamlined order management).

**Independent Test**: Can be fully tested by: (1) scanning restaurant QR code, (2) browsing menu, (3) adding items to cart, (4) confirming cash payment, (5) receiving order confirmation with real-time status updates. Delivers complete end-to-end ordering experience without requiring any other features.

**Acceptance Scenarios**:

1. **Given** customer scans QR code on table, **When** code is scanned, **Then** restaurant menu loads instantly (under 2 seconds) with categories, items, prices, and photos displayed as HTML page
2. **Given** customer is viewing menu, **When** customer taps food items, **Then** items are added to cart with running total displayed, and page updates without full reload
3. **Given** customer has items in cart, **When** customer taps "Place Order", **Then** order confirmation page appears showing order summary, total amount, and order number
4. **Given** order confirmation page is displayed, **When** customer confirms "I will pay with cash", **Then** order is created with status "Pending Payment" and customer sees order number
5. **Given** order is created, **When** order is created, **Then** customer sees order number and real-time status: "Pending Payment → Paid → Preparing → Ready"
6. **Given** order status changes in kitchen, **When** status updates to "Ready", **Then** customer receives real-time update without needing to refresh page
7. **Given** customer has completed order, **When** customer pays cash to staff and staff marks order as paid, **Then** customer sees status update to "Paid" and then "Preparing"

---

### User Story 2 - Kitchen Staff Fulfills Orders (Priority: P2)

Marcus works in the kitchen and needs to see incoming orders, mark payments received, prepare food, and notify customers when orders are ready - all through a simple HTML interface without handling complex POS systems.

**Why this priority**: Without kitchen fulfillment, customers receive order confirmations but never get their food. This completes the operational loop and enables restaurants to actually use the system. It must be simple enough that staff can learn it in under 5 minutes.

**Independent Test**: Can be tested by: (1) kitchen staff viewing kitchen display HTML page, (2) seeing new orders appear automatically, (3) marking orders as "Paid" when cash received, (4) marking orders as "Preparing", (5) marking orders as "Ready", (6) verifying customers receive real-time updates. Delivers complete kitchen workflow independently from menu setup or analytics.

**Acceptance Scenarios**:

1. **Given** kitchen display HTML page is open, **When** customer completes order, **Then** new order appears on kitchen page with audible alert within 5 seconds
2. **Given** new order appears, **When** staff views order, **Then** order shows: order number, items with quantities, total amount, timestamp, payment status
3. **Given** customer pays cash to staff, **When** staff marks order as "Paid", **Then** order status changes to "Paid" and customer sees update in real-time
4. **Given** staff starts preparing food, **When** staff taps order, **Then** order status changes to "Preparing" and customer sees update in real-time
5. **Given** food is ready, **When** staff taps "Mark Ready", **Then** order status changes to "Ready" and customer receives real-time update
6. **Given** order is marked ready, **When** customer picks up food, **Then** order moves to completed queue and disappears from active orders
7. **Given** multiple orders exist, **When** staff views kitchen display, **Then** orders are sorted by time received (oldest first) with clear visual priority

---

### User Story 3 - Owner Sets Up Restaurant Menu (Priority: P3)

Linda owns a small restaurant and wants to accept orders through a simple system. She needs to set up her menu and start taking orders within 10 minutes without technical knowledge, using an HTML-based interface.

**Why this priority**: Without menu setup, there's nothing for customers to order. However, this is P3 because restaurants typically set up menus once and update infrequently. The critical path is customer ordering → kitchen fulfillment. Menu setup is a prerequisite but not part of the transaction flow.

**Independent Test**: Can be tested by: (1) owner signing up, (2) entering restaurant name, (3) creating menu categories, (4) adding items with names, descriptions, prices, photos, (5) generating customer-facing QR code. Delivers complete menu management independently from order processing.

**Acceptance Scenarios**:

1. **Given** restaurant owner visits signup page, **When** owner enters restaurant name, **Then** account is created in under 2 minutes with no email verification
2. **Given** owner has account, **When** owner creates menu categories (Appetizers, Mains, Desserts, Drinks), **Then** categories appear in menu structure immediately on HTML page
3. **Given** category exists, **When** owner adds menu item with name, description, price (in local currency), **Then** item appears in category within 30 seconds
4. **Given** menu item exists, **When** owner uploads photo, **Then** photo is optimized and displayed in menu within 10 seconds
5. **Given** menu is complete, **When** owner generates customer QR code, **Then** printable QR code and shareable link are provided
6. **Given** menu changes needed, **When** owner edits prices or adds items, **Then** changes appear in customer-facing menu within 5 seconds
7. **Given** owner wants to temporarily disable item, **When** owner marks item as "out of stock", **Then** item is hidden from customer menu but retained in system

---

### User Story 4 - Owner Views Sales Dashboard (Priority: P4)

Linda needs to see daily sales, order count, top-selling items, and payment status to understand her business performance through a simple HTML dashboard.

**Why this priority**: Basic analytics are essential for business operation but not for MVP functionality. Restaurants can operate for days with just orders flowing through. Analytics become important after the system proves reliable, making this lowest priority for initial launch.

**Independent Test**: Can be tested by: (1) owner logging into dashboard HTML page, (2) viewing today's sales and order count, (3) seeing average order value, (4) viewing top-selling items, (5) confirming payment status. Delivers business insights independently from transaction processing.

**Acceptance Scenarios**:

1. **Given** owner logs into dashboard, **When** dashboard HTML page loads, **Then** owner sees: orders today, total sales, average order value
2. **Given** orders have been placed, **When** owner views order history, **Then** orders show: time, items, amount, payment status, order status
3. **Given** multiple items sold, **When** owner views "Top Items", **Then** items are ranked by quantity sold with revenue per item
4. **Given** owner wants daily summary, **When** owner views dashboard, **Then** daily sales totals and order counts display for current day
5. **Given** owner wants weekly summary, **When** owner selects date range, **Then** sales trends and totals display for selected period

---

### Edge Cases

- **What happens when customer abandons order before confirming cash payment?** System must allow customer to return to cart, modify items, and place order again. Unconfirmed orders are not created in system.
- **What happens when customer's network drops after ordering?** Order status must be retrievable via order number lookup (no account needed), real-time updates resume when network reconnects
- **What happens when kitchen HTML page loses connection?** Orders must queue and sync automatically when connection restored, with clear visual indicator of offline status
- **What happens when menu item has no photo?** System must show placeholder image or text-only card that remains visually consistent with menu design
- **What happens when two customers order at same time?** System must handle concurrent orders without conflicts, preserving order sequence by timestamp
- **What happens when customer abandons cart before payment?** Cart is ephemeral (session-only), no persistent cart storage, no cleanup needed
- **What happens when restaurant needs to refund order?** Refunds are handled manually by restaurant owner - owner processes cash refund directly. System does not track or process refunds in v1.0. Owner can view order details to understand refund context.
- **What happens when owner wants to close restaurant temporarily?** System displays "Currently Closed" banner on menu HTML page with custom message and expected reopening hours. Menu remains visible for browsing but ordering is disabled. Owner can toggle open/closed status from dashboard.
- **What happens to photos if restaurant hits storage limits?** System enforces minimal limits: maximum 2MB per photo upload, maximum 100 photos per restaurant. All photos are automatically compressed to 300KB optimized version for display. If restaurant reaches 100 photo limit, owner must delete existing photos before uploading new ones.
- **What happens when staff marks order as paid but customer hasn't paid yet?** System allows staff to mark orders as paid immediately. If mistake occurs, staff can manually adjust order status. System does not validate actual cash receipt - relies on staff verification.
- **What happens when customer confirms cash payment but never pays?** Order remains in "Pending Payment" status indefinitely. Kitchen staff can see unpaid orders and follow up with customers. Owner can cancel unpaid orders after reasonable time period.

## Requirements *(mandatory)*

### Functional Requirements

**Customer Ordering:**
- **FR-001**: System MUST display restaurant menu as HTML page organized by categories (Appetizers, Mains, Desserts, Drinks) with item names, descriptions, prices, and photos
- **FR-002**: System MUST allow customers to browse menu without account creation, login, or personal information
- **FR-003**: System MUST provide shopping cart where customers can add/remove items, adjust quantities, and see running total, with page updates via hypermedia (no full page reload)
- **FR-004**: System MUST display order confirmation HTML page when customer initiates order, showing order summary, total amount, and order number
- **FR-005**: System MUST allow customer to confirm cash payment, creating order with status "Pending Payment" immediately upon confirmation
- **FR-006**: System MUST assign unique order number to each order for tracking and kitchen reference
- **FR-007**: System MUST provide real-time order status updates (Pending Payment → Paid → Preparing → Ready) visible to customer without page refresh using hypermedia techniques
- **FR-008**: System MUST send real-time update to customer when order status changes to "Ready"
- **FR-009**: System MUST work as Progressive Web App (PWA) - installable to home screen, works offline for menu browsing

**Kitchen Operations:**
- **FR-010**: System MUST display all orders on kitchen display HTML page in chronological order (oldest first)
- **FR-011**: System MUST show order details: order number, items with quantities, total amount, timestamp, payment status
- **FR-012**: System MUST provide audible/visual alert when new order arrives
- **FR-013**: System MUST allow kitchen staff to mark orders as "Paid" when cash is received
- **FR-014**: System MUST allow kitchen staff to mark orders as "Preparing" or "Ready" with single action
- **FR-015**: System MUST automatically update customer's order status in real-time when kitchen changes order state
- **FR-016**: System MUST move completed orders (marked Ready and picked up) to archived queue after 1 hour
- **FR-017**: System MUST handle kitchen display HTML page offline - queue status changes and sync when connection restored

**Restaurant Management:**
- **FR-018**: System MUST allow restaurant owner to create account with minimal information: restaurant name
- **FR-019**: System MUST allow owner to create menu categories and add items within each category through HTML interface
- **FR-020**: System MUST accept menu item details: name (required), description (optional), price in local currency (required), photo (optional)
- **FR-021**: System MUST optimize and store uploaded photos with limits: maximum 2MB per photo upload, maximum 100 photos per restaurant, automatic compression to 300KB optimized version for display
- **FR-022**: System MUST allow owner to edit menu items, update prices, mark items as out of stock through HTML interface
- **FR-023**: System MUST generate customer-facing QR code linking to restaurant menu for printing or sharing
- **FR-024**: System MUST reflect menu changes in customer-facing HTML view within 5 seconds
- **FR-025**: System MUST provide HTML dashboard showing: orders today, total sales, average order value
- **FR-026**: System MUST display order history with: timestamp, items, amount, payment status, fulfillment status
- **FR-027**: System MUST show top-selling items ranked by quantity sold
- **FR-028**: System MUST allow owner to toggle restaurant open/closed status from dashboard
- **FR-029**: System MUST display "Currently Closed" banner on customer menu HTML page when restaurant is closed, showing custom message and expected reopening hours
- **FR-030**: System MUST disable ordering functionality when restaurant is marked as closed, while keeping menu visible for browsing

**Payment Processing:**
- **FR-031**: System MUST support payment method abstraction architecture - designed to support multiple payment types (cash initially, Lightning Network in future) without requiring refactoring
- **FR-032**: System MUST support cash payment confirmation flow - customer confirms intent to pay cash, order created with "Pending Payment" status and payment method type "cash"
- **FR-034**: System MUST validate that orders marked as "Paid" can proceed to "Preparing" status
- **FR-035**: System MUST display prices in local currency for customer clarity
- **FR-036**: System MUST track payment status for each order: Pending Payment, Paid, Not Paid (if cancelled)
- **FR-037**: System MUST track payment method type for each order (e.g., "cash", future: "lightning") to support multiple payment methods

**Hypermedia UI:**
- **FR-038**: System MUST return HTML pages for all customer-facing routes instead of JSON responses
- **FR-039**: System MUST return HTML pages for all kitchen display routes instead of JSON responses
- **FR-040**: System MUST return HTML pages for all owner dashboard routes instead of JSON responses
- **FR-041**: System MUST support hypermedia-driven interactions - form submissions, links, and real-time updates without requiring JavaScript frameworks
- **FR-042**: System MUST provide real-time updates using Server-Sent Events (SSE) or similar hypermedia techniques for order status changes
- **FR-043**: System MUST allow page updates (cart, order status) without full page reloads using hypermedia techniques

**System Reliability:**
- **FR-044**: System MUST handle concurrent orders from multiple customers without conflicts
- **FR-045**: System MUST maintain order sequence integrity based on order creation timestamp
- **FR-046**: System MUST log all order creation, payment status changes, and fulfillment status changes for troubleshooting
- **FR-047**: System MUST provide order lookup by order number for customers who lost connection (no account needed)
- **FR-048**: System MUST continue functioning if external services (photos, analytics) temporarily fail, displaying placeholder images for missing photos and disabling analytics views gracefully
- **FR-049**: System MUST prevent new photo uploads when restaurant reaches 100 photo limit, requiring owner to delete existing photos first

### Key Entities

- **Restaurant**: Represents single restaurant tenant with name, menu structure, owner credentials, creation timestamp, open/closed status
- **Menu Category**: Logical grouping of menu items (e.g., Appetizers, Mains) with name, display order, active status
- **Menu Item**: Food/drink item with name, description, price (local currency), photo URL, category association, availability status (in stock / out of stock)
- **Order**: Customer purchase record with unique order number, timestamp, items with quantities, total amount (local currency), payment method type (e.g., "cash", future: "lightning"), payment status (pending payment/paid/not paid), fulfillment status (pending payment/paid/preparing/ready/completed)
- **Cart**: Ephemeral session-only object containing selected items and quantities - not persisted to database

### Assumptions

- Photos will be stored in cloud storage (assumed AWS S3 or similar) with CDN delivery
- Single restaurant owner per account initially (no multi-user management)
- Order modifications/cancellations not supported in v1.0 (aligns with simplicity principle)
- Refunds handled manually by owner - system does not process refunds
- Customer can have multiple active orders simultaneously (no limit)
- Kitchen display remains always-on during business hours
- Menu item modifications (extra cheese, no onions) will be added in future version - v1.0 is fixed items only
- System timezone set to restaurant's local timezone for all timestamps
- Photo storage limits: 2MB max upload, 100 photos max per restaurant, 300KB optimized display version
- Cash payment confirmation relies on customer honesty and staff verification - system does not validate actual cash receipt
- Staff can mark orders as paid immediately when cash is received - no separate payment verification step required
- HTML pages will use hypermedia techniques (forms, links, SSE) for interactions - minimal JavaScript required
- Server-rendered HTML templates will be used instead of JSON API endpoints
- Payment method abstraction architecture will be designed from the start to support multiple payment types (cash initially, Lightning Network in future) - Lightning support will be added as a new payment method type without requiring architectural refactoring

## Success Criteria *(mandatory)*

### Measurable Outcomes

**Speed & Performance:**
- **SC-001**: Customer completes entire ordering flow (browse → select → confirm cash payment → receive order number) in under 2 minutes on mobile device
- **SC-002**: Cash payment confirmation completes immediately upon customer confirmation (no external payment processing delay)
- **SC-003**: Menu HTML pages load in under 2 seconds on 3G mobile networks
- **SC-004**: Order status updates appear in customer view within 5 seconds of kitchen staff action
- **SC-005**: New orders appear on kitchen display HTML page within 5 seconds of order creation

**Setup & Usability:**
- **SC-006**: Non-technical restaurant owner completes full menu setup (categories + 20 items with photos) in under 10 minutes using HTML interface
- **SC-007**: Kitchen staff can fulfill orders confidently after 5-minute demonstration with no additional training
- **SC-008**: Customer can order food without instructions, documentation, or staff assistance
- **SC-009**: All user interactions work without requiring JavaScript frameworks - hypermedia-driven UI functions with standard HTML forms and links

**Adoption & Volume:**
- **SC-010**: 10 restaurants actively using system daily within 3 months of launch
- **SC-011**: System processes 1,000+ successful orders
- **SC-012**: Each restaurant processes average of 20+ orders per day after first week

**Reliability:**
- **SC-013**: 99% system uptime measured across all components (menu display, order processing, kitchen display)
- **SC-014**: Zero orders lost due to system errors (all orders appear in kitchen queue)
- **SC-015**: HTML pages render correctly across major browsers (Chrome, Safari, Firefox) on mobile and desktop

**User Satisfaction:**
- **SC-016**: After first successful order, customers ask "How do I install this on my phone?" (PWA adoption indicator)
- **SC-017**: Restaurant owners demonstrate system to other business owners within first week (word-of-mouth indicator)
- **SC-018**: Kitchen staff prefer BitMerchant over previous POS system for order management (qualitative survey)
- **SC-019**: Customers describe ordering experience as "faster than traditional ordering" (qualitative feedback)
- **SC-020**: Restaurant owners report simplified order management compared to previous systems (qualitative feedback)

**Business Impact:**
- **SC-021**: Restaurant operates full business day (breakfast through dinner) using only BitMerchant with no additional tools needed
- **SC-022**: Cash payment flow reduces payment processing complexity compared to digital payment integrations


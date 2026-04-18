# Feature Specification: BitMerchant v1.0 - Lightning Payment Platform for Restaurants

**Feature Branch**: `001-lightning-restaurant-platform`  
**Created**: 2025-11-08  
**Status**: Draft  
**Input**: User description: "BitMerchant v1.0: Lightning Payment Platform for Restaurants"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Customer Orders Food with Lightning Payment (Priority: P1) 🎯

Sarah walks into a cafe, sees a QR code on the table, and wants to order food using her Lightning wallet without creating an account or sharing personal information.

**Why this priority**: This is the core value proposition - instant Bitcoin payments with zero friction. Without this working flawlessly, the entire platform has no purpose. This delivers immediate value to both customers (fast, private payments) and restaurants (instant settlement, lower fees).

**Independent Test**: Can be fully tested by: (1) scanning restaurant QR code, (2) browsing menu, (3) adding items to cart, (4) paying with any Lightning wallet, (5) receiving order confirmation with real-time status updates. Delivers complete end-to-end ordering experience without requiring any other features.

**Acceptance Scenarios**:

1. **Given** customer scans QR code on table, **When** code is scanned, **Then** restaurant menu loads instantly (under 2 seconds) with categories, items, prices, and photos
2. **Given** customer is viewing menu, **When** customer taps food items, **Then** items are added to cart with running total displayed
3. **Given** customer has items in cart, **When** customer taps "Pay with Lightning", **Then** Lightning invoice QR code appears with amount in satoshis
4. **Given** Lightning QR code is displayed, **When** customer scans with any Lightning wallet and completes payment, **Then** payment confirmation appears within 10 seconds
5. **Given** payment is confirmed, **When** order is created, **Then** customer sees order number and real-time status: "Paid → Preparing → Ready"
6. **Given** order status changes in kitchen, **When** status updates to "Ready", **Then** customer receives push notification without needing to refresh
7. **Given** customer has completed order, **When** customer leaves restaurant, **Then** no additional interaction required (no checkout, no table number entry, no tip screen)

---

### User Story 2 - Kitchen Staff Fulfills Orders (Priority: P2)

Marcus works in the kitchen and needs to see incoming orders, prepare food, and notify customers when orders are ready - all without handling payments or learning complex POS systems.

**Why this priority**: Without kitchen fulfillment, customers receive payment confirmations but never get their food. This completes the operational loop and enables restaurants to actually use the system. It must be simple enough that staff can learn it in under 5 minutes.

**Independent Test**: Can be tested by: (1) kitchen staff viewing kitchen display tablet, (2) seeing new paid orders appear automatically, (3) marking orders as "Preparing", (4) marking orders as "Ready", (5) verifying customers receive notifications. Delivers complete kitchen workflow independently from menu setup or analytics.

**Acceptance Scenarios**:

1. **Given** kitchen display is open, **When** customer completes Lightning payment, **Then** new order appears on kitchen tablet with audible alert within 5 seconds
2. **Given** new order appears, **When** staff views order, **Then** order shows: order number, items with quantities, modifications, timestamp
3. **Given** staff starts preparing food, **When** staff taps order, **Then** order status changes to "Preparing" and customer sees update in real-time
4. **Given** food is ready, **When** staff taps "Mark Ready", **Then** order status changes to "Ready" and customer receives push notification
5. **Given** order is marked ready, **When** customer picks up food, **Then** order moves to completed queue and disappears from active orders
6. **Given** multiple orders exist, **When** staff views kitchen display, **Then** orders are sorted by time received (oldest first) with clear visual priority

---

### User Story 3 - Owner Sets Up Restaurant Menu (Priority: P3)

Linda owns a small restaurant and wants to accept Bitcoin payments but finds existing solutions too complex. She needs to set up her menu and start taking orders within 10 minutes without technical knowledge.

**Why this priority**: Without menu setup, there's nothing for customers to order. However, this is P3 because restaurants typically set up menus once and update infrequently. The critical path is customer ordering → kitchen fulfillment. Menu setup is a prerequisite but not part of the transaction flow.

**Independent Test**: Can be tested by: (1) owner signing up, (2) entering restaurant name and Bitcoin address, (3) creating menu categories, (4) adding items with names, descriptions, prices, photos, (5) generating customer-facing QR code. Delivers complete menu management independently from order processing.

**Acceptance Scenarios**:

1. **Given** restaurant owner visits signup page, **When** owner enters restaurant name and Lightning address, **Then** account is created in under 2 minutes with no email verification
2. **Given** owner has account, **When** owner creates menu categories (Appetizers, Mains, Desserts, Drinks), **Then** categories appear in menu structure immediately
3. **Given** category exists, **When** owner adds menu item with name, description, price (in local currency), **Then** item appears in category within 30 seconds
4. **Given** menu item exists, **When** owner uploads photo, **Then** photo is optimized and displayed in menu within 10 seconds
5. **Given** menu is complete, **When** owner generates customer QR code, **Then** printable QR code and shareable link are provided
6. **Given** menu changes needed, **When** owner edits prices or adds items, **Then** changes appear in customer-facing menu within 5 seconds
7. **Given** owner wants to temporarily disable item, **When** owner marks item as "out of stock", **Then** item is hidden from customer menu but retained in system

---

### User Story 4 - Owner Views Sales Dashboard (Priority: P4)

Linda needs to see daily sales, order count, top-selling items, and settlement status to understand her business performance without complex analytics tools.

**Why this priority**: Basic analytics are essential for business operation but not for MVP functionality. Restaurants can operate for days with just orders flowing through. Analytics become important after the system proves reliable, making this lowest priority for initial launch.

**Independent Test**: Can be tested by: (1) owner logging into dashboard, (2) viewing today's sales in Bitcoin and local currency, (3) seeing order count and average order value, (4) viewing top-selling items, (5) confirming daily settlement status. Delivers business insights independently from transaction processing.

**Acceptance Scenarios**:

1. **Given** owner logs into dashboard, **When** dashboard loads, **Then** owner sees: orders today, total sales (BTC and fiat), average order value
2. **Given** orders have been placed, **When** owner views order history, **Then** orders show: time, items, amount, payment status, order status
3. **Given** multiple items sold, **When** owner views "Top Items", **Then** items are ranked by quantity sold with revenue per item
4. **Given** end of day occurs, **When** owner checks settlements, **Then** daily Bitcoin settlement status shows: amount, destination address, settlement time
5. **Given** owner wants weekly summary, **When** owner selects date range, **Then** sales trends and totals display for selected period

---

### Edge Cases

- **What happens when Lightning payment fails?** System must detect failure within 30 seconds, show clear error message, allow customer to retry or generate new invoice, and never create order for unpaid invoices
- **What happens when customer's network drops after ordering?** Order status must be retrievable via order number lookup (no account needed), push notifications resume when network reconnects
- **What happens when kitchen tablet loses connection?** Orders must queue and sync automatically when connection restored, with clear visual indicator of offline status
- **What happens when menu item has no photo?** System must show placeholder image or text-only card that remains visually consistent with menu design
- **What happens when two customers order at same time?** System must handle concurrent orders without conflicts, preserving order sequence by timestamp
- **What happens when customer abandons cart before payment?** Cart is ephemeral (session-only), no persistent cart storage, no cleanup needed
- **What happens when restaurant needs to refund order?** Refunds are handled manually by restaurant owner - owner sends Lightning payment directly to customer using external wallet. System does not track or process refunds in v1.0. Owner can view order details to obtain customer's Lightning address if needed for refund.
- **What happens when owner wants to close restaurant temporarily?** System displays "Currently Closed" banner on menu with custom message and expected reopening hours. Menu remains visible for browsing but ordering and payment are disabled. Owner can toggle open/closed status from dashboard.
- **What happens to photos if restaurant hits storage limits?** System enforces minimal limits: maximum 2MB per photo upload, maximum 100 photos per restaurant. All photos are automatically compressed to 300KB optimized version for display. If restaurant reaches 100 photo limit, owner must delete existing photos before uploading new ones.

## Requirements *(mandatory)*

### Functional Requirements

**Customer Ordering:**
- **FR-001**: System MUST display restaurant menu organized by categories (Appetizers, Mains, Desserts, Drinks) with item names, descriptions, prices, and photos
- **FR-002**: System MUST allow customers to browse menu without account creation, login, or personal information
- **FR-003**: System MUST provide shopping cart where customers can add/remove items, adjust quantities, and see running total
- **FR-004**: System MUST generate Lightning Network invoice when customer initiates payment, displaying QR code and amount in satoshis
- **FR-005**: System MUST detect Lightning payment completion within 10 seconds and create order immediately upon confirmation
- **FR-006**: System MUST assign unique order number to each paid order for tracking and kitchen reference
- **FR-007**: System MUST provide real-time order status updates (Paid → Preparing → Ready) visible to customer without page refresh
- **FR-008**: System MUST send push notification to customer when order status changes to "Ready"
- **FR-009**: System MUST work as Progressive Web App (PWA) - installable to home screen, works offline for menu browsing

**Kitchen Operations:**
- **FR-010**: System MUST display all paid orders on kitchen display tablet in chronological order (oldest first)
- **FR-011**: System MUST show order details: order number, items with quantities, modifications, timestamp
- **FR-012**: System MUST provide audible/visual alert when new paid order arrives
- **FR-013**: System MUST allow kitchen staff to mark orders as "Preparing" or "Ready" with single tap
- **FR-014**: System MUST automatically update customer's order status in real-time when kitchen changes order state
- **FR-015**: System MUST move completed orders (marked Ready and picked up) to archived queue after 1 hour
- **FR-016**: System MUST handle kitchen display tablet offline - queue status changes and sync when connection restored

**Restaurant Management:**
- **FR-017**: System MUST allow restaurant owner to create account with minimal information: restaurant name, Lightning address
- **FR-018**: System MUST allow owner to create menu categories and add items within each category
- **FR-019**: System MUST accept menu item details: name (required), description (optional), price in local currency (required), photo (optional)
- **FR-020**: System MUST optimize and store uploaded photos with limits: maximum 2MB per photo upload, maximum 100 photos per restaurant, automatic compression to 300KB optimized version for display
- **FR-021**: System MUST allow owner to edit menu items, update prices, mark items as out of stock
- **FR-022**: System MUST generate customer-facing QR code linking to restaurant menu for printing or sharing
- **FR-023**: System MUST reflect menu changes in customer-facing view within 5 seconds
- **FR-024**: System MUST provide dashboard showing: orders today, total sales (BTC and fiat equivalent), average order value
- **FR-025**: System MUST display order history with: timestamp, items, amount, payment status, fulfillment status
- **FR-026**: System MUST show top-selling items ranked by quantity sold
- **FR-027**: System MUST display daily Bitcoin settlement status: amount, destination address, settlement timestamp
- **FR-039**: System MUST allow owner to toggle restaurant open/closed status from dashboard
- **FR-040**: System MUST display "Currently Closed" banner on customer menu when restaurant is closed, showing custom message and expected reopening hours
- **FR-041**: System MUST disable ordering and payment functionality when restaurant is marked as closed, while keeping menu visible for browsing

**Payment Processing:**
- **FR-028**: System MUST integrate with Lightning Network for payment processing (Strike API)
- **FR-029**: System MUST validate Lightning payment before creating order - no orders for unpaid invoices
- **FR-030**: System MUST handle payment failures gracefully - detect within 30 seconds, show clear error, allow retry
- **FR-031**: System MUST settle Bitcoin payments to restaurant's Lightning address daily (assumed end-of-day automatic settlement)
- **FR-032**: System MUST display prices in local fiat currency for customer clarity while processing payments in Bitcoin
- **FR-033**: System MUST convert fiat prices to satoshis at current exchange rate at time of invoice generation

**System Reliability:**
- **FR-034**: System MUST handle concurrent orders from multiple customers without conflicts
- **FR-035**: System MUST maintain order sequence integrity based on payment confirmation timestamp
- **FR-036**: System MUST log all payment attempts, successes, and failures for troubleshooting
- **FR-037**: System MUST provide order lookup by order number for customers who lost connection (no account needed)
- **FR-038**: System MUST continue functioning if external services (photos, analytics) temporarily fail
- **FR-042**: System MUST prevent new photo uploads when restaurant reaches 100 photo limit, requiring owner to delete existing photos first

### Key Entities

- **Restaurant**: Represents single restaurant tenant with name, Lightning address, menu structure, owner credentials, creation timestamp
- **Menu Category**: Logical grouping of menu items (e.g., Appetizers, Mains) with name, display order, active status
- **Menu Item**: Food/drink item with name, description, price (fiat), photo URL, category association, availability status (in stock / out of stock)
- **Order**: Customer purchase record with unique order number, timestamp, items with quantities, total amount (satoshis), payment status (pending/paid/failed), fulfillment status (paid/preparing/ready/completed), Lightning invoice ID
- **Payment**: Lightning Network transaction record with invoice ID, amount (satoshis), fiat equivalent, status, restaurant association, settlement status, timestamp
- **Cart**: Ephemeral session-only object containing selected items and quantities - not persisted to database

### Assumptions

- Photos will be stored in cloud storage (assumed AWS S3 or similar) with CDN delivery
- Exchange rate will be fetched from reliable source (assumed Strike API provides conversion)
- Single restaurant owner per account initially (no multi-user management)
- Settlement happens once daily at end of business day (assumed 11:59 PM local time)
- Order modifications/cancellations not supported in v1.0 (aligns with SLC simplicity principle)
- Refunds handled manually by owner using external Lightning wallet - system does not process refunds
- Customer can have multiple active orders simultaneously (no limit)
- Kitchen display remains always-on during business hours
- Menu item modifications (extra cheese, no onions) will be added in future version - v1.0 is fixed items only
- System timezone set to restaurant's local timezone for all timestamps
- Photo storage limits: 2MB max upload, 100 photos max per restaurant, 300KB optimized display version

## Success Criteria *(mandatory)*

### Measurable Outcomes

**Speed & Performance:**
- **SC-001**: Customer completes entire ordering flow (browse → select → pay → confirm) in under 2 minutes on mobile device
- **SC-002**: Lightning payment completes in under 10 seconds from invoice generation to order confirmation
- **SC-003**: Menu pages load in under 2 seconds on 3G mobile networks
- **SC-004**: Order status updates appear in customer view within 5 seconds of kitchen staff action
- **SC-005**: New paid orders appear on kitchen display within 5 seconds of payment confirmation

**Setup & Usability:**
- **SC-006**: Non-technical restaurant owner completes full menu setup (categories + 20 items with photos) in under 10 minutes
- **SC-007**: Kitchen staff can fulfill orders confidently after 5-minute demonstration with no additional training
- **SC-008**: Customer can order food without instructions, documentation, or staff assistance

**Adoption & Volume:**
- **SC-009**: 10 restaurants actively using system daily within 3 months of launch
- **SC-010**: System processes 1,000+ successful Lightning payment orders
- **SC-011**: Each restaurant processes average of 20+ orders per day after first week

**Reliability:**
- **SC-012**: 99% payment success rate (successful payment / total payment attempts)
- **SC-013**: 99% system uptime measured across all components (menu display, payment processing, kitchen display)
- **SC-014**: Zero orders lost due to system errors (all paid orders appear in kitchen queue)

**User Satisfaction:**
- **SC-015**: After first successful order, customers ask "How do I install this on my phone?" (PWA adoption indicator)
- **SC-016**: Restaurant owners demonstrate system to other business owners within first week (word-of-mouth indicator)
- **SC-017**: Kitchen staff prefer BitMerchant over previous POS system for order management (qualitative survey)
- **SC-018**: Customers describe payment experience as "faster than credit card" (qualitative feedback)

**Business Impact:**
- **SC-019**: Restaurants report lower effective payment processing fees compared to credit card processors (2-3% baseline)
- **SC-020**: Restaurant operates full business day (breakfast through dinner) using only BitMerchant with no additional tools needed


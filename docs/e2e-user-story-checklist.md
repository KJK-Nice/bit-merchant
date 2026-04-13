# E2E User Story Checklist

This checklist tracks end-to-end user journeys for Playwright coverage.

## Existing Coverage

- [x] As a customer, I can open menu, add item to cart, checkout, place cash order, and see order status.
- [x] As a merchant owner, I can sign up with passkey and access dashboard, admin, QR, and kitchen pages.
- [x] As a user, I am redirected to the correct host surface (public/customer/merchant) based on route intent.
- [x] As a platform, customer and merchant sessions are isolated by host cookies.

## Next User Stories To Add

### Auth & Access

- [ ] As a returning merchant owner, I can log in with an existing passkey and land on the correct default page.
- [x] As an owner, I can invite kitchen staff and the invited user can complete onboarding from invite link.
- [x] As a kitchen staff user, I cannot access owner-only admin/dashboard routes.
- [x] As an unauthenticated user, I am redirected to login when opening protected merchant routes.

### Multi-Restaurant

- [ ] As a user with multiple restaurant memberships, I am prompted to select a restaurant after login.
- [x] As a multi-restaurant user, selecting a restaurant updates active session context and routes me by role.
- [ ] As an owner, I can create a new restaurant and my active context switches to the new restaurant.

### Customer Ordering Resilience

- [ ] As a customer, I can recover my order via order lookup when returning later.
- [ ] As a customer, I can remove an item from cart and totals update before checkout.
- [ ] As a customer, I cannot place an order when cart is empty and I see a clear validation message.

### Kitchen Operations

- [ ] As kitchen staff, I can mark an order paid -> preparing -> ready from the kitchen screen.
- [ ] As a customer, I can observe order status updates while kitchen transitions the order state.

### Admin & Menu Management

- [ ] As an owner, I can toggle menu item availability and customers immediately see the availability state.
- [ ] As an owner, I can update QR/table settings and generated QR links remain valid for customer menu access.

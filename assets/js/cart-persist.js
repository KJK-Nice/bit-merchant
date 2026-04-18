// Persists Datastar cartItemQty signals to localStorage so the cart survives
// page reloads and is available for display when the user is offline.
(function () {
  var KEY = 'bm_cart_v1';

  // Save cart state whenever Datastar patches signals containing cartItemQty.
  document.addEventListener('datastar-signal-patch', function (e) {
    try {
      if (e.detail && e.detail.cartItemQty !== undefined) {
        localStorage.setItem(KEY, JSON.stringify(e.detail.cartItemQty));
      }
    } catch (_) {}
  });

  // When offline, restore the last-known cart state before the server fetch fires.
  if (!navigator.onLine) {
    try {
      var saved = localStorage.getItem(KEY);
      if (saved) {
        var cartItemQty = JSON.parse(saved);
        document.dispatchEvent(
          new CustomEvent('datastar-signal-patch', {
            detail: { cartItemQty: cartItemQty },
            bubbles: false,
          })
        );
      }
    } catch (_) {}
  }
})();

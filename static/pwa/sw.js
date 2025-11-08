// Service Worker for BitMerchant PWA
// Provides offline support for menu browsing

const CACHE_NAME = 'bitmerchant-v1';
const CACHE_URLS = [
  '/',
  '/static/pwa/manifest.json',
  '/static/css/main.css',
  '/static/js/datastar.js'
];

// Install event - cache static assets
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => cache.addAll(CACHE_URLS))
      .then(() => self.skipWaiting())
  );
});

// Activate event - clean up old caches
self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames
          .filter((name) => name !== CACHE_NAME)
          .map((name) => caches.delete(name))
      );
    }).then(() => self.clients.claim())
  );
});

// Fetch event - serve from cache, fallback to network
self.addEventListener('fetch', (event) => {
  // Only cache GET requests for menu browsing
  if (event.request.method !== 'GET') {
    return;
  }

  // Don't cache API endpoints or SSE streams
  if (event.request.url.includes('/api/') || 
      event.request.url.includes('/stream') ||
      event.request.url.includes('/payment/')) {
    return;
  }

  event.respondWith(
    caches.match(event.request)
      .then((cachedResponse) => {
        if (cachedResponse) {
          return cachedResponse;
        }

        return fetch(event.request).then((response) => {
          // Don't cache non-successful responses
          if (!response || response.status !== 200 || response.type !== 'basic') {
            return response;
          }

          const responseToCache = response.clone();
          caches.open(CACHE_NAME).then((cache) => {
            cache.put(event.request, responseToCache);
          });

          return response;
        });
      })
      .catch(() => {
        // Offline fallback - return cached menu page if available
        if (event.request.destination === 'document') {
          return caches.match('/');
        }
      })
  );
});


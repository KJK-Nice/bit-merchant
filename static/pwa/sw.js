const CACHE_NAME = 'bitmerchant-v2';
const URLS_TO_CACHE = [
  '/',
  '/menu',
  '/static/pwa/manifest.json',
  '/static/pwa/icon.svg',
  '/assets/js/input.min.js',
  '/assets/css/output.css'
];

const EXTERNAL_URLS = [
  'https://cdn.jsdelivr.net/gh/starfederation/datastar@1.0.0-RC.6/bundles/datastar.js'
];

self.addEventListener('install', event => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then(cache => {
        console.log('Opened cache');
        
        // Cache local resources (critical) - fail install if any fail
        const cacheLocal = cache.addAll(URLS_TO_CACHE);

        // Cache external resources (best effort) - don't fail install
        const cacheExternal = Promise.all(
          EXTERNAL_URLS.map(url => 
            fetch(url)
              .then(response => {
                if (response.ok) {
                  return cache.put(url, response);
                }
                console.warn('Skipping external resource cache (not ok):', url);
              })
              .catch(err => console.warn('Failed to cache external resource:', url, err))
          )
        );

        return Promise.all([cacheLocal, cacheExternal]);
      })
  );
});

self.addEventListener('fetch', event => {
  event.respondWith(
    caches.match(event.request)
      .then(response => {
        if (response) {
          return response;
        }
        return fetch(event.request);
      })
  );
});

self.addEventListener('activate', event => {
  const cacheWhitelist = [CACHE_NAME];
  event.waitUntil(
    caches.keys().then(cacheNames => {
      return Promise.all(
        cacheNames.map(cacheName => {
          if (cacheWhitelist.indexOf(cacheName) === -1) {
            return caches.delete(cacheName);
          }
        })
      );
    })
  );
});

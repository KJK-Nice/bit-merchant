package server

import (
	"net/http"
	"text/template"

	"github.com/labstack/echo/v4"
)

// BuildHash is injected at build time via -ldflags "-X bitmerchant/internal/common/server.BuildHash=<git-hash>".
var BuildHash = "dev"

var swTmpl = template.Must(template.New("sw").Parse(swSource))

func serveSW(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/javascript; charset=utf-8")
	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Response().Header().Set("Service-Worker-Allowed", "/")
	return swTmpl.Execute(c.Response(), map[string]string{"Version": BuildHash})
}

func serveOffline(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	return c.HTML(http.StatusServiceUnavailable, offlineHTML)
}

func serveKillSwitch(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/javascript; charset=utf-8")
	c.Response().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	return c.String(http.StatusOK, killSwitchJS)
}

const swSource = `
const VERSION = '{{.Version}}';
const CACHE_NAME = 'bitmerchant-' + VERSION;
const PRECACHE_URLS = [
  '/',
  '/static/pwa/manifest.json',
  '/static/pwa/icon.svg',
  '/static/pwa/icons/icon-192.png',
  '/static/pwa/icons/icon-512.png',
  '/assets/js/input.min.js',
  '/assets/css/output.css'
];

// Never cache auth, merchant surfaces, or API paths.
const DENY_PREFIXES = ['/merchant', '/admin', '/kitchen', '/auth', '/api'];

self.addEventListener('install', event => {
  event.waitUntil(
    caches.open(CACHE_NAME).then(cache => cache.addAll(PRECACHE_URLS))
  );
});

self.addEventListener('activate', event => {
  event.waitUntil(
    caches.keys()
      .then(names => Promise.all(
        names.filter(n => n !== CACHE_NAME).map(n => caches.delete(n))
      ))
      .then(() => self.clients.claim())
  );
});

self.addEventListener('fetch', event => {
  const { request } = event;
  const url = new URL(request.url);

  // Pass navigations through — canonical-host 302 redirects must not be wrapped.
  if (request.mode === 'navigate') return;

  // Only intercept same-origin GETs outside the deny-list.
  if (request.method !== 'GET') return;
  if (url.origin !== self.location.origin) return;
  if (DENY_PREFIXES.some(p => url.pathname.startsWith(p))) return;

  if (url.pathname.startsWith('/static/')) {
    // Cache-first for static assets.
    event.respondWith(
      caches.match(request).then(cached => cached || fetch(request).then(res => {
        caches.open(CACHE_NAME).then(c => c.put(request, res.clone()));
        return res;
      }))
    );
    return;
  }

  if (url.pathname.startsWith('/assets/')) {
    // Stale-while-revalidate for compiled CSS/JS.
    event.respondWith(
      caches.open(CACHE_NAME).then(cache =>
        cache.match(request).then(cached => {
          const fresh = fetch(request).then(res => {
            cache.put(request, res.clone());
            return res;
          });
          return cached || fresh;
        })
      )
    );
    return;
  }
});

self.addEventListener('message', event => {
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting();
  }
});
`

const killSwitchJS = `
// Emergency SW kill switch — unregisters all service workers and clears all caches.
(async function () {
  if (!('serviceWorker' in navigator)) return;
  const regs = await navigator.serviceWorker.getRegistrations();
  await Promise.all(regs.map(r => r.unregister()));
  const keys = await caches.keys();
  await Promise.all(keys.map(k => caches.delete(k)));
  console.log('[sw-kill] All service workers unregistered and caches cleared.');
})();
`

const offlineHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0, viewport-fit=cover">
  <meta name="theme-color" content="#000000">
  <title>Offline — BitMerchant</title>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
    html, body { height: 100%; font-family: system-ui, -apple-system, sans-serif; background: #fff; color: #111; }
    .page { min-height: 100%; display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 2rem; text-align: center; gap: 1rem; }
    h1 { font-size: 1.5rem; font-weight: 700; }
    p  { color: #666; max-width: 24rem; }
    button { margin-top: .5rem; padding: .6rem 1.4rem; border: none; border-radius: .5rem; background: #000; color: #fff; font-size: .9rem; cursor: pointer; }
    button:hover { background: #333; }
    @media (prefers-color-scheme: dark) {
      html, body { background: #0a0a0a; color: #f5f5f5; }
      p { color: #999; }
      button { background: #fff; color: #000; }
      button:hover { background: #ddd; }
    }
  </style>
</head>
<body>
  <div class="page">
    <h1>You're offline</h1>
    <p>Check your connection and try again.</p>
    <button onclick="window.location.reload()">Retry</button>
  </div>
</body>
</html>`

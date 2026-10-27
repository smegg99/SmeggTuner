// frontend/web-public/sw.js
// The web edition's offline cache. scripts/web.sh stamps __BUILD__ with VERSION and a hash of the build, so every
// deploy gets a new cache and the old one is deleted on activate.
const CACHE = 'smeggtuner-web-__BUILD__'
const PREFIX = 'smeggtuner-web-'
const SHELL = ['/', '/engine/smeggtuner.wasm', '/engine/wasm_exec.js', '/manifest.webmanifest']

self.addEventListener('install', (e) => {
  self.skipWaiting()
  e.waitUntil(caches.open(CACHE).then(c => c.addAll(SHELL)))
})

self.addEventListener('activate', (e) => {
  e.waitUntil(caches.keys()
    .then(names => Promise.all(names.filter(n => n.startsWith(PREFIX) && n !== CACHE).map(n => caches.delete(n))))
    .then(() => self.clients.claim()))
})

// Pages go to the network first so a deploy shows at once; hashed assets and the engine come from the cache first.
self.addEventListener('fetch', (e) => {
  const req = e.request
  const url = new URL(req.url)
  if (req.method !== 'GET' || url.origin !== self.location.origin) return
  if (req.mode === 'navigate') {
    e.respondWith(fetch(req)
      .then((res) => {
        const copy = res.clone()
        caches.open(CACHE).then(c => c.put('/', copy))
        return res
      })
      .catch(() => caches.match('/')))
    return
  }
  e.respondWith(caches.match(req).then(hit => hit || fetch(req).then((res) => {
    if (res.ok) {
      const copy = res.clone()
      caches.open(CACHE).then(c => c.put(req, copy))
    }
    return res
  })))
})

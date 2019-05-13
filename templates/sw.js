
const C_CACHE_VERSION = '{{.CacheVersion}}';
const C_SERVICE_WORKER_VERSION = '{{.ServiceWorkerVersion}}';
const C_APPLICATION_VERSION = '{{.BuildVersion}}';
const C_PLAYLIST_VERSION = '{{.PlaylistVersion}}';

self.addEventListener('install', function (event) {

    console.log("Installing ServiceWorker " + C_SERVICE_WORKER_VERSION + "...");

    event.waitUntil(
        caches.open(C_CACHE_VERSION).then(function (cache) {
            return cache.addAll([
                '/',
                // '/api/playlist',
                '/manifest.json',

                '/static/js/vendor/jquery.min.js',
                '/static/js/vendor/bootstrap.min.js',
                '/static/js/vendor/mithril.js',
                '/static/js/vendor/popper.min.js',
                '/static/js/vendor/underscore.min.js',
                '/static/js/app.js',
                '/static/images/background_stripes.png',
                '/static/images/explode-guitar-1.png',
                '/static/images/header.png',
                '/static/images/192.png',
                '/static/images/512.png',
                '/static/css/vendor/bootstrap.lux.min.css',
                '/static/vendor/font-awesome/css/all.css',
                '/static/css/anton.css',
                '/static/css/styles.css',
                '/static/vendor/font-awesome/webfonts/fa-solid-900.woff2',
                '/static/vendor/font-awesome/webfonts/fa-regular-400.woff2',
                '/static/fonts/1Ptgg87LROyAm3K8-C8CSKlvPfE.woff2',
                '/static/fonts/1Ptgg87LROyAm3K9-C8CSKlvPfE.woff2',
                '/static/fonts/1Ptgg87LROyAm3Kz-C8CSKlv.woff2',

            ]);
        })
    );

    self.skipWaiting();

});

self.addEventListener('activate', function (event) {
    console.log("Activating ServiceWorker " + C_SERVICE_WORKER_VERSION + "...");
    event.waitUntil(caches.keys().then(function (keys) {
        keys.forEach(function (k) {
            if (k != C_CACHE_VERSION) {
                console.debug("Deleting old cache: " + k + "...");
                caches.delete(k);
            }
        })
    }));
});

self.addEventListener('fetch', function (event) {
    event.respondWith(caches.match(event.request).then(function (response) {
        // caches.match() always resolves
        // but in case of success response will have value
        if (response !== undefined) {
            return response;
        } else {
            return fetch(event.request, { credentials: 'include' }).then(function (response) {
                // response may be used only once
                // we need to save clone to put one copy in cache
                // and serve second one
                let responseClone = response.clone();

                caches.open(C_CACHE_VERSION).then(function (cache) {
                    cache.put(event.request, responseClone);
                });
                return response;
            }).catch(function () {
                return caches.match('/static/images/explode-guitar-1.png');
            });
        }
    }));
});
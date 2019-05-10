self.addEventListener('install', function (event) {
    event.waitUntil(
        caches.open('v1').then(function (cache) {
            return cache.addAll([
                '',
                '/',
                '/api/playlist',
                '/mainfest.json',

                '/static/js/vendor/jquery.min.js',
                '/static/js/vendor/bootstrap.min.js',
                '/static/js/vendor/mithril.js',
                '/static/js/vendor/popper.min.js',
                '/static/js/vendor/underscore.min.js',

                '/static/js/app.js',
                '/static/css/styles.css',
                '/static/images/background.jpg',
                '/static/images/explode-guitar-1.png',
                '/static/images/header.jpg',
            ]);
        })
    );
});

self.addEventListener('fetch', function (event) {
    event.respondWith(caches.match(event.request).then(function (response) {
        // caches.match() always resolves
        // but in case of success response will have value
        if (response !== undefined) {
            return response;
        } else {
            return fetch(event.request).then(function (response) {
                // response may be used only once
                // we need to save clone to put one copy in cache
                // and serve second one
                let responseClone = response.clone();

                caches.open('v1').then(function (cache) {
                    cache.put(event.request, responseClone);
                });
                return response;
            }).catch(function () {
                return caches.match('/sw-test/gallery/myLittleVader.jpg');
            });
        }
    }));
});
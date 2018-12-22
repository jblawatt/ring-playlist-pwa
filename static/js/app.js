

; (function (exports) {


    const DISPLAY_KT = 1;
    const DISPLAY_TK = 2;

    exports.data = [];
    exports.favorites = {};
    exports.display = DISPLAY_KT;
    exports.filter = null;
    exports.filterFav = false;

    function setFavorite(id) {
        localStorage.setItem('favorite-' + id, 1);
    }

    function clearFavorite(id) {
        localStorage.removeItem('favorite-' + id);
    }

    function isFavorite(id) {
        return localStorage.getItem('favorite-' + id) == '1';
    }

    m.request({
        method: "GET",
        url: "/api/playlist"
    }).then(function (result) {
        exports.data = result.data;
    });

    const DisplaySwitch = {
        view: function (vnode) {
            return m('div.btn-toolbar',
                m('div.button-group.mr-2',
                    m('button.btn.btn-md.btn-tab', {
                        class: exports.display == DISPLAY_KT ? 'btn-secondary' : 'btn-primary',
                        onclick: function () { exports.display = DISPLAY_KT },
                    }, 'Künstler, Titel'),
                    m('button.btn.btn-md.btn-tab', {
                        class: exports.display == DISPLAY_TK ? 'btn-secondary' : 'btn-primary',
                        onclick: function () { exports.display = DISPLAY_TK },
                    }, 'Titel, Künstler'),
                ),
                m('div.button-group', m('button.btn.btn-md.btn-tab', {
                    class: exports.filterFav ? 'btn-secondary' : 'btn-primary',
                    onclick: function () {
                        exports.filterFav = !exports.filterFav;
                    }
                }, m('i.fa-heart.fa')))
            )
        }
    }

    const FavoriteButton = {
        view: function (vnode) {
            var item = vnode.attrs.item;
            return m(
                'button.btn.btn-link.float-right.btn-sm.btn-favorite',
                {
                    onclick: function (e) {
                        if (isFavorite(item.h)) {
                            clearFavorite(item.h);
                        } else {
                            setFavorite(item.h);
                        }
                    },
                    style: "font-size: 1em"
                },
                m('i.fa-heart', { class: isFavorite(item.h) ? "fa" : "far" })
            )


        }
    }

    const FilterField = {
        view: function (vnode) {
            return m('div.form-group', m('form.form-inline', { onreset: function () { exports.filter = ""; } }, m('div.input-group mb-3'
                , m('input.form-control', {
                    placeholder: 'Was suchste denn?',
                    oninput: _.throttle(function (evt) {
                        exports.filter = evt.target.value;
                        console.log(exports.filter);
                    }, 100)
                })
                , m('div.input-group-append', m('button.btn.btn-primary', { type: 'reset', style: "font-size:1.5em;" }, m.trust('<i class="fas fa-eraser"></i>')))
            )));
        }
    }

    function doFilter(item) {
        if (exports.filterFav) {
            if (!isFavorite(item.h)) return false;
        }
        var f = (exports.filter || "").trim().toLowerCase();
        if (!f) {
            return true;
        }
        return item.r.toLowerCase().indexOf(f) > -1;
    }

    const Playlist = {
        view: function (vnode) {

            let filtered = exports.data.filter(doFilter);

            if (filtered.length == 0) {
                if (exports.data.length == 0) {
                    return m('ul.list-group.list-group-flush.bg-dark', m('li.list-group-item.bg-dark',
                        m.trust('<i class="fas fa-circle-notch fa-spin"></i> kommt sofort...')
                    ));
                }
                return m('ul.list-group.list-group-flush.bg-dark', m('li.list-group-item.bg-dark.text-center',
                    m('p', 'sorry, da gibbet nichts zu'),
                    m('img', { src: '/static/images/explode-guitar-1.png', style: "width:100%" })

                ));
            }

            return m('ul.list-group.list-group-flush.bg-dark', filtered.map(function (item) {
                return m('li.list-group-item.bg-dark', { dataSearch: item.r }, [
                    m('div.line-header', m('span.line-header__text', m.trust('<small>No.</small> ' + item.n))),
                    m(FavoriteButton, { item: item }),
                    exports.display == DISPLAY_KT ? m('span.text-primary', item.a) : m('span.text-primary', item.tt),
                    exports.display == DISPLAY_KT ? m('small', ' mit') : m('small', ' von'),
                    m('br'),
                    exports.display == DISPLAY_KT ? m('span.text-secondary', item.tt) : m('span.text-secondary', item.a),
                    m('small', " (" + item.t + ")")
                ])
            }));
        }
    }

    m.mount(document.getElementById("content"), Playlist);
    m.mount(document.getElementById("display-switch"), DisplaySwitch);
    m.mount(document.getElementById("filter-form"), FilterField);



})(window);

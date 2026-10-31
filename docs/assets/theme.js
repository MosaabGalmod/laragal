// Loaded synchronously in <head> so the saved theme applies before first paint (no flash).
// External file instead of an inline script keeps the CSP free of 'unsafe-inline'.
(function () {
  var KEY = 'laragal_site_theme';
  var root = document.documentElement;

  function stored() {
    try { return localStorage.getItem(KEY); } catch (e) { return null; }
  }

  function apply(theme) {
    root.setAttribute('data-theme', theme);
  }

  var initial = stored();
  if (initial !== 'dark' && initial !== 'light') {
    initial = window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }
  apply(initial);

  document.addEventListener('DOMContentLoaded', function () {
    // Mobile menu is a native <details>; close it on link click, outside click and Escape.
    var menu = document.querySelector('details.menu');
    if (menu) {
      menu.addEventListener('click', function (e) {
        if (e.target.closest('a')) menu.open = false;
      });
      document.addEventListener('click', function (e) {
        if (menu.open && !menu.contains(e.target)) menu.open = false;
      });
      document.addEventListener('keydown', function (e) {
        if (e.key === 'Escape' && menu.open) {
          menu.open = false;
          menu.querySelector('summary').focus();
        }
      });
    }

    var btn = document.getElementById('themeToggle');
    if (!btn) return;
    btn.setAttribute('aria-pressed', String(root.getAttribute('data-theme') === 'dark'));
    btn.addEventListener('click', function () {
      var next = root.getAttribute('data-theme') === 'dark' ? 'light' : 'dark';
      apply(next);
      btn.setAttribute('aria-pressed', String(next === 'dark'));
      try { localStorage.setItem(KEY, next); } catch (e) {}
    });
  });
})();

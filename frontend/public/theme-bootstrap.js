/* FOUC-аас сэргийлэх: React hydrate хийхээс ӨМНӨ загвар/хэл/харагдацыг сонгож
   <html>-д тавина. Түрүүлэх дараалал: хэрэглэгчийн localStorage сонголт →
   админы сайт-default (window.__SITE_APPEARANCE__, layout.tsx-аас) → template
   fallback. src/lib/preferences.ts-тэй ижил түлхүүр (gerege.*) ашиглана.
   <head> доторх блоклогч script тул body зурахаас өмнө ажиллана. */
(function () {
  try {
    var root = document.documentElement;
    var SA = window.__SITE_APPEARANCE__ || {};
    var oneOf = function (v, list, fb) { return list.indexOf(v) !== -1 ? v : fb; };
    var PRESETS = ['cobalt', 'teal', 'violet', 'emerald', 'amber'];

    // Загвар (theme): хэрэглэгч → сайт-default → light.
    var theme = localStorage.getItem('gerege.theme');
    if (['light', 'dark', 'system'].indexOf(theme) === -1) {
      theme = oneOf(SA.theme, ['light', 'dark', 'system'], 'light');
    }
    var effective = theme;
    if (theme === 'system') {
      effective =
        window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    }
    if (effective === 'dark') root.setAttribute('data-theme', 'dark');
    else root.removeAttribute('data-theme'); // SSR dark тавьсан байвал буцаана
    root.setAttribute('data-theme-pref', theme);

    // Хэл.
    var lang = localStorage.getItem('gerege.lang');
    if (lang !== 'mn' && lang !== 'en') lang = 'mn';
    root.setAttribute('lang', lang);

    // Өнгөний хослол (accent): хэрэглэгч зөвхөн preset сонгодог; байхгүй бол
    // сайт-default (preset ЭСВЭЛ '#rrggbb' custom hex).
    var accent = localStorage.getItem('gerege.accent');
    if (PRESETS.indexOf(accent) === -1) accent = SA.accent || 'cobalt';
    if (typeof accent === 'string' && accent.charAt(0) === '#') {
      root.style.setProperty('--dan-blue-base', accent);
      root.setAttribute('data-accent', 'custom');
    } else {
      root.setAttribute('data-accent', PRESETS.indexOf(accent) !== -1 ? accent : 'cobalt');
      root.style.removeProperty('--dan-blue-base');
    }

    // Фонт ба нягтрал: хэрэглэгч → сайт-default → template fallback.
    var font = localStorage.getItem('gerege.font');
    if (['inter', 'serif', 'system'].indexOf(font) === -1) font = oneOf(SA.font, ['inter', 'serif', 'system'], 'inter');
    root.setAttribute('data-font', font);

    var style = localStorage.getItem('gerege.style');
    if (['comfortable', 'compact'].indexOf(style) === -1) {
      style = oneOf(SA.style, ['comfortable', 'compact'], 'comfortable');
    }
    root.setAttribute('data-style', style);
  } catch (e) {}
})();

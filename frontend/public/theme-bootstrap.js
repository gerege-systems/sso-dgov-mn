/* FOUC-аас сэргийлэх: React hydrate хийхээс ӨМНӨ localStorage-аас загвар/хэлийг
   сонгож <html>-д тавина — буруу загвар анивчихгүй. src/lib/preferences.ts-тэй
   ижил түлхүүр (gerege.theme / gerege.lang) ашиглана. <head> доторх блоклогч
   script тул body зурахаас өмнө ажиллана. */
(function () {
  try {
    var theme = localStorage.getItem('gerege.theme');
    if (theme !== 'light' && theme !== 'dark' && theme !== 'system') theme = 'light';
    var effective = theme;
    if (theme === 'system') {
      effective =
        window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches
          ? 'dark'
          : 'light';
    }
    if (effective === 'dark') document.documentElement.setAttribute('data-theme', 'dark');
    document.documentElement.setAttribute('data-theme-pref', theme);
    var lang = localStorage.getItem('gerege.lang');
    if (lang !== 'mn' && lang !== 'en') lang = 'mn';
    document.documentElement.setAttribute('lang', lang);

    /* Харагдац — өнгө/фонт/нягтрал. preferences.ts дахь DEFAULTS/VALID-тэй нийцүүлнэ. */
    var accent = localStorage.getItem('gerege.accent');
    if (['cobalt', 'teal', 'violet', 'emerald', 'amber'].indexOf(accent) === -1) accent = 'cobalt';
    document.documentElement.setAttribute('data-accent', accent);
    var font = localStorage.getItem('gerege.font');
    if (['inter', 'serif', 'system'].indexOf(font) === -1) font = 'inter';
    document.documentElement.setAttribute('data-font', font);
    var style = localStorage.getItem('gerege.style');
    if (['comfortable', 'compact'].indexOf(style) === -1) style = 'comfortable';
    document.documentElement.setAttribute('data-style', style);
  } catch (e) {}
})();

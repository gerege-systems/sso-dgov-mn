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
  } catch (e) {}
})();

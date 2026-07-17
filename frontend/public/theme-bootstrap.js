/* FOUC-аас сэргийлэх: React hydrate хийхээс ӨМНӨ загвар/хэл/харагдацыг сонгож
   <html>-д тавина. Энэ бол блоклогч (head доторх sync) script тул body зурахаас
   өмнө ажиллана.

   ХАРАГДАЦЫГ ХОЁР ХҮРЭЭНД САЛГАВ:
   - Нэвтэрсэн апп (/me, /admin, /manager, …) → ХЭРЭГЛЭГЧийн өөрийн сонголт
     (localStorage 'gerege.*'), байхгүй бол template default. Админы сайт-
     тохиргоог эндэ ХЭРЭГЛЭХГҮЙ.
   - Нийтийн хуудас (/, /login, /auth, /oauth, /sso, …) → АДМИНы сайт-тохиргоо
     (window.__SITE_APPEARANCE__). Хэрэглэгчийн localStorage-ыг эндэ үл хэрэгснэ.
   Хэл (mn/en) нь харагдацаас тусдаа — хаана ч localStorage-аар ажиллана. */
(function () {
  try {
    var root = document.documentElement;
    var SA = window.__SITE_APPEARANCE__ || {};
    var PRESETS = ['cobalt', 'teal', 'violet', 'emerald', 'amber'];
    // Нэвтэрсэн апп-ын талбар уу?
    var authed = /^\/(me|admin|manager|profile|settings)(\/|$)/.test(location.pathname);

    // Тухайн түлхүүрийн утгыг хүрээнд нь тохируулан сонгоно.
    var pick = function (key, list, tmpl) {
      if (authed) {
        var v = localStorage.getItem('gerege.' + key);
        return list.indexOf(v) !== -1 ? v : tmpl;
      }
      return list.indexOf(SA[key]) !== -1 ? SA[key] : tmpl;
    };

    // Загвар (theme).
    var theme = pick('theme', ['light', 'dark', 'system'], 'light');
    var effective = theme;
    if (theme === 'system') {
      effective =
        window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
    }
    if (effective === 'dark') root.setAttribute('data-theme', 'dark');
    else root.removeAttribute('data-theme');
    root.setAttribute('data-theme-pref', theme);

    // Хэл — харагдацаас тусдаа, хаана ч localStorage-аар.
    var lang = localStorage.getItem('gerege.lang');
    if (lang !== 'mn' && lang !== 'en') lang = 'mn';
    root.setAttribute('lang', lang);

    // Өнгөний хослол (accent). Нэвтэрсэн апп-д хэрэглэгч зөвхөн preset сонгодог;
    // нийтийн хуудсанд админ preset ЭСВЭЛ '#rrggbb' custom hex өгч болно.
    var accent;
    if (authed) {
      accent = localStorage.getItem('gerege.accent');
      if (PRESETS.indexOf(accent) === -1) accent = 'cobalt';
    } else {
      accent = SA.accent || 'cobalt';
    }
    if (typeof accent === 'string' && accent.charAt(0) === '#') {
      root.style.setProperty('--dan-blue-base', accent);
      root.setAttribute('data-accent', 'custom');
    } else {
      root.setAttribute('data-accent', PRESETS.indexOf(accent) !== -1 ? accent : 'cobalt');
      root.style.removeProperty('--dan-blue-base');
    }

    // Фонт ба нягтрал.
    root.setAttribute('data-font', pick('font', ['inter', 'serif', 'system'], 'inter'));
    root.setAttribute('data-style', pick('style', ['comfortable', 'compact'], 'comfortable'));
  } catch (e) {}
})();

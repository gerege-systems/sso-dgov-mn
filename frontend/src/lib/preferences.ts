"use client";

import { useCallback, useEffect, useState } from 'react';

export type ThemePref = 'light' | 'dark' | 'system';
export type LangPref = 'mn' | 'en';

const KEYS = { theme: 'gerege.theme', lang: 'gerege.lang' } as const;
const VALID = {
  theme: new Set<ThemePref>(['light', 'dark', 'system']),
  lang: new Set<LangPref>(['mn', 'en']),
};

const read = <T extends string>(key: 'theme' | 'lang', fallback: T, valid: Set<T>): T => {
  if (typeof window === 'undefined') return fallback;
  try {
    const v = localStorage.getItem(KEYS[key]) as T | null;
    return v && valid.has(v) ? v : fallback;
  } catch {
    return fallback;
  }
};

const applyTheme = (value: ThemePref) => {
  if (typeof document === 'undefined') return;
  const effective: 'light' | 'dark' =
    value === 'system'
      ? window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
      : value;
  const root = document.documentElement;
  if (effective === 'dark') root.setAttribute('data-theme', 'dark');
  else root.removeAttribute('data-theme');
  root.setAttribute('data-theme-pref', value);
};

const applyLang = (value: LangPref) => {
  if (typeof document === 'undefined') return;
  document.documentElement.setAttribute('lang', value);
};

/**
 * gerege theme-ийн тохиргоог (загвар + хэл) localStorage-д уншиж/бичээд <html>
 * дээр тусгана. me.dgov.mn-ийн site.js-тэй ижил зарчмаар ажиллана.
 */
export function usePreferences() {
  // Inline-bootstrap утгаас эхэлснээр SSR markup эхний client render-тэй тохирно
  // (hydration зөрөхгүй). useEffect-д localStorage-аас дахин синк хийнэ.
  const [theme, setThemeState] = useState<ThemePref>('light');
  const [lang, setLangState] = useState<LangPref>('mn');

  useEffect(() => {
    setThemeState(read('theme', 'light', VALID.theme));
    setLangState(read('lang', 'mn', VALID.lang));
  }, []);

  // OS загвар солигдоход "system" дээр байвал дахин тусгана.
  useEffect(() => {
    if (theme !== 'system' || typeof window === 'undefined') return;
    const mql = window.matchMedia('(prefers-color-scheme: dark)');
    const handler = () => applyTheme('system');
    mql.addEventListener('change', handler);
    return () => mql.removeEventListener('change', handler);
  }, [theme]);

  const setTheme = useCallback((value: ThemePref) => {
    if (!VALID.theme.has(value)) return;
    setThemeState(value);
    try { localStorage.setItem(KEYS.theme, value); } catch {}
    applyTheme(value);
  }, []);

  const setLang = useCallback((value: LangPref) => {
    if (!VALID.lang.has(value)) return;
    setLangState(value);
    try { localStorage.setItem(KEYS.lang, value); } catch {}
    applyLang(value);
  }, []);

  return { theme, setTheme, lang, setLang };
}

/** Жижиг toast туслах — globals.css дахь .toast класс ашиглана. */
export function showToast(message: string) {
  if (typeof document === 'undefined') return;
  let el = document.querySelector<HTMLDivElement>('.toast[data-app-toast]');
  if (!el) {
    el = document.createElement('div');
    el.className = 'toast';
    el.dataset.appToast = '1';
    el.setAttribute('role', 'status');
    el.setAttribute('aria-live', 'polite');
    document.body.appendChild(el);
  }
  el.textContent = message;
  requestAnimationFrame(() => el!.classList.add('is-visible'));
  window.clearTimeout((el as unknown as { _t: number })._t);
  (el as unknown as { _t: number })._t = window.setTimeout(() => el!.classList.remove('is-visible'), 1800);
}

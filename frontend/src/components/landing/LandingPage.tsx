// Gerege Systems Development Team & Claude AI, 2026
"use client";

import React from 'react';
import {
  Fingerprint, Sparkles, ShieldCheck, Network, Waypoints, Users,
  Terminal, Layers, Bot, CheckCircle2, ArrowRight, ChevronRight,
  LogIn, Languages, KeyRound, ScrollText, Globe, Gauge, ShieldAlert,
  Menu, X,
} from 'lucide-react';
import { useLang } from '@/lib/lang';
import LoginForm from '@/app/login/LoginForm';
import { landingCopy, type LandingCopy } from './copy';
import { deepMerge } from '@/lib/theme';
import { DOCS_URL } from '@/lib/links';

// «Бүх боломж» хэсгийн жижиг картуудын icon-ууд (copy.ts-ийн items дараалалтай нэг эрэмбэ).
const EVERYTHING_ICONS = [Fingerprint, Globe, KeyRound, ScrollText, CheckCircle2, Languages, Gauge, ShieldAlert];

interface Props {
  /** LoginForm-д дамжуулах — нэвтэрсний дараа буцах зам. */
  next: string;
  notice?: string;
  googleLink?: boolean;
  googleError?: boolean;
  /** Идэвхтэй theme-ийн landing текст/цэс (mn/en) — copy.ts default дээр давхарлана. */
  themeLanding?: { mn?: Partial<LandingCopy>; en?: Partial<LandingCopy> };
}

/**
 * Government SSO нүүр (landing) — нэвтрээгүй зочдод харагдах маркетингийн
 * нүүр. Government SSO-ийн бүх чадварыг харуулж, hero-ийн баруун талд ОДОО БАЙГАА нэвтрэх
 * картыг (SigninShell доторх `.signin-card` + LoginForm) шигтгэв. Брэнд токен
 * (blue + gold, light/dark) дээр найруулав.
 */
export default function LandingPage({ next, notice, googleLink, googleError, themeLanding }: Props) {
  const { lang, setLang } = useLang();
  // Mobile (<900px)-д цэсний холбоосууд hamburger доор нээгдэнэ.
  const [menuOpen, setMenuOpen] = React.useState(false);
  // Идэвхтэй theme-ийн текст байвал copy.ts default дээр гүн merge хийнэ.
  const override = themeLanding?.[lang];
  const t = override ? deepMerge(landingCopy[lang], override) : landingCopy[lang];
  const brand = t.brand || 'Government SSO';

  return (
    <div className="lp">
      {/* ---------- Nav ---------- */}
      <header className="lp-nav">
        <div className="lp-nav__inner">
          <a className="lp-nav__brand" href="#top">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img className="lp-nav__mark" src="/brand.webp" alt="" aria-hidden="true" />
            <span className="lp-nav__name">{brand}</span>
          </a>

          <nav className="lp-nav__links" aria-label="Хэсгүүд">
            <a href="#features">{t.nav.features}</a>
            <a href="#security">{t.nav.security}</a>
            <a href="#tech">{t.nav.tech}</a>
            <a href={DOCS_URL} target="_blank" rel="noreferrer">{t.nav.docs}</a>
          </nav>

          <div className="lp-nav__actions">
            <button
              type="button"
              className="lp-lang"
              onClick={() => setLang(lang === 'mn' ? 'en' : 'mn')}
              aria-label="Хэл солих"
            >
              <Languages size={15} strokeWidth={2} />
              <span>{lang === 'mn' ? 'EN' : 'МН'}</span>
            </button>
            <a className="lp-btn lp-btn--gold lp-btn--sm" href="#login">
              <LogIn size={16} strokeWidth={2} />
              <span>{t.nav.login}</span>
            </a>
            <button
              type="button"
              className="lp-burger"
              onClick={() => setMenuOpen((o) => !o)}
              aria-label={menuOpen ? 'Цэс хаах' : 'Цэс нээх'}
              aria-expanded={menuOpen}
              aria-controls="lp-mobile-menu"
            >
              {menuOpen ? <X size={19} strokeWidth={2} /> : <Menu size={19} strokeWidth={2} />}
            </button>
          </div>
        </div>

        {/* Mobile цэс — hamburger дарахад nav-ийн доор задарна (≥900px-д CSS нуана). */}
        {menuOpen && (
          <nav className="lp-mobile-menu" id="lp-mobile-menu" aria-label="Хэсгүүд">
            <a href="#features" onClick={() => setMenuOpen(false)}>{t.nav.features}</a>
            <a href="#security" onClick={() => setMenuOpen(false)}>{t.nav.security}</a>
            <a href="#tech" onClick={() => setMenuOpen(false)}>{t.nav.tech}</a>
            <a href={DOCS_URL} target="_blank" rel="noreferrer" onClick={() => setMenuOpen(false)}>{t.nav.docs}</a>
          </nav>
        )}
      </header>

      <main id="top">
        {/* ---------- Hero (login card шигтгэсэн) ---------- */}
        <section className="lp-hero">
          <div className="lp-hero__pattern" aria-hidden="true" />
          <div className="lp-hero__inner">
            <div className="lp-hero__copy">
              <span className="lp-eyebrow">
                <span className="lp-eyebrow__dot" />
                {t.hero.badge}
              </span>
              <h1 className="lp-hero__title">
                {t.hero.titleLead}{' '}
                <span className="lp-accent">{t.hero.titleAccent}</span>{' '}
                {t.hero.titleTail}
              </h1>
              <p className="lp-hero__lede">{t.hero.lede}</p>

              <div className="lp-hero__cta">
                <a className="lp-btn lp-btn--gold lp-btn--lg" href="#login">
                  {t.hero.ctaLogin}
                  <ArrowRight size={18} strokeWidth={2} />
                </a>
                <a className="lp-btn lp-btn--outline lp-btn--lg" href="#features">
                  {t.hero.ctaExplore}
                </a>
              </div>

              <div className="lp-hero__statrow">
                {t.hero.stats.map((s) => (
                  <div className="lp-stat" key={s.label}>
                    <span className="lp-stat__value">{s.value}</span>
                    <span className="lp-stat__label">{s.label}</span>
                  </div>
                ))}
              </div>
            </div>

            {/* ОДОО БАЙГАА нэвтрэх карт — hero-ийн баруун талд шигтгэв */}
            <div className="lp-hero__visual">
              <section id="login" className="signin-card lp-hero__login" aria-labelledby="login-title">
                {/* Нүүр дээр автомат фокус хийхгүй — энэ нь маркетингийн хуудас
                    бөгөөд карт нь mobile-д hero-гийн доор байрладаг. */}
                <LoginForm
                  next={next}
                  notice={notice}
                  googleLink={googleLink}
                  googleError={googleError}
                  autoFocus={false}
                />
              </section>
            </div>
          </div>
        </section>

        {/* ---------- Гол давуу талууд (bento) ---------- */}
        <section className="lp-section" id="features">
          <div className="lp-container">
            <header className="lp-head">
              <h2>{t.advantages.heading}</h2>
              <p>{t.advantages.sub}</p>
            </header>

            <div className="lp-bento">
              {/* eID (wide) */}
              <article className="lp-card lp-card--wide">
                <div className="lp-card__top">
                  <span className="lp-card__icon"><Fingerprint size={26} strokeWidth={1.75} /></span>
                  <span className="lp-card__tag">{t.advantages.eidTag}</span>
                </div>
                <div>
                  <h3>{t.advantages.eidTitle}</h3>
                  <p>{t.advantages.eidBody}</p>
                </div>
              </article>

              {/* Google холболт (dark) */}
              <article className="lp-card lp-card--dark">
                <span className="lp-card__icon lp-card__icon--onDark"><KeyRound size={26} strokeWidth={1.75} /></span>
                <div>
                  <h3>{t.advantages.googleTitle}</h3>
                  <p>{t.advantages.googleBody}</p>
                </div>
              </article>

              {/* Аюулгүй байдал (muted) */}
              <article className="lp-card lp-card--muted" id="security">
                <span className="lp-card__icon"><ShieldCheck size={26} strokeWidth={1.75} /></span>
                <div>
                  <h3>{t.advantages.secTitle}</h3>
                  <p>{t.advantages.secBody}</p>
                </div>
              </article>

              {/* SSO / OIDC provider (wide split) */}
              <article className="lp-card lp-card--wide lp-card--split">
                <div>
                  <span className="lp-card__icon"><Network size={26} strokeWidth={1.75} /></span>
                  <h3>{t.advantages.ssoTitle}</h3>
                  <p>{t.advantages.ssoBody}</p>
                </div>
                <span className="lp-card__ghost" aria-hidden="true"><Network size={120} strokeWidth={1} /></span>
              </article>

              {/* Sign-relay (wide split) */}
              <article className="lp-card lp-card--wide lp-card--split">
                <div>
                  <span className="lp-card__icon"><ScrollText size={26} strokeWidth={1.75} /></span>
                  <h3>{t.advantages.signTitle}</h3>
                  <p>{t.advantages.signBody}</p>
                </div>
                <span className="lp-card__ghost" aria-hidden="true"><Waypoints size={120} strokeWidth={1} /></span>
              </article>

              {/* Зөвшөөрөл санах */}
              <article className="lp-card">
                <span className="lp-card__icon"><Users size={26} strokeWidth={1.75} /></span>
                <div>
                  <h3>{t.advantages.consentTitle}</h3>
                  <p>{t.advantages.consentBody}</p>
                </div>
              </article>
            </div>
          </div>
        </section>

        {/* ---------- Технологи split ---------- */}
        <section className="lp-section lp-section--alt" id="tech">
          <div className="lp-container lp-split">
            <div className="lp-split__copy">
              <h2>{t.tech.heading}</h2>
              <p className="lp-split__sub">{t.tech.sub}</p>
              <ul className="lp-feature-list">
                <li>
                  <span className="lp-feature-list__icon"><Terminal size={20} strokeWidth={1.9} /></span>
                  <div>
                    <h4>{t.tech.backendTitle}</h4>
                    <p>{t.tech.backendBody}</p>
                  </div>
                </li>
                <li>
                  <span className="lp-feature-list__icon"><Layers size={20} strokeWidth={1.9} /></span>
                  <div>
                    <h4>{t.tech.frontendTitle}</h4>
                    <p>{t.tech.frontendBody}</p>
                  </div>
                </li>
                <li>
                  <span className="lp-feature-list__icon"><Bot size={20} strokeWidth={1.9} /></span>
                  <div>
                    <h4>{t.tech.aiTitle}</h4>
                    <p>{t.tech.aiBody}</p>
                  </div>
                </li>
              </ul>
            </div>

            {/* Итгэлийн баталгааны карт */}
            <div className="lp-deploy">
              <div className="lp-deploy__bar">
                <span>{t.tech.trustTitle}</span>
                <span className="lp-deploy__badge">{t.tech.trustBadge}</span>
              </div>
              <ul className="lp-deploy__list">
                {t.tech.trustItems.map((item) => (
                  <li key={item}>
                    <span className="lp-deploy__check"><CheckCircle2 size={18} strokeWidth={2} /></span>
                    <span className="lp-deploy__label">{item}</span>
                    <span className="lp-deploy__state">Active</span>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </section>

        {/* ---------- Бүх боломж ---------- */}
        <section className="lp-section">
          <div className="lp-container">
            <header className="lp-head">
              <h2>{t.everything.heading}</h2>
              <p>{t.everything.sub}</p>
            </header>
            <div className="lp-grid">
              {t.everything.items.map((it, i) => {
                const Icon = EVERYTHING_ICONS[i] ?? Sparkles;
                return (
                  <article className="lp-mini" key={it.title}>
                    <span className="lp-mini__icon"><Icon size={20} strokeWidth={1.9} /></span>
                    <h4>{it.title}</h4>
                    <p>{it.body}</p>
                  </article>
                );
              })}
            </div>
          </div>
        </section>

        {/* ---------- Эцсийн CTA ---------- */}
        <section className="lp-cta">
          <div className="lp-cta__pattern" aria-hidden="true" />
          <div className="lp-container lp-cta__inner">
            <h2>{t.cta.title}</h2>
            <p>{t.cta.sub}</p>
            <div className="lp-cta__buttons">
              <a className="lp-btn lp-btn--gold lp-btn--lg" href="#login">
                {t.cta.ctaLogin}
                <ArrowRight size={18} strokeWidth={2} />
              </a>
              <a className="lp-btn lp-btn--glass lp-btn--lg" href="#features">
                <ChevronRight size={18} strokeWidth={2} />
                {t.cta.ctaExplore}
              </a>
            </div>
            <p className="lp-cta__tagline">{t.cta.tagline}</p>
          </div>
        </section>
      </main>

      {/* ---------- Footer ---------- */}
      <footer className="lp-footer">
        <div className="lp-container lp-footer__inner">
          <div className="lp-footer__brand">
            <div className="lp-footer__mark">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img src="/brand.webp" alt="" aria-hidden="true" />
              <span>{brand}</span>
            </div>
            <p>{t.footer.tagline}</p>
          </div>
          <nav className="lp-footer__links" aria-label="Footer">
            {t.footer.links.map((l) => (
              <a href="#top" key={l}>{l}</a>
            ))}
          </nav>
          <p className="lp-footer__copy">{t.footer.copyright}</p>
        </div>
      </footer>
    </div>
  );
}

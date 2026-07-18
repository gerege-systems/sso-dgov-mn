import React from 'react';
import Link from 'next/link';

interface Props {
  /** topbar баруун талын нэмэлт навигаци (сонголттой). Нийтийн хуудасны
   *  харагдацыг админ удирддаг тул энд per-user харагдац солигч байхгүй. */
  rightNav?: React.ReactNode;
  hideFooter?: boolean;
  children: React.ReactNode;
}

/**
 * Анонимос бүрхүүл — landing (/) болон auth хуудаснуудад. Rail / UserMenu / session
 * байхгүй. Брэнд topbar + төвлөрсөн агуулга + footer.
 */
export default function SigninShell({ rightNav, hideFooter, children }: Props) {
  return (
    <div className="signin-shell">
      <header className="signin-shell__nav">
        <Link className="topbar__brand" href="/">
          {/* eslint-disable-next-line @next/next/no-img-element */}
          <img className="topbar__brand-mark" src="/brand.webp" alt="Төрийн нэгдсэн нэвтрэлт" />
          <div className="topbar__brand-text">
            <span className="topbar__brand-name">Төрийн нэгдсэн нэвтрэлт</span>
          </div>
        </Link>
        {rightNav}
      </header>

      <main className="signin-shell__body">{children}</main>

      {!hideFooter && (
        <footer className="signin-footer" style={{ justifyContent: 'center', textAlign: 'center' }}>
          <span>© 2026 Gerege Systems · <span className="mono">Government SSO</span></span>
        </footer>
      )}
    </div>
  );
}

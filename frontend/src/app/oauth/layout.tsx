// eID based AI enabled Government Template Platform V3.0
// OIDC provider (RP-facing) SSO дэлгэцийн base shell — UI-г sso.dgov.mn-ээс
// хуулж авав (base.html). dan-ий өөрийн апп login (/login)-оос ТУСДАА, 3 дагч
// RP-д зориулсан нэгдсэн нэвтрэлтийн дэлгэц.
import Link from 'next/link';
import type { ReactNode } from 'react';

export default function OAuthLayout({ children }: { children: ReactNode }) {
  return (
    <>
      {/* sso.dgov.mn-ээс хуулсан стиль (public/oauth/style.css). */}
      <link rel="stylesheet" href="/oauth/style.css" />
      <div className="bg" style={{ minHeight: '100vh', display: 'flex', flexDirection: 'column' }}>
        <header className="topbar">
          <Link className="brand" href="/oauth">
            <span className="brand-name">DAN — Засгийн газрын нэгдсэн нэвтрэлт</span>
          </Link>
        </header>
        <main className="shell">
          <div className="card-wrap">{children}</div>
        </main>
        <footer className="footer">
          dan.dgov.mn — Government Single Sign On · eID Mongolia + Ory Hydra
        </footer>
      </div>
    </>
  );
}

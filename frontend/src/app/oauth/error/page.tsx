// eID based AI enabled Government Template Platform V3.0
// OIDC provider алдааны хуудас — Hydra URLS_ERROR энд чиглүүлнэ.
import Link from 'next/link';

export const dynamic = 'force-dynamic';

export default async function OAuthErrorPage(props: {
  searchParams: Promise<{ error?: string; error_description?: string }>;
}) {
  const sp = await props.searchParams;
  return (
    <div className="card">
      <div className="eyebrow">Алдаа</div>
      <h1 className="title">Нэвтрэлтийн алдаа</h1>
      <div className="alert alert-error">{sp.error_description || sp.error || 'Тодорхойгүй алдаа'}</div>
      <p className="hint" style={{ marginTop: 16 }}>
        <Link href="/">← Нүүр хуудас руу буцах</Link>
      </p>
    </div>
  );
}

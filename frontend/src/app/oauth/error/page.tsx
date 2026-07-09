// eID based AI enabled Government Template Platform V3.0
// OIDC provider алдааны хуудас — Hydra URLS_ERROR энд чиглүүлнэ.
import Link from 'next/link';

export const dynamic = 'force-dynamic';

export default async function OAuthErrorPage(props: {
  searchParams: Promise<{ error?: string; error_description?: string }>;
}) {
  const sp = await props.searchParams;
  return (
    <main style={{ maxWidth: 480, margin: '0 auto', padding: 24 }}>
      <h1 style={{ fontSize: 20, fontWeight: 600, color: '#b00020' }}>Нэвтрэлтийн алдаа</h1>
      <p style={{ marginTop: 8 }}>{sp.error_description || sp.error || 'Тодорхойгүй алдаа'}</p>
      <p style={{ marginTop: 16 }}>
        <Link href="/">Нүүр хуудас руу буцах</Link>
      </p>
    </main>
  );
}

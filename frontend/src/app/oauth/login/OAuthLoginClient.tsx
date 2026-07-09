// eID based AI enabled Government Template Platform V3.0
'use client';

import { useEffect, useState } from 'react';
import { postJSON } from '@/lib/client';

export default function OAuthLoginClient({ challenge }: { challenge: string }) {
  const [error, setError] = useState('');

  useEffect(() => {
    let done = false;
    (async () => {
      const r = await postJSON<{ redirect_to?: string }>('/api/provider/login/accept', {
        login_challenge: challenge,
      });
      if (done) return;
      if (r.ok && r.data?.redirect_to) {
        window.location.href = r.data.redirect_to;
      } else {
        setError(r.ok ? 'Нэвтрэлт амжилтгүй боллоо' : r.message || 'Алдаа гарлаа');
      }
    })();
    return () => {
      done = true;
    };
  }, [challenge]);

  return (
    <main style={{ display: 'grid', placeItems: 'center', minHeight: '60vh', padding: 24 }}>
      {error ? (
        <p style={{ color: '#b00020' }}>{error}</p>
      ) : (
        <p>Нэвтрэлтийг баталгаажуулж байна…</p>
      )}
    </main>
  );
}

// eID based AI enabled Government Template Platform V3.0
'use client';

import { useEffect, useState } from 'react';
import { getJSON, postJSON } from '@/lib/client';

// Backend provider.ConsentInfo (json tag-гүй тул талбарын нэр том үсгээр).
type ConsentInfo = {
  Challenge: string;
  ClientID: string;
  ClientName: string;
  RequestedScope: string[] | null;
  Skip: boolean;
};

const scopeLabels: Record<string, string> = {
  openid: 'Нэвтрэлт (таны танигч)',
  profile: 'Нэр, профайл',
  email: 'И-мэйл хаяг',
  nationalid: 'Регистрийн дугаар',
  phone: 'Утасны дугаар',
  offline_access: 'Офлайн (үргэлжилсэн) хандалт',
};

export default function ConsentClient({ challenge }: { challenge: string }) {
  const [info, setInfo] = useState<ConsentInfo | null>(null);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  useEffect(() => {
    let done = false;
    (async () => {
      try {
        const data = await getJSON<ConsentInfo>(
          `/api/provider/consent?consent_challenge=${encodeURIComponent(challenge)}`,
        );
        if (done) return;
        if (data.Skip) {
          await submit('accept', data.RequestedScope ?? []);
          return;
        }
        setInfo(data);
      } catch {
        if (!done) setError('Consent мэдээлэл авахад алдаа гарлаа');
      }
    })();
    return () => {
      done = true;
    };
  }, [challenge]);

  async function submit(action: 'accept' | 'reject', scopes: string[]) {
    setBusy(true);
    const path =
      action === 'accept' ? '/api/provider/consent/accept' : '/api/provider/consent/reject';
    const r = await postJSON<{ redirect_to?: string }>(path, {
      consent_challenge: challenge,
      grant_scope: scopes,
    });
    if (r.ok && r.data?.redirect_to) {
      window.location.href = r.data.redirect_to;
    } else {
      setError(r.ok ? 'Амжилтгүй боллоо' : r.message || 'Алдаа гарлаа');
      setBusy(false);
    }
  }

  const wrap = { maxWidth: 420, margin: '0 auto', padding: 24 } as const;
  if (error) return <main style={wrap}><p style={{ color: '#b00020' }}>{error}</p></main>;
  if (!info) return <main style={wrap}><p>Ачааллаж байна…</p></main>;

  const scopes = info.RequestedScope ?? [];
  return (
    <main style={wrap}>
      <h1 style={{ fontSize: 20, fontWeight: 600, marginBottom: 8 }}>
        {info.ClientName || info.ClientID}
      </h1>
      <p style={{ marginBottom: 12 }}>Дараах мэдээлэлд хандах зөвшөөрөл хүсэж байна:</p>
      <ul style={{ marginBottom: 20, paddingLeft: 18 }}>
        {scopes.map((s) => (
          <li key={s}>{scopeLabels[s] ?? s}</li>
        ))}
      </ul>
      <div style={{ display: 'flex', gap: 12 }}>
        <button disabled={busy} onClick={() => submit('accept', scopes)}>
          Зөвшөөрөх
        </button>
        <button disabled={busy} onClick={() => submit('reject', [])}>
          Татгалзах
        </button>
      </div>
    </main>
  );
}

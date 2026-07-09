// eID based AI enabled Government Template Platform V3.0
// UI-г sso.dgov.mn-ээс (logout.html) хуулав. Гарах товч дарахад ЭХЛЭЭД dan-ий
// өөрийн session-ыг цэвэрлээд (/api/auth/logout), дараа нь Hydra logout challenge-
// ыг accept хийж (/api/provider/logout/accept) RP руу буцна — "dan дээрээ ирж
// logout хийгээд буцна".
'use client';

import { useState } from 'react';
import { postJSON } from '@/lib/client';

export default function OAuthLogoutClient({ challenge }: { challenge: string }) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  async function confirmLogout() {
    if (!challenge) {
      window.location.href = '/';
      return;
    }
    setBusy(true);
    // 1) dan-ий өөрийн session-ыг цэвэрлэнэ (refresh blacklist + cookie устгах).
    await postJSON('/api/auth/logout', undefined).catch(() => {});
    // 2) Hydra logout challenge-ыг accept хийж RP руу буцна.
    const r = await postJSON<{ redirect_to?: string }>('/api/provider/logout/accept', {
      logout_challenge: challenge,
    });
    if (r.ok && r.data?.redirect_to) {
      window.location.href = r.data.redirect_to;
    } else {
      setError(r.ok ? 'Гарах амжилтгүй боллоо' : r.message || 'Алдаа гарлаа');
      setBusy(false);
    }
  }

  return (
    <div className="card">
      <div className="eyebrow">Гарах</div>
      <h1 className="title">Та үнэхээр гарах уу?</h1>
      <p className="sub">
        Энэ нь таныг <strong>dan.dgov.mn</strong>-аас гаргахаас гадна холбогдсон үйлчилгээ рүү
        single-logout сигнал явуулна.
      </p>

      <div className="alert alert-info">
        Single Sign-Out — нэг товшилтоор бүх RP-аас гарна. Дахин нэвтрэхдээ eID Mongolia аппаараа
        дахин баталгаажуулна.
      </div>

      {error && <div className="alert alert-error">{error}</div>}

      <div className="actions">
        <a className="btn btn-secondary" href="/">
          Цуцлах
        </a>
        <button className="btn btn-primary" type="button" disabled={busy} onClick={confirmLogout}>
          Тийм, гарна
        </button>
      </div>
    </div>
  );
}

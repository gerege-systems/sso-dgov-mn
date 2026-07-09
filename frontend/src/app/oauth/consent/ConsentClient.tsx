// eID based AI enabled Government Template Platform V3.0
// UI-г sso.dgov.mn-ээс (consent.html) хуулав; wiring нь dan-ий /api/provider.
'use client';

import { useEffect, useMemo, useRef, useState } from 'react';
import { getJSON, postJSON } from '@/lib/client';

type ConsentInfo = {
  ClientID: string;
  ClientName: string;
  RequestedScope: string[] | null;
  Skip: boolean;
};

// scopeHelp — sso.dgov.mn-ий тайлбаруудыг хуулав.
function scopeHelp(s: string): string {
  switch (s) {
    case 'openid':
      return 'Таныг танихад зориулсан үндсэн ID';
    case 'profile':
      return 'Овог нэр, тохиргооны хэл';
    case 'email':
      return 'И-мэйл хаяг';
    case 'phone':
      return 'Утасны дугаар';
    case 'nationalid':
      return 'Регистрийн дугаар (нууцлалд анхаарч хандана)';
    case 'roles':
      return 'Албан тушаалын эрх';
    default:
      return '';
  }
}

export default function ConsentClient({ challenge }: { challenge: string }) {
  const [info, setInfo] = useState<ConsentInfo | null>(null);
  const [checked, setChecked] = useState<Record<string, boolean>>({});
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);
  const mounted = useRef(true);

  const scopes = useMemo(() => info?.RequestedScope ?? [], [info]);

  useEffect(() => {
    mounted.current = true;
    (async () => {
      try {
        const data = await getJSON<ConsentInfo>(
          `/api/provider/consent?consent_challenge=${encodeURIComponent(challenge)}`,
        );
        if (!mounted.current) return;
        if (data.Skip) {
          await submit('accept', data.RequestedScope ?? []);
          return;
        }
        setInfo(data);
        setChecked(Object.fromEntries((data.RequestedScope ?? []).map((s) => [s, true])));
      } catch {
        if (mounted.current) setError('Consent мэдээлэл авахад алдаа гарлаа');
      }
    })();
    return () => {
      mounted.current = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [challenge]);

  async function submit(action: 'accept' | 'reject', grantScope: string[]) {
    setBusy(true);
    const path = action === 'accept' ? '/api/provider/consent/accept' : '/api/provider/consent/reject';
    const r = await postJSON<{ redirect_to?: string }>(path, {
      consent_challenge: challenge,
      grant_scope: grantScope,
    });
    if (r.ok && r.data?.redirect_to) {
      window.location.href = r.data.redirect_to;
    } else {
      setError(r.ok ? 'Амжилтгүй боллоо' : r.message || 'Алдаа гарлаа');
      setBusy(false);
    }
  }

  if (error) {
    return (
      <div className="card">
        <div className="alert alert-error">{error}</div>
      </div>
    );
  }
  if (!info) {
    return (
      <div className="card">
        <div className="status status-running">
          <div className="spinner" aria-hidden="true" />
          <span className="status-text">Ачааллаж байна…</span>
        </div>
      </div>
    );
  }

  const grant = scopes.filter((s) => checked[s]);
  return (
    <div className="card">
      <span className="rp-chip">
        <span className="rp-chip-icon">RP</span>
        <span>{info.ClientName || info.ClientID}</span>
      </span>
      <div className="eyebrow">Хандах эрх олгох</div>
      <h1 className="title">
        Энэхүү үйлчилгээ танаас
        <br />
        дараах мэдээлэлд хандах эрх хүсэж байна
      </h1>
      <p className="sub">
        Шаардлагатай scope-уудыг тэмдэглээд <strong>Зөвшөөрөх</strong> дарна уу, эсвэл бүхэлд нь{' '}
        <strong>Татгалзаж</strong> болно.
      </p>

      <div className="scope-form">
        <ul className="scopes">
          {scopes.map((s) => (
            <li key={s}>
              <label>
                <input
                  type="checkbox"
                  name="scope"
                  value={s}
                  checked={!!checked[s]}
                  onChange={(e) => setChecked((c) => ({ ...c, [s]: e.target.checked }))}
                />
                <span className="scope-key">{s}</span>
                <span className="scope-desc">{scopeHelp(s)}</span>
              </label>
            </li>
          ))}
        </ul>

        <div className="actions">
          <button
            className="btn btn-danger"
            type="button"
            disabled={busy}
            onClick={() => submit('reject', [])}
          >
            Татгалзах
          </button>
          <button
            className="btn btn-primary"
            type="button"
            disabled={busy}
            onClick={() => submit('accept', grant)}
          >
            Зөвшөөрөх
          </button>
        </div>
      </div>

      <p className="hint">
        Client ID: <code>{info.ClientID}</code>
      </p>
    </div>
  );
}

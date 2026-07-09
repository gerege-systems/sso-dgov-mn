// eID based AI enabled Government Template Platform V3.0
// UI-г sso.dgov.mn-ээс (login_initiate / login_verify / login_qr) хуулж, eID-г
// dan-ий backend (/api/auth/eid/*)-д, challenge accept-ыг /api/provider-д холбов.
'use client';

import { useCallback, useEffect, useRef, useState, type FormEvent } from 'react';
import { QRCodeSVG } from 'qrcode.react';
import { getJSON, postJSON } from '@/lib/client';

type LoginInfo = { ClientID: string; ClientName: string; RequestedScope: string[] | null };
type StartData = { session_id: string; verification_code?: string; device_link_url?: string };
type Method = 'id' | 'qr';
type Phase = 'form' | 'waiting' | 'accepting' | 'expired' | 'refused' | 'error';

const POLL_MS = 2500;

export default function OAuthLoginClient({
  challenge,
  hasSession,
}: {
  challenge: string;
  hasSession: boolean;
}) {
  const [info, setInfo] = useState<LoginInfo | null>(null);
  const [method, setMethod] = useState<Method>('id');
  const [phase, setPhase] = useState<Phase>(hasSession ? 'accepting' : 'form');
  const [nationalId, setNationalId] = useState('');
  const [start, setStart] = useState<StartData | null>(null);
  const [error, setError] = useState('');
  const mounted = useRef(true);

  // Hydra login challenge-ыг accept (dan session аль хэдийн тогтсон).
  const accept = useCallback(async () => {
    setPhase('accepting');
    const r = await postJSON<{ redirect_to?: string }>('/api/provider/login/accept', {
      login_challenge: challenge,
    });
    if (r.ok && r.data?.redirect_to) {
      window.location.href = r.data.redirect_to;
    } else {
      setError(r.ok ? 'Нэвтрэлт амжилтгүй боллоо' : r.message || 'Алдаа гарлаа');
      setPhase('error');
    }
  }, [challenge]);

  // Нэвтрэлтийг цуцлах — Hydra login-ыг reject хийж RP руу access_denied-ээр буцна.
  const cancel = useCallback(async () => {
    const r = await postJSON<{ redirect_to?: string }>('/api/provider/login/reject', {
      login_challenge: challenge,
    });
    window.location.href = r.ok && r.data?.redirect_to ? r.data.redirect_to : '/';
  }, [challenge]);

  useEffect(() => {
    mounted.current = true;
    getJSON<LoginInfo>(`/api/provider/login?login_challenge=${encodeURIComponent(challenge)}`)
      .then((d) => {
        if (mounted.current) setInfo(d);
      })
      .catch(() => {});
    if (hasSession) void accept();
    return () => {
      mounted.current = false;
    };
  }, [challenge, hasSession, accept]);

  const poll = useCallback(
    async (sessionId: string) => {
      if (!mounted.current) return;
      const r = await postJSON<{ state?: string }>('/api/auth/eid/poll', { session_id: sessionId });
      if (!mounted.current) return;
      const state = r.ok ? r.data?.state : undefined;
      if (state === 'COMPLETE') {
        void accept();
        return;
      }
      if (state === 'EXPIRED') {
        setPhase('expired');
        return;
      }
      if (state === 'REFUSED') {
        setPhase('refused');
        return;
      }
      setTimeout(() => poll(sessionId), POLL_MS);
    },
    [accept],
  );

  const beginId = async (e: FormEvent) => {
    e.preventDefault();
    setError('');
    const r = await postJSON<StartData>('/api/auth/eid/start-id', {
      national_id: nationalId.trim(),
      callbackUrl: '',
    });
    if (!r.ok || !r.data?.session_id) {
      setError(r.ok ? 'eID нэвтрэлтийг эхлүүлж чадсангүй' : r.message || 'Алдаа гарлаа');
      return;
    }
    setStart(r.data);
    setPhase('waiting');
    void poll(r.data.session_id);
  };

  const beginQr = useCallback(async () => {
    setError('');
    const r = await postJSON<StartData>('/api/auth/eid/start', { callbackUrl: '' });
    if (!r.ok || !r.data?.session_id) {
      setError(r.ok ? 'QR код үүсгэж чадсангүй' : r.message || 'Алдаа гарлаа');
      return;
    }
    setStart(r.data);
    setPhase('waiting');
    void poll(r.data.session_id);
  }, [poll]);

  const switchMethod = (m: Method) => {
    if (m === method && phase === 'form') return;
    setMethod(m);
    setStart(null);
    setError('');
    setPhase('form');
    if (m === 'qr') void beginQr();
  };

  const vcode = start?.verification_code;

  // Бүх дэлгэц дээр тавих толгой — аль RP-ээс нэвтэрч буй.
  const rpHeader = info?.ClientName ? (
    <div className="rp-head">
      <h1 className="rp-title">{info.ClientName}</h1>
      <p className="rp-tagline">DAN — нэгдсэн нэвтрэлтээр нэвтрэх гэж байна</p>
    </div>
  ) : null;

  // --- accepting / redirecting ---
  if (phase === 'accepting') {
    return (
      <div className="card">
        {rpHeader}
        <div className="status status-running">
          <div className="spinner" aria-hidden="true" />
          <span className="status-text">Нэвтрэлтийг баталгаажуулж байна…</span>
        </div>
      </div>
    );
  }

  // --- waiting for eID approval ---
  if (phase === 'waiting') {
    return (
      <div className="card">
        {rpHeader}
        <div className="eyebrow">eID апп дээр баталгаажуулна уу</div>
        <h1 className="title">
          Доорх дугаар утсан дээр
          <br />
          харагдаж байгаа эсэхийг шалга
        </h1>

        {method === 'qr' && start?.device_link_url && (
          <div className="qr-wrap">
            <QRCodeSVG value={start.device_link_url} size={210} level="M" includeMargin={false} />
          </div>
        )}

        {vcode && (
          <div className="vcode" aria-label="Verification code">
            {vcode.split('').map((c, i) => (
              <span className="vcode-digit" key={i}>
                {c}
              </span>
            ))}
          </div>
        )}

        <p className="sub">
          Гар утсаа авч <strong>eID Mongolia</strong> апп руу орно уу. Баталгаажуулах дугаар нь
          дээрхтэй <strong>яг ижил</strong> байх ёстой. Тохирвол <em>Зөвшөөрөх</em> дарна уу.
        </p>

        <div id="status" className="status status-running">
          <div className="spinner" aria-hidden="true" />
          <span className="status-text">eID апп-аас баталгаажуулалт хүлээж байна…</span>
        </div>

        <div className="actions" style={{ marginTop: 8 }}>
          <button className="btn btn-ghost" type="button" onClick={() => switchMethod(method)}>
            Дахин эхлүүлэх
          </button>
          <button className="btn btn-ghost" type="button" onClick={cancel}>
            Цуцлах
          </button>
        </div>
      </div>
    );
  }

  // --- expired / refused / error ---
  if (phase === 'expired' || phase === 'refused' || phase === 'error') {
    const msg =
      phase === 'expired'
        ? 'Хугацаа дууссан. Дахин эхлүүлнэ үү.'
        : phase === 'refused'
          ? 'Нэвтрэх хүсэлт цуцлагдсан байна.'
          : error || 'Алдаа гарлаа.';
    return (
      <div className="card">
        {rpHeader}
        <div className="alert alert-error">{msg}</div>
        <div className="actions actions-fill" style={{ marginTop: 12 }}>
          <button className="btn btn-primary" type="button" onClick={() => switchMethod('id')}>
            Дахин нэвтрэх
          </button>
        </div>
      </div>
    );
  }

  // --- form (choose method + input) ---
  return (
    <div className="card card-relative">
      {rpHeader}

      <div className="method-tabs" role="tablist" aria-label="Нэвтрэх арга">
        <button
          type="button"
          className={`method-tab${method === 'id' ? ' is-active' : ''}`}
          role="tab"
          aria-selected={method === 'id'}
          onClick={() => switchMethod('id')}
        >
          РД / Civil ID
        </button>
        <button
          type="button"
          className={`method-tab${method === 'qr' ? ' is-active' : ''}`}
          role="tab"
          aria-selected={method === 'qr'}
          onClick={() => switchMethod('qr')}
        >
          QR код
        </button>
      </div>

      <p className="sub">
        Регистрийн дугаараа оруулна уу. Дараа нь утсан дээрх <strong>eID Mongolia</strong> аппад
        баталгаажуулна.
      </p>

      {error && <div className="alert alert-error">{error}</div>}

      <form onSubmit={beginId} className="form" autoComplete="off">
        <label className="field">
          <span className="field-label">Регистрийн дугаар</span>
          <input
            type="text"
            name="national_id"
            value={nationalId}
            onChange={(e) => setNationalId(e.target.value)}
            placeholder="АА00000000"
            pattern="[A-Za-zА-Яа-я0-9]{6,16}"
            required
            autoFocus
            inputMode="text"
            className="input input-mono"
          />
        </label>
        <div className="actions actions-fill">
          <button className="btn btn-primary" type="submit">
            Үргэлжлүүлэх →
          </button>
        </div>
      </form>

      <div className="actions" style={{ marginTop: 4 }}>
        <button className="btn btn-ghost" type="button" onClick={cancel}>
          Цуцлах
        </button>
      </div>
    </div>
  );
}

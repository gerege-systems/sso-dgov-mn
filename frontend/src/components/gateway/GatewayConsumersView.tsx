"use client";

import React, { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2, Inbox, X, KeyRound, Ban, Copy, Check } from 'lucide-react';
import { getJSON, sendJSON, postJSON } from '@/lib/client';
import type { GwConsumer, GwKey } from '@/lib/gatewayTypes';
import { Loading, EnabledChip, Tags, fmtDateTime, splitList } from './gwShared';

const empty = { username: '', custom_id: '', tags: '' };

export default function GatewayConsumersView() {
  const qc = useQueryClient();
  const [adding, setAdding] = useState(false);
  const [form, setForm] = useState(empty);
  const [openId, setOpenId] = useState<string | null>(null);
  const [err, setErr] = useState('');

  const q = useQuery({ queryKey: ['gw-consumers'], queryFn: () => getJSON<GwConsumer[]>('/api/gateway/consumers') });
  const items = q.data ?? [];

  const refresh = () => qc.invalidateQueries({ queryKey: ['gw-consumers'] });

  const create = async () => {
    setErr('');
    const res = await sendJSON('/api/gateway/consumers', 'POST', {
      username: form.username, custom_id: form.custom_id, tags: splitList(form.tags), enabled: true,
    });
    if (res.ok) { setForm(empty); setAdding(false); await refresh(); }
    else setErr(res.message || 'Үүсгэхэд алдаа гарлаа.');
  };

  const remove = async (c: GwConsumer) => {
    if (!window.confirm(`"${c.username}" хэрэглэгчийг устгах уу? Бүх key нь мөн устана.`)) return;
    setErr('');
    const res = await sendJSON(`/api/gateway/consumers/${c.id}`, 'DELETE');
    if (res.ok) { if (openId === c.id) setOpenId(null); await refresh(); }
    else setErr(res.message || 'Устгахад алдаа гарлаа.');
  };

  return (
    <>
      {err && <div className="alert alert--danger" role="alert">{err}</div>}

      <div style={{ marginBottom: 14, display: 'flex', justifyContent: 'flex-end' }}>
        <button className="btn btn--primary btn--sm" type="button" onClick={() => setAdding((a) => !a)}>
          {adding ? <><X size={14} /> Болих</> : <><Plus size={14} /> Хэрэглэгч нэмэх</>}
        </button>
      </div>

      {adding && (
        <section className="card" style={{ padding: 18, marginBottom: 16 }}>
          <div className="card__head"><div className="card__title"><h2>Шинэ API хэрэглэгч</h2></div></div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px,1fr))', gap: 12 }}>
            <label>Нэр<input className="input" value={form.username} onChange={(e) => setForm({ ...form, username: e.target.value })} placeholder="mobile-app" /></label>
            <label>Custom ID<input className="input" value={form.custom_id} onChange={(e) => setForm({ ...form, custom_id: e.target.value })} placeholder="app-001" /></label>
            <label>Tag<input className="input" value={form.tags} onChange={(e) => setForm({ ...form, tags: e.target.value })} placeholder="internal" /></label>
          </div>
          <div style={{ marginTop: 12 }}>
            <button className="btn btn--primary btn--sm" type="button" onClick={create} disabled={!form.username}>Хадгалах</button>
          </div>
        </section>
      )}

      {q.isPending && <Loading />}
      {!q.isPending && items.length === 0 && (
        <div className="card" style={{ padding: 24 }}><p className="muted"><Inbox size={15} /> Хэрэглэгч алга.</p></div>
      )}

      {items.length > 0 && (
        <div className="card users-table-wrap">
          <table className="users-table">
            <thead><tr><th>Нэр</th><th>Custom ID</th><th>Tag</th><th>Key</th><th>Төлөв</th><th aria-label="actions" /></tr></thead>
            <tbody>
              {items.map((c) => (
                <tr key={c.id}>
                  <td>{c.username}</td>
                  <td className="mono muted">{c.custom_id || '—'}</td>
                  <td><Tags tags={c.tags} /></td>
                  <td>
                    <button className="btn btn--ghost btn--sm" type="button" onClick={() => setOpenId(openId === c.id ? null : c.id)}>
                      <KeyRound size={14} /> {c.key_count}
                    </button>
                  </td>
                  <td><EnabledChip enabled={c.enabled} /></td>
                  <td className="users-table__actions">
                    <button className="btn btn--ghost btn--sm" type="button" title="Устгах" onClick={() => remove(c)}><Trash2 size={14} /></button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {openId && <KeysPanel consumerId={openId} onClose={() => setOpenId(null)} onChanged={refresh} />}
    </>
  );
}

// KeysPanel нь нэг consumer-ийн API key-үүдийг удирдана. Шинэ key үүсгэхэд
// plaintext НЭГ удаа л буцдаг тул түүнийг тод хайрцагт харуулна.
function KeysPanel({ consumerId, onClose, onChanged }: { consumerId: string; onClose: () => void; onChanged: () => void }) {
  const qc = useQueryClient();
  const [label, setLabel] = useState('');
  const [created, setCreated] = useState<GwKey | null>(null);
  const [copied, setCopied] = useState(false);
  const [err, setErr] = useState('');

  const keyQ = useKeys(consumerId);
  const keys = keyQ.data ?? [];

  const invalidate = async () => {
    await qc.invalidateQueries({ queryKey: ['gw-keys', consumerId] });
    onChanged();
  };

  const create = async () => {
    setErr(''); setCreated(null); setCopied(false);
    const res = await postJSON<GwKey>(`/api/gateway/consumers/${consumerId}/keys`, { label });
    if (res.ok && res.data) { setCreated(res.data); setLabel(''); await invalidate(); }
    else setErr(res.message || 'Key үүсгэхэд алдаа гарлаа.');
  };

  const revoke = async (k: GwKey) => {
    setErr('');
    const res = await postJSON(`/api/gateway/keys/${k.id}/revoke`, {});
    if (res.ok) await invalidate(); else setErr(res.message || 'Revoke алдаа.');
  };

  const remove = async (k: GwKey) => {
    if (!window.confirm('Энэ key-г бүрмөсөн устгах уу?')) return;
    setErr('');
    const res = await sendJSON(`/api/gateway/keys/${k.id}`, 'DELETE');
    if (res.ok) await invalidate(); else setErr(res.message || 'Устгах алдаа.');
  };

  const copy = () => {
    if (!created?.plaintext) return;
    navigator.clipboard?.writeText(created.plaintext).then(() => { setCopied(true); }).catch(() => {});
  };

  return (
    <section className="card" style={{ padding: 18, marginTop: 16, borderTop: '3px solid var(--dan-blue-text, #2563eb)' }}>
      <div className="card__head card__head--with-sub">
        <div className="card__title"><KeyRound size={18} style={{ color: 'var(--dan-blue-text)' }} /><h2>API key-үүд</h2></div>
        <button className="btn btn--ghost btn--sm" type="button" onClick={onClose}><X size={14} /> Хаах</button>
      </div>

      {err && <div className="alert alert--danger" role="alert">{err}</div>}

      {created?.plaintext && (
        <div className="alert" style={{ background: 'var(--surface-2,#f3f4f6)', borderLeft: '3px solid var(--success,#16a34a)' }}>
          <p style={{ margin: '0 0 6px', fontWeight: 600 }}>Шинэ key үүслээ — энэ нь зөвхөн ОДОО харагдана, хадгалаарай:</p>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <code className="mono" style={{ wordBreak: 'break-all' }}>{created.plaintext}</code>
            <button className="btn btn--ghost btn--sm" type="button" onClick={copy}>{copied ? <Check size={14} /> : <Copy size={14} />}</button>
          </div>
        </div>
      )}

      <div style={{ display: 'flex', gap: 8, alignItems: 'flex-end', margin: '8px 0 14px' }}>
        <label style={{ flex: 1, maxWidth: 280 }}>Label (сонголттой)
          <input className="input" value={label} onChange={(e) => setLabel(e.target.value)} placeholder="production" />
        </label>
        <button className="btn btn--primary btn--sm" type="button" onClick={create}><Plus size={14} /> Key үүсгэх</button>
      </div>

      {keyQ.isPending && <Loading />}
      {!keyQ.isPending && keys.length === 0 && <p className="muted"><Inbox size={15} /> Key алга.</p>}

      {keys.length > 0 && (
        <table className="users-table">
          <thead><tr><th>Prefix</th><th>Label</th><th>Төлөв</th><th>Үүссэн</th><th aria-label="actions" /></tr></thead>
          <tbody>
            {keys.map((k) => (
              <tr key={k.id}>
                <td className="mono">{k.prefix}…</td>
                <td>{k.label || <span className="muted">—</span>}</td>
                <td>{k.revoked ? <span className="chip chip--danger">Цуцлагдсан</span> : <span className="chip chip--success">Идэвхтэй</span>}</td>
                <td className="mono muted">{fmtDateTime(k.created_at)}</td>
                <td className="users-table__actions">
                  {!k.revoked && <button className="btn btn--ghost btn--sm" type="button" title="Цуцлах" onClick={() => revoke(k)}><Ban size={14} /></button>}
                  <button className="btn btn--ghost btn--sm" type="button" title="Устгах" onClick={() => remove(k)}><Trash2 size={14} /></button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </section>
  );
}

function useKeys(consumerId: string) {
  return useQuery({
    queryKey: ['gw-keys', consumerId],
    queryFn: () => getJSON<GwKey[]>(`/api/gateway/consumers/${consumerId}/keys`),
  });
}

"use client";

import React, { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2, Power, Inbox, X } from 'lucide-react';
import { getJSON, sendJSON } from '@/lib/client';
import type { GwRoute, GwService } from '@/lib/gatewayTypes';
import { Loading, EnabledChip, Methods, splitList } from './gwShared';

const METHODS = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'];
const empty = { service_id: '', name: '', methods: ['GET'] as string[], paths: '', strip_path: true };

export default function GatewayRoutesView() {
  const qc = useQueryClient();
  const [adding, setAdding] = useState(false);
  const [form, setForm] = useState(empty);
  const [err, setErr] = useState('');

  const routesQ = useQuery({ queryKey: ['gw-routes'], queryFn: () => getJSON<GwRoute[]>('/api/gateway/routes') });
  const servicesQ = useQuery({ queryKey: ['gw-services'], queryFn: () => getJSON<GwService[]>('/api/gateway/services') });
  const items = routesQ.data ?? [];
  const services = servicesQ.data ?? [];

  const refresh = () => qc.invalidateQueries({ queryKey: ['gw-routes'] });

  const toggleMethod = (m: string) =>
    setForm((f) => ({ ...f, methods: f.methods.includes(m) ? f.methods.filter((x) => x !== m) : [...f.methods, m] }));

  const create = async () => {
    setErr('');
    const res = await sendJSON('/api/gateway/routes', 'POST', {
      service_id: form.service_id, name: form.name, methods: form.methods,
      paths: splitList(form.paths), strip_path: form.strip_path, enabled: true,
    });
    if (res.ok) { setForm(empty); setAdding(false); await refresh(); }
    else setErr(res.message || 'Үүсгэхэд алдаа гарлаа.');
  };

  const toggle = async (r: GwRoute) => {
    setErr('');
    const res = await sendJSON(`/api/gateway/routes/${r.id}`, 'PUT', {
      service_id: r.service_id, name: r.name, methods: r.methods, paths: r.paths,
      strip_path: r.strip_path, preserve_host: r.preserve_host, enabled: !r.enabled,
    });
    if (res.ok) await refresh(); else setErr(res.message || 'Шинэчлэхэд алдаа гарлаа.');
  };

  const remove = async (r: GwRoute) => {
    if (!window.confirm(`"${r.name}" маршрутыг устгах уу?`)) return;
    setErr('');
    const res = await sendJSON(`/api/gateway/routes/${r.id}`, 'DELETE');
    if (res.ok) await refresh(); else setErr(res.message || 'Устгахад алдаа гарлаа.');
  };

  return (
    <>
      {err && <div className="alert alert--danger" role="alert">{err}</div>}

      <div style={{ marginBottom: 14, display: 'flex', justifyContent: 'flex-end' }}>
        <button className="btn btn--primary btn--sm" type="button" onClick={() => setAdding((a) => !a)} disabled={services.length === 0}>
          {adding ? <><X size={14} /> Болих</> : <><Plus size={14} /> Маршрут нэмэх</>}
        </button>
      </div>

      {services.length === 0 && !servicesQ.isPending && (
        <div className="alert alert--danger" role="alert">Эхлээд дор хаяж нэг сервис үүсгэнэ үү.</div>
      )}

      {adding && (
        <section className="card" style={{ padding: 18, marginBottom: 16 }}>
          <div className="card__head"><div className="card__title"><h2>Шинэ маршрут</h2></div></div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px,1fr))', gap: 12 }}>
            <label>Нэр<input className="input" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="pay-charge" /></label>
            <label>Сервис
              <select className="input" value={form.service_id} onChange={(e) => setForm({ ...form, service_id: e.target.value })}>
                <option value="">— сонгох —</option>
                {services.map((s) => <option key={s.id} value={s.id}>{s.name}</option>)}
              </select>
            </label>
            <label>Зам (зай/таслалаар)<input className="input" value={form.paths} onChange={(e) => setForm({ ...form, paths: e.target.value })} placeholder="/pay/charge" /></label>
          </div>
          <div style={{ marginTop: 12 }}>
            <span className="muted" style={{ fontSize: 13 }}>Method:</span>
            <div className="segmented" style={{ marginTop: 6, flexWrap: 'wrap' }}>
              {METHODS.map((m) => (
                <button key={m} type="button" className={`segmented__item${form.methods.includes(m) ? ' is-active' : ''}`} onClick={() => toggleMethod(m)}>{m}</button>
              ))}
            </div>
          </div>
          <label style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 12 }}>
            <input type="checkbox" checked={form.strip_path} onChange={(e) => setForm({ ...form, strip_path: e.target.checked })} /> strip_path
          </label>
          <div style={{ marginTop: 12 }}>
            <button className="btn btn--primary btn--sm" type="button" onClick={create} disabled={!form.name || !form.service_id || !form.paths.trim()}>Хадгалах</button>
          </div>
        </section>
      )}

      {routesQ.isPending && <Loading />}
      {!routesQ.isPending && items.length === 0 && (
        <div className="card" style={{ padding: 24 }}><p className="muted"><Inbox size={15} /> Маршрут алга.</p></div>
      )}

      {items.length > 0 && (
        <div className="card users-table-wrap">
          <table className="users-table">
            <thead><tr><th>Нэр</th><th>Method</th><th>Зам</th><th>Сервис</th><th>Төлөв</th><th aria-label="actions" /></tr></thead>
            <tbody>
              {items.map((r) => (
                <tr key={r.id}>
                  <td>{r.name}</td>
                  <td><Methods methods={r.methods} /></td>
                  <td className="mono">{r.paths.join(', ')}</td>
                  <td>{r.service_name}</td>
                  <td><EnabledChip enabled={r.enabled} /></td>
                  <td className="users-table__actions">
                    <button className="btn btn--ghost btn--sm" type="button" title={r.enabled ? 'Идэвхгүй болгох' : 'Идэвхжүүлэх'} onClick={() => toggle(r)}><Power size={14} /></button>
                    <button className="btn btn--ghost btn--sm" type="button" title="Устгах" onClick={() => remove(r)}><Trash2 size={14} /></button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </>
  );
}

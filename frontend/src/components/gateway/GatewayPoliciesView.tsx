"use client";

import React, { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2, Power, Inbox, X } from 'lucide-react';
import { getJSON, sendJSON } from '@/lib/client';
import type { GwPolicy, GwPolicyType, GwRoute } from '@/lib/gatewayTypes';
import { Loading, EnabledChip } from './gwShared';

const TYPES: { value: GwPolicyType; label: string; sample: string }[] = [
  { value: 'rate-limit', label: 'Rate limit', sample: '{"limit":60,"window":"minute"}' },
  { value: 'key-auth', label: 'Key auth', sample: '{"key_in":"header","header_name":"x-api-key"}' },
  { value: 'cors', label: 'CORS', sample: '{"origins":["https://web.dgov.mn"],"methods":["GET","POST"]}' },
  { value: 'ip-restriction', label: 'IP restriction', sample: '{"allow":["10.0.0.0/8"]}' },
  { value: 'request-transform', label: 'Request transform', sample: '{"add_headers":{"x-env":"prod"}}' },
];

export default function GatewayPoliciesView() {
  const qc = useQueryClient();
  const [adding, setAdding] = useState(false);
  const [type, setType] = useState<GwPolicyType>('rate-limit');
  const [routeId, setRouteId] = useState('');
  const [config, setConfig] = useState(TYPES[0].sample);
  const [err, setErr] = useState('');

  const policiesQ = useQuery({ queryKey: ['gw-policies'], queryFn: () => getJSON<GwPolicy[]>('/api/gateway/policies') });
  const routesQ = useQuery({ queryKey: ['gw-routes'], queryFn: () => getJSON<GwRoute[]>('/api/gateway/routes') });
  const items = policiesQ.data ?? [];
  const routes = routesQ.data ?? [];

  const refresh = () => qc.invalidateQueries({ queryKey: ['gw-policies'] });

  const pickType = (t: GwPolicyType) => {
    setType(t);
    setConfig(TYPES.find((x) => x.value === t)?.sample ?? '{}');
  };

  const create = async () => {
    setErr('');
    let parsed: unknown;
    try { parsed = config.trim() ? JSON.parse(config) : {}; }
    catch { setErr('Config JSON буруу байна.'); return; }
    const res = await sendJSON('/api/gateway/policies', 'POST', {
      type, route_id: routeId || undefined, config: parsed, enabled: true,
    });
    if (res.ok) { setAdding(false); setRouteId(''); await refresh(); }
    else setErr(res.message || 'Үүсгэхэд алдаа гарлаа.');
  };

  const toggle = async (p: GwPolicy) => {
    setErr('');
    const res = await sendJSON(`/api/gateway/policies/${p.id}`, 'PUT', {
      type: p.type, route_id: p.route_id || undefined, config: p.config ?? {}, enabled: !p.enabled,
    });
    if (res.ok) await refresh(); else setErr(res.message || 'Шинэчлэхэд алдаа гарлаа.');
  };

  const remove = async (p: GwPolicy) => {
    if (!window.confirm('Энэ бодлогыг устгах уу?')) return;
    setErr('');
    const res = await sendJSON(`/api/gateway/policies/${p.id}`, 'DELETE');
    if (res.ok) await refresh(); else setErr(res.message || 'Устгахад алдаа гарлаа.');
  };

  return (
    <>
      {err && <div className="alert alert--danger" role="alert">{err}</div>}

      <div style={{ marginBottom: 14, display: 'flex', justifyContent: 'flex-end' }}>
        <button className="btn btn--primary btn--sm" type="button" onClick={() => setAdding((a) => !a)}>
          {adding ? <><X size={14} /> Болих</> : <><Plus size={14} /> Бодлого нэмэх</>}
        </button>
      </div>

      {adding && (
        <section className="card" style={{ padding: 18, marginBottom: 16 }}>
          <div className="card__head"><div className="card__title"><h2>Шинэ бодлого</h2></div></div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px,1fr))', gap: 12 }}>
            <label>Төрөл
              <select className="input" value={type} onChange={(e) => pickType(e.target.value as GwPolicyType)}>
                {TYPES.map((t) => <option key={t.value} value={t.value}>{t.label}</option>)}
              </select>
            </label>
            <label>Маршрут (хоосон = global)
              <select className="input" value={routeId} onChange={(e) => setRouteId(e.target.value)}>
                <option value="">— Global —</option>
                {routes.map((r) => <option key={r.id} value={r.id}>{r.name}</option>)}
              </select>
            </label>
          </div>
          <label style={{ display: 'block', marginTop: 12 }}>Config (JSON)
            <textarea className="input mono" rows={4} value={config} onChange={(e) => setConfig(e.target.value)} style={{ fontFamily: 'var(--font-mono, monospace)' }} />
          </label>
          <div style={{ marginTop: 12 }}>
            <button className="btn btn--primary btn--sm" type="button" onClick={create}>Хадгалах</button>
          </div>
        </section>
      )}

      {policiesQ.isPending && <Loading />}
      {!policiesQ.isPending && items.length === 0 && (
        <div className="card" style={{ padding: 24 }}><p className="muted"><Inbox size={15} /> Бодлого алга.</p></div>
      )}

      {items.length > 0 && (
        <div className="card users-table-wrap">
          <table className="users-table">
            <thead><tr><th>Төрөл</th><th>Хамрах хүрээ</th><th>Config</th><th>Төлөв</th><th aria-label="actions" /></tr></thead>
            <tbody>
              {items.map((p) => (
                <tr key={p.id}>
                  <td><span className="badge badge--primary">{p.type}</span></td>
                  <td>{p.route_name || <span className="chip chip--neutral">Global</span>}</td>
                  <td className="mono" style={{ maxWidth: 320, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                    {JSON.stringify(p.config)}
                  </td>
                  <td><EnabledChip enabled={p.enabled} /></td>
                  <td className="users-table__actions">
                    <button className="btn btn--ghost btn--sm" type="button" title={p.enabled ? 'Идэвхгүй болгох' : 'Идэвхжүүлэх'} onClick={() => toggle(p)}><Power size={14} /></button>
                    <button className="btn btn--ghost btn--sm" type="button" title="Устгах" onClick={() => remove(p)}><Trash2 size={14} /></button>
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

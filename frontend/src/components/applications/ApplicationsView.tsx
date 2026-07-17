"use client";

import React, { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Plus, Trash2, Inbox, X, Copy, Check, Settings2, RefreshCw, Pencil } from 'lucide-react';
import { getJSON, sendJSON, postJSON } from '@/lib/client';
import type { GwService } from '@/lib/gatewayTypes';
import type { Application, AppType } from '@/lib/applicationTypes';
import { APP_TYPES, needsRedirect, isConfidential } from '@/lib/applicationTypes';
import { useT } from '@/lib/lang';
import { Loading, EnabledChip, Tags, splitList } from '../gateway/gwShared';

const empty = { name: '', app_type: 'web' as AppType, redirect_uris: '', tags: '' };

// useServices нь service picker-т зориулж бүх gateway service-ийг татна.
function useServices() {
  return useQuery({ queryKey: ['gw-services'], queryFn: () => getJSON<GwService[]>('/api/gateway/services') });
}

export default function ApplicationsView() {
  const qc = useQueryClient();
  const { T, lang } = useT();
  const [adding, setAdding] = useState(false);
  const [form, setForm] = useState(empty);
  const [pickedServices, setPickedServices] = useState<string[]>([]);
  const [created, setCreated] = useState<Application | null>(null);
  const [openId, setOpenId] = useState<string | null>(null);
  const [err, setErr] = useState('');

  const q = useQuery({ queryKey: ['applications'], queryFn: () => getJSON<Application[]>('/api/applications') });
  const items = q.data ?? [];
  const svcQ = useServices();
  const services = svcQ.data ?? [];

  const typeLabel = (t: AppType) => {
    const found = APP_TYPES.find((x) => x.value === t);
    return found ? (lang === 'en' ? found.en : found.mn) : t;
  };

  const refresh = () => qc.invalidateQueries({ queryKey: ['applications'] });

  const togglePick = (id: string) =>
    setPickedServices((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]));

  const resetForm = () => { setForm(empty); setPickedServices([]); };

  const create = async () => {
    setErr(''); setCreated(null);
    const res = await postJSON<Application>('/api/applications', {
      name: form.name,
      app_type: form.app_type,
      redirect_uris: needsRedirect(form.app_type) ? splitList(form.redirect_uris) : [],
      tags: splitList(form.tags),
      service_ids: pickedServices,
      enabled: true,
    });
    if (res.ok && res.data) { setCreated(res.data); resetForm(); setAdding(false); await refresh(); }
    else setErr(res.message || T('apps.createErr'));
  };

  const remove = async (a: Application) => {
    if (!window.confirm(T('apps.deleteConfirm').replace('{name}', a.name))) return;
    setErr('');
    const res = await sendJSON(`/api/applications/${a.id}`, 'DELETE');
    if (res.ok) { if (openId === a.id) setOpenId(null); await refresh(); }
    else setErr(res.message || T('apps.deleteErr'));
  };

  return (
    <>
      {err && <div className="alert alert--danger" role="alert">{err}</div>}

      {created && <SecretBox app={created} onClose={() => setCreated(null)} clientIdLabel={T('apps.clientId')} secretLabel={T('apps.secretOnce')} />}

      <div style={{ marginBottom: 14, display: 'flex', justifyContent: 'flex-end' }}>
        <button className="btn btn--primary btn--sm" type="button" onClick={() => setAdding((a) => !a)}>
          {adding ? <><X size={14} /> {T('apps.cancel')}</> : <><Plus size={14} /> {T('apps.add')}</>}
        </button>
      </div>

      {adding && (
        <section className="card" style={{ padding: 18, marginBottom: 16 }}>
          <div className="card__head"><div className="card__title"><h2>{T('apps.new')}</h2></div></div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px,1fr))', gap: 12 }}>
            <label>{T('apps.name')}
              <input className="input" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} placeholder="mobile-app" />
            </label>
            <label>{T('apps.type')}
              <select className="input" value={form.app_type} onChange={(e) => setForm({ ...form, app_type: e.target.value as AppType })}>
                {APP_TYPES.map((t) => <option key={t.value} value={t.value}>{lang === 'en' ? t.en : t.mn}</option>)}
              </select>
            </label>
            <label>{T('apps.tags')}
              <input className="input" value={form.tags} onChange={(e) => setForm({ ...form, tags: e.target.value })} placeholder="internal" />
            </label>
          </div>

          {needsRedirect(form.app_type) && (
            <label style={{ display: 'block', marginTop: 12 }}>{T('apps.redirectUris')}
              <textarea className="input" rows={3} value={form.redirect_uris}
                onChange={(e) => setForm({ ...form, redirect_uris: e.target.value })}
                placeholder="https://app.example.mn/callback, myapp://auth" />
              <span className="muted" style={{ fontSize: 12 }}>{T('apps.redirectHint')}</span>
            </label>
          )}

          <div style={{ marginTop: 12 }}>
            <span className="sidepanel__group-label" style={{ display: 'block', marginBottom: 6 }}>{T('apps.services')}</span>
            <ServiceChecklist services={services} loading={svcQ.isPending} checked={pickedServices} onToggle={togglePick} emptyLabel={T('apps.noServices')} />
          </div>

          <div style={{ marginTop: 12 }}>
            <button className="btn btn--primary btn--sm" type="button" onClick={create} disabled={!form.name}>{T('apps.save')}</button>
          </div>
        </section>
      )}

      {q.isPending && <Loading />}
      {!q.isPending && items.length === 0 && (
        <div className="card" style={{ padding: 24 }}><p className="muted"><Inbox size={15} /> {T('apps.empty')}</p></div>
      )}

      {items.length > 0 && (
        <div className="card users-table-wrap">
          <table className="users-table">
            <thead><tr><th>{T('apps.name')}</th><th>{T('apps.clientId')}</th><th>{T('apps.type')}</th><th>{T('apps.tags')}</th><th>{T('apps.services')}</th><th>{T('apps.status')}</th><th aria-label="actions" /></tr></thead>
            <tbody>
              {items.map((a) => (
                <tr key={a.id}>
                  <td>{a.name}</td>
                  <td className="mono muted" style={{ wordBreak: 'break-all' }}>{a.client_id}</td>
                  <td><span className="badge badge--primary" style={{ fontSize: 11 }}>{typeLabel(a.app_type)}</span></td>
                  <td><Tags tags={a.tags} /></td>
                  <td><span className="muted" style={{ display: 'inline-flex', alignItems: 'center', gap: 6 }}><Settings2 size={14} /> {a.service_ids.length}</span></td>
                  <td><EnabledChip enabled={a.enabled} /></td>
                  <td className="users-table__actions">
                    <button className="btn btn--ghost btn--sm" type="button" title={T('apps.edit')} onClick={() => setOpenId(openId === a.id ? null : a.id)}><Pencil size={14} /></button>
                    <button className="btn btn--ghost btn--sm" type="button" title={T('apps.delete')} onClick={() => remove(a)}><Trash2 size={14} /></button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {openId && (() => {
        const app = items.find((a) => a.id === openId);
        if (!app) return null;
        return (
          <EditPanel
            app={app}
            services={services}
            servicesLoading={svcQ.isPending}
            onClose={() => setOpenId(null)}
            onChanged={refresh}
          />
        );
      })()}
    </>
  );
}

// ServiceChecklist нь gateway service-үүдийг checkbox жагсаалтаар үзүүлнэ.
function ServiceChecklist({ services, loading, checked, onToggle, emptyLabel }: {
  services: GwService[]; loading: boolean; checked: string[]; onToggle: (id: string) => void; emptyLabel: string;
}) {
  if (loading) return <Loading />;
  if (services.length === 0) return <p className="muted"><Inbox size={15} /> {emptyLabel}</p>;
  return (
    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px,1fr))', gap: 6 }}>
      {services.map((s) => (
        <label key={s.id} style={{ display: 'flex', alignItems: 'center', gap: 8, cursor: 'pointer' }}>
          <input type="checkbox" checked={checked.includes(s.id)} onChange={() => onToggle(s.id)} />
          <span>{s.name}</span>
        </label>
      ))}
    </div>
  );
}

// SecretBox нь client_id + secret-ийг НЭГ удаагийн тод хайрцагт (copy-той) харуулна.
function SecretBox({ app, onClose, clientIdLabel, secretLabel }: {
  app: Application; onClose: () => void; clientIdLabel: string; secretLabel: string;
}) {
  const [copied, setCopied] = useState('');
  const copy = (val: string, which: string) => {
    navigator.clipboard?.writeText(val).then(() => setCopied(which)).catch(() => {});
  };
  return (
    <div className="alert" style={{ background: 'var(--surface-2,#f3f4f6)', borderLeft: '3px solid var(--success,#16a34a)' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <p style={{ margin: '0 0 8px', fontWeight: 600 }}>{secretLabel}</p>
        <button className="btn btn--ghost btn--sm" type="button" onClick={onClose}><X size={14} /></button>
      </div>
      <div style={{ display: 'grid', gap: 8 }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <span className="muted" style={{ minWidth: 80 }}>{clientIdLabel}</span>
          <code className="mono" style={{ wordBreak: 'break-all', flex: 1 }}>{app.client_id}</code>
          <button className="btn btn--ghost btn--sm" type="button" onClick={() => copy(app.client_id, 'id')}>{copied === 'id' ? <Check size={14} /> : <Copy size={14} />}</button>
        </div>
        {app.secret && (
          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <span className="muted" style={{ minWidth: 80 }}>secret</span>
            <code className="mono" style={{ wordBreak: 'break-all', flex: 1 }}>{app.secret}</code>
            <button className="btn btn--ghost btn--sm" type="button" onClick={() => copy(app.secret!, 'secret')}>{copied === 'secret' ? <Check size={14} /> : <Copy size={14} />}</button>
          </div>
        )}
      </div>
    </div>
  );
}

// EditPanel нь нэг application-ийн бүх талбар (нэр/төрөл/redirect/tag/төлөв/
// service) засах + secret rotate-ийг удирдана.
function EditPanel({ app, services, servicesLoading, onClose, onChanged }: {
  app: Application; services: GwService[]; servicesLoading: boolean;
  onClose: () => void; onChanged: () => void;
}) {
  const { T, lang } = useT();
  const [name, setName] = useState(app.name);
  const [appType, setAppType] = useState<AppType>(app.app_type);
  const [redirects, setRedirects] = useState((app.redirect_uris ?? []).join(', '));
  const [tags, setTags] = useState((app.tags ?? []).join(', '));
  const [enabled, setEnabled] = useState(app.enabled);
  const [checked, setChecked] = useState<string[]>(app.service_ids);
  const [rotated, setRotated] = useState<Application | null>(null);
  const [saving, setSaving] = useState(false);
  const [saved, setSaved] = useState(false);
  const [err, setErr] = useState('');

  const toggle = (id: string) =>
    setChecked((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]));

  const save = async () => {
    setErr(''); setSaved(false); setSaving(true);
    const res = await sendJSON(`/api/applications/${app.id}`, 'PUT', {
      name,
      app_type: appType,
      redirect_uris: needsRedirect(appType) ? splitList(redirects) : [],
      tags: splitList(tags),
      service_ids: checked,
      enabled,
    });
    setSaving(false);
    if (res.ok) { setSaved(true); onChanged(); }
    else setErr(res.message || T('apps.saveErr'));
  };

  const rotate = async () => {
    if (!window.confirm(T('apps.rotateConfirm'))) return;
    setErr(''); setRotated(null);
    const res = await postJSON<Application>(`/api/applications/${app.id}/rotate-secret`, {});
    if (res.ok && res.data) { setRotated(res.data); onChanged(); }
    else setErr(res.message || T('apps.rotateErr'));
  };

  return (
    <section className="card" style={{ padding: 18, marginTop: 16, borderTop: '3px solid var(--dan-blue-text, #2563eb)' }}>
      <div className="card__head card__head--with-sub">
        <div className="card__title"><Pencil size={18} style={{ color: 'var(--dan-blue-text)' }} /><h2>{app.name}</h2></div>
        <button className="btn btn--ghost btn--sm" type="button" onClick={onClose}><X size={14} /> {T('apps.close')}</button>
      </div>

      {err && <div className="alert alert--danger" role="alert">{err}</div>}
      {saved && <div className="alert alert--success" role="status">{T('apps.saved')}</div>}
      {rotated && <SecretBox app={rotated} onClose={() => setRotated(null)} clientIdLabel={T('apps.clientId')} secretLabel={T('apps.secretOnce')} />}

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px,1fr))', gap: 12, marginTop: 8 }}>
        <label>{T('apps.name')}
          <input className="input" value={name} onChange={(e) => setName(e.target.value)} />
        </label>
        <label>{T('apps.type')}
          <select className="input" value={appType} onChange={(e) => setAppType(e.target.value as AppType)}>
            {APP_TYPES.map((t) => <option key={t.value} value={t.value}>{lang === 'en' ? t.en : t.mn}</option>)}
          </select>
        </label>
        <label>{T('apps.tags')}
          <input className="input" value={tags} onChange={(e) => setTags(e.target.value)} />
        </label>
      </div>

      {needsRedirect(appType) && (
        <label style={{ display: 'block', marginTop: 12 }}>{T('apps.redirectUris')}
          <textarea className="input" rows={3} value={redirects} onChange={(e) => setRedirects(e.target.value)} />
          <span className="muted" style={{ fontSize: 12 }}>{T('apps.redirectHint')}</span>
        </label>
      )}

      <label style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 12, cursor: 'pointer' }}>
        <input type="checkbox" checked={enabled} onChange={(e) => setEnabled(e.target.checked)} />
        <span>{T('apps.enabledField')}</span>
      </label>

      <div style={{ marginTop: 12 }}>
        <span className="sidepanel__group-label" style={{ display: 'block', marginBottom: 6 }}>{T('apps.services')}</span>
        <ServiceChecklist services={services} loading={servicesLoading} checked={checked} onToggle={toggle} emptyLabel={T('apps.noServices')} />
      </div>

      <div style={{ marginTop: 14, display: 'flex', gap: 8, flexWrap: 'wrap' }}>
        <button className="btn btn--primary btn--sm" type="button" onClick={save} disabled={saving || !name}>{T('apps.save')}</button>
        {isConfidential(app.app_type) && (
          <button className="btn btn--ghost btn--sm" type="button" onClick={rotate}><RefreshCw size={14} /> {T('apps.rotateSecret')}</button>
        )}
      </div>
    </section>
  );
}

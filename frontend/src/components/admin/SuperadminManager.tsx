"use client";

import { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Trash2, Loader2, Plus, Save, X, ShieldPlus } from 'lucide-react';
import { useT } from '@/lib/lang';
import { getJSON, sendJSON } from '@/lib/client';
import { ROLE_SUPERADMIN, roleLabel } from '@/lib/types';

interface AdminUser {
  id: string;
  username: string;
  full_name?: string;
  full_name_en?: string;
  email: string;
  role_id: number;
  active: boolean;
  created_at: string;
}

interface Props {
  currentUserId: string;
}

const emptyForm = { username: '', email: '', password: '', first_name: '', last_name: '' };

export default function SuperadminManager({ currentUserId }: Props) {
  const { T, lang } = useT();
  const queryClient = useQueryClient();
  const [actionError, setActionError] = useState('');
  const [adding, setAdding] = useState(false);
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState(emptyForm);
  const [grantId, setGrantId] = useState('');

  // Админ түвшний бүртгэлүүд (super admin + admin).
  const adminsQuery = useQuery({
    queryKey: ['superadmin-admins'],
    queryFn: () => getJSON<AdminUser[]>('/api/superadmin/admins'),
  });
  // Эрх олгох боломжтой (админ биш) хэрэглэгчид — /admin/users-ийн эхний хуудас.
  const usersQuery = useQuery({
    queryKey: ['superadmin-grantable'],
    queryFn: () => getJSON<AdminUser[]>('/api/admin/users?limit=200'),
  });

  const admins = adminsQuery.data ?? null;
  const grantable = (usersQuery.data ?? []).filter(
    (u) => u.role_id !== ROLE_SUPERADMIN && !(admins ?? []).some((a) => a.id === u.id),
  );
  const loadError = adminsQuery.isError ? (adminsQuery.error as Error).message || T('superadmin.loadError') : '';
  const error = actionError || loadError;

  const reload = async () => {
    await queryClient.invalidateQueries({ queryKey: ['superadmin-admins'] });
    await queryClient.invalidateQueries({ queryKey: ['superadmin-grantable'] });
  };

  const createAdmin = async () => {
    setActionError('');
    setSaving(true);
    const res = await sendJSON('/api/superadmin/admins', 'POST', form);
    setSaving(false);
    if (res.ok) {
      setForm(emptyForm);
      setAdding(false);
      await reload();
    } else {
      setActionError(res.message || T('superadmin.actionError'));
    }
  };

  const grantAdmin = async () => {
    if (!grantId) return;
    setActionError('');
    const res = await sendJSON(`/api/superadmin/admins/${grantId}/grant`, 'PUT');
    if (res.ok) {
      setGrantId('');
      await reload();
    } else {
      setActionError(res.message || T('superadmin.actionError'));
    }
  };

  const revokeAdmin = async (u: AdminUser) => {
    if (!window.confirm(T('superadmin.revokeConfirm'))) return;
    setActionError('');
    const res = await sendJSON(`/api/superadmin/admins/${u.id}`, 'DELETE');
    if (res.ok) await reload();
    else setActionError(res.message || T('superadmin.actionError'));
  };

  const fmtDate = (iso: string) => {
    try {
      return new Date(iso).toLocaleDateString(lang === 'en' ? 'en-US' : 'mn-MN', {
        year: 'numeric', month: 'short', day: 'numeric',
      });
    } catch {
      return iso;
    }
  };

  return (
    <div className="users">
      {error && <div className="alert alert--danger" role="alert">{error}</div>}

      {/* Шинэ админ үүсгэх + байгаа хэрэглэгчид эрх олгох */}
      <div className="card" style={{ padding: 16, marginBottom: 16, display: 'grid', gap: 16 }}>
        {/* Байгаа хэрэглэгчид админ эрх олгох */}
        <div className="field">
          <label className="field__label">{T('superadmin.promoteTitle')}</label>
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
            <select className="input" value={grantId} onChange={(e) => setGrantId(e.target.value)} style={{ minWidth: 260 }}>
              <option value="">{grantable.length ? T('superadmin.promoteSelect') : T('superadmin.noUsers')}</option>
              {grantable.map((u) => (
                <option key={u.id} value={u.id}>
                  {(lang === 'en' ? u.full_name_en : u.full_name)?.trim() || u.username} — {u.email}
                </option>
              ))}
            </select>
            <button className="btn btn--primary" type="button" onClick={grantAdmin} disabled={!grantId}>
              <ShieldPlus size={16} strokeWidth={2} />
              <span>{T('superadmin.promote')}</span>
            </button>
          </div>
        </div>

        {/* Шинэ админ бүртгэл үүсгэх */}
        {!adding ? (
          <button className="btn btn--secondary" type="button" onClick={() => setAdding(true)} style={{ justifySelf: 'start' }}>
            <Plus size={16} strokeWidth={2} />
            <span>{T('superadmin.addAdmin')}</span>
          </button>
        ) : (
          <div style={{ display: 'grid', gap: 8 }}>
            <label className="field__label">{T('superadmin.createTitle')}</label>
            <div style={{ display: 'grid', gap: 8, gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))' }}>
              <input className="input" placeholder={T('superadmin.username')} value={form.username}
                onChange={(e) => setForm({ ...form, username: e.target.value })} />
              <input className="input" type="email" placeholder={T('superadmin.email')} value={form.email}
                onChange={(e) => setForm({ ...form, email: e.target.value })} />
              <input className="input" type="password" placeholder={T('superadmin.password')} value={form.password}
                onChange={(e) => setForm({ ...form, password: e.target.value })} />
              <input className="input" placeholder={T('superadmin.lastName')} value={form.last_name}
                onChange={(e) => setForm({ ...form, last_name: e.target.value })} />
              <input className="input" placeholder={T('superadmin.firstName')} value={form.first_name}
                onChange={(e) => setForm({ ...form, first_name: e.target.value })} />
            </div>
            <div style={{ display: 'flex', gap: 8 }}>
              <button className="btn btn--primary" type="button" onClick={createAdmin}
                disabled={saving || !form.username || !form.email || form.password.length < 8}>
                {saving ? <Loader2 size={16} strokeWidth={2} className="spin" /> : <Save size={16} strokeWidth={2} />}
                <span>{T('common.create')}</span>
              </button>
              <button className="btn btn--ghost" type="button" onClick={() => { setAdding(false); setForm(emptyForm); }}>
                <X size={16} strokeWidth={2} />
                <span>{T('common.cancel')}</span>
              </button>
            </div>
          </div>
        )}
      </div>

      {adminsQuery.isPending && (
        <div className="muted" style={{ display: 'flex', gap: 8, alignItems: 'center', padding: 16 }}>
          <Loader2 size={16} strokeWidth={2} className="spin" />
          <span>{T('superadmin.loading')}</span>
        </div>
      )}

      {admins !== null && admins.length === 0 && !error && (
        <div className="card" style={{ padding: 24 }}><p className="muted">{T('superadmin.empty')}</p></div>
      )}

      {admins !== null && admins.length > 0 && (
        <div className="card users-table-wrap">
          <table className="users-table">
            <thead>
              <tr>
                <th>{T('users.col.name')}</th>
                <th>{T('users.col.email')}</th>
                <th>{T('users.col.role')}</th>
                <th>{T('users.col.status')}</th>
                <th>{T('users.col.created')}</th>
                <th aria-label="actions" />
              </tr>
            </thead>
            <tbody>
              {admins.map((u) => {
                const isSelf = u.id === currentUserId;
                const isSuper = u.role_id === ROLE_SUPERADMIN;
                const name = (lang === 'en' ? (u.full_name_en?.trim() || u.full_name?.trim()) : u.full_name?.trim());
                return (
                  <tr key={u.id}>
                    <td>
                      {name || u.username}
                      {isSelf && <span className="chip chip--neutral" style={{ marginLeft: 8 }}>{T('users.you')}</span>}
                      {name && <div className="muted mono" style={{ fontSize: 12 }}>@{u.username}</div>}
                    </td>
                    <td className="mono">{u.email}</td>
                    <td>
                      <span className={isSuper ? 'chip chip--success' : 'chip chip--neutral'}>{roleLabel(u.role_id, lang)}</span>
                    </td>
                    <td>
                      {u.active
                        ? <span className="chip chip--success">{T('users.active')}</span>
                        : <span className="chip chip--neutral">{T('users.inactive')}</span>}
                    </td>
                    <td className="mono">{fmtDate(u.created_at)}</td>
                    <td className="users-table__actions">
                      {/* Super admin-г болон өөрийгөө хасахгүй. */}
                      {!isSuper && !isSelf && (
                        <button className="btn btn--ghost btn--sm" type="button" onClick={() => revokeAdmin(u)} title={T('superadmin.revoke')}>
                          <Trash2 size={14} strokeWidth={2} />
                        </button>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

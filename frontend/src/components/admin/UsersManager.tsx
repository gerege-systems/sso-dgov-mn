"use client";

import { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { Trash2, Loader2, Check, Ban, ChevronLeft, ChevronRight } from 'lucide-react';
import { useT } from '@/lib/lang';
import { getJSON, sendJSON } from '@/lib/client';
import { ROLE_SUPERADMIN } from '@/lib/types';

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
interface RoleItem {
  id: number;
  key: string;
  name: string;
}

interface Props {
  currentUserId: string;
  /** readOnly бол зөвхөн харах горим. */
  readOnly?: boolean;
}

// Server-side pagination — backend /admin/users нь offset/limit дэмждэг
// (хамгийн ихдээ 200). total тоо буцаадаггүй тул "дараах" товчийг буцсан
// мөрийн тоо хуудасны хэмжээнээс бага үед идэвхгүй болгоно.
const PAGE_SIZE = 50;

export default function UsersManager({ currentUserId, readOnly = false }: Props) {
  const { T, lang, tRole } = useT();
  const queryClient = useQueryClient();
  const [page, setPage] = useState(0);
  const [actionError, setActionError] = useState('');

  const usersQuery = useQuery({
    queryKey: ['admin-users', page],
    queryFn: () => getJSON<AdminUser[]>(`/api/admin/users?offset=${page * PAGE_SIZE}&limit=${PAGE_SIZE}`),
  });
  const rolesQuery = useQuery({
    queryKey: ['rbac-roles'],
    queryFn: () => getJSON<RoleItem[]>('/api/rbac/roles'),
  });

  const items = usersQuery.data ?? null;
  const roles = rolesQuery.data ?? [];
  const loadError = usersQuery.isError ? (usersQuery.error as Error).message || T('users.loadError') : '';
  const error = actionError || loadError;

  const mutate = async (url: string, method: 'PUT' | 'DELETE', body?: unknown) => {
    setActionError('');
    const res = await sendJSON(url, method, body);
    if (res.ok) {
      await queryClient.invalidateQueries({ queryKey: ['admin-users'] });
    } else {
      setActionError(res.message || T('users.actionError'));
    }
  };

  const changeRole = (id: string, roleId: number) => mutate(`/api/admin/users/${id}/role`, 'PUT', { role_id: roleId });
  const toggleActive = (u: AdminUser) => mutate(`/api/admin/users/${u.id}/active`, 'PUT', { active: !u.active });
  const remove = (id: string) => {
    if (!window.confirm(T('users.deleteConfirm'))) return;
    void mutate(`/api/admin/users/${id}`, 'DELETE');
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

  const hasPrev = page > 0;
  const hasNext = (items?.length ?? 0) === PAGE_SIZE;

  return (
    <div className="users">
      {error && <div className="alert alert--danger" role="alert">{error}</div>}

      {usersQuery.isPending && (
        <div className="muted" style={{ display: 'flex', gap: 8, alignItems: 'center', padding: 16 }}>
          <Loader2 size={16} strokeWidth={2} className="spin" />
          <span>{T('users.loading')}</span>
        </div>
      )}

      {items !== null && items.length === 0 && !error && (
        <div className="card" style={{ padding: 24 }}><p className="muted">{T('users.empty')}</p></div>
      )}

      {items !== null && items.length > 0 && (
        <div className="card users-table-wrap">
          <table className="users-table">
            <thead>
              <tr>
                <th>{T('users.col.name')}</th>
                <th>{T('users.col.email')}</th>
                <th>{T('users.col.role')}</th>
                <th>{T('users.col.status')}</th>
                <th>{T('users.col.created')}</th>
                {!readOnly && <th aria-label="actions" />}
              </tr>
            </thead>
            <tbody>
              {items.map((u) => {
                const isSelf = u.id === currentUserId;
                // Super admin бүртгэлийг энэ хуудаснаас өөрчлөхгүй (backend ч
                // хаадаг) — зөвхөн /admin/superadmin-аар удирдана.
                const isProtected = isSelf || u.role_id === ROLE_SUPERADMIN;
                return (
                  <tr key={u.id}>
                    <td>
                      {(() => {
                        const name = (lang === 'en' ? (u.full_name_en?.trim() || u.full_name?.trim()) : u.full_name?.trim());
                        return (
                          <>
                            {name || u.username}
                            {isSelf && <span className="chip chip--neutral" style={{ marginLeft: 8 }}>{T('users.you')}</span>}
                            {name && <div className="muted mono" style={{ fontSize: 12 }}>@{u.username}</div>}
                          </>
                        );
                      })()}
                    </td>
                    <td className="mono">{u.email}</td>
                    <td>
                      {readOnly || isProtected ? (
                        <span>{(() => {
                          const r = roles.find((x) => x.id === u.role_id);
                          return r ? tRole(r.key, r.name) : u.role_id;
                        })()}</span>
                      ) : (
                        <select
                          className="input users-table__role"
                          value={u.role_id}
                          onChange={(e) => changeRole(u.id, Number(e.target.value))}
                        >
                          {roles
                            .filter((r) => r.id !== ROLE_SUPERADMIN)
                            .map((r) => <option key={r.id} value={r.id}>{tRole(r.key, r.name)}</option>)}
                        </select>
                      )}
                    </td>
                    <td>
                      {u.active
                        ? <span className="chip chip--success">{T('users.active')}</span>
                        : <span className="chip chip--neutral">{T('users.inactive')}</span>}
                    </td>
                    <td className="mono">{fmtDate(u.created_at)}</td>
                    {!readOnly && (
                      <td className="users-table__actions">
                        {!isProtected && (
                          <>
                            <button
                              className="btn btn--ghost btn--sm"
                              type="button"
                              onClick={() => toggleActive(u)}
                              title={u.active ? T('users.deactivate') : T('users.activate')}
                            >
                              {u.active ? <Ban size={14} strokeWidth={2} /> : <Check size={14} strokeWidth={2} />}
                            </button>
                            <button className="btn btn--ghost btn--sm" type="button" onClick={() => remove(u.id)} title={T('common.delete')}>
                              <Trash2 size={14} strokeWidth={2} />
                            </button>
                          </>
                        )}
                      </td>
                    )}
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {items !== null && (hasPrev || hasNext) && (
        <div className="pager">
          <button
            className="btn btn--secondary btn--sm"
            type="button"
            disabled={!hasPrev || usersQuery.isFetching}
            onClick={() => setPage((p) => Math.max(0, p - 1))}
          >
            <ChevronLeft size={14} strokeWidth={2} />
            <span>{T('common.prev')}</span>
          </button>
          <span className="pager__page">{T('common.page')} {page + 1}</span>
          <button
            className="btn btn--secondary btn--sm"
            type="button"
            disabled={!hasNext || usersQuery.isFetching}
            onClick={() => setPage((p) => p + 1)}
          >
            <span>{T('common.next')}</span>
            <ChevronRight size={14} strokeWidth={2} />
          </button>
        </div>
      )}
    </div>
  );
}

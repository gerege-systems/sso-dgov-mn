"use client";

import React, { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { X, Inbox } from 'lucide-react';
import { getJSON, postJSON } from '@/lib/client';
import type { GovApplication } from '@/lib/govTypes';
import { Loading, ApplicationStatus, fmtDate } from './govShared';

export default function GovApplicationsView() {
  const qc = useQueryClient();
  const [err, setErr] = useState('');
  const q = useQuery({ queryKey: ['gov-applications'], queryFn: () => getJSON<GovApplication[]>('/api/gov/applications') });
  const items = q.data ?? [];

  const cancel = async (a: GovApplication) => {
    if (!window.confirm(`"${a.service_name}" хүсэлтийг цуцлах уу?`)) return;
    setErr('');
    const res = await postJSON(`/api/gov/applications/${a.id}/cancel`, {});
    if (res.ok) await qc.invalidateQueries({ queryKey: ['gov-applications'] });
    else setErr(res.message || 'Цуцлахад алдаа гарлаа.');
  };

  const canCancel = (s: string) => s === 'submitted' || s === 'in_review';

  return (
    <>
      {err && <div className="alert alert--danger" role="alert" style={{ marginBottom: 14 }}>{err}</div>}
      {q.isPending && <Loading />}
      {!q.isPending && items.length === 0 && (
        <div className="card" style={{ padding: 24 }}><p className="muted"><Inbox size={15} /> Хүсэлт алга. «Үйлчилгээ» хэсгээс хүсэлт гаргана уу.</p></div>
      )}
      {items.length > 0 && (
        <div className="card users-table-wrap" style={{ margin: 0 }}>
          <table className="users-table">
            <thead><tr><th>Үйлчилгээ</th><th>Лавлах №</th><th>Төлөв</th><th>Огноо</th><th aria-label="actions" /></tr></thead>
            <tbody>
              {items.map((a) => (
                <tr key={a.id}>
                  <td>{a.service_name}{a.note && <div className="muted" style={{ fontSize: 12 }}>{a.note}</div>}</td>
                  <td className="mono">{a.reference_no}</td>
                  <td><ApplicationStatus status={a.status} /></td>
                  <td className="mono muted">{fmtDate(a.submitted_at)}</td>
                  <td className="users-table__actions">
                    {canCancel(a.status) && (
                      <button className="btn btn--ghost btn--sm" type="button" title="Цуцлах" onClick={() => cancel(a)}><X size={14} /></button>
                    )}
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

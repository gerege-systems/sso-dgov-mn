"use client";

// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { FileSignature, LogIn } from 'lucide-react';
import { useT } from '@/lib/lang';
import { formatTS } from '@/lib/format';
import { pkiGet, type PkiActItem } from '@/lib/pki';

// Иргэний eID үйл ажиллагааны (нэвтрэлт + гарын үсэг) лог. Backend
// /api/me/eid/activity (limit/offset дэмждэг). Шүүлтүүр: бүгд/нэвтрэлт/гарын үсэг.
type Filter = 'all' | 'AUTHENTICATION' | 'SIGNATURE';
const PAGE = 20;

export default function EidLogsView({ show }: { show: boolean }) {
  const { T, lang } = useT();
  const [filter, setFilter] = useState<Filter>('all');
  const [limit, setLimit] = useState(PAGE);

  const q = useQuery({
    queryKey: ['eid-pki-logs', limit],
    queryFn: () => pkiGet<{ sessions: PkiActItem[]; total: number }>(`/api/me/eid/activity?limit=${limit}&offset=0`),
    enabled: show,
  });

  if (!show) return null;
  const forbidden = q.data?.status === 403;
  const all = q.data?.data?.sessions ?? [];
  const total = q.data?.data?.total ?? all.length;
  const rows = all.filter((a) => (filter === 'all' ? true : a.flow === filter));

  if (forbidden) {
    return <section className="card"><p className="muted" style={{ padding: '4px 2px' }}>{T('me.pki.pending')}</p></section>;
  }

  return (
    <>
      <div className="segmented segmented--tall" role="tablist" style={{ display: 'flex', marginBottom: 16 }}>
        {(['all', 'AUTHENTICATION', 'SIGNATURE'] as Filter[]).map((f) => (
          <button key={f} type="button" role="tab" aria-selected={filter === f}
            className={`segmented__item${filter === f ? ' is-active' : ''}`} style={{ flex: 1 }}
            onClick={() => setFilter(f)}>
            <span>{f === 'all' ? T('eid.logs.all') : f === 'AUTHENTICATION' ? T('eid.logs.auth') : T('eid.logs.sign')}</span>
          </button>
        ))}
      </div>

      <section className="card" aria-label={T('eid.logs.title')}>
        <div className="card__head card__head--with-sub">
          <div className="card__title"><h2>{T('eid.logs.title')}</h2></div>
          <span className="card__sub">{total} {T('eid.logs.records')}</span>
        </div>
        {rows.length === 0 ? (
          <p className="muted" style={{ padding: '4px 2px' }}>{T('me.pki.none')}</p>
        ) : (
          <div className="pki-list">
            {rows.map((a, i) => (
              <div key={a.session_id || i} className="pki-row">
                {a.flow === 'SIGNATURE' ? <FileSignature size={15} /> : <LogIn size={15} />}
                <span className="pki-row__main">
                  {a.doc_text || (lang === 'en' ? a.flow : a.flow === 'SIGNATURE' ? 'Гарын үсэг' : 'Нэвтрэлт')}
                </span>
                <span className={`badge badge--${a.outcome === 'OK' ? 'success' : 'warning'}`}>{a.outcome}</span>
                {a.timestamp && <span className="pki-row__meta mono">{formatTS(a.timestamp)}</span>}
              </div>
            ))}
          </div>
        )}
        {all.length < total && (
          <button className="btn btn--secondary btn--block" type="button" style={{ marginTop: 12 }}
            onClick={() => setLimit((l) => l + PAGE)} disabled={q.isFetching}>
            {q.isFetching ? T('users.loading') : T('common.next')}
          </button>
        )}
      </section>
    </>
  );
}

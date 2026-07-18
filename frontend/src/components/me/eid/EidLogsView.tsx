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

  // Төрлөөр (Нэвтрэлт / Гарын үсэг) бүлэглэнэ. filter='all' үед хоёулаа, эс бөгөөс
  // сонгосон нэг бүлэг. Хоосон бүлгийг харуулахгүй.
  const groups: { flow: 'AUTHENTICATION' | 'SIGNATURE'; items: PkiActItem[] }[] =
    filter === 'all'
      ? (['AUTHENTICATION', 'SIGNATURE'] as const)
          .map((f) => ({ flow: f, items: rows.filter((a) => a.flow === f) }))
          .filter((g) => g.items.length > 0)
      : [{ flow: filter as 'AUTHENTICATION' | 'SIGNATURE', items: rows }];

  const isSign = (f: string) => f === 'SIGNATURE';
  const flowLabel = (f: string) => (isSign(f) ? T('eid.logs.sign') : T('eid.logs.auth'));
  // Нэвтрэлт — cobalt/цэнхэр, Гарын үсэг — ногоон; icon-оор ялгана.
  const flowColor = (f: string) => (isSign(f) ? 'var(--success, #16a34a)' : 'var(--accent, #6366f1)');

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
          <div style={{ display: 'grid', gap: 18 }}>
            {groups.map((g) => {
              const Icon = isSign(g.flow) ? FileSignature : LogIn;
              const color = flowColor(g.flow);
              return (
                <div key={g.flow} style={{ display: 'grid', gap: 6 }}>
                  {/* Бүлгийн толгой — төрөл + icon + тоо */}
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, padding: '2px 2px 6px' }}>
                    <span style={{
                      display: 'inline-flex', alignItems: 'center', justifyContent: 'center',
                      width: 26, height: 26, borderRadius: 8, color, background: 'color-mix(in oklab, currentColor 14%, transparent)',
                    }}>
                      <Icon size={15} />
                    </span>
                    <span style={{ fontWeight: 600 }}>{flowLabel(g.flow)}</span>
                    <span className="badge badge--neutral" style={{ marginLeft: 2 }}>{g.items.length}</span>
                  </div>
                  <div className="pki-list">
                    {g.items.map((a, i) => (
                      <div key={a.session_id || i} className="pki-row" style={{ alignItems: 'flex-start' }}>
                        <Icon size={15} style={{ color, marginTop: 2, flexShrink: 0 }} />
                        <div className="pki-row__main" style={{ display: 'grid', gap: 2, minWidth: 0 }}>
                          <span style={{ fontWeight: 500 }}>
                            {a.doc_text || (lang === 'en' ? flowLabel(g.flow) : flowLabel(g.flow))}
                          </span>
                          <span className="muted mono" style={{ fontSize: 11 }}>
                            {flowLabel(g.flow)}{a.session_id ? ` · ${a.session_id.slice(0, 8)}` : ''}
                          </span>
                        </div>
                        <span className={`badge badge--${a.outcome === 'OK' ? 'success' : 'warning'}`}>{a.outcome}</span>
                        {a.timestamp && <span className="pki-row__meta mono">{formatTS(a.timestamp)}</span>}
                      </div>
                    ))}
                  </div>
                </div>
              );
            })}
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

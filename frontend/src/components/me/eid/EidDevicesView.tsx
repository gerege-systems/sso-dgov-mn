"use client";

// Government Template Platform V3.0
// Gerege Systems Development Team болон Claude AI хамтран бүтээв, 2026.

import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Smartphone } from 'lucide-react';
import { useT } from '@/lib/lang';
import { formatTS } from '@/lib/format';
import { pkiGet, type PkiDeviceItem } from '@/lib/pki';

// Иргэний eID-д холбогдсон төхөөрөмжийн тусгай хуудас. Backend
// /api/me/eid/devices (бодит PKI — платформ, элсэлт, идэвх, идэвхгүйжсэн огноо).
export default function EidDevicesView({ show }: { show: boolean }) {
  const { T } = useT();
  const q = useQuery({
    queryKey: ['eid-pki-devices'],
    queryFn: () => pkiGet<{ devices: PkiDeviceItem[]; active_count: number; total: number }>('/api/me/eid/devices'),
    enabled: show,
  });

  if (!show) return null;
  const forbidden = q.data?.status === 403;
  const list = q.data?.data?.devices ?? [];
  const activeCount = list.filter((d) => d.active).length;

  if (forbidden) {
    return <section className="card"><p className="muted" style={{ padding: '4px 2px' }}>{T('me.pki.pending')}</p></section>;
  }

  return (
    <>
      <div className="pki-tiles" style={{ marginBottom: 16 }}>
        <div className="pki-tile pki-tile--success"><div className="pki-tile__icon"><Smartphone size={18} /></div><div className="pki-tile__value">{activeCount}</div><div className="pki-tile__label">{T('me.pki.active')}</div></div>
        <div className="pki-tile"><div className="pki-tile__icon"><Smartphone size={18} /></div><div className="pki-tile__value">{list.length}</div><div className="pki-tile__label">{T('eid.certs.total')}</div></div>
      </div>

      <section className="card" aria-label={T('eid.devices.title')}>
        <div className="card__head card__head--with-sub">
          <div className="card__title"><Smartphone size={18} strokeWidth={2} style={{ color: 'var(--dan-blue-text)' }} /><h2>{T('eid.devices.title')}</h2></div>
          <span className="card__sub">{T('me.pki.sub')} <span className="mono">eidmongolia.mn/v3</span></span>
        </div>
        {list.length === 0 ? (
          <p className="muted" style={{ padding: '4px 2px' }}>{T('me.pki.none')}</p>
        ) : (
          <div>
            {list.map((d) => (
              <div key={d.document_number} className="defrow">
                <span className="defrow__label">
                  <Smartphone size={13} style={{ verticalAlign: 'middle', marginRight: 6 }} />
                  <span className="mono">{d.document_number.slice(0, 16)}…</span>
                  {d.platform && <span className="chip chip--neutral" style={{ marginLeft: 8 }}>{d.platform}</span>}
                </span>
                <span className="defrow__value" style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                  <span className={`badge badge--${d.active ? 'success' : 'danger'}`}>{d.active ? T('me.pki.active') : T('me.pki.inactive')}</span>
                  <span className="mono" style={{ fontSize: 12, color: 'var(--muted)' }}>
                    {d.active ? (d.enrolled_at ? formatTS(d.enrolled_at) : '') : (d.deactivated_at ? formatTS(d.deactivated_at) : '')}
                  </span>
                </span>
              </div>
            ))}
          </div>
        )}
      </section>
    </>
  );
}

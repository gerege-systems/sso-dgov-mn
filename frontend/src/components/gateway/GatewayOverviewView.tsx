"use client";

import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { Activity, AlertTriangle, Timer, KeyRound, Server, Route as RouteIcon, Inbox } from 'lucide-react';
import { getJSON } from '@/lib/client';
import type { GwOverview } from '@/lib/gatewayTypes';
import { Loading } from './gwShared';

function StatCard({ icon, value, label }: { icon: React.ReactNode; value: React.ReactNode; label: string }) {
  return (
    <div className="card stat-card" style={{ margin: 0 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8, color: 'var(--dan-blue-text)' }}>{icon}</div>
      <div className="stat-card__value">{value}</div>
      <div className="stat-card__label">{label}</div>
    </div>
  );
}

export default function GatewayOverviewView() {
  const q = useQuery({ queryKey: ['gw-overview'], queryFn: () => getJSON<GwOverview>('/api/gateway/overview') });

  if (q.isPending) return <Loading />;
  if (q.isError) return <div className="alert alert--danger" role="alert">{(q.error as Error).message}</div>;
  const o = q.data!;
  const errPct = (o.error_rate * 100).toFixed(1);
  const maxBucket = Math.max(1, ...o.status_buckets.map((b) => b.count));
  const maxRoute = Math.max(1, ...o.top_routes.map((t) => t.count));

  return (
    <>
      <div className="stat-grid">
        <StatCard icon={<Activity size={18} />} value={o.requests_24h} label="Хүсэлт (24ц)" />
        <StatCard icon={<AlertTriangle size={18} />} value={`${errPct}%`} label={`Алдааны хувь · ${o.errors_24h} ширхэг`} />
        <StatCard icon={<Timer size={18} />} value={`${o.p95_latency_ms}ms`} label={`p95 латент · дунд ${o.avg_latency_ms}ms`} />
        <StatCard icon={<KeyRound size={18} />} value={o.active_keys} label="Идэвхтэй API key" />
        <StatCard icon={<Server size={18} />} value={o.services} label="Сервис" />
        <StatCard icon={<RouteIcon size={18} />} value={o.routes} label="Маршрут" />
      </div>

      <div className="card-grid" style={{ gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))' }}>
        <section className="card">
          <div className="card__head"><div className="card__title"><Activity size={18} style={{ color: 'var(--dan-blue-text)' }} /><h2>Статусын хуваарилалт (24ц)</h2></div></div>
          <div>
            {o.status_buckets.length === 0 && (
              <div className="defrow"><span className="defrow__value muted"><Inbox size={15} /> Өгөгдөл алга.</span></div>
            )}
            {o.status_buckets.map((b) => (
              <div className="defrow" key={b.class}>
                <span className="defrow__label mono">{b.class}</span>
                <span className="defrow__value" style={{ display: 'flex', alignItems: 'center', gap: 8, flex: 1, justifyContent: 'flex-end' }}>
                  <span style={{ flex: 1, height: 6, background: 'var(--surface-2, #eee)', borderRadius: 4, overflow: 'hidden', maxWidth: 160 }}>
                    <span style={{ display: 'block', height: '100%', width: `${(b.count / maxBucket) * 100}%`, background: b.class.startsWith('2') ? 'var(--success,#16a34a)' : b.class.startsWith('4') || b.class.startsWith('5') ? 'var(--danger,#dc2626)' : 'var(--muted)' }} />
                  </span>
                  <span className="mono" style={{ minWidth: 36, textAlign: 'right' }}>{b.count}</span>
                </span>
              </div>
            ))}
          </div>
        </section>

        <section className="card">
          <div className="card__head"><div className="card__title"><RouteIcon size={18} style={{ color: 'var(--dan-blue-text)' }} /><h2>Топ маршрут (24ц)</h2></div></div>
          <div>
            {o.top_routes.length === 0 && (
              <div className="defrow"><span className="defrow__value muted"><Inbox size={15} /> Өгөгдөл алга.</span></div>
            )}
            {o.top_routes.map((t) => (
              <div className="defrow" key={t.route_name}>
                <span className="defrow__label">{t.route_name}</span>
                <span className="defrow__value" style={{ display: 'flex', alignItems: 'center', gap: 8, flex: 1, justifyContent: 'flex-end' }}>
                  <span style={{ flex: 1, height: 6, background: 'var(--surface-2, #eee)', borderRadius: 4, overflow: 'hidden', maxWidth: 160 }}>
                    <span style={{ display: 'block', height: '100%', width: `${(t.count / maxRoute) * 100}%`, background: 'var(--dan-blue-text, #2563eb)' }} />
                  </span>
                  <span className="mono" style={{ minWidth: 36, textAlign: 'right' }}>{t.count}</span>
                </span>
              </div>
            ))}
          </div>
        </section>
      </div>
    </>
  );
}

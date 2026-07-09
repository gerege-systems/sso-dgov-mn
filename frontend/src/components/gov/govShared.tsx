"use client";

import React from 'react';
import { Loader2 } from 'lucide-react';
import type { GovStatus } from '@/lib/govTypes';

// Иргэний Төрийн үйлчилгээний view-уудын хуваалцсан туслахууд. UI мөрүүд (eid
// view-уудын адил) монголоор; gerege-theme классуудыг ашиглана.

export function fmtDate(iso: string | null): string {
  if (!iso) return '—';
  try {
    return new Date(iso).toLocaleDateString('mn-MN', { year: 'numeric', month: 'short', day: 'numeric' });
  } catch {
    return iso;
  }
}

export function fmtDateTime(iso: string | null): string {
  if (!iso) return '—';
  try {
    return new Date(iso).toLocaleString('mn-MN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
  } catch {
    return iso;
  }
}

export function money(n: number, currency = 'MNT'): string {
  const formatted = new Intl.NumberFormat('mn-MN').format(n);
  return currency === 'MNT' ? `${formatted}₮` : `${formatted} ${currency}`;
}

const STATUS_LABEL: Record<GovStatus, string> = {
  submitted: 'Илгээсэн',
  in_review: 'Хянагдаж буй',
  approved: 'Зөвшөөрсөн',
  rejected: 'Татгалзсан',
  completed: 'Дууссан',
  cancelled: 'Цуцалсан',
};

export function ApplicationStatus({ status }: { status: GovStatus }) {
  const klass =
    status === 'approved' || status === 'completed' ? 'chip--success'
      : status === 'rejected' || status === 'cancelled' ? 'chip--danger'
        : 'chip--neutral';
  return <span className={`chip ${klass}`}>{STATUS_LABEL[status] ?? status}</span>;
}

export function Loading({ label = 'Ачаалж байна…' }: { label?: string }) {
  return (
    <div className="muted" style={{ display: 'flex', gap: 8, alignItems: 'center', padding: 16 }}>
      <Loader2 size={16} strokeWidth={2} className="spin" />
      <span>{label}</span>
    </div>
  );
}

export function EmptyRow({ text }: { text: string }) {
  return (
    <div className="defrow"><span className="defrow__value muted">{text}</span></div>
  );
}

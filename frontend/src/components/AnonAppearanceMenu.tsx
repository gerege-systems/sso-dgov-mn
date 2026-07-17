"use client";

import React, { useEffect, useRef, useState } from 'react';
import { Palette } from 'lucide-react';
import AppearanceControls from './AppearanceControls';
import { useT } from '@/lib/lang';

/**
 * Нэвтрэхээс өмнөх хуудаснуудын харагдац солигч — topbar-д Palette товч дарахад
 * өнгө/фонт/нягтралын удирдлагыг popover-оор нээнэ. AnonThemeToggle-ийн хажууд
 * зэрэгцэн байрлана.
 */
export default function AnonAppearanceMenu() {
  const { T } = useT();
  const [open, setOpen] = useState(false);
  const wrapRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    if (!open) return;
    const onDocClick = (e: MouseEvent) => {
      if (!wrapRef.current?.contains(e.target as Node)) setOpen(false);
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        setOpen(false);
        triggerRef.current?.focus();
      }
    };
    document.addEventListener('click', onDocClick);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('click', onDocClick);
      document.removeEventListener('keydown', onKey);
    };
  }, [open]);

  return (
    <div ref={wrapRef} className={`anon-appearance${open ? ' is-open' : ''}`}>
      <button
        ref={triggerRef}
        type="button"
        className="anon-appearance__trigger"
        aria-haspopup="dialog"
        aria-expanded={open}
        aria-label={T('appearance.title')}
        title={T('appearance.title')}
        onClick={(e) => { e.stopPropagation(); setOpen((v) => !v); }}
      >
        <Palette size={16} strokeWidth={2} />
      </button>

      {open && (
        <div className="anon-appearance__popover" role="dialog" aria-label={T('appearance.title')}>
          <div className="anon-appearance__title">{T('appearance.title')}</div>
          <AppearanceControls bare />
        </div>
      )}
    </div>
  );
}

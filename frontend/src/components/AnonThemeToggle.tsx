"use client";

import React from 'react';
import { Sun, Moon, Monitor } from 'lucide-react';
import SegmentedControl from './SegmentedControl';
import { usePreferences } from '@/lib/preferences';

/** Нэвтрэхээс өмнөх хуудаснуудын дээд талын загвар солигч. */
export default function AnonThemeToggle() {
  const { theme, setTheme } = usePreferences();
  return (
    <SegmentedControl
      ariaLabel="Загвар сонгох"
      value={theme}
      onChange={setTheme}
      options={[
        { value: 'light',  icon: <Sun size={14} strokeWidth={2} />,     ariaLabel: 'Гэгээн' },
        { value: 'dark',   icon: <Moon size={14} strokeWidth={2} />,    ariaLabel: 'Харанхуй' },
        { value: 'system', icon: <Monitor size={14} strokeWidth={2} />, ariaLabel: 'Систем' },
      ]}
    />
  );
}

import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import SettingsNote from '@gerege/ui-core/components/admin/SettingsNote';
import AiPromptsManager from '@gerege/ui-core/components/admin/AiPromptsManager';
import { fetchMe, fetchMyPermissions } from '@gerege/ui-core/lib/api';

export const dynamic = 'force-dynamic';
export const metadata = { title: 'Тохиргоо — Админ' };

export default async function AdminSettingsPage() {
  const me = await fetchMe();
  if (!me) redirect('/login?next=/admin/settings');
  const perms = await fetchMyPermissions();
  if (!perms.includes('settings.manage')) redirect('/');

  return (
    <>
      <PageHead eyebrowKey="sys.admin" titleKey="nav.settings" subKey="admin.settings.sub" />
      <AiPromptsManager />
      <SettingsNote />
    </>
  );
}

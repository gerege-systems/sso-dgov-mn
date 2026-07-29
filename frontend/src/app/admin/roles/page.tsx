import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import RolesManager from '@gerege/ui-core/components/admin/RolesManager';
import { fetchMe, fetchMyPermissions } from '@gerege/ui-core/lib/api';

export const dynamic = 'force-dynamic';
export const metadata = { title: 'Эрх (RBAC) — Админ' };

export default async function AdminRolesPage() {
  const me = await fetchMe();
  if (!me) redirect('/login?next=/admin/roles');
  const perms = await fetchMyPermissions();
  if (!perms.includes('roles.manage')) redirect('/');

  return (
    <>
      <PageHead eyebrowKey="sys.admin" titleKey="nav.roles" subKey="admin.roles.sub" />
      <RolesManager />
    </>
  );
}

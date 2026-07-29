import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import SecurityViewer from '@gerege/ui-core/components/admin/SecurityViewer';
import { fetchMe } from '@gerege/ui-core/lib/api';
import { isAdminLevel } from '@gerege/ui-core/lib/types';

export const dynamic = 'force-dynamic';
export const metadata = { title: 'Аюулгүй байдал — Админ' };

export default async function AdminSecurityPage() {
  const me = await fetchMe();
  if (!me) redirect('/login?next=/admin/security');
  if (!isAdminLevel(me.roleId)) redirect('/');

  return (
    <>
      <PageHead eyebrowKey="sys.admin" titleKey="security.title" subKey="security.sub" />
      <SecurityViewer />
    </>
  );
}

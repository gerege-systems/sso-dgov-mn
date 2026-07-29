import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import AuditViewer from '@gerege/ui-core/components/admin/AuditViewer';
import { fetchMe } from '@gerege/ui-core/lib/api';
import { isAdminLevel } from '@gerege/ui-core/lib/types';

export const dynamic = 'force-dynamic';
export const metadata = { title: 'Аудит лог — Админ' };

export default async function AdminAuditPage() {
  const me = await fetchMe();
  if (!me) redirect('/login?next=/admin/audit');
  if (!isAdminLevel(me.roleId)) redirect('/');

  return (
    <>
      <PageHead eyebrowKey="sys.admin" titleKey="audit.title" subKey="audit.sub" />
      <AuditViewer />
    </>
  );
}

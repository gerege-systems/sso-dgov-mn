import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import SuperadminManager from '@gerege/ui-core/components/admin/SuperadminManager';
import { fetchMe } from '@gerege/ui-core/lib/api';
import { isSuperAdmin } from '@gerege/ui-core/lib/types';

export const dynamic = 'force-dynamic';
export const metadata = { title: 'Супер админ — Админуудыг удирдах' };

export default async function SuperadminPage() {
  const me = await fetchMe();
  if (!me) redirect('/login?next=/admin/superadmin');
  // Зөвхөн super admin — энгийн admin ч хандахгүй (least-privilege).
  if (!isSuperAdmin(me.roleId)) redirect('/');

  return (
    <>
      <PageHead eyebrowKey="sys.admin" titleKey="nav.superadmin" subKey="superadmin.sub" />
      <SuperadminManager currentUserId={me.id} />
    </>
  );
}

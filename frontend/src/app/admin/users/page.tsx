import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import UsersManager from '@gerege/ui-core/components/admin/UsersManager';
import { fetchMe, fetchMyPermissions } from '@gerege/ui-core/lib/api';

export const dynamic = 'force-dynamic';
export const metadata = { title: 'Хэрэглэгчид — Админ' };

export default async function AdminUsersPage() {
  const me = await fetchMe();
  if (!me) redirect('/login?next=/admin/users');
  const perms = await fetchMyPermissions();
  if (!perms.includes('users.manage')) redirect('/');

  return (
    <>
      <PageHead eyebrowKey="sys.admin" titleKey="nav.users" subKey="admin.users.sub" />
      <UsersManager currentUserId={me.id} currentUserRoleId={me.roleId} />
    </>
  );
}

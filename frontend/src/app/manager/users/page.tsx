import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import UsersManager from '@gerege/ui-core/components/admin/UsersManager';
import { fetchMe, fetchMyPermissions } from '@gerege/ui-core/lib/api';

export const dynamic = 'force-dynamic';
export const metadata = { title: 'Хэрэглэгчид — Менежер' };

export default async function ManagerUsersPage() {
  const me = await fetchMe();
  if (!me) redirect('/login?next=/manager/users');
  const perms = await fetchMyPermissions();
  if (!perms.includes('users.manage')) redirect('/');

  return (
    <>
      <PageHead eyebrowKey="sys.manager" titleKey="nav.users" subKey="manager.users.sub" />
      <UsersManager currentUserId={me.id} currentUserRoleId={me.roleId} />
    </>
  );
}

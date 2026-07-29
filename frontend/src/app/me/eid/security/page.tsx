import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import EidSecurityView from '@gerege/ui-core/components/me/eid/EidSecurityView';
import { fetchMe } from '@gerege/ui-core/lib/api';
import { pageTitle } from '@/brand.config';

export const dynamic = 'force-dynamic';
export const metadata = { title: pageTitle('eID аюулгүй байдал') };

export default async function EidSecurityPage() {
  const me = await fetchMe();
  if (!me) redirect('/login?next=/me/eid/security');
  return (
    <>
      <PageHead eyebrowKey="sys.user" titleKey="eid.security.title" subKey="eid.security.sub" />
      <EidSecurityView show={!!me.eid} />
    </>
  );
}

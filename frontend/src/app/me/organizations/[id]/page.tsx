import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import OrgDetail from '@gerege/ui-core/components/me/OrgDetail';
import { fetchMe } from '@gerege/ui-core/lib/api';
import { pageTitle } from '@/brand.config';

export const dynamic = 'force-dynamic';
export const metadata = { title: pageTitle('Байгууллага') };

export default async function MeOrganizationDetailPage(props: { params: Promise<{ id: string }> }) {
  const params = await props.params;
  const me = await fetchMe();
  if (!me) redirect(`/login?next=/me/organizations/${params.id}`);

  return (
    <>
      <PageHead eyebrowKey="sys.user" titleKey="org.title" subKey="org.detail" />
      <OrgDetail orgId={params.id} currentUserId={me.id} />
    </>
  );
}

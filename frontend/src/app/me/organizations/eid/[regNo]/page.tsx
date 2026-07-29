import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import OrgManageView from '@gerege/ui-core/components/me/OrgManageView';
import { fetchMe } from '@gerege/ui-core/lib/api';
import { pageTitle } from '@/brand.config';

export const dynamic = 'force-dynamic';
export const metadata = { title: pageTitle('Байгууллага') };

// eID-д бүртгэлтэй, төлөөлдөг байгууллагын удирдах дэлгэц (гарын үсэг зурагч + салгах).
export default async function MeEidOrgManagePage(props: { params: Promise<{ regNo: string }> }) {
  const params = await props.params;
  const me = await fetchMe();
  if (!me) redirect(`/login?next=/me/organizations/eid/${params.regNo}`);

  return (
    <>
      <PageHead eyebrowKey="sys.user" titleKey="org.title" subKey="org.detail" />
      <OrgManageView regNo={decodeURIComponent(params.regNo)} />
    </>
  );
}

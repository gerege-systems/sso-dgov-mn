import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import GovApplicationsView from '@gerege/ui-core/components/gov/GovApplicationsView';
import { fetchMe } from '@gerege/ui-core/lib/api';
import { pageTitle } from '@/brand.config';

export const dynamic = 'force-dynamic';
export const metadata = { title: pageTitle('Миний хүсэлт') };

export default async function MeApplicationsPage() {
  const me = await fetchMe();
  if (!me) redirect('/');
  return (
    <>
      <PageHead eyebrowKey="group.govServices" titleKey="nav.govApplications" subKey="gov.applications.sub" />
      <GovApplicationsView />
    </>
  );
}

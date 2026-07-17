import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@/components/PageHead';
import GovApplicationsView from '@/components/gov/GovApplicationsView';
import { fetchMe } from '@/lib/api';

export const dynamic = 'force-dynamic';
export const metadata = { title: 'Миний хүсэлт — DAN-Government SSO' };

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

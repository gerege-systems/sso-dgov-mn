import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import GovNotificationsView from '@gerege/ui-core/components/gov/GovNotificationsView';
import { fetchMe } from '@gerege/ui-core/lib/api';
import { pageTitle } from '@/brand.config';

export const dynamic = 'force-dynamic';
export const metadata = { title: pageTitle('Мэдэгдэл') };

export default async function MeNotificationsPage() {
  const me = await fetchMe();
  if (!me) redirect('/');
  return (
    <>
      <PageHead eyebrowKey="group.govServices" titleKey="nav.govNotifications" subKey="gov.notifications.sub" />
      <GovNotificationsView />
    </>
  );
}

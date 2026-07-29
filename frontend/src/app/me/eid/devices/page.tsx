import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import EidDevicesView from '@gerege/ui-core/components/me/eid/EidDevicesView';
import { fetchMe } from '@gerege/ui-core/lib/api';
import { pageTitle } from '@/brand.config';

export const dynamic = 'force-dynamic';
export const metadata = { title: pageTitle('Төхөөрөмж') };

export default async function EidDevicesPage() {
  const me = await fetchMe();
  if (!me) redirect('/login?next=/me/eid/devices');
  return (
    <>
      <PageHead eyebrowKey="sys.user" titleKey="eid.devices.title" subKey="eid.devices.sub" />
      <EidDevicesView show={!!me.eid} />
    </>
  );
}

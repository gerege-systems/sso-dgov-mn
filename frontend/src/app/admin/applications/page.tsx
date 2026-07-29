import React from 'react';
import PageHead from '@gerege/ui-core/components/PageHead';
import ApplicationsView from '@gerege/ui-core/components/applications/ApplicationsView';
import { requireGatewayAccess } from '../gateway/guard';

export const dynamic = 'force-dynamic';
export const metadata = { title: 'Applications' };

export default async function Page() {
  await requireGatewayAccess();
  return (
    <>
      <PageHead eyebrowKey="group.gateway" titleKey="nav.applications" subKey="apps.sub" />
      <ApplicationsView />
    </>
  );
}

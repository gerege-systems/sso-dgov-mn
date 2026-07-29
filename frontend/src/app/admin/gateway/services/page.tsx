import React from 'react';
import PageHead from '@gerege/ui-core/components/PageHead';
import GatewayServicesView from '@gerege/ui-core/components/gateway/GatewayServicesView';
import { requireGatewayAccess } from '../guard';

export const dynamic = 'force-dynamic';
export const metadata = { title: 'API Gateway — Сервисүүд' };

export default async function Page() {
  await requireGatewayAccess();
  return (
    <>
      <PageHead eyebrowKey="group.gateway" titleKey="nav.gwServices" subKey="gateway.services.sub" />
      <GatewayServicesView />
    </>
  );
}

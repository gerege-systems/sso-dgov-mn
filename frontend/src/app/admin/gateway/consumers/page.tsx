import React from 'react';
import PageHead from '@/components/PageHead';
import GatewayConsumersView from '@/components/gateway/GatewayConsumersView';
import { requireGatewayAccess } from '../guard';

export const dynamic = 'force-dynamic';
export const metadata = { title: 'API Gateway — Хэрэглэгчид' };

export default async function Page() {
  await requireGatewayAccess();
  return (
    <>
      <PageHead eyebrowKey="group.gateway" titleKey="nav.gwConsumers" subKey="gateway.consumers.sub" />
      <GatewayConsumersView />
    </>
  );
}

import React from 'react';
import PageHead from '@/components/PageHead';
import GatewayRoutesView from '@/components/gateway/GatewayRoutesView';
import { requireGatewayAccess } from '../guard';

export const dynamic = 'force-dynamic';
export const metadata = { title: 'API Gateway — Маршрутууд' };

export default async function Page() {
  await requireGatewayAccess();
  return (
    <>
      <PageHead eyebrowKey="group.gateway" titleKey="nav.gwRoutes" subKey="gateway.routes.sub" />
      <GatewayRoutesView />
    </>
  );
}

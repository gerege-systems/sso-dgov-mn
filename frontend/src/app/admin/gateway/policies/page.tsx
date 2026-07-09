import React from 'react';
import PageHead from '@/components/PageHead';
import GatewayPoliciesView from '@/components/gateway/GatewayPoliciesView';
import { requireGatewayAccess } from '../guard';

export const dynamic = 'force-dynamic';
export const metadata = { title: 'API Gateway — Бодлогууд' };

export default async function Page() {
  await requireGatewayAccess();
  return (
    <>
      <PageHead eyebrowKey="group.gateway" titleKey="nav.gwPolicies" subKey="gateway.policies.sub" />
      <GatewayPoliciesView />
    </>
  );
}

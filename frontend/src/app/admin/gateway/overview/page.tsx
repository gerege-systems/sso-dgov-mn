import React from 'react';
import PageHead from '@gerege/ui-core/components/PageHead';
import GatewayOverviewView from '@gerege/ui-core/components/gateway/GatewayOverviewView';
import { requireGatewayAccess } from '../guard';

export const dynamic = 'force-dynamic';
export const metadata = { title: 'API Gateway — Тойм' };

export default async function Page() {
  await requireGatewayAccess();
  return (
    <>
      <PageHead eyebrowKey="group.gateway" titleKey="nav.gwOverview" subKey="gateway.overview.sub" />
      <GatewayOverviewView />
    </>
  );
}

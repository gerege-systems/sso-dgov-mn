import { pageTitle } from '@/brand.config';
import React from 'react';
import PageHead from '@gerege/ui-core/components/PageHead';
import RegistryOverviewView from '@gerege/ui-core/components/registry/RegistryOverviewView';
import { requireRegistryAccess } from './guard';

export const dynamic = 'force-dynamic';
export const metadata = { title: pageTitle('Үйлчилгээний регистр') };

export default async function Page() {
  await requireRegistryAccess();
  return (
    <>
      <PageHead eyebrowKey="group.registry" titleKey="nav.registryOverview" subKey="registry.overview.sub" />
      <RegistryOverviewView />
    </>
  );
}

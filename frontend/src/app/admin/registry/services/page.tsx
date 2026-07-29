import { pageTitle } from '@/brand.config';
import React from 'react';
import PageHead from '@gerege/ui-core/components/PageHead';
import RegistryServicesView from '@gerege/ui-core/components/registry/RegistryServicesView';
import { requireRegistryAccess } from '../guard';

export const dynamic = 'force-dynamic';
export const metadata = { title: pageTitle('Үйлчилгээний паспорт') };

export default async function Page() {
  await requireRegistryAccess();
  return (
    <>
      <PageHead eyebrowKey="group.registry" titleKey="nav.registryServices" subKey="registry.services.sub" />
      <RegistryServicesView />
    </>
  );
}

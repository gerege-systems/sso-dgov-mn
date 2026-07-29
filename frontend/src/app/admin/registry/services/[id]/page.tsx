import { pageTitle } from '@/brand.config';
import React from 'react';
import PageHead from '@gerege/ui-core/components/PageHead';
import RegistryServiceDetailView from '@gerege/ui-core/components/registry/RegistryServiceDetailView';
import { requireRegistryAccess } from '../../guard';

export const dynamic = 'force-dynamic';
export const metadata = { title: pageTitle('Паспортын дэлгэрэнгүй') };

export default async function Page(props: { params: Promise<{ id: string }> }) {
  await requireRegistryAccess();
  const { id } = await props.params;
  return (
    <>
      <PageHead eyebrowKey="group.registry" titleKey="nav.registryServices" subKey="registry.detail.sub" />
      <RegistryServiceDetailView id={id} />
    </>
  );
}

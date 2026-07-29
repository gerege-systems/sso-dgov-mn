import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import GovReferencesView from '@gerege/ui-core/components/gov/GovReferencesView';
import { fetchMe } from '@gerege/ui-core/lib/api';
import { pageTitle } from '@/brand.config';

export const dynamic = 'force-dynamic';
export const metadata = { title: pageTitle('Лавлагаа') };

export default async function MeReferencesPage() {
  const me = await fetchMe();
  if (!me) redirect('/');
  return (
    <>
      <PageHead eyebrowKey="group.govServices" titleKey="nav.govReferences" subKey="gov.references.sub" />
      <GovReferencesView />
    </>
  );
}

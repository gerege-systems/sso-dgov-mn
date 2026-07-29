import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import EidCertificatesView from '@gerege/ui-core/components/me/eid/EidCertificatesView';
import { fetchMe } from '@gerege/ui-core/lib/api';
import { pageTitle } from '@/brand.config';

export const dynamic = 'force-dynamic';
export const metadata = { title: pageTitle('Гэрчилгээ') };

export default async function EidCertificatesPage() {
  const me = await fetchMe();
  if (!me) redirect('/login?next=/me/eid/certificates');
  return (
    <>
      <PageHead eyebrowKey="sys.user" titleKey="eid.certs.title" subKey="eid.certs.sub" />
      <EidCertificatesView show={!!me.eid} />
    </>
  );
}

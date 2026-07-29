import React from 'react';
import { redirect } from 'next/navigation';
import PageHead from '@gerege/ui-core/components/PageHead';
import OrgRepsCard from '@gerege/ui-core/components/me/OrgRepsCard';
import ImageUploadCard from '@gerege/ui-core/components/me/ImageUploadCard';
import { fetchMe } from '@gerege/ui-core/lib/api';
import { pageTitle } from '@/brand.config';

export const dynamic = 'force-dynamic';
export const metadata = { title: pageTitle('Байгууллага') };

export default async function MeOrganizationsPage() {
  const me = await fetchMe();
  if (!me) redirect('/login?next=/me/organizations');

  return (
    <>
      <PageHead eyebrowKey="sys.user" titleKey="org.title" subKey="org.sub" />
      {/* eID-д бүртгэлтэй, төлөөлдөг байгууллагууд (eidmongolia.mn) */}
      <OrgRepsCard show={!!me.eid} />
      {/* Хувь хүний гарын үсгийн зураг (Google Drive-д хадгална). */}
      <ImageUploadCard
        titleKey="me.assets.signatureTitle"
        hintKey="me.assets.signatureHint"
        path="/api/me/signature"
        queryKey={['my-signature']}
        canEdit
      />
    </>
  );
}

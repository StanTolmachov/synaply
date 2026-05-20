import React from 'react';
import { languages } from '@/lib/languages';
import TermsOfServiceClient from './TermsOfServiceClient';
import { Metadata } from 'next';
import { getTranslations } from 'next-intl/server';

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }): Promise<Metadata> {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: 'common' });
  const baseUrl = "https://synaply.me";

  const languageAlternates: Record<string, string> = {};
  Object.keys(languages).forEach((langCode) => {
    const code = langCode.toLowerCase();
    languageAlternates[code] = `${baseUrl}/${code}/terms`;
  });
  languageAlternates['x-default'] = `${baseUrl}/terms`;
  
  return {
    title: `${t('terms')} | Synaply`,
    description: 'Terms of Service and Privacy Policy for Synaply.',
    alternates: {
      canonical: `${baseUrl}/${locale}/terms`,
      languages: languageAlternates,
    },
    openGraph: {
      title: `${t('terms')} | Synaply`,
      description: 'Terms of Service and Privacy Policy for Synaply.',
      url: `${baseUrl}/${locale}/terms`,
      type: 'website',
    },
  };
}

export default async function TermsOfService({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  const t = await getTranslations({ locale, namespace: 'common' });
  const baseUrl = "https://synaply.me";

  const breadcrumbJsonLd = {
    '@context': 'https://schema.org',
    '@type': 'BreadcrumbList',
    itemListElement: [
      {
        '@type': 'ListItem',
        position: 1,
        name: 'Home',
        item: `${baseUrl}/${locale}`,
      },
      {
        '@type': 'ListItem',
        position: 2,
        name: t('terms'),
        item: `${baseUrl}/${locale}/terms`,
      },
    ],
  };

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }}
      />
      <TermsOfServiceClient />
    </>
  );
}

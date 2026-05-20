import { Metadata } from 'next';
import { getTranslations } from 'next-intl/server';
import { languages } from '@/lib/languages';
import PublicListsClient from './PublicListsClient';

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }): Promise<Metadata> {
  const { locale } = await params;
  const tLists = await getTranslations({ locale, namespace: 'dashboard.public_lists' });
  const baseUrl = "https://synaply.me";

  const languageAlternates: Record<string, string> = {};
  Object.keys(languages).forEach((langCode) => {
    const code = langCode.toLowerCase();
    languageAlternates[code] = `${baseUrl}/${code}/public-lists`;
  });
  languageAlternates['x-default'] = `${baseUrl}/public-lists`;

  return {
    title: `${tLists('title')} | Synaply`,
    description: tLists('subtitle'),
    alternates: {
      canonical: `${baseUrl}/${locale}/public-lists`,
      languages: languageAlternates,
    },
    openGraph: {
      title: `${tLists('title')} | Synaply`,
      description: tLists('subtitle'),
      url: `${baseUrl}/${locale}/public-lists`,
      type: 'website',
    },
  };
}

export default async function PublicListsPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  const tLists = await getTranslations({ locale, namespace: 'dashboard.public_lists' });
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
        name: tLists('title'),
        item: `${baseUrl}/${locale}/public-lists`,
      },
    ],
  };

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }}
      />
      <PublicListsClient />
    </>
  );
}

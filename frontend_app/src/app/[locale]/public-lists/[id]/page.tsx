import { Metadata } from 'next';
import { getTranslations } from 'next-intl/server';
import { languages } from '@/lib/languages';
import PublicListDetailClient from './PublicListDetailClient';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

async function getPublicList(id: string) {
  try {
    const response = await fetch(`${API_BASE}/public-lists/${id}`, {
      next: { revalidate: 3600 }
    });
    if (!response.ok) return null;
    return await response.json();
  } catch (error) {
    console.error('Error fetching public list for metadata:', error);
    return null;
  }
}

export async function generateMetadata({ params }: { params: Promise<{ id: string, locale: string }> }): Promise<Metadata> {
  const { id, locale } = await params;
  const list = await getPublicList(id);
  
  if (!list) {
    return {
      title: 'List Not Found | Synaply',
    };
  }

  const baseUrl = "https://synaply.me";

  const languageAlternates: Record<string, string> = {};
  Object.keys(languages).forEach((langCode) => {
    const code = langCode.toLowerCase();
    languageAlternates[code] = `${baseUrl}/${code}/public-lists/${id}`;
  });
  languageAlternates['x-default'] = `${baseUrl}/public-lists/${id}`;

  return {
    title: `${list.title} | Synaply`,
    description: list.description || `Explore this public word list: ${list.title} (${list.source_lang} to ${list.target_lang})`,
    alternates: {
      canonical: `${baseUrl}/${locale}/public-lists/${id}`,
      languages: languageAlternates,
    },
    openGraph: {
      title: `${list.title} | Synaply`,
      description: list.description,
      url: `${baseUrl}/${locale}/public-lists/${id}`,
      type: 'article',
      images: [
        {
          url: "/opengraph-image.png",
          width: 1200,
          height: 630,
          alt: list.title,
        },
      ],
    },
    twitter: {
      card: 'summary_large_image',
      title: `${list.title} | Synaply`,
      description: list.description,
      images: ["/opengraph-image.png"],
    },
  };
}

export default async function PublicListDetailPage({ params }: { params: Promise<{ id: string, locale: string }> }) {
  const { id, locale } = await params;
  const list = await getPublicList(id);
  const t = await getTranslations({ locale, namespace: 'dashboard.public_lists' });
  const tCommon = await getTranslations({ locale, namespace: 'common' });
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
        name: t('title'),
        item: `${baseUrl}/${locale}/public-lists`,
      },
      {
        '@type': 'ListItem',
        position: 3,
        name: list?.title || id,
        item: `${baseUrl}/${locale}/public-lists/${id}`,
      },
    ],
  };

  const jsonLd = list ? {
    '@context': 'https://schema.org',
    '@type': 'ItemList',
    name: list.title,
    description: list.description,
    numberOfItems: list.words?.length || 0,
    itemListElement: (list.words || []).slice(0, 10).map((word: any, index: number) => ({
      '@type': 'ListItem',
      position: index + 1,
      name: word.original,
      description: word.translation,
    })),
  } : null;

  return (
    <>
      {jsonLd && (
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
        />
      )}
      {breadcrumbJsonLd && (
        <script
          type="application/ld+json"
          dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }}
        />
      )}
      <PublicListDetailClient id={id} />
    </>
  );
}

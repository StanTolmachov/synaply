import React from 'react';
import HelpPageClient from './HelpPageClient';
import { Metadata } from 'next';
import { getTranslations } from 'next-intl/server';
import { languages } from '@/lib/languages';

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }): Promise<Metadata> {
  const { locale } = await params;
  const tHelp = await getTranslations({ locale, namespace: 'help' });
  const baseUrl = "https://synaply.me";

  const languageAlternates: Record<string, string> = {};
  Object.keys(languages).forEach((langCode) => {
    const code = langCode.toLowerCase();
    languageAlternates[code] = `${baseUrl}/${code}/help`;
  });
  languageAlternates['x-default'] = `${baseUrl}/help`;
  
  return {
    title: `${tHelp('title')} | Synaply`,
    description: 'Learn how to use Synaply effectively with our help guide.',
    alternates: {
      canonical: `${baseUrl}/${locale}/help`,
      languages: languageAlternates,
    },
    openGraph: {
      title: `${tHelp('title')} | Synaply`,
      description: 'Learn how to use Synaply effectively with our help guide.',
      url: `${baseUrl}/${locale}/help`,
      type: 'website',
    },
  };
}

export default async function HelpPage({ params }: { params: Promise<{ locale: string }> }) {
  const { locale } = await params;
  const tHelp = await getTranslations({ locale, namespace: 'help' });
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
        name: tHelp('title'),
        item: `${baseUrl}/${locale}/help`,
      },
    ],
  };

  const faqJsonLd = {
    '@context': 'https://schema.org',
    '@type': 'FAQPage',
    mainEntity: [
      {
        '@type': 'Question',
        name: 'What is Synaply?',
        acceptedAnswer: {
          '@type': 'Answer',
          text: 'Synaply is an AI-powered language learning platform that uses scientific spaced repetition (FSRS) to help you memorize vocabulary efficiently by building neural synapses.'
        }
      },
      {
        '@type': 'Question',
        name: 'How does the scientific algorithm work?',
        acceptedAnswer: {
          '@type': 'Answer',
          text: 'Our Neural FSRS+ algorithm predicts memory decay and schedules reviews at the optimal time to strengthen synaptic pathways in your brain.'
        }
      }
    ]
  };

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(breadcrumbJsonLd) }}
      />
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(faqJsonLd) }}
      />
      <HelpPageClient />
    </>
  );
}

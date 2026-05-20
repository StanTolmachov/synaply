import { MetadataRoute } from 'next';
import { languages } from '@/lib/languages';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

async function getPublicLists() {
  try {
    const response = await fetch(`${API_BASE}/public-lists`, {
      next: { revalidate: 3600 } // Cache for 1 hour
    });
    if (!response.ok) return [];
    const data = await response.json();
    return Array.isArray(data) ? data : (data?.lists || []);
  } catch (error) {
    console.error('Error fetching public lists for sitemap:', error);
    return [];
  }
}

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const baseUrl = process.env.NEXT_PUBLIC_APP_URL || 'https://synaply.me';
  const locales = Object.keys(languages).map(lang => lang.toLowerCase());
  
  const staticRoutes = ['', '/terms', '/help', '/public-lists'];
  const publicLists = await getPublicLists();
  
  const sitemapItems: MetadataRoute.Sitemap = [];

  // Добавляем основные страницы для каждой локали
  locales.forEach((locale) => {
    staticRoutes.forEach((route) => {
      const languageAlternates: Record<string, string> = {};
      locales.forEach((l) => {
        languageAlternates[l] = `${baseUrl}/${l}${route}`;
      });
      // x-default points to the root/canonical URL (which redirects)
      const xDefault = `${baseUrl}${route}`;

      sitemapItems.push({
        url: `${baseUrl}/${locale}${route}`,
        lastModified: new Date(),
        changeFrequency: route === '' ? 'daily' : 'weekly',
        priority: route === '' ? 1.0 : 0.7,
        // @ts-ignore - Next.js supports alternates in sitemap since 14.2
        alternates: {
          languages: {
            ...languageAlternates,
            'x-default': xDefault
          }
        }
      });
    });

    // Добавляем динамические ссылки на публичные списки
    publicLists.forEach((list: any) => {
      const languageAlternates: Record<string, string> = {};
      locales.forEach((l) => {
        languageAlternates[l] = `${baseUrl}/${l}/public-lists/${list.id}`;
      });
      const xDefault = `${baseUrl}/public-lists/${list.id}`;

      sitemapItems.push({
        url: `${baseUrl}/${locale}/public-lists/${list.id}`,
        lastModified: new Date(list.created_at || new Date()),
        changeFrequency: 'monthly',
        priority: 0.6,
        // @ts-ignore
        alternates: {
          languages: {
            ...languageAlternates,
            'x-default': xDefault
          }
        }
      });
    });
  });

  return sitemapItems;
}

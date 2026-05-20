'use client';

import React, { createContext, useContext, ReactNode } from 'react';
import { useTranslations, useLocale } from 'next-intl';
import { useRouter, usePathname, routing } from '@/i18n/routing';

interface I18nContextProps {
  lang: string;
  setLang: (lang: string, saveToLocal?: boolean) => void;
  resetToSaved: () => void;
  t: (path: string, params?: Record<string, any>) => any;
  tHtml: (path: string, params?: Record<string, any>) => string;
}

const I18nContext = createContext<I18nContextProps | undefined>(undefined);

export function I18nProvider({ children }: { children: ReactNode }) {
  const locale = useLocale();
  const tNext = useTranslations();
  const router = useRouter();
  const pathname = usePathname();

  const resetToSaved = () => {
    if (typeof window === 'undefined') return;
    const savedLang = localStorage.getItem('source_lang');
    if (savedLang) {
      const lowerSaved = savedLang.toLowerCase();
      // Only redirect if different from current locale AND it's a valid locale
      if (lowerSaved !== locale.toLowerCase()) {
        const isValidLocale = routing.locales.includes(lowerSaved as any);
        if (isValidLocale) {
          // Check where we are actually going
          const currentUrlPath = window.location.pathname;
          
          // If we are already on a specific page (e.g. /dashboard),
          // pathname will be /dashboard.
          // If we are still on root (unmounting landing), pathname is /.
          const targetPath = `/${lowerSaved}${pathname}`;
          
          if (currentUrlPath !== targetPath && currentUrlPath !== `${targetPath}/`) {
            window.location.replace(targetPath);
          }
        }
      }
    }
  };

  const syncLocaleWithSaved = () => {
    if (typeof window === 'undefined') return;
    const savedLang = localStorage.getItem('source_lang');
    if (savedLang) {
      const lowerSaved = savedLang.toLowerCase();
      // Don't sync on landing page (handled by LandingPage component)
      if (pathname === '/') {
        return;
      }
      
      if (lowerSaved !== locale.toLowerCase()) {
        const isValidLocale = routing.locales.includes(lowerSaved as any);
        if (isValidLocale) {
          // IMPORTANT: Check if we are already on the correct locale in the URL
          // pathname from usePathname() doesn't include locale.
          // window.location.pathname DOES include locale.
          const currentUrlPath = window.location.pathname;
          const targetPath = `/${lowerSaved}${pathname}`;
          
          if (currentUrlPath !== targetPath && currentUrlPath !== `${targetPath}/`) {
             window.location.replace(targetPath);
          }
        }
      }
    }
  };

  // Run sync on mount and when pathname/locale changes
  React.useEffect(() => {
    syncLocaleWithSaved();
  }, [pathname, locale]);

  const setLang = (newLang: string, saveToLocal: boolean = true) => {
    if (typeof window === 'undefined') return;
    const lowerLang = newLang.toLowerCase();
    
    if (saveToLocal) {
      localStorage.setItem('source_lang', lowerLang);
    }

    // Only redirect if different and valid
    if (lowerLang !== locale.toLowerCase()) {
      const isValidLocale = routing.locales.includes(lowerLang as any);
      if (isValidLocale) {
        // Use window.location.replace to avoid triggering router.replace in a loop
        // and to fully reload the page with the new locale if necessary.
        const currentUrlPath = window.location.pathname;
        const targetPath = pathname === '/' ? `/${lowerLang}` : `/${lowerLang}${pathname}`;
        
        if (currentUrlPath !== targetPath && currentUrlPath !== `${targetPath}/`) {
          window.location.replace(targetPath);
        }
      }
    }
  };

  // Адаптер для старого t() метода, чтобы не переписывать все компоненты сразу
  const t = (path: string, params?: Record<string, any>): any => {
    try {
      const htmlHandlers: Record<string, (chunks: ReactNode) => any> = {
        strong: (chunks) => <strong>{chunks}</strong>,
        b: (chunks) => <b>{chunks}</b>,
        i: (chunks) => <i>{chunks}</i>,
        em: (chunks) => <em>{chunks}</em>,
        u: (chunks) => <u>{chunks}</u>,
        br: () => <br/>
      };

      // next-intl useTranslations() t function handles nested keys with dots by default
      // but if it fails, we want to return the path as a fallback
      // result is a React element if rich text is used, or a string otherwise
      const result = tNext.rich(path as any, { 
        ...htmlHandlers,
        ...params
      });
      
      // If the result is the same as the path, it might be because next-intl 
      // returns the key if it's missing (depending on configuration)
      return result;
    } catch (e) {
      // If rich fails, try simple t
      try {
        return tNext(path as any, params);
      } catch (e2) {
        // Если ключ не найден, возвращаем сам путь
        return path;
      }
    }
  };

  // Специальная версия t для dangerouslySetInnerHTML
  const tHtml = (path: string, params?: Record<string, any>): string => {
    try {
      const stringHandlers: Record<string, (chunks: ReactNode) => string> = {
        strong: (chunks) => `<strong>${chunks}</strong>`,
        b: (chunks) => `<b>${chunks}</b>`,
        i: (chunks) => `<i>${chunks}</i>`,
        em: (chunks) => `<em>${chunks}</em>`,
        u: (chunks) => `<u>${chunks}</u>`,
        br: () => `<br/>`
      };

      // next-intl's rich text expects <tag></tag> or <tag/>.
      // If the source JSON has <br/>, we need to make sure next-intl
      // processes it as a rich text tag.
      
      const result = (tNext.rich as any)(path, { 
        ...stringHandlers,
        ...params
      });
      
      return typeof result === 'string' ? result : (Array.isArray(result) ? result.join('') : String(result));
    } catch (e) {
      try {
        const result = (tNext as any)(path, params);
        return typeof result === 'string' ? result : path;
      } catch (e2) {
        return path;
      }
    }
  };

  return (
    <I18nContext.Provider value={{ lang: locale, setLang, resetToSaved, t, tHtml }}>
      {children}
    </I18nContext.Provider>
  );
}

export function useTranslation() {
  const context = useContext(I18nContext);
  if (context === undefined) {
    throw new Error('useTranslation must be used within an I18nProvider');
  }
  return context;
}

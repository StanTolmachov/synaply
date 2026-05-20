"use client";

import React from 'react';
import { usePathname, Link } from '@/i18n/routing';
import { useTranslation } from './I18nContext';

const Footer = () => {
  const { t } = useTranslation();
  const pathname = usePathname();
  const isLandingPage = pathname === '/' || pathname === '';

  const year = typeof window === 'undefined' ? 2026 : (new Date().getFullYear());

  return (
    <footer className={`border-t border-gray-200 dark:border-gray-800 p-8 mt-8 ${isLandingPage ? 'bg-white dark:bg-gray-950' : 'bg-transparent'}`}>
      <div className="max-w-7xl mx-auto flex flex-col items-center">
        <div className="mb-4 text-center text-sm text-gray-600 dark:text-gray-400">
          <p>© {year} Synaply. {t('common.all_rights_reserved')}</p>
        </div>
        <nav className="mb-4 flex justify-center gap-6 sm:gap-8 flex-wrap" aria-label="Footer Navigation">
          <Link href="/terms" className="text-sm text-gray-600 dark:text-gray-400 hover:text-synaply-blue dark:hover:text-synaply-cyan no-underline transition-colors">{t('common.terms')}</Link>
          <Link href="/help" className="text-sm text-gray-600 dark:text-gray-400 hover:text-synaply-blue dark:hover:text-synaply-cyan no-underline transition-colors">{t('common.help')}</Link>
          <Link href="/public-lists" className="text-sm text-gray-600 dark:text-gray-400 hover:text-synaply-blue dark:hover:text-synaply-cyan no-underline transition-colors">{t('dashboard.public_lists.title')}</Link>
          <a href="/sitemap.xml" className="text-sm text-gray-600 dark:text-gray-400 hover:text-synaply-blue dark:hover:text-synaply-cyan no-underline transition-colors">Sitemap</a>
        </nav>
        <div className="mb-4 text-center text-sm text-gray-600 dark:text-gray-400">
          <p key="help-text">
            {t('common.footer_help_text')}
            {' '}
            <a href="mailto:support@tolmachov.dev" className="text-synaply-blue dark:text-synaply-cyan hover:underline">support@tolmachov.dev</a>
          </p>
        </div>
      </div>
    </footer>
  );
};

export default Footer;

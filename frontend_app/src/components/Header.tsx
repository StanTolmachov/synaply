"use client";

import React, { useState, useEffect } from 'react';
import Image from 'next/image';
import { usePathname, useRouter, Link } from '@/i18n/routing';
import { useScore } from './ScoreContext';
import { sendGAEvent } from '@next/third-parties/google';
import { LogOut, HelpCircle, User, WifiOff, Globe, Layers, LayoutDashboard, Menu, X, Star, BookOpen, Brain, List, ShieldCheck } from 'lucide-react';
import { ThemeToggle } from './ThemeToggle';
import { useTranslation } from './I18nContext';
import { fetchApi } from '@/lib/api';

export function Header() {
  const { t } = useTranslation();
  const { score } = useScore();
  const pathname = usePathname();
  const router = useRouter();
  const [isOffline, setIsOffline] = useState(false);
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isProfileOpen, setIsProfileOpen] = useState(false);
  const [isAdmin, setIsAdmin] = useState(false);

  useEffect(() => {
    const checkAdmin = async () => {
      const token = localStorage.getItem('token');
      if (!token) return;
      try {
        const userData = await fetchApi('/words/GetMe');
        if (userData && userData.role === 'admin') {
          setIsAdmin(true);
        }
      } catch (err) {
        // Silently fail
      }
    };
    checkAdmin();
  }, [pathname]);

  useEffect(() => {
    setIsOffline(!navigator.onLine);
    const handleOnline = () => setIsOffline(false);
    const handleOffline = () => setIsOffline(true);

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, []);

  // Close menu when route changes
  useEffect(() => {
    setIsMenuOpen(false);
    setIsProfileOpen(false);
  }, [pathname]);

  // Handle outside click for dropdown
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as HTMLElement;
      if (isProfileOpen && !target.closest('.profile-dropdown')) {
        setIsProfileOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, [isProfileOpen]);

  const isLandingPage = pathname === '/' || pathname === '';
  const isAuthPage = pathname.endsWith('/login') || pathname.endsWith('/register');
  const isHelpPage = pathname.endsWith('/help');

  if (isAuthPage || isLandingPage) {
    return null;
  }

  const handleLogout = () => {
    sendGAEvent('event', 'logout', {});
    localStorage.removeItem('token');
    router.push('/login');
  };

  const NavItems = () => (
    <>
      <Link
        href="/dashboard"
        className="text-gray-600 dark:text-gray-300 hover:text-synaply-blue dark:hover:text-synaply-cyan flex items-center text-sm font-medium transition-colors py-2 md:py-0"
      >
        <LayoutDashboard className="w-4 h-4 mr-2 md:mr-1 text-synaply-blue dark:text-synaply-cyan" />
        <span className="md:hidden lg:inline">{t('common.dashboard')}</span>
      </Link>
      <Link
        href="/lesson"
        className="text-gray-600 dark:text-gray-300 hover:text-synaply-blue dark:hover:text-synaply-cyan flex items-center text-sm font-medium transition-colors py-2 md:py-0"
      >
        <BookOpen className="w-4 h-4 mr-2 md:mr-1 text-green-500" />
        <span className="md:hidden lg:inline">{t('common.review')}</span>
      </Link>
      <Link
        href="/practice"
        className="text-gray-600 dark:text-gray-300 hover:text-synaply-blue dark:hover:text-synaply-cyan flex items-center text-sm font-medium transition-colors py-2 md:py-0"
      >
        <Brain className="w-4 h-4 mr-2 md:mr-1 text-purple-500" />
        <span className="md:hidden lg:inline">{t('common.practice')}</span>
      </Link>
      <Link
        href="/words"
        className="text-gray-600 dark:text-gray-300 hover:text-synaply-blue dark:hover:text-synaply-cyan flex items-center text-sm font-medium transition-colors py-2 md:py-0"
      >
        <List className="w-4 h-4 mr-2 md:mr-1 text-amber-500" />
        <span className="md:hidden lg:inline">{t('common.words')}</span>
      </Link>
    </>
  );

  const ProfileItems = () => (
    <>
      <Link
        href="/profile"
        className="text-gray-600 dark:text-gray-300 hover:text-synaply-blue dark:hover:text-synaply-cyan flex items-center text-sm font-medium transition-colors py-2 px-3 hover:bg-synaply-blue/5 dark:hover:bg-synaply-cyan/10 rounded-md md:flex"
      >
        <User className="w-4 h-4 mr-2 md:mr-1 text-gray-400" />
        {t('common.profile')}
      </Link>
      {isAdmin && (
        <Link
          href="/admin/stats"
          className="text-gray-600 dark:text-gray-300 hover:text-synaply-blue dark:hover:text-synaply-cyan flex items-center text-sm font-medium transition-colors py-2 px-3 hover:bg-synaply-blue/5 dark:hover:bg-synaply-cyan/10 rounded-md md:flex"
        >
          <ShieldCheck className="w-4 h-4 mr-2 text-synaply-blue dark:text-synaply-cyan" />
          {t('common.admin_stats')}
        </Link>
      )}
      <Link
        href="/public-lists"
        className="text-gray-600 dark:text-gray-300 hover:text-synaply-blue dark:hover:text-synaply-cyan flex items-center text-sm font-medium transition-colors py-2 px-3 hover:bg-synaply-blue/5 dark:hover:bg-synaply-cyan/10 rounded-md md:flex"
      >
        <List className="w-4 h-4 mr-2 md:mr-1 text-orange-500" />
        {t('dashboard.public_lists.title')}
      </Link>
      <Link
        href="/help"
        className="text-gray-600 dark:text-gray-300 hover:text-synaply-blue dark:hover:text-synaply-cyan flex items-center text-sm font-medium transition-colors py-2 px-3 hover:bg-synaply-blue/5 dark:hover:bg-synaply-cyan/10 rounded-md md:flex"
      >
        <HelpCircle className="w-4 h-4 mr-2 md:mr-1 text-blue-500" />
        {t('common.help')}
      </Link>
      <div className="h-px bg-gray-100 dark:bg-gray-800 my-1 hidden md:block" />
      <button
        onClick={handleLogout}
        className="text-gray-600 dark:text-gray-300 hover:text-red-600 dark:hover:text-red-400 flex items-center text-sm font-medium transition-colors py-2 px-3 hover:bg-red-50 dark:hover:bg-red-900/10 rounded-md text-left w-full md:flex"
      >
        <LogOut className="w-4 h-4 mr-2 md:mr-1 text-gray-400" />
        {t('common.sign_out')}
      </button>
    </>
  );

  return (
    <header className="bg-white dark:bg-gray-900 shadow-sm border-b border-gray-200 dark:border-gray-800 sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16 items-center">
          {/* Logo Section */}
          <div className="flex-shrink-0 flex items-center">
            <Link href="/" className="flex items-center group">
              <Image
                src="/logo-Header.png"
                alt="Synaply Logo"
                width={32}
                height={40}
                className="w-auto h-8 mr-2"
              />
              <span className="text-xl font-bold tracking-tight text-synaply-blue dark:text-blue-400">
                synaply<span className="text-synaply-cyan dark:text-blue-300">.me</span>
              </span>
              <span className="ml-2 px-1.5 py-0.5 text-[10px] font-bold uppercase tracking-wider text-synaply-blue bg-synaply-blue/5 dark:bg-blue-900/30 border border-synaply-blue/10 dark:border-blue-800 rounded">{t('common.beta')}</span>
            </Link>
          </div>

          {/* Desktop Navigation */}
          <div className="hidden md:flex flex-1 items-center justify-end space-x-4 lg:space-x-8">
            {isOffline && (
              <div className="flex items-center text-amber-600 dark:text-amber-500 bg-amber-50 dark:bg-amber-900/20 px-2 py-1 rounded text-xs font-medium animate-pulse border border-amber-100 dark:border-amber-900/50">
                <WifiOff className="w-3.5 h-3.5 mr-1.5" />
                {t('common.offline_mode')}
              </div>
            )}
            
            <div className="hidden md:flex items-center bg-blue-50 dark:bg-blue-900/30 px-3 py-1 rounded-full border border-blue-100 dark:border-blue-800 shadow-sm">
              <Star className="w-3.5 h-3.5 mr-1.5 text-blue-500 fill-blue-500" />
              <span className="text-sm font-medium text-gray-700 dark:text-gray-200">
                {t('common.score')}: <span className="font-bold text-blue-700 dark:text-blue-400">{score}</span>
              </span>
            </div>

            <nav className="hidden md:flex items-center space-x-4 lg:space-x-6">
              <NavItems />
            </nav>
            
            <div className="hidden md:flex items-center border-l border-gray-200 dark:border-gray-800 pl-4 h-6 space-x-4">
              <ThemeToggle />
              
              <div className="relative profile-dropdown">
                <button
                  onClick={() => setIsProfileOpen(!isProfileOpen)}
                  className="flex items-center space-x-2 p-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                  aria-label="User menu"
                >
                  <div className="w-8 h-8 rounded-full bg-blue-100 dark:bg-blue-900/50 flex items-center justify-center border border-blue-200 dark:border-blue-800">
                    <User className="w-5 h-5 text-blue-600 dark:text-blue-400" />
                  </div>
                </button>

                {isProfileOpen && (
                  <div className="absolute right-0 mt-2 w-48 bg-white dark:bg-gray-900 rounded-lg shadow-xl border border-gray-200 dark:border-gray-800 py-2 z-50 animate-in fade-in zoom-in-95 duration-100">
                    <div className="px-4 py-2 border-b border-gray-100 dark:border-gray-800 mb-1">
                      <p className="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">{t('common.profile')}</p>
                    </div>
                    <div className="px-2">
                      <ProfileItems />
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Mobile Actions (Score & Menu Toggle) */}
          <div className="flex md:hidden items-center space-x-3">
            {isOffline && (
              <WifiOff className="w-5 h-5 text-amber-600 dark:text-amber-500 animate-pulse" />
            )}
            
            <div className="flex items-center bg-blue-50 dark:bg-blue-900/30 px-2.5 py-1 rounded-full border border-blue-100 dark:border-blue-800">
              <Star className="w-3 h-3 mr-1 text-blue-500 fill-blue-500" />
              <span className="text-xs font-bold text-blue-700 dark:text-blue-400">{score}</span>
            </div>

            <ThemeToggle />
            
            <button
              onClick={() => setIsMenuOpen(!isMenuOpen)}
              className="p-2 rounded-md text-gray-500 hover:text-gray-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
              aria-label="Toggle menu"
            >
              {isMenuOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
            </button>
          </div>
        </div>
      </div>

      {/* Mobile Menu */}
      {isMenuOpen && (
        <div className="md:hidden bg-white dark:bg-gray-900 border-t border-gray-100 dark:border-gray-800 shadow-lg animate-in slide-in-from-top-4 duration-200 overflow-y-auto max-h-[calc(100vh-4rem)]">
          <nav className="px-4 pt-2 pb-6 space-y-1">
            <div className="space-y-1">
              <NavItems />
            </div>
            <div className="h-px bg-gray-100 dark:bg-gray-800 my-2" />
            <ProfileItems />
          </nav>
        </div>
      )}
    </header>
  );
}
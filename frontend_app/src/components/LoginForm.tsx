'use client';

import { useState } from 'react';
import Image from 'next/image';
import { useRouter, Link } from '@/i18n/routing';
import { fetchApi } from '@/lib/api';
import { sendGAEvent } from '@next/third-parties/google';
import { useTranslation } from '@/components/I18nContext';

export default function LoginForm() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const router = useRouter();
  const { t, setLang } = useTranslation();

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    try {
      sendGAEvent('event', 'login', { method: 'email' });
      const data = await fetchApi('/users/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      });
      localStorage.setItem('token', data.token);
      if (data.source_lang) {
        setLang(data.source_lang);
      }
      router.push('/dashboard');
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Oops! We couldn't sign you in. Please check your credentials.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-mesh px-4 sm:px-6 lg:px-8 py-12 relative overflow-hidden">
      <div className="absolute inset-0 mesh-grid opacity-30 pointer-events-none"></div>

      <div className="w-full max-w-md space-y-8 bg-mesh backdrop-blur-md p-8 rounded-2xl shadow-2xl border border-white/20 dark:border-gray-800/50 relative z-10 overflow-hidden">
        <div className="absolute inset-0 mesh-grid opacity-20 pointer-events-none"></div>
        <div className="relative z-10">
          <div className="flex justify-center mb-6">
            <Link href="/" className="flex items-center gap-3 group transition-transform hover:scale-105">
              <Image
                src="/logo-Header.png"
                alt="Synaply Logo"
                width={38}
                height={48}
                className="w-auto h-12"
              />
              <span className="text-2xl font-bold tracking-tight text-synaply-blue dark:text-blue-400">
                synaply<span className="text-synaply-cyan dark:text-blue-300">.me</span>
              </span>
              <span className="px-1.5 py-0.5 text-[10px] font-bold uppercase tracking-wider text-synaply-blue bg-synaply-blue/5 dark:bg-blue-900/30 border border-synaply-blue/10 dark:border-blue-800 rounded">{t('common.beta')}</span>
            </Link>
          </div>
          <h2 className="text-center text-3xl font-extrabold tracking-tight synaply-gradient-text pb-1">
            {t('login.title')}
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600 dark:text-gray-400">
            {t('login.no_account')}{' '}
            <Link href="/register" className="font-semibold text-synaply-blue dark:text-synaply-cyan hover:opacity-80 transition-opacity">
              {t('login.register')}
            </Link>
          </p>
        </div>
        <form className="mt-8 space-y-5" onSubmit={handleLogin}>
          <div className="space-y-4">
            <div>
              <label className="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-1">{t('login.email')}</label>
              <input
                type="email"
                required
                className="block w-full rounded-xl border border-gray-200 dark:border-gray-800 px-4 py-2.5 text-gray-900 dark:text-gray-100 dark:bg-gray-800/20 focus:border-synaply-blue dark:focus:border-synaply-cyan focus:outline-none focus:ring-2 focus:ring-synaply-blue/20 dark:focus:ring-synaply-cyan/20 transition-all sm:text-sm"
                placeholder="you@example.com"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </div>
            <div>
              <label className="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-1">{t('login.password')}</label>
              <input
                type="password"
                required
                className="block w-full rounded-xl border border-gray-200 dark:border-gray-800 px-4 py-2.5 text-gray-900 dark:text-gray-100 dark:bg-gray-800/20 focus:border-synaply-blue dark:focus:border-synaply-cyan focus:outline-none focus:ring-2 focus:ring-synaply-blue/20 dark:focus:ring-synaply-cyan/20 transition-all sm:text-sm"
                placeholder="••••••••"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
          </div>

          {error && (
            <div className="text-xs text-red-600 dark:text-red-400 font-semibold bg-red-50 dark:bg-red-900/20 p-3 rounded-lg border border-red-100 dark:border-red-900/30">
              {error}
            </div>
          )}

          <div className="pt-2">
            <button
              type="submit"
              disabled={loading}
              className="group relative flex w-full justify-center rounded-xl synaply-gradient-bg py-3 px-4 text-sm font-bold text-white shadow-lg shadow-synaply-blue-shadow hover:shadow-synaply-purple-shadow focus:outline-none focus:ring-2 focus:ring-synaply-purple focus:ring-offset-2 disabled:opacity-50 transition-all active:scale-[0.98]"
            >
              {loading ? (
                <span className="flex items-center space-x-2">
                  <svg className="animate-spin h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  <span>{t('login.signing_in')}</span>
                </span>
              ) : (
                t('login.sign_in')
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

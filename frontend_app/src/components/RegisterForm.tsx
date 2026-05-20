'use client';

import { useState, useEffect } from 'react';
import Image from 'next/image';
import { useRouter, Link } from '@/i18n/routing';
import { fetchApi } from '@/lib/api';
import { sendGAEvent } from '@next/third-parties/google';
import { useTranslation } from '@/components/I18nContext';

export default function RegisterForm() {
  const [formData, setFormData] = useState({
    email: '',
    password: '',
    first_name: '',
    last_name: '',
    source_lang: '',
    target_lang: '',
  });
  const [languages, setLanguages] = useState<{ source: { code: string; name: string }[], target: { code: string; name: string }[] }>({ source: [], target: [] });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [acceptedTerms, setAcceptedTerms] = useState(false);
  const router = useRouter();
  const { t, setLang } = useTranslation();

  useEffect(() => {
    fetchApi('/users/lang')
      .then((data) => {
        setLanguages(data);
        if (data.source?.length > 0 && data.target?.length > 0) {
          setFormData((prev) => ({
            ...prev,
            source_lang: data.source[0].code,
            target_lang: data.target[0].code,
          }));
        }
      })
      .catch((err) => console.error('Failed to load languages', err));
  }, []);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    setFormData({ ...formData, [e.target.name]: e.target.value });
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');

    if (!acceptedTerms) {
      setError('You must agree to the Terms of Service to register');
      setLoading(false);
      return;
    }

    if (formData.source_lang === formData.target_lang) {
      setError('Learning language cannot be the same as your native language');
      setLoading(false);
      return;
    }

    try {
      sendGAEvent('event', 'sign_up', { method: 'email', source_lang: formData.source_lang, target_lang: formData.target_lang });
      await fetchApi('/users/create', {
        method: 'POST',
        body: JSON.stringify(formData),
      });
      // Optionally login immediately or redirect
      const loginData = await fetchApi('/users/login', {
        method: 'POST',
        body: JSON.stringify({ email: formData.email, password: formData.password }),
      });
      localStorage.setItem('token', loginData.token);
      if (loginData.source_lang) {
        setLang(loginData.source_lang);
      }
      router.push('/dashboard');
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "We couldn't create your account. Please try again in a moment!");
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
            {t('register.title')}
          </h2>
          <p className="mt-2 text-center text-sm text-gray-600 dark:text-gray-400">
            {t('register.has_account')}{' '}
            <Link href="/login" className="font-semibold text-synaply-blue dark:text-synaply-cyan hover:opacity-80 transition-opacity">
              {t('register.login')}
            </Link>
          </p>
        </div>
        <form className="mt-8 space-y-5" onSubmit={handleRegister}>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-1">{t('register.first_name')}</label>
              <input
                type="text"
                name="first_name"
                required
                className="block w-full rounded-xl border border-gray-200 dark:border-gray-800 px-4 py-2.5 text-gray-900 dark:text-gray-100 dark:bg-gray-800/20 focus:border-synaply-blue dark:focus:border-synaply-cyan focus:outline-none focus:ring-2 focus:ring-synaply-blue/20 dark:focus:ring-synaply-cyan/20 transition-all sm:text-sm"
                value={formData.first_name}
                onChange={handleChange}
              />
            </div>
            <div>
              <label className="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-1">{t('register.last_name')}</label>
              <input
                type="text"
                name="last_name"
                required
                className="block w-full rounded-xl border border-gray-200 dark:border-gray-800 px-4 py-2.5 text-gray-900 dark:text-gray-100 dark:bg-gray-800/20 focus:border-synaply-blue dark:focus:border-synaply-cyan focus:outline-none focus:ring-2 focus:ring-synaply-blue/20 dark:focus:ring-synaply-cyan/20 transition-all sm:text-sm"
                value={formData.last_name}
                onChange={handleChange}
              />
            </div>
          </div>
          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-1">{t('register.email')}</label>
            <input
              type="email"
              name="email"
              required
              className="block w-full rounded-xl border border-gray-200 dark:border-gray-800 px-4 py-2.5 text-gray-900 dark:text-gray-100 dark:bg-gray-800/20 focus:border-synaply-blue dark:focus:border-synaply-cyan focus:outline-none focus:ring-2 focus:ring-synaply-blue/20 dark:focus:ring-synaply-cyan/20 transition-all sm:text-sm"
              value={formData.email}
              onChange={handleChange}
            />
          </div>
          <div>
            <label className="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-1">{t('register.password')}</label>
            <input
              type="password"
              name="password"
              required
              minLength={8}
              className="block w-full rounded-xl border border-gray-200 dark:border-gray-800 px-4 py-2.5 text-gray-900 dark:text-gray-100 dark:bg-gray-800/20 focus:border-synaply-blue dark:focus:border-synaply-cyan focus:outline-none focus:ring-2 focus:ring-synaply-blue/20 dark:focus:ring-synaply-cyan/20 transition-all sm:text-sm"
              value={formData.password}
              onChange={handleChange}
            />
            <p className="text-[10px] text-gray-500 dark:text-gray-500 mt-1.5 px-1 uppercase tracking-tight">{t('register.password_hint')}</p>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-1">{t('profile.interface_lang')}</label>
              <select
                name="source_lang"
                className="block w-full rounded-xl border border-gray-200 dark:border-gray-800 px-4 py-2.5 text-gray-900 dark:text-gray-100 dark:bg-gray-800/20 focus:border-synaply-blue dark:focus:border-synaply-cyan focus:outline-none focus:ring-2 focus:ring-synaply-blue/20 dark:focus:ring-synaply-cyan/20 transition-all sm:text-sm"
                value={formData.source_lang}
                onChange={handleChange}
              >
                {languages.source?.map((lang) => (
                  <option key={lang.code} value={lang.code}>{lang.name}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400 mb-1">{t('profile.learning_lang')}</label>
              <select
                name="target_lang"
                className="block w-full rounded-xl border border-gray-200 dark:border-gray-800 px-4 py-2.5 text-gray-900 dark:text-gray-100 dark:bg-gray-800/20 focus:border-synaply-blue dark:focus:border-synaply-cyan focus:outline-none focus:ring-2 focus:ring-synaply-blue/20 dark:focus:ring-synaply-cyan/20 transition-all sm:text-sm"
                value={formData.target_lang}
                onChange={handleChange}
              >
                {languages.target?.map((lang) => (
                  <option key={lang.code} value={lang.code}>{lang.name}</option>
                ))}
              </select>
            </div>
          </div>

          {error && (
            <div className="text-xs text-red-600 dark:text-red-400 font-semibold bg-red-50 dark:bg-red-900/20 p-3 rounded-lg border border-red-100 dark:border-red-900/30">
              {error}
            </div>
          )}

          <div className="flex items-start space-x-3 pt-1">
            <div className="flex items-center h-5">
              <input
                type="checkbox"
                id="terms"
                name="terms"
                className="h-4 w-4 rounded border-gray-300 dark:border-gray-700 text-synaply-blue focus:ring-synaply-blue dark:bg-gray-800/20 transition-all cursor-pointer"
                checked={acceptedTerms}
                onChange={(e) => setAcceptedTerms(e.target.checked)}
                required
              />
            </div>
            <label htmlFor="terms" className="text-xs text-gray-600 dark:text-gray-400 leading-normal cursor-pointer">
              {t('register.terms_acceptance_prefix')}
              <Link href="/terms" className="font-semibold text-synaply-blue dark:text-synaply-cyan hover:underline mx-1">
                {t('register.terms_acceptance_link')}
              </Link>
              {t('register.terms_acceptance_suffix')}
            </label>
          </div>

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
                  <span>{t('register.signing_up')}</span>
                </span>
              ) : (
                t('register.sign_up')
              )}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}

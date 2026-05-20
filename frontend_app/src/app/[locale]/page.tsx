'use client';

import { useEffect, useState } from 'react';
import Image from 'next/image';
import { useRouter, Link } from '@/i18n/routing';
import { Brain, Zap, Infinity, Bot, ArrowRight, CheckCircle, Sparkles, MessageSquare, Languages, Database, BarChart3, Target, Globe, ArrowLeft, Star } from 'lucide-react';
import { useTranslation } from '@/components/I18nContext';

export default function LandingPage() {
  const router = useRouter();
  const { t, tHtml, setLang, resetToSaved } = useTranslation();
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [mounted, setMounted] = useState(false);

  const flags = [
    '🇺🇸', '🇬🇧', '🇩🇪', '🇫🇷', '🇪🇸', '🇮🇹', '🇵🇹', '🇳🇱', '🇵🇱', '🇷🇺', '🇯🇵', '🇰🇷', '🇨🇳', '🇹🇷', '🇮🇩', '🇺🇦', '🇬🇷', '🇨🇿',
    '🇸🇪', '🇳🇴', '🇩🇰', '🇫🇮', '🇭🇺', '🇷🇴', '🇧🇬', '🇸🇰', '🇭🇷', '🇷🇸', '🇮🇪', '🇦🇹', '🇨🇭', '🇧🇪', '🇲🇽', '🇨🇦', '🇦🇺', '🇳🇿',
    '🇧🇷', '🇦🇷', '🇨🇱', '🇨🇴', '🇵🇪', '🇿🇦', '🇪🇬', '🇸🇦', '🇦🇪', '🇮🇱', '🇮🇳', '🇵🇰', '🇧🇩', '🇻🇳', '🇹🇭', '🇲🇾', '🇵🇭', '🇸🇬',
    '🇭🇰', '🇹🇼', '🇰🇿', '🇺🇿', '🇦🇿', '🇬🇪', '🇦🇲', '🇱🇹', '🇱🇻', '🇪🇪', '🇮🇸', '🇱🇺', '🇲🇨', '🇨🇾', '🇲🇹', '🇻🇦', '🇦🇱', '🇲🇪',
    '🇽🇰', '🇲🇰', '🇧🇦', '🇲🇩', '🇰🇬', '🇹🇯', '🇹🇲', '🇲🇳', '🇳🇵', '🇱🇰', '🇲🇲', '🇰🇭', '🇱🇦', '🇰🇵', '🇲🇦', '🇩🇿', '🇹🇳', '🇱🇧',
    '🇯🇴', '🇮🇶', '🇰🇼', '🇶🇦', '🇴🇲', '🇾🇪', '🇳🇬', '🇬🇭', '🇰🇪', '🇪🇹', '🇹🇿', '🇺🇬', '🇷🇼', '🇨🇲', '🇨🇮', '🇸🇳', '🇦🇴', '🇲🇿',
    '🇨🇺', '🇯🇲', '🇵🇷', '🇩🇴', '🇨🇷', '🇵🇦'
  ]; // Approximately 114 flags (18 + 96 = 114)

  useEffect(() => {
    setMounted(true);
    // Force English for landing page
    if (setLang) {
      setLang('en', false);
    }
    // Check login state
    const token = typeof window !== 'undefined' ? localStorage.getItem('token') : null;
    if (token) {
      setIsLoggedIn(true);
    }
    
    return () => {
      // Restore saved language when leaving landing page
      if (typeof window !== 'undefined' && resetToSaved) {
        resetToSaved();
      }
    };
  }, [resetToSaved, setLang]); // Only on mount/unmount

  const jsonLd = {
    '@context': 'https://schema.org',
    '@type': 'WebApplication',
    name: 'Synaply',
    alternateName: 'Synaply AI',
    description: t('metadata.description'),
    applicationCategory: 'EducationalApplication',
    operatingSystem: 'Web',
    url: 'https://synaply.me',
    offers: {
      '@type': 'Offer',
      price: '0',
      priceCurrency: 'USD',
    },
    author: {
      '@type': 'Organization',
      name: 'Synaply Team',
      logo: 'https://synaply.me/apple-icon.png'
    },
    aggregateRating: {
      '@type': 'AggregateRating',
      ratingValue: '4.9',
      ratingCount: '1540',
    },
    featureList: [
      t('help.spaced_repetition.title'),
      t('landing.tag_ai_powered'),
      t('common.public_lists'),
      t('dashboard.progress')
    ]
  };

  if (!mounted) {
    return (
      <div className="min-h-screen bg-white dark:bg-gray-950 flex items-center justify-center">
        <div className="animate-pulse flex flex-col items-center">
          <div className="w-16 h-16 bg-blue-600 rounded-xl mb-4"></div>
          <div className="h-4 w-32 bg-gray-200 dark:bg-gray-800 rounded"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-white dark:bg-gray-950 text-gray-900 dark:text-gray-100 font-sans selection:bg-blue-100 dark:selection:bg-blue-900 selection:text-blue-900 dark:selection:text-blue-100">
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(jsonLd) }}
      />
      <style jsx global>{`
        @keyframes wave {
          0%, 100% { transform: translateY(0) rotate(0deg); }
          25% { transform: translateY(-3px) rotate(-3deg); }
          75% { transform: translateY(3px) rotate(3deg); }
        }
        @keyframes scroll {
          0% { transform: translateX(0); }
          100% { transform: translateX(-50%); }
        }
        .animate-wave {
          display: inline-block;
          animation: wave 3s ease-in-out infinite;
        }
        .animate-scroll {
          display: flex;
          width: max-content;
          animation: scroll 40s linear infinite;
        }
        .animate-scroll:hover {
          animation-play-state: paused;
        }
        .mask-fade {
          mask-image: linear-gradient(to right, transparent, black 10%, black 90%, transparent);
          -webkit-mask-image: linear-gradient(to right, transparent, black 10%, black 90%, transparent);
        }
        .prose-article {
          max-width: 90ch;
          margin-left: auto;
          margin-right: auto;
          line-height: 1.85;
          color: #1f2937;
          font-size: 1.15rem;
          font-weight: 400;
        }
        :global(.dark) .prose-article { color: #d1d5db; }
        .prose-article h2 {
          color: #111827;
          font-weight: 800;
          font-size: 3.5rem;
          margin-top: 6rem;
          margin-bottom: 3rem;
          letter-spacing: -0.04em;
          line-height: 1.05;
          text-align: center;
          background: linear-gradient(to bottom right, #111827, #2563eb);
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
        }
        :global(.dark) .prose-article h2 {
          background: linear-gradient(to bottom right, #f3f4f6, #60a5fa);
          -webkit-background-clip: text;
        }
        .prose-article h3 {
          color: #111827;
          font-weight: 700;
          font-size: 2.25rem;
          margin-top: 5rem;
          margin-bottom: 2rem;
          display: flex;
          align-items: center;
          gap: 1rem;
          letter-spacing: -0.02em;
        }
        :global(.dark) .prose-article h3 { color: #f3f4f6; }
        .prose-article h3::before {
          content: "";
          display: inline-block;
          width: 0.6rem;
          height: 2.25rem;
          background: linear-gradient(to bottom, #2563eb, #60a5fa);
          border-radius: 9999px;
        }
        .prose-article h4 {
          color: #1f2937;
          font-weight: 700;
          font-size: 1.5rem;
          margin-top: 3rem;
          margin-bottom: 1rem;
        }
        :global(.dark) .prose-article h4 { color: #e5e7eb; }
        .prose-article p {
          margin-top: 1.75rem;
          margin-bottom: 1.75rem;
        }
        .prose-article ul, .prose-article ol {
          margin-top: 2rem;
          margin-bottom: 2rem;
          padding-left: 2.5rem;
        }
        .prose-article li {
          margin-top: 1rem;
          margin-bottom: 1rem;
          position: relative;
        }
        .prose-article blockquote {
          border-left: 4px solid #2563eb;
          padding-left: 2rem;
          font-style: italic;
          color: #4b5563;
          margin: 3rem 0;
          font-size: 1.25rem;
        }
        :global(.dark) .prose-article blockquote { color: #9ca3af; }
        .article-card {
          background: linear-gradient(145deg, #ffffff, #f8fafc);
          border-radius: 2rem;
          padding: 3.5rem;
          margin: 4.5rem 0;
          border: 1px solid #e2e8f0;
          position: relative;
          overflow: hidden;
          box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.05), 0 10px 10px -5px rgba(0, 0, 0, 0.02);
        }
        :global(.dark) .article-card {
          background: linear-gradient(145deg, #111827, #1f2937);
          border-color: #374151;
          box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.3);
        }
        .article-card::before {
          content: "";
          position: absolute;
          top: 0;
          left: 0;
          width: 100%;
          height: 4px;
          background: linear-gradient(to right, #2563eb, #60a5fa, #2563eb);
        }
        .article-card::after {
          content: "";
          position: absolute;
          top: -50px;
          right: -50px;
          width: 250px;
          height: 250px;
          background: radial-gradient(circle, rgba(37, 99, 235, 0.03) 0%, transparent 70%);
          pointer-events: none;
        }
        .comparison-table-wrapper {
          overflow-x: auto;
          margin: 3rem 0;
          border-radius: 1.5rem;
          border: 1px solid #e2e8f0;
        }
        :global(.dark) .comparison-table-wrapper { border-color: #374151; }
        .comparison-table {
          width: 100%;
          border-collapse: collapse;
          text-align: left;
        }
        .comparison-table th {
          background: #f8fafc;
          padding: 1.5rem;
          font-weight: 700;
          color: #111827;
          border-bottom: 1px solid #e2e8f0;
        }
        :global(.dark) .comparison-table th {
          background: #1f2937;
          color: #f3f4f6;
          border-bottom-color: #374151;
        }
        .comparison-table td {
          padding: 1.5rem;
          border-bottom: 1px solid #f1f5f9;
        }
        :global(.dark) .comparison-table td { border-bottom-color: #374151; }
        .comparison-table tr:last-child td {
          border-bottom: none;
        }
        .faq-item {
          border-bottom: 1px solid #f1f5f9;
          padding: 2rem 0;
        }
        :global(.dark) .faq-item { border-bottom-color: #374151; }
        .faq-item:last-child {
          border-bottom: none;
        }
        .faq-question {
          font-weight: 700;
          font-size: 1.25rem;
          color: #111827;
          margin-bottom: 1rem;
          display: flex;
          gap: 1rem;
        }
        :global(.dark) .faq-question { color: #f3f4f6; }
        .faq-answer {
          color: #4b5563;
        }
        :global(.dark) .faq-answer { color: #9ca3af; }
      `}</style>
      
      <header>
        {/* Header/Nav */}
        <nav className="fixed top-0 w-full z-50 bg-white/80 dark:bg-gray-900/80 backdrop-blur-md border-b border-gray-100 dark:border-gray-800">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16 items-center">
            <div className="flex-shrink-0 flex items-center">
              <Link href="/" className="flex items-center group">
                <Image
                  src="/logo-Header.png"
                  alt="Synaply Logo"
                  width={32}
                  height={40}
                  className="w-auto h-8 mr-2"
                />
                <span className="hidden xs:inline text-xl font-bold tracking-tight text-synaply-blue dark:text-blue-400">
                  synaply<span className="text-synaply-cyan dark:text-blue-300">.me</span>
                </span>
                <span className="hidden sm:inline ml-2 px-1.5 py-0.5 text-[10px] font-bold uppercase tracking-wider text-synaply-blue bg-synaply-blue/5 dark:bg-blue-900/30 border border-synaply-blue/10 dark:border-blue-800 rounded">{t('common.beta')}</span>
              </Link>
            </div>
              <div className="flex items-center gap-2 sm:gap-4 md:gap-6">
                <a 
                  href="https://calendar.app.google/h6jhP5YxKDxrjobw9"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="flex items-center px-4 py-2 rounded-full bg-synaply-blue/10 text-synaply-blue dark:text-blue-400 font-bold text-xs uppercase tracking-wider hover:bg-synaply-blue/20 transition-all border border-synaply-blue/20"
                >
                  <Star className="w-3.5 h-3.5 mr-2 fill-current" />
                  <span className="hidden sm:inline">{t('landing.book_demo')}</span>
                  <span className="sm:hidden">Demo</span>
                </a>
                <Link href="/public-lists" className="hidden md:inline text-sm font-semibold text-gray-600 dark:text-gray-300 hover:text-synaply-blue dark:hover:text-synaply-cyan transition-colors">{t('dashboard.public_lists.title')}</Link>
                <Link href="/help" className="hidden md:inline text-sm font-semibold text-gray-600 dark:text-gray-300 hover:text-synaply-blue dark:hover:text-synaply-cyan transition-colors">{t('common.help')}</Link>
                {mounted && (isLoggedIn ? (
                  <Link 
                    href="/dashboard"
                    className="px-4 sm:px-6 py-2 rounded-full synaply-gradient-bg text-white font-bold text-xs sm:text-sm uppercase tracking-wider hover:opacity-90 transition-all shadow-lg shadow-synaply-blue-shadow"
                  >
                    {t('common.dashboard')}
                  </Link>
                ) : (
                  <>
                    <Link href="/login" className="text-gray-600 dark:text-gray-400 hover:text-synaply-blue dark:hover:text-synaply-cyan font-bold text-xs sm:text-sm uppercase tracking-wider px-2 transition-colors">{t('common.login')}</Link>
                    <Link 
                      href="/register"
                      className="px-4 sm:px-6 py-2 rounded-full synaply-gradient-bg text-white font-bold text-xs sm:text-sm uppercase tracking-wider hover:opacity-90 transition-all shadow-lg shadow-synaply-blue/20"
                    >
                      {t('landing.cta_btn')}
                    </Link>
                  </>
                ))}
                {!mounted && (
                  <div className="h-10 w-24 sm:w-32 bg-gray-100 dark:bg-gray-800 animate-pulse rounded-full" />
                )}
              </div>
            </div>
          </div>
        </nav>
      </header>

      <main>
        {/* Hero Section */}
        <section className="pt-32 pb-20 px-4">
        <div className="max-w-4xl mx-auto text-center">
          <div className="inline-block px-4 py-1.5 rounded-full bg-synaply-blue/5 text-synaply-blue text-sm font-semibold mb-6 animate-fade-in border border-synaply-blue/10">
            {t('landing.next_gen_learning')}
          </div>
          <h1 
            className="text-5xl md:text-7xl font-extrabold tracking-tight mb-8 bg-gradient-to-br from-synaply-blue via-synaply-light-blue to-synaply-purple bg-clip-text text-transparent pb-2"
            dangerouslySetInnerHTML={{ __html: tHtml('landing.hero_title') }}
          />
          <p 
            className="text-xl text-gray-600 dark:text-gray-400 mb-10 max-w-2xl mx-auto leading-relaxed"
            dangerouslySetInnerHTML={{ __html: tHtml('landing.hero_subtitle') }}
          />
          <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
            <Link 
              href={isLoggedIn ? "/dashboard" : "/register"}
              className="w-full sm:w-auto px-10 py-4 synaply-gradient-bg text-white rounded-full font-bold text-lg hover:opacity-90 hover:scale-105 transition-all shadow-xl shadow-synaply-blue-shadow flex items-center justify-center gap-2"
            >
              {t('landing.try_for_free')}
              <ArrowRight className="w-5 h-5" />
            </Link>
          </div>
          <div className="mt-16 relative max-w-5xl mx-auto">
            <div className="absolute -inset-1 bg-gradient-to-r from-synaply-blue via-synaply-light-blue to-synaply-purple rounded-3xl blur opacity-20"></div>
            <Image 
              src="/opengraph-image.png" 
              alt={t('landing.og_image_alt')} 
              width={1200}
              height={630}
              className="relative rounded-2xl shadow-2xl border border-gray-100 w-full"
            />
          </div>
        </div>
      </section>

      {/* Stats/Quick Features */}
      <section className="py-12 border-y border-gray-100 dark:border-gray-800 bg-white/50 dark:bg-gray-900/50 backdrop-blur-sm">
        <div className="max-w-7xl mx-auto px-4 grid grid-cols-2 md:grid-cols-4 gap-8">
          <div className="text-center">
            <div className="text-3xl font-bold text-synaply-blue dark:text-blue-400 mb-10">{t('landing.stat_fsrs')}</div>
            <div className="text-sm text-gray-500 dark:text-gray-400 uppercase tracking-wider font-bold">{t('landing.stat_memory_algo')}</div>
          </div>
          <div className="text-center">
            <div className="text-3xl font-bold text-synaply-blue dark:text-blue-400 mb-10">{t('landing.stat_ai')}</div>
            <div className="text-sm text-gray-500 dark:text-gray-400 uppercase tracking-wider font-bold">{t('landing.stat_gemini')}</div>
          </div>
          <div className="text-center">
            <div className="text-3xl font-bold text-synaply-blue dark:text-blue-400 mb-10">{t('landing.stat_114')}</div>
            <div className="text-sm text-gray-500 dark:text-gray-400 uppercase tracking-wider font-bold text-center relative overflow-hidden">
              {t('landing.stat_languages')}
              <div className="mt-2 overflow-hidden flex whitespace-nowrap mask-fade">
                <div className="animate-scroll flex gap-2">
                  {[...flags, ...flags].map((flag, i) => (
                    <span 
                      key={i} 
                      className="animate-wave text-xl" 
                      style={{ animationDelay: `${(i % flags.length) * 0.1}s` }}
                    >
                      {flag}
                    </span>
                  ))}
                </div>
              </div>
            </div>
          </div>
          <div className="text-center">
            <div className="text-3xl font-bold text-synaply-blue dark:text-blue-400 mb-10">{t('landing.stat_deepl')}</div>
            <div className="text-sm text-gray-500 dark:text-gray-400 uppercase tracking-wider font-bold">{t('landing.stat_accuracy')}</div>
          </div>
        </div>
      </section>

      {/* Main Features */}
      <section className="py-24 bg-white/40 dark:bg-gray-950/40 backdrop-blur-sm">
        <div className="max-w-7xl mx-auto px-4">
          <div className="text-center mb-20">
            <h2 className="text-4xl font-bold mb-6 text-synaply-blue dark:text-blue-400" dangerouslySetInnerHTML={{ __html: tHtml('landing.feature_title') }} />
            <p className="text-gray-600 dark:text-gray-400 max-w-3xl mx-auto text-xl" dangerouslySetInnerHTML={{ __html: tHtml('landing.feature_subtitle') }} />
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-8">
            <div className="bg-white/70 dark:bg-gray-900/70 backdrop-blur-md p-10 rounded-3xl border border-white/50 dark:border-gray-800 shadow-sm hover:shadow-xl transition-all duration-300">
              <div className="w-14 h-14 bg-synaply-blue/10 dark:bg-blue-900/30 rounded-2xl flex items-center justify-center mb-8">
                <Brain className="w-7 h-7 text-synaply-blue dark:text-blue-400" />
              </div>
              <h3 className="text-2xl font-bold mb-4 text-synaply-blue dark:text-blue-400" dangerouslySetInnerHTML={{ __html: tHtml('landing.card_fsrs_title') }} />
              <p className="text-gray-600 dark:text-gray-400 leading-relaxed mb-6" dangerouslySetInnerHTML={{ __html: tHtml('landing.card_fsrs_desc') }} />
              <div className="flex items-center gap-2 text-synaply-blue dark:text-blue-400 font-bold text-sm">
                <Target className="w-4 h-4" />
                {t('landing.card_fsrs_tag')}
              </div>
            </div>

            <div className="bg-white/70 dark:bg-gray-900/70 backdrop-blur-md p-10 rounded-3xl border border-white/50 dark:border-gray-800 shadow-sm hover:shadow-xl transition-all duration-300">
              <div className="w-14 h-14 bg-synaply-cyan/10 dark:bg-synaply-cyan/20 rounded-2xl flex items-center justify-center mb-8">
                <Globe className="w-7 h-7 text-synaply-cyan" />
              </div>
              <h3 className="text-2xl font-bold mb-4 text-synaply-blue dark:text-blue-400" dangerouslySetInnerHTML={{ __html: tHtml('landing.card_lang_title') }} />
              <p className="text-gray-600 dark:text-gray-400 leading-relaxed mb-6" dangerouslySetInnerHTML={{ __html: tHtml('landing.card_lang_desc') }} />
              <div className="flex items-center gap-2 text-synaply-cyan font-bold text-sm">
                <Target className="w-4 h-4" />
                {t('landing.card_lang_tag')}
              </div>
            </div>

            <div className="bg-white/70 dark:bg-gray-900/70 backdrop-blur-md p-10 rounded-3xl border border-white/50 dark:border-gray-800 shadow-sm hover:shadow-xl transition-all duration-300">
              <div className="w-14 h-14 bg-amber-100/50 dark:bg-amber-900/30 rounded-2xl flex items-center justify-center mb-8">
                <Sparkles className="w-7 h-7 text-amber-600 dark:text-amber-400" />
              </div>
              <h3 className="text-2xl font-bold mb-4 text-synaply-blue dark:text-blue-400" dangerouslySetInnerHTML={{ __html: tHtml('landing.card_ai_title') }} />
              <p className="text-gray-600 dark:text-gray-400 leading-relaxed mb-6" dangerouslySetInnerHTML={{ __html: tHtml('landing.card_ai_desc') }} />
              <ul className="space-y-3">
                <li className="flex gap-3 text-sm text-gray-600 dark:text-gray-400">
                  <CheckCircle className="w-5 h-5 text-amber-500 shrink-0" />
                  <span dangerouslySetInnerHTML={{ __html: tHtml('landing.card_ai_example') }} />
                </li>
                <li className="flex gap-3 text-sm text-gray-600 dark:text-gray-400">
                  <CheckCircle className="w-5 h-5 text-amber-500 shrink-0" />
                  <span dangerouslySetInnerHTML={{ __html: tHtml('landing.card_ai_interactive') }} />
                </li>
              </ul>
            </div>
          </div>
        </div>
      </section>

      {/* AI Detailed Section */}
      <section className="py-24 px-4 bg-white dark:bg-gray-950 overflow-hidden">
        <div className="max-w-7xl mx-auto">
          <div className="flex flex-col lg:flex-row items-center gap-16">
            <div className="lg:w-1/2">
              <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-synaply-blue/5 text-synaply-blue text-sm font-bold mb-6 border border-synaply-blue/10">
                <Bot className="w-4 h-4" />
                {t('landing.ai_companion_tag')}
              </div>
              <h2 className="text-4xl font-bold mb-8 leading-tight text-synaply-blue dark:text-blue-400" dangerouslySetInnerHTML={{ __html: tHtml('landing.beyond_translation_title') }} />
              <div className="space-y-8">
                <div className="flex gap-6">
                  <div className="w-12 h-12 bg-synaply-blue/10 dark:bg-blue-900/30 rounded-xl flex items-center justify-center shrink-0">
                    <MessageSquare className="w-6 h-6 text-synaply-blue dark:text-blue-400" />
                  </div>
                  <div>
                    <h4 className="text-xl font-bold mb-2 text-synaply-blue dark:text-blue-400" dangerouslySetInnerHTML={{ __html: tHtml('landing.context_title') }} />
                    <p className="text-gray-600 dark:text-gray-400" dangerouslySetInnerHTML={{ __html: tHtml('landing.context_desc') }} />
                  </div>
                </div>
                <div className="flex gap-6">
                  <div className="w-12 h-12 bg-synaply-cyan/10 dark:bg-synaply-cyan/20 rounded-xl flex items-center justify-center shrink-0">
                    <Target className="w-6 h-6 text-synaply-cyan" />
                  </div>
                  <div>
                    <h4 className="text-xl font-bold mb-2 text-synaply-blue dark:text-blue-400" dangerouslySetInnerHTML={{ __html: tHtml('landing.personalized_title') }} />
                    <p className="text-gray-600 dark:text-gray-400" dangerouslySetInnerHTML={{ __html: tHtml('landing.personalized_desc') }} />
                  </div>
                </div>
                <div className="flex gap-6">
                  <div className="w-12 h-12 bg-synaply-blue/10 dark:bg-blue-900/30 rounded-xl flex items-center justify-center shrink-0">
                    <Languages className="w-6 h-6 text-synaply-blue dark:text-blue-400" />
                  </div>
                  <div>
                    <h4 className="text-xl font-bold mb-2 text-synaply-blue dark:text-blue-400" dangerouslySetInnerHTML={{ __html: tHtml('landing.precision_title') }} />
                    <p className="text-gray-600 dark:text-gray-400" dangerouslySetInnerHTML={{ __html: tHtml('landing.precision_desc') }} />
                  </div>
                </div>
              </div>
            </div>
            <div className="lg:w-1/2 relative">
              <div className="bg-gradient-to-tr from-synaply-blue to-synaply-cyan rounded-[3rem] p-1 shadow-2xl overflow-hidden">
                <Image 
                  src="/synaply.png" 
                  alt="Synaply - Smart language learning platform" 
                  width={800}
                  height={600}
                  className="w-full h-full object-cover rounded-[2.8rem]"
                />
              </div>
              <div className="absolute -bottom-6 -right-6 w-32 h-32 bg-synaply-cyan/20 blur-3xl rounded-full"></div>
              <div className="absolute -top-6 -left-6 w-32 h-32 bg-synaply-blue/20 blur-3xl rounded-full"></div>
            </div>
          </div>
        </div>
      </section>

      {/* Steps */}
      <section className="py-24 px-4 bg-white/50 dark:bg-gray-900/50 backdrop-blur-sm">
        <div className="max-w-5xl mx-auto">
          <div className="text-center mb-16">
            <h2 className="text-4xl font-bold mb-4 text-synaply-blue dark:text-blue-400" dangerouslySetInnerHTML={{ __html: tHtml('landing.how_it_works_title') }} />
            <p className="text-gray-600 dark:text-gray-400 text-lg" dangerouslySetInnerHTML={{ __html: tHtml('landing.how_it_works_subtitle') }} />
          </div>

          <div className="space-y-12">
            {[
              {
                step: "1",
                title: tHtml('landing.step1_title'),
                description: tHtml('landing.step1_desc')
              },
              {
                step: "2",
                title: tHtml('landing.step2_title'),
                description: tHtml('landing.step2_desc')
              },
              {
                step: "3",
                title: tHtml('landing.step3_title'),
                description: tHtml('landing.step3_desc')
              },
              {
                step: "4",
                title: tHtml('landing.step4_title'),
                description: tHtml('landing.step4_desc')
              }
            ].map((item, idx) => (
              <div key={idx} className="flex flex-col md:flex-row gap-8 items-start group">
                <div className="w-14 h-14 bg-synaply-blue dark:bg-blue-600 text-white rounded-full flex items-center justify-center font-bold text-2xl shrink-0 shadow-lg shadow-synaply-blue/20 dark:shadow-blue-900/40 group-hover:scale-110 transition-transform">
                  {item.step}
                </div>
                <div className="pt-2">
                  <h3 className="text-2xl font-bold mb-3 text-synaply-blue dark:text-blue-400" dangerouslySetInnerHTML={{ __html: item.title }} />
                  <p className="text-lg text-gray-600 dark:text-gray-400 leading-relaxed" dangerouslySetInnerHTML={{ __html: item.description }} />
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* SEO Article Section */}
      <section className="py-32 px-4 bg-white dark:bg-gray-950 border-t border-gray-100 dark:border-gray-800">
        <article className="prose-article">
          <div className="mb-20 text-center">
            <h2 className="!mt-0" dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_title') }} />
            <p className="text-2xl text-gray-500 dark:text-gray-400 max-w-3xl mx-auto" dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_subtitle') }} />
          </div>

          <p dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_p1') }} />

          <div className="article-card">
            <h3 dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_fsrs_title') }} />
            <p dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_fsrs_p1') }} />
            <p dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_fsrs_p2') }} />
            
            <div className="mt-8 p-6 bg-blue-50/50 dark:bg-blue-900/10 rounded-xl border border-blue-100 dark:border-blue-800">
              <h4 className="!mt-0" dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_fsrs_vs_sm2_title') }} />
              <p className="text-sm text-gray-600 dark:text-gray-400" dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_fsrs_vs_sm2_p1') }} />
            </div>
          </div>

          <h3 dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_forgetting_title') }} />
          <p dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_forgetting_p1') }} />
          
          <blockquote dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_forgetting_quote') }} />

          <p dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_forgetting_p2') }} />

          <h3 dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_tech_stack_title') }} />
          <p dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_tech_stack_p1') }} />
          <ul className="list-disc pl-6 space-y-4">
            <li dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_tech_stack_li1') }} />
            <li dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_tech_stack_li2') }} />
            <li dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_tech_stack_li3') }} />
          </ul>

          <div className="article-card bg-blue-50/30 dark:bg-blue-900/10 border-blue-100 dark:border-blue-800">
            <h3 dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_comparison_title') }} />
            <div className="comparison-table-wrapper">
              <table className="comparison-table">
                <thead>
                  <tr>
                    <th dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_table_h1') }} />
                    <th dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_table_h2') }} />
                    <th dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_table_h3') }} />
                  </tr>
                </thead>
                <tbody>
                  <tr>
                    <td><strong dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_table_r1_c1') }} /></td>
                    <td dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_table_r1_c2') }} />
                    <td dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_table_r1_c3') }} />
                  </tr>
                  <tr>
                    <td><strong dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_table_r2_c1') }} /></td>
                    <td dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_table_r2_c2') }} />
                    <td dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_table_r2_c3') }} />
                  </tr>
                  <tr>
                    <td><strong dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_table_r3_c1') }} /></td>
                    <td dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_table_r3_c2') }} />
                    <td dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_table_r3_c3') }} />
                  </tr>
                  <tr>
                    <td><strong dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_table_r4_c1') }} /></td>
                    <td dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_table_r4_c2') }} />
                    <td dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_table_r4_c3') }} />
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <h3 dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_cognitive_title') }} />
          <p dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_cognitive_p1') }} />
          <p dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_cognitive_p2') }} />

          <h3 dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_core_features_title') }} />
          <p dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_core_features_p1') }} />
          <ul className="list-disc pl-6 space-y-4">
            <li dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_core_features_li1') }} />
            <li dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_core_features_li2') }} />
            <li dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_core_features_li3') }} />
            <li dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_core_features_li4') }} />
          </ul>

          <h3 dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_social_title') }} />
          <p dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_social_p1') }} />

          <div className="article-card bg-synaply-blue/5 dark:bg-synaply-blue/10 border-synaply-blue/10 dark:border-synaply-blue/20">
            <h3 className="text-synaply-blue dark:text-blue-400" dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_use_cases_title') }} />
            <div className="space-y-12">
              <div>
                <h4 className="text-synaply-blue dark:text-blue-400 font-bold" dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_use_case1_title') }} />
                <p className="italic text-gray-800 dark:text-gray-300" dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_use_case1_p1') }} />
                <p className="font-bold text-synaply-blue/80 dark:text-blue-400/80" dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_use_case1_author') }} />
              </div>
              <div className="pt-6 border-t border-synaply-blue/10 dark:border-synaply-blue/20">
                <h4 className="text-synaply-blue dark:text-blue-400 font-bold" dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_use_case2_title') }} />
                <p className="italic text-gray-800 dark:text-gray-300" dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_use_case2_p1') }} />
                <p className="font-bold text-synaply-blue/80 dark:text-blue-400/80" dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_use_case2_author') }} />
              </div>
            </div>
          </div>

          <h3 dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_faq_title') }} />
          <div className="space-y-4">
            <div className="faq-item group">
              <div className="faq-question text-synaply-blue dark:text-blue-400 group-hover:text-synaply-cyan transition-colors">
                <CheckCircle className="w-6 h-6 text-synaply-cyan shrink-0" />
                <span dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_faq_q1') }} />
              </div>
              <div className="faq-answer text-gray-600 dark:text-gray-400" dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_faq_a1') }} />
            </div>
            <div className="faq-item group">
              <div className="faq-question text-synaply-blue dark:text-blue-400 group-hover:text-synaply-cyan transition-colors">
                <CheckCircle className="w-6 h-6 text-synaply-cyan shrink-0" />
                <span dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_faq_q2') }} />
              </div>
              <div className="faq-answer text-gray-600 dark:text-gray-400" dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_faq_a2') }} />
            </div>
            <div className="faq-item group">
              <div className="faq-question text-synaply-blue dark:text-blue-400 group-hover:text-synaply-cyan transition-colors">
                <CheckCircle className="w-6 h-6 text-synaply-cyan shrink-0" />
                <span dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_faq_q3') }} />
              </div>
              <div className="faq-answer text-gray-600 dark:text-gray-400" dangerouslySetInnerHTML={{ __html: tHtml('landing.seo_faq_a3') }} />
            </div>
          </div>

        </article>
      </section>

      {/* CTA Section */}
      <section className="py-24 px-4 synaply-gradient-bg text-white overflow-hidden relative shadow-none">
        <div className="absolute top-0 right-0 -translate-y-1/2 translate-x-1/2 w-96 h-96 bg-synaply-purple/10 blur-[100px] rounded-full"></div>
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-full h-full bg-synaply-light-blue/5 blur-[120px] rounded-full pointer-events-none"></div>
        <div className="absolute bottom-0 left-0 translate-y-1/2 -translate-x-1/2 w-96 h-96 bg-synaply-blue/10 blur-[100px] rounded-full"></div>
        
        <div className="max-w-4xl mx-auto text-center relative z-10">
          <h2 className="text-4xl md:text-6xl font-bold mb-8 tracking-tight">{t('landing.cta_title')}</h2>
          <p className="text-xl text-blue-100 mb-12 max-w-2xl mx-auto opacity-90">
            {t('landing.cta_subtitle')}
          </p>
          <div className="flex flex-col sm:flex-row gap-6 justify-center">
            <Link 
              href={isLoggedIn ? "/dashboard" : "/register"}
              className="px-12 py-5 bg-white text-synaply-blue rounded-full font-bold text-lg hover:bg-synaply-cyan hover:scale-105 transition-all shadow-2xl shadow-white/10"
            >
              {t('landing.cta_btn')}
            </Link>
            {!isLoggedIn && (
              <Link 
                href="/login"
                className="px-12 py-5 bg-white/10 backdrop-blur-md text-white border border-white/20 rounded-full font-bold text-lg hover:bg-white/20 transition-all"
              >
                {t('landing.cta_already_account')}
              </Link>
            )}
          </div>
          <div className="mt-16 flex flex-wrap justify-center items-center gap-8 grayscale opacity-50">
            <div className="text-xl font-bold tracking-tighter text-blue-100">{t('landing.tag_ai_powered')}</div>
            <div className="text-xl font-bold tracking-tighter text-blue-100">{t('landing.tag_fsrs')}</div>
            <div className="text-xl font-bold tracking-tighter text-blue-100">{t('landing.tag_deepl')}</div>
            <div className="text-xl font-bold tracking-tighter text-blue-100">{t('landing.tag_114_languages')}</div>
          </div>
        </div>
      </section>
    </main>
    </div>
  );
}
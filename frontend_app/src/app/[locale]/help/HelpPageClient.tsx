'use client';

import React from 'react';
import { Link } from '@/i18n/routing';
import { ArrowLeft, BookOpen, Plus, Star, Sparkles, GraduationCap, List, Lightbulb } from 'lucide-react';
import { useTranslation } from '@/components/I18nContext';

const HelpPageClient = () => {
  const { t, tHtml } = useTranslation();

  return (
    <div className="min-h-screen bg-transparent py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-4xl mx-auto relative z-10">
        <div className="mb-8 flex items-center justify-between">
          <Link href="/dashboard" className="flex items-center text-synaply-blue dark:text-synaply-cyan hover:opacity-80 font-bold transition-all uppercase tracking-wider text-sm">
            <ArrowLeft className="w-5 h-5 mr-2" />
            {t('common.back_to_dashboard')}
          </Link>
          <h1 className="text-3xl font-extrabold synaply-gradient-text">{t('help.title')}</h1>
        </div>

        <div className="space-y-12">
          {/* Section: Best Practice */}
          <section className="bg-white/10 dark:bg-synaply-blue/20 backdrop-blur-md rounded-2xl shadow-xl p-8 text-gray-900 dark:text-white border border-white/20 dark:border-synaply-blue/30 overflow-hidden relative">
            <div className="absolute top-0 right-0 p-4 opacity-10">
              <Lightbulb className="w-24 h-24" />
            </div>
            <h2 className="text-2xl font-bold mb-6 flex items-center uppercase tracking-wider">
              <Lightbulb className="w-6 h-6 mr-3 text-yellow-500 dark:text-yellow-300" />
              {t('help.best_practice.title')}
            </h2>
            <div className="space-y-4 text-gray-700 dark:text-white/90 relative z-10">
              <p className="text-lg leading-relaxed font-medium">
                {t('help.best_practice.strategy_desc')}
              </p>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mt-6">
                <div className="bg-white/50 dark:bg-white/10 backdrop-blur-sm rounded-xl p-5 border border-white/20 shadow-sm">
                  <div className="text-2xl font-bold mb-2 text-synaply-blue dark:text-white">{t('help.best_practice.step1_title')}</div>
                  <p className="text-sm leading-relaxed" dangerouslySetInnerHTML={{ __html: tHtml('help.best_practice.step1_desc') }} />
                </div>
                <div className="bg-white/50 dark:bg-white/10 backdrop-blur-sm rounded-xl p-5 border border-white/20 shadow-sm">
                  <div className="text-2xl font-bold mb-2 text-synaply-blue dark:text-white">{t('help.best_practice.step2_title')}</div>
                  <p className="text-sm leading-relaxed" dangerouslySetInnerHTML={{ __html: tHtml('help.best_practice.step2_desc') }} />
                </div>
                <div className="bg-white/50 dark:bg-white/10 backdrop-blur-sm rounded-xl p-5 border border-white/20 shadow-sm">
                  <div className="text-2xl font-bold mb-2 text-synaply-blue dark:text-white">{t('help.best_practice.step3_title')}</div>
                  <p className="text-sm leading-relaxed" dangerouslySetInnerHTML={{ __html: tHtml('help.best_practice.step3_desc') }} />
                </div>
              </div>
            </div>
          </section>

          {/* Section: Getting Started */}
          <section className="bg-white/70 dark:bg-gray-900/40 backdrop-blur-md rounded-2xl shadow-xl border border-white/20 dark:border-gray-800/50 p-8 transition-all">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-6 flex items-center uppercase tracking-wider">
              <Star className="w-6 h-6 mr-3 text-yellow-500" />
              {t('help.getting_started.title')}
            </h2>
            <p className="text-gray-600 dark:text-gray-400 mb-4 text-lg">
              {t('help.getting_started.desc')}
            </p>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mt-8">
              <div className="p-5 bg-synaply-blue/5 dark:bg-synaply-cyan/10 rounded-xl border border-synaply-blue/10 dark:border-synaply-cyan/20">
                <h3 className="font-bold text-synaply-blue dark:text-synaply-cyan mb-2 uppercase tracking-wide text-sm">{t('help.getting_started.step1_title')}</h3>
                <p className="text-sm text-gray-600 dark:text-gray-400 leading-relaxed">{t('help.getting_started.step1_desc')}</p>
              </div>
              <div className="p-5 bg-purple-50/50 dark:bg-purple-900/20 rounded-xl border border-purple-100 dark:border-purple-900/30">
                <h3 className="font-bold text-purple-900 dark:text-purple-300 mb-2 uppercase tracking-wide text-sm">{t('help.getting_started.step2_title')}</h3>
                <p className="text-sm text-purple-800 dark:text-purple-400 leading-relaxed">{t('help.getting_started.step2_desc')}</p>
              </div>
            </div>
          </section>

          {/* Section: Adding Words */}
          <section className="bg-white/70 dark:bg-gray-900/40 backdrop-blur-md rounded-2xl shadow-xl border border-white/20 dark:border-gray-800/50 p-8 transition-all">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-6 flex items-center uppercase tracking-wider">
              <Plus className="w-6 h-6 mr-3 text-synaply-blue dark:text-synaply-cyan" />
              {t('help.adding_words.title')}
            </h2>
            <div className="space-y-6">
              {[
                { step: 1, title: t('help.adding_words.manual.title'), desc: t('help.adding_words.manual.desc') },
                { step: 2, title: t('help.adding_words.auto_translate.title'), desc: t('help.adding_words.auto_translate.desc'), isHtml: true },
                { step: 3, title: t('help.adding_words.ai_list.title'), desc: t('help.adding_words.ai_list.desc') },
                { step: 4, title: t('help.adding_words.ai_insights.title'), desc: t('help.adding_words.ai_insights.desc') }
              ].map((item) => (
                <div key={item.step} className="flex items-start group">
                  <div className="flex-shrink-0 w-10 h-10 bg-synaply-blue/10 dark:bg-synaply-cyan/10 rounded-xl flex items-center justify-center font-bold text-synaply-blue dark:text-synaply-cyan mr-5 border border-synaply-blue/20 dark:border-synaply-cyan/30 group-hover:scale-110 transition-transform">
                    {item.step}
                  </div>
                  <div>
                    <h3 className="font-bold text-gray-900 dark:text-gray-100 text-lg mb-1">{item.title}</h3>
                    {item.isHtml ? (
                      <p className="text-gray-600 dark:text-gray-400 text-sm leading-relaxed" dangerouslySetInnerHTML={{ __html: tHtml(item.desc) }} />
                    ) : (
                      <p className="text-gray-600 dark:text-gray-400 text-sm leading-relaxed">{item.desc}</p>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </section>

          {/* Section: Vocabulary Review */}
          <section className="bg-white/70 dark:bg-gray-900/40 backdrop-blur-md rounded-2xl shadow-xl border border-white/20 dark:border-gray-800/50 p-8 transition-all">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-6 flex items-center uppercase tracking-wider">
              <BookOpen className="w-6 h-6 mr-3 text-green-600" />
              {t('help.spaced_repetition.title')}
            </h2>
            <p className="text-gray-600 dark:text-gray-400 mb-6 italic text-lg leading-relaxed">
              {t('help.spaced_repetition.desc')}
            </p>
            <div className="space-y-4">
              <div className="border-l-4 border-green-200 dark:border-green-600 pl-6 py-2 bg-green-50/30 dark:bg-green-900/10 rounded-r-xl">
                <h3 className="font-bold text-gray-900 dark:text-gray-100 mb-3">{t('help.spaced_repetition.process_title')}</h3>
                <ul className="list-none text-sm text-gray-600 dark:text-gray-400 space-y-3">
                  {[
                    t('help.spaced_repetition.step1'),
                    t('help.spaced_repetition.step2'),
                    t('help.spaced_repetition.step3'),
                    t('help.spaced_repetition.step4')
                  ].map((step, idx) => (
                    <li key={idx} className="flex items-center">
                      <div className="w-1.5 h-1.5 rounded-full bg-green-500 mr-3 shrink-0" />
                      <span dangerouslySetInnerHTML={{ __html: tHtml(step) }} />
                    </li>
                  ))}
                </ul>
              </div>
            </div>
          </section>

          {/* Section: AI Practice */}
          <section className="bg-white/70 dark:bg-gray-900/40 backdrop-blur-md rounded-2xl shadow-xl border border-white/20 dark:border-gray-800/50 p-8 transition-all">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-6 flex items-center uppercase tracking-wider">
              <Sparkles className="w-6 h-6 mr-3 text-purple-600" />
              {t('help.ai_practice.title')}
            </h2>
            <p className="text-gray-600 dark:text-gray-400 mb-6 text-lg">
              {t('help.ai_practice.desc')}
            </p>
            <div className="grid grid-cols-1 gap-6">
              <div className="p-6 bg-white/50 dark:bg-gray-800/50 rounded-2xl border border-white/20 dark:border-gray-700 shadow-inner">
                <h3 className="font-bold text-gray-900 dark:text-gray-100 mb-4 uppercase tracking-wide text-sm">{t('help.ai_practice.how_it_works')}</h3>
                <ol className="space-y-4">
                  {[
                    t('help.ai_practice.step1'),
                    t('help.ai_practice.step2'),
                    t('help.ai_practice.step3'),
                    t('help.ai_practice.step4')
                  ].map((step, idx) => (
                    <li key={idx} className="flex items-start text-sm text-gray-600 dark:text-gray-400">
                      <span className="font-bold text-purple-600 mr-3">{idx + 1}.</span>
                      <span dangerouslySetInnerHTML={{ __html: tHtml(step) }} />
                    </li>
                  ))}
                </ol>
              </div>
            </div>
          </section>

          {/* Section: Managing Words */}
          <section className="bg-white/70 dark:bg-gray-900/40 backdrop-blur-md rounded-2xl shadow-xl border border-white/20 dark:border-gray-800/50 p-8 transition-all">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-6 flex items-center uppercase tracking-wider">
              <List className="w-6 h-6 mr-3 text-synaply-blue dark:text-synaply-cyan" />
              {t('help.dictionary_management.title')}
            </h2>
            <div className="space-y-4 text-gray-600 dark:text-gray-400 text-sm">
              <p className="flex items-center p-3 bg-white/30 dark:bg-gray-800/30 rounded-xl" dangerouslySetInnerHTML={{ __html: `<span class="mr-3 p-2 bg-white dark:bg-gray-800 rounded-lg shadow-sm"><svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-search"><circle cx="11" cy="11" r="8"/><path d="m21 21-4.3-4.3"/></svg></span>` + tHtml('help.dictionary_management.search') }} />
              <p className="flex items-center p-3 bg-white/30 dark:bg-gray-800/30 rounded-xl" dangerouslySetInnerHTML={{ __html: `<span class="mr-3 p-2 bg-white dark:bg-gray-800 rounded-lg shadow-sm"><svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-layers"><path d="m12.83 2.18a2 2 0 0 0-1.66 0L2.6 6.08a1 1 0 0 0 0 1.83l8.58 3.91a2 2 0 0 0 1.66 0l8.58-3.9a1 1 0 0 0 0-1.83Z"/><path d="m22 17.65-9.17 4.16a2 2 0 0 1-1.66 0L2 17.65"/><path d="m22 12.65-9.17 4.16a2 2 0 0 1-1.66 0L2 12.65"/></svg></span>` + tHtml('help.dictionary_management.sorting') }} />
              <p className="flex items-center p-3 bg-white/30 dark:bg-gray-800/30 rounded-xl" dangerouslySetInnerHTML={{ __html: `<span class="mr-3 p-2 bg-white dark:bg-gray-800 rounded-lg shadow-sm text-red-500"><svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-x-circle"><circle cx="12" cy="12" r="10"/><path d="m15 9-6 6"/><path d="m9 9 6 6"/></svg></span>` + tHtml('help.dictionary_management.delete_all') }} />
            </div>
          </section>

          {/* Section: Progress */}
          <section className="bg-white/70 dark:bg-gray-900/40 backdrop-blur-md rounded-2xl shadow-xl border border-white/20 dark:border-gray-800/50 p-8 transition-all mb-10">
            <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-6 flex items-center uppercase tracking-wider">
              <GraduationCap className="w-6 h-6 mr-3 text-synaply-blue dark:text-synaply-cyan" />
              {t('help.progress.title')}
            </h2>
            <p className="text-gray-600 dark:text-gray-400 text-lg leading-relaxed">
              {t('help.progress.desc')}
            </p>
          </section>
        </div>
      </div>
    </div>
  );
};

export default HelpPageClient;

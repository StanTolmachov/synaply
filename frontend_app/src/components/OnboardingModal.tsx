"use client";

import React, { useState, useEffect } from 'react';
import { X, Lightbulb, Rocket, CheckCircle2 } from 'lucide-react';
import { useTranslation } from './I18nContext';

interface OnboardingModalProps {
  onClose: () => void;
}

export function OnboardingModal({ onClose }: OnboardingModalProps) {
  const { t, tHtml } = useTranslation();
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    // Small delay to ensure smooth transition
    const timer = setTimeout(() => setIsVisible(true), 100);
    return () => clearTimeout(timer);
  }, []);

  const handleClose = () => {
    setIsVisible(false);
    setTimeout(onClose, 300); // Wait for animation
  };

  return (
    <div className={`fixed inset-0 z-50 flex items-center justify-center p-4 transition-opacity duration-300 ${isVisible ? 'opacity-100' : 'opacity-0'}`}>
      {/* Backdrop */}
      <div className="absolute inset-0 bg-gray-900/60 backdrop-blur-sm" onClick={handleClose} />
      
      {/* Modal Content */}
      <div className={`relative w-full max-w-2xl max-h-[calc(100dvh-2rem)] overflow-y-auto transform rounded-3xl bg-white shadow-2xl transition-all duration-300 ${isVisible ? 'scale-100 translate-y-0' : 'scale-95 translate-y-4'} [&::-webkit-scrollbar]:hidden [-ms-overflow-style:none] [scrollbar-width:none]`}>
        <div className="sticky right-4 top-4 z-10 flex justify-end">
          <button
            onClick={handleClose}
            className="m-2 rounded-full p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-600 transition-colors bg-white/80 backdrop-blur-sm"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        <div className="p-8 sm:p-12 pt-0 sm:pt-0">
          <div className="mb-8 flex items-center gap-3">
            <div className="rounded-2xl bg-blue-600 p-3 shadow-lg shadow-blue-200">
              <Rocket className="h-8 w-8 text-white" />
            </div>
            <div>
              <h2 className="text-3xl font-bold text-gray-900 leading-tight">{t('dashboard.onboarding.welcome')}</h2>
              <p className="text-blue-600 font-medium text-lg">{t('dashboard.onboarding.journey_begins')}</p>
            </div>
          </div>

          <div className="mb-10 rounded-2xl bg-blue-50/50 border border-blue-100 p-6">
            <h3 className="mb-4 flex items-center text-xl font-bold text-blue-900">
              <Lightbulb className="mr-3 h-6 w-6 text-yellow-500" />
              {t('dashboard.onboarding.best_practice')}
            </h3>
            <p className="mb-6 text-gray-700 leading-relaxed text-lg font-medium">
              {t('dashboard.onboarding.strategy_desc')}
            </p>
            
            <div className="grid gap-6">
              <div className="flex gap-4 items-start">
                <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-blue-600 font-bold text-white shadow-md">1</div>
                <div>
                  <h4 className="font-bold text-gray-900 text-lg">{t('dashboard.onboarding.step1_title')}</h4>
                  <p className="text-gray-600" dangerouslySetInnerHTML={{ __html: tHtml('dashboard.onboarding.step1_desc') }} />
                </div>
              </div>
              
              <div className="flex gap-4 items-start">
                <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-blue-600 font-bold text-white shadow-md">2</div>
                <div>
                  <h4 className="font-bold text-gray-900 text-lg">{t('dashboard.onboarding.step2_title')}</h4>
                  <p className="text-gray-600" dangerouslySetInnerHTML={{ __html: tHtml('dashboard.onboarding.step2_desc') }} />
                </div>
              </div>
              
              <div className="flex gap-4 items-start">
                <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-blue-600 font-bold text-white shadow-md">3</div>
                <div>
                  <h4 className="font-bold text-gray-900 text-lg">{t('dashboard.onboarding.step3_title')}</h4>
                  <p className="text-gray-600" dangerouslySetInnerHTML={{ __html: tHtml('dashboard.onboarding.step3_desc') }} />
                </div>
              </div>
            </div>
          </div>

          <button
            onClick={handleClose}
            className="flex w-full items-center justify-center gap-2 rounded-2xl bg-blue-600 py-4 text-xl font-bold text-white shadow-xl shadow-blue-200 hover:bg-blue-700 hover:scale-[1.02] active:scale-[0.98] transition-all"
          >
            <CheckCircle2 className="h-6 w-6" />
            {t('dashboard.onboarding.got_it')}
          </button>
        </div>
      </div>
    </div>
  );
}

import { TrendingUp, Target, Send, AlertCircle, BarChart2 } from 'lucide-react';
import { useState } from 'react';
import { motion } from 'framer-motion';
import type { CampaignResult, MatchedProduct, EventItem } from '../types';

interface AnalyticsProps {
  campaignResult: CampaignResult | null;
  selectedMatch: {
    product: MatchedProduct;
    event: EventItem;
  } | null;
  isLoading: boolean;
  onGenerateCampaign: (productId: number) => void;
  imageExample: { url: string; prompt: string } | null;
  imageLoading: boolean;
  imageError: string;
  onGenerateImage: () => void;
}

export function Analytics({
  campaignResult,
  selectedMatch,
  isLoading,
  onGenerateCampaign,
  imageExample,
  imageLoading,
  imageError,
  onGenerateImage,
}: AnalyticsProps) {
  const [error, setError] = useState('');

  const handleGenerate = async () => {
    if (!selectedMatch) return;
    setError('');
    try {
      await onGenerateCampaign(selectedMatch.product.product_id);
    } catch (e) {
      setError('Не удалось сгенерировать кампанию');
    }
  };

  function formatRuDate(value: string) {
    const parsed = new Date(value);
    if (Number.isNaN(parsed.getTime())) return value;
    return new Intl.DateTimeFormat('ru-RU', {
      day: 'numeric',
      month: 'long',
      year: 'numeric',
    }).format(parsed);
  }

  const formatPlanDates = (dates: unknown) => {
    if (!dates) return null;
    if (typeof dates === 'string' || typeof dates === 'number') return String(dates);
    if (typeof dates !== 'object') return null;

    const record = dates as Record<string, string | number>;
    const mapped = [
      { key: 'event_start', label: 'Событие', value: record.event_start },
      { key: 'pre_event_start', label: 'Прогрев старт', value: record.pre_event_start },
      { key: 'pre_event_end', label: 'Прогрев финиш', value: record.pre_event_end },
      { key: 'progrev_start', label: 'Прогрев старт', value: record.progrev_start },
      { key: 'progrev_end', label: 'Прогрев финиш', value: record.progrev_end },
    ]
      .filter((item) => item.value)
      .map((item) => {
        const raw = String(item.value);
        return `${item.label}: ${formatRuDate(raw)}`;
      });

    return mapped.length > 0 ? mapped.join(' · ') : null;
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h2 className="text-2xl font-semibold text-gray-900 mb-1">Маркетинг</h2>
        <p className="text-sm text-gray-600">Генерация маркетингового контента</p>
      </div>

      {/* Stats cards */}
      <div className="grid grid-cols-4 gap-4 mb-6">
        <div className="bg-green-600 rounded-xl p-5 text-white">
          <Target className="w-6 h-6 mb-3" />
          <p className="text-3xl font-semibold">{campaignResult?.content?.post_text ? 1 : 0}</p>
          <p className="text-xs text-green-100">Кампаний создано</p>
        </div>
        <div className="bg-blue-600 rounded-xl p-5 text-white">
          <TrendingUp className="w-6 h-6 mb-3" />
          <p className="text-3xl font-semibold">
            {campaignResult?.match?.score != null
              ? campaignResult.match.score.toFixed(3)
              : '—'}
          </p>
          <p className="text-xs text-blue-100">Relevance Score</p>
        </div>
        <div className="bg-purple-600 rounded-xl p-5 text-white">
          <BarChart2 className="w-6 h-6 mb-3" />
          <p className="text-3xl font-semibold">
            {campaignResult?.plan ? (campaignResult.plan as any).objective ? '✓' : 'Draft' : '—'}
          </p>
          <p className="text-xs text-purple-100">Статус плана</p>
        </div>
        <div className="bg-amber-600 rounded-xl p-5 text-white">
          <Send className="w-6 h-6 mb-3" />
          <p className="text-3xl font-semibold">
            {campaignResult?.plan_approved ? '✓' : campaignResult?.plan ? 'Pending' : '—'}
          </p>
          <p className="text-xs text-amber-100">Валидация</p>
        </div>
      </div>

      {/* Match info */}
      {selectedMatch && !campaignResult && !isLoading && (
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          className="bg-white rounded-xl border border-gray-200 p-5 mb-6"
        >
          <div className="flex items-center gap-2 mb-3">
            <Target className="w-5 h-5 text-green-600" />
            <h3 className="font-semibold text-sm text-gray-900">Текущий тандем товар-событие</h3>
          </div>
          <div className="grid grid-cols-2 gap-4 text-sm">
            <div>
              <p className="text-gray-500 text-xs">Товар</p>
              <p className="font-medium text-gray-900">{selectedMatch.product.product_name}</p>
              <p className="text-xs text-gray-500">
                {selectedMatch.product.category} · {selectedMatch.product.price.toLocaleString()} ₽
              </p>
            </div>
            <div>
              <p className="text-gray-500 text-xs">Событие</p>
              <p className="font-medium text-gray-900">{selectedMatch.event.holiday_info}</p>
              <p className="text-xs text-gray-500">
                {formatRuDate(selectedMatch.event.event_date)} · {selectedMatch.event.category}
              </p>
            </div>
          </div>
        </motion.div>
      )}

      {/* Generate campaign button */}
      {selectedMatch && !campaignResult && !isLoading && (
        <div className="mb-6">
          <button
            onClick={handleGenerate}
            disabled={isLoading}
            className="w-full bg-gradient-to-r from-green-600 to-emerald-600 text-white py-3 rounded-xl font-medium text-sm hover:from-green-700 hover:to-emerald-700 transition-all disabled:opacity-50 flex items-center justify-center gap-2"
          >
            <Send className="w-4 h-4" />
            Создать маркетинговую кампанию
          </button>
        </div>
      )}

      {/* Error */}
      {error && (
        <div className="mb-4 text-sm text-red-600 bg-red-50 border border-red-200 rounded-lg p-3">
          <AlertCircle className="w-4 h-4 inline mr-2" />
          {error}
        </div>
      )}

      {/* Loading */}
      {isLoading && (
        <div className="bg-blue-50 border border-blue-200 rounded-xl p-6 mb-6">
          <div className="flex items-center gap-3">
            <div className="w-6 h-6 border-2 border-blue-600 border-t-transparent rounded-full animate-spin" />
            <div>
              <p className="text-sm font-medium text-blue-800">
                Генерируем маркетинговый план…
              </p>
              <p className="text-xs text-blue-600">
                Подготовка плана и контента
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Campaign result */}
      {campaignResult && (
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          className="space-y-4"
        >
          {/* Plan */}
          {campaignResult.plan && (
            <div className="bg-white rounded-xl border border-gray-200 p-5">
              <div className="flex items-center gap-2 mb-3">
                <Target className="w-5 h-5 text-green-600" />
                <h3 className="font-semibold text-sm">Маркетинговый план</h3>
                {campaignResult.plan_approved && (
                  <span className="ml-auto text-xs bg-green-100 text-green-700 px-2 py-0.5 rounded">
                    ✓ Одобрено
                  </span>
                )}
              </div>
              <div className="space-y-2 text-sm">
                {campaignResult.plan.objective && (
                  <p><span className="font-medium text-gray-700">Цель:</span> {campaignResult.plan.objective}</p>
                )}
                {campaignResult.plan.mecanica && (
                  <p><span className="font-medium text-gray-700">Механика:</span> {campaignResult.plan.mecanica}</p>
                )}
                {campaignResult.plan.hypothesis && (
                  <p><span className="font-medium text-gray-700">Гипотеза:</span> {campaignResult.plan.hypothesis}</p>
                )}
                {campaignResult.plan.dates && (
                  <p>
                    <span className="font-medium text-gray-700">Даты:</span>{' '}
                    {formatPlanDates(campaignResult.plan.dates) ?? '—'}
                  </p>
                )}
                {campaignResult.plan.target_audience && (
                  <p><span className="font-medium text-gray-700">Целевая аудитория:</span> {campaignResult.plan.target_audience}</p>
                )}
                {campaignResult.plan.warming && (
                  <p><span className="font-medium text-gray-700">Прогрев:</span> {campaignResult.plan.warming}</p>
                )}
              </div>
            </div>
          )}

          {campaignResult.plan?.promotions && (
            <div className="bg-white rounded-xl border border-gray-200 p-5">
              <div className="flex items-center gap-2 mb-3">
                <Target className="w-5 h-5 text-green-600" />
                <h3 className="font-semibold text-sm">Рекомендации по акциям</h3>
              </div>
              <div className="space-y-3 text-sm">
                {campaignResult.plan.promotions.map((item: any, i: number) => {
                  const product = typeof item === 'string' ? '' : item.product;
                  const promo = typeof item === 'string' ? item : item.promo;
                  const reason = typeof item === 'string' ? '' : item.reason;
                  return (
                    <div key={i} className="bg-gray-50 rounded-lg p-3">
                      {product && <p className="text-gray-900 font-medium">{product}</p>}
                      <p className="text-gray-700">{promo}</p>
                      {reason && (
                        <p className="text-xs text-gray-500 mt-1">{reason}</p>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* Content */}
          {campaignResult.content && (
            <div className="bg-white rounded-xl border border-gray-200 p-5">
              <div className="flex items-center gap-2 mb-3">
                <Send className="w-5 h-5 text-purple-600" />
                <h3 className="font-semibold text-sm">Контент</h3>
              </div>
              <div className="space-y-3 text-sm">
                {campaignResult.content.post_text && (
                  <div className="bg-gray-50 rounded-lg p-3">
                    <p className="text-gray-500 text-xs mb-1">Пост</p>
                    <p className="text-gray-800">{campaignResult.content.post_text}</p>
                  </div>
                )}
                {campaignResult.content.story_bullets && campaignResult.content.story_bullets.length > 0 && (
                  <div className="bg-gray-50 rounded-lg p-3">
                    <p className="text-gray-500 text-xs mb-1">Сторис</p>
                    <ul className="list-disc list-inside space-y-1 text-gray-700">
                      {campaignResult.content.story_bullets.map((bullet, i) => (
                        <li key={i}>{bullet}</li>
                      ))}
                    </ul>
                  </div>
                )}
                {campaignResult.content.push_title && (
                  <div className="bg-gray-50 rounded-lg p-3">
                    <p className="text-gray-500 text-xs mb-1">Push-уведомление</p>
                    <p className="font-medium text-gray-900">{campaignResult.content.push_title}</p>
                    <p className="text-gray-600">{campaignResult.content.push_text}</p>
                  </div>
                )}
              </div>
            </div>
          )}

          <div className="bg-white rounded-xl border border-gray-200 p-5">
            <div className="flex items-center gap-2 mb-3">
              <Send className="w-5 h-5 text-purple-600" />
              <h3 className="font-semibold text-sm">Пример рекламного фото</h3>
            </div>
            <p className="text-xs text-gray-500 mb-3">Это пример, как может выглядеть</p>
            <button
              onClick={onGenerateImage}
              disabled={imageLoading || !campaignResult.image_prompt}
              className="w-full bg-gradient-to-r from-purple-600 to-indigo-600 text-white py-2.5 rounded-lg font-medium text-sm hover:from-purple-700 hover:to-indigo-700 transition-all disabled:opacity-50"
            >
              {imageLoading ? 'Генерация…' : 'Сгенерировать рекламное фото'}
            </button>
            {imageError && (
              <p className="text-xs text-red-600 mt-2">{imageError}</p>
            )}
            {imageExample && (
              <div className="mt-4">
                <div className="border border-gray-200 rounded-lg overflow-hidden bg-gray-50">
                  <img
                    src={imageExample.url}
                    alt="Пример рекламного фото"
                    className="w-full max-h-[60vh] object-contain"
                  />
                </div>
                <p className="text-[11px] text-gray-400 mt-2">Промт: {imageExample.prompt}</p>
              </div>
            )}
          </div>

        </motion.div>
      )}

      {/* Placeholder when nothing selected */}
      {!selectedMatch && !isLoading && (
        <div className="bg-gray-50 rounded-xl border border-gray-200 p-12 text-center">
          <Target className="w-8 h-8 text-gray-300 mx-auto mb-3" />
          <p className="text-sm text-gray-500">
            Выберите товар и событие на вкладках «События» и «Товары»,
            затем нажмите «Создать маркетинговую кампанию»
          </p>
        </div>
      )}
    </div>
  );
}
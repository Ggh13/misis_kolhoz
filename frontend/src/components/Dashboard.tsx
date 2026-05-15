import { TrendingUp, Rocket, CheckCircle } from 'lucide-react';
import { motion } from 'framer-motion';
import type { WorkflowState } from '../types';

interface DashboardProps {
  workflow: WorkflowState;
}

export function Dashboard({ workflow }: DashboardProps) {
  const { selectedEventIds, matchedProducts, selectedMatch, campaignResult, isLoading } = workflow;
  const hasEvents = selectedEventIds.length > 0;

  const stats = [
    { label: 'Выбрано событий', value: selectedEventIds.length > 0 ? String(selectedEventIds.length) : '—', change: 'для анализа', positive: selectedEventIds.length > 0 },
    { label: 'Найдено товаров', value: matchedProducts.length > 0 ? String(matchedProducts.length) : '—', change: 'по векторам', positive: matchedProducts.length > 0 },
    { label: 'Кампания', value: campaignResult ? (campaignResult.plan_approved ? '✓ Готова' : 'Draft') : '—', change: campaignResult ? (campaignResult.retry_count ?? 0) + ' итераций' : 'Нет данных', positive: !!campaignResult?.plan_approved },
  ];

  return (
    <div className="p-6">
      <div className="mb-6">
        <h2 className="text-2xl font-semibold text-gray-900 mb-1">Дашборд</h2>
        <p className="text-sm text-gray-600">Статус ML-пайплайна</p>
      </div>

      <div className="grid grid-cols-3 gap-4 mb-6">
        {stats.map((stat, i) => (
          <motion.div key={i} initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: i * 0.08 }} className="bg-white rounded-xl border border-gray-200 p-5">
            <p className="text-sm text-gray-600 mb-2">{stat.label}</p>
            <div className="flex items-end justify-between">
              <p className="text-3xl font-semibold text-gray-900">{stat.value}</p>
              <span className={`text-xs font-medium ${stat.positive ? 'text-green-600' : 'text-gray-500'}`}>{stat.change}</span>
            </div>
          </motion.div>
        ))}
      </div>

      <div className="grid grid-cols-2 gap-4 mb-6">
        <div className="bg-white rounded-xl border border-gray-200 p-5">
          <div className="flex items-center gap-2 mb-4">
            <Rocket className="w-5 h-5 text-green-600" />
            <h3 className="text-base font-semibold text-gray-900">ML-пайплайн</h3>
            {isLoading && <span className="ml-auto flex items-center gap-1 text-xs text-blue-600"><span className="w-3 h-3 border-2 border-blue-600 border-t-transparent rounded-full animate-spin" />Работает...</span>}
          </div>
          <div className="space-y-2 text-sm">
            {[
              ['Выбор событий', hasEvents],
              ['Embed + Vector Search', matchedProducts.length > 0],
              ['Выбор товара', !!selectedMatch],
              ['Генерация кампании', !!campaignResult],
            ].map(([label, done], i) => (
              <div key={i} className="flex items-center gap-2">
                <span className="text-gray-500 text-xs w-5">{i + 1}</span>
                <span className={done ? 'text-green-600 font-medium' : 'text-gray-400'}>{label}</span>
                {done && <CheckCircle className="w-3 h-3 text-green-500 ml-auto" />}
              </div>
            ))}
          </div>
          {!hasEvents && (
            <button onClick={() => { const el = document.querySelector('[data-tab="events"]'); if (el) (el as HTMLElement).click(); }} className="mt-4 w-full bg-green-600 text-white py-2 rounded-lg text-sm hover:bg-green-700 transition-colors">
              Начать с выбора событий
            </button>
          )}
        </div>

        <div className="bg-white rounded-xl border border-gray-200 p-5">
          <div className="flex items-center gap-2 mb-3">
            <TrendingUp className="w-5 h-5 text-green-600" />
            <h3 className="text-base font-semibold text-gray-900">Тренды</h3>
          </div>
          <p className="text-sm text-gray-700">
            {hasEvents ? `Анализ ${selectedEventIds.length} событий` : 'Выберите события для начала анализа'}
          </p>
          {matchedProducts.length > 0 && (
            <div className="mt-3">
              <p className="text-xs text-gray-500 mb-1">Товары найдены: {matchedProducts.length}</p>
              <div className="flex flex-wrap gap-1">
                {matchedProducts.slice(0, 5).map((p) => (
                  <span key={p.id} className={`text-xs px-2 py-0.5 rounded ${selectedMatch?.product.id === p.id ? 'bg-green-100 text-green-700 font-medium' : 'bg-gray-100 text-gray-600'}`}>{p.product_name}</span>
                ))}
              </div>
            </div>
          )}
          {campaignResult?.content?.post_text && (
            <div className="mt-3 bg-purple-50 border border-purple-100 rounded-lg p-3">
              <p className="text-xs text-purple-700 font-medium mb-1">Готовый пост:</p>
              <p className="text-sm text-purple-900">{campaignResult.content.post_text}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
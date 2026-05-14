import { Sparkles, Search, BarChart2 } from 'lucide-react';
import { motion } from 'framer-motion';
import type { EventItem, MatchedProduct, MatchInfo } from '../types';

function formatRuDate(value: string) {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' }).format(parsed);
}

interface SemanticMatcherProps {
  selectedEvents: EventItem[];
  matchedProducts: MatchedProduct[];
  selectedMatch: {
    product: MatchedProduct;
    event: EventItem;
    match: MatchInfo;
  } | null;
  isLoading: boolean;
}

export function SemanticMatcher({
  selectedEvents,
  matchedProducts,
  selectedMatch,
  isLoading,
}: SemanticMatcherProps) {
  const hasResults = matchedProducts.length > 0;

  return (
    <div className="p-6">
      <div className="mb-6">
        <h2 className="text-2xl font-semibold text-gray-900 mb-1">Сопоставление товаров</h2>
        <p className="text-sm text-gray-600">
          Выбранные события → эмбеддинг → поиск по векторам
        </p>
      </div>

      {selectedEvents.length === 0 && !hasResults && (
        <div className="bg-gradient-to-r from-purple-600 to-pink-600 rounded-xl p-6 text-white mb-6">
          <div className="flex items-center gap-2 mb-2">
            <Sparkles className="w-5 h-5" />
            <h3 className="text-lg font-semibold">AI-движок сопоставления</h3>
          </div>
          <p className="text-purple-100 text-sm">Выберите события в календаре, чтобы запустить векторный поиск товаров</p>
        </div>
      )}

      {isLoading && (
        <div className="bg-blue-50 border border-blue-200 rounded-xl p-4 mb-6">
          <div className="flex items-center gap-3">
            <div className="w-6 h-6 border-2 border-blue-600 border-t-transparent rounded-full animate-spin" />
            <div>
              <p className="text-sm font-medium text-blue-800">Выполняется анализ…</p>
              <p className="text-xs text-blue-600">Эмбеддинг событий → векторный поиск товаров</p>
            </div>
          </div>
        </div>
      )}

      {/* Selected events */}
      {selectedEvents.length > 0 && (
        <motion.div initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }} className="bg-white rounded-xl border border-gray-200 p-4 mb-4">
          <div className="flex items-center gap-2 mb-2">
            <Search className="w-4 h-4 text-green-600" />
            <h3 className="font-semibold text-sm">Выбранные события ({selectedEvents.length})</h3>
          </div>
          <div className="space-y-1">
            {selectedEvents.map((e) => (
              <div key={e.id} className="text-xs text-gray-700 border-l-2 border-green-400 pl-2 py-0.5">
                <span className="font-medium">{e.holiday_info}</span>
                <span className="text-gray-400 ml-1">· {formatRuDate(e.event_date)}</span>
                <span className="text-gray-400 ml-1">· {e.category}</span>
              </div>
            ))}
          </div>
        </motion.div>
      )}

      {/* Results Summary */}
      {hasResults && (
        <motion.div initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }} className="bg-white rounded-xl border border-gray-200 p-4 mb-4">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center gap-2">
              <BarChart2 className="w-4 h-4 text-purple-600" />
              <h3 className="font-semibold text-sm">Результаты сопоставления</h3>
            </div>
            <span className="text-xs text-gray-500">Найдено: {matchedProducts.length}</span>
          </div>
          {selectedMatch && (
            <p className="text-xs text-green-600">
              Выбран: <span className="font-medium">{selectedMatch.product.product_name}</span>
              {' · '}Score: {selectedMatch.match.score.toFixed(3)}
            </p>
          )}
        </motion.div>
      )}

      {/* Product match graph */}
      <div className="bg-white rounded-xl border border-gray-200 p-6 min-h-48 relative overflow-hidden">
        <div className="flex items-center gap-2 mb-4">
          <Sparkles className="w-5 h-5 text-purple-600" />
          <h3 className="text-base font-semibold text-gray-900">Граф сопоставления</h3>
        </div>
        {hasResults ? (
          <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="space-y-2">
            {matchedProducts.slice(0, 5).map((p, i) => (
              <div key={p.id} className={`flex items-center justify-between rounded-lg px-3 py-2 text-sm ${selectedMatch?.product.id === p.id ? 'bg-purple-50 border border-purple-300' : 'bg-gray-50 border border-gray-100'}`}>
                <div>
                  <span className="font-medium text-gray-900">{p.product_name}</span>
                  <span className="text-xs text-gray-500 ml-2">{p.category}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-xs text-gray-500">{p.price.toLocaleString()} ₽</span>
                  {i === 0 && <span className="text-xs bg-purple-100 text-purple-700 px-1.5 py-0.5 rounded">Топ-1</span>}
                </div>
              </div>
            ))}
          </motion.div>
        ) : (
          <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="text-sm text-gray-500 text-center py-8">
            Граф покажет топологию сопоставления после выбора событий
          </motion.div>
        )}
      </div>
    </div>
  );
}
import { Sparkles, Search, Package } from 'lucide-react';
import { motion } from 'framer-motion';
import type { EventItem, MatchedProduct } from '../types';

function formatRuDate(value: string) {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' }).format(parsed);
}

interface SemanticMatcherProps {
  selectedEvents: EventItem[];
  eventResults: Record<number, MatchedProduct[]>;
  selectedMatch: {
    product: MatchedProduct;
    event: EventItem;
    match: { id: string; score: number };
  } | null;
  isLoading: boolean;
}

export function SemanticMatcher({
  selectedEvents,
  eventResults,
  selectedMatch,
  isLoading,
}: SemanticMatcherProps) {
  const eventsWithResults = selectedEvents.filter((e) => (eventResults[e.id]?.length ?? 0) > 0);

  return (
    <div className="p-6">
      <div className="mb-6">
        <h2 className="text-2xl font-semibold text-gray-900 mb-1">Сопоставление товаров</h2>
        <p className="text-sm text-gray-600">
          Выбранные события → эмбеддинг → поиск по векторам
        </p>
      </div>

      {selectedEvents.length === 0 && (
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

      {selectedEvents.length > 0 && (
        <motion.div initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }} className="bg-white rounded-xl border border-gray-200 p-4 mb-6">
          <div className="flex items-center gap-2 mb-2">
            <Search className="w-4 h-4 text-green-600" />
            <h3 className="font-semibold text-sm">Выбранные события ({selectedEvents.length})</h3>
          </div>
          <div className="space-y-1">
            {selectedEvents.map((e) => {
              const count = eventResults[e.id]?.length ?? 0;
              return (
                <div key={e.id} className="text-xs text-gray-700 border-l-2 border-green-400 pl-2 py-0.5 flex justify-between items-center">
                  <div>
                    <span className="font-medium">{e.holiday_info}</span>
                    <span className="text-gray-400 ml-1">· {formatRuDate(e.event_date)}</span>
                    <span className="text-gray-400 ml-1">· {e.category}</span>
                  </div>
                  {count > 0 && (
                    <span className="text-[11px] bg-purple-100 text-purple-700 px-2 py-0.5 rounded-full shrink-0">
                      {count} товаров
                    </span>
                  )}
                </div>
              );
            })}
          </div>
        </motion.div>
      )}

      {eventsWithResults.length > 0 && (
        <div className="space-y-6">
          {eventsWithResults.map((event) => {
            const products = eventResults[event.id] ?? [];
            return (
              <motion.div
                key={event.id}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                className="bg-white rounded-xl border border-gray-200 overflow-hidden"
              >
                <div className="bg-gradient-to-r from-green-50 to-emerald-50 border-b border-gray-200 px-5 py-3">
                  <div className="flex items-center justify-between">
                    <div>
                      <h3 className="font-semibold text-sm text-gray-900">{event.holiday_info}</h3>
                      <p className="text-xs text-gray-500 mt-0.5">
                        {formatRuDate(event.event_date)} · {event.category}
                      </p>
                    </div>
                    <span className="text-xs bg-white text-purple-700 px-2.5 py-1 rounded-full border border-purple-200 font-medium">
                      {products.length} {products.length === 1 ? 'товар' : 'товаров'}
                    </span>
                  </div>
                  {event.about && (
                    <p className="text-xs text-gray-500 mt-1.5 italic line-clamp-2">{event.about}</p>
                  )}
                </div>

                <div className="divide-y divide-gray-100">
                  {products.map((p, i) => (
                    <div
                      key={p.id}
                      className={`flex items-center justify-between px-5 py-3 text-sm transition-colors ${
                        selectedMatch?.product.id === p.id && selectedMatch?.event.id === event.id
                          ? 'bg-purple-50'
                          : 'hover:bg-gray-50'
                      }`}
                    >
                      <div className="flex items-center gap-3 min-w-0">
                        <Package className="w-4 h-4 text-gray-300 shrink-0" />
                        <div className="min-w-0">
                          <span className="font-medium text-gray-900 truncate block">{p.product_name}</span>
                          <span className="text-xs text-gray-400">{p.category}</span>
                        </div>
                      </div>
                      <div className="flex items-center gap-3 shrink-0">
                        <span className="text-sm font-semibold text-gray-800">{p.price.toLocaleString()} ₽</span>
                        {i === 0 && (
                          <span className="text-[11px] bg-purple-100 text-purple-700 px-1.5 py-0.5 rounded font-medium">
                            Топ-1
                          </span>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </motion.div>
            );
          })}
        </div>
      )}

      {selectedEvents.length > 0 && eventsWithResults.length === 0 && !isLoading && (
        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="text-sm text-gray-500 text-center py-12 bg-white rounded-xl border border-gray-200">
          <Package className="w-10 h-10 text-gray-200 mx-auto mb-3" />
          <p className="font-medium text-gray-600">Результаты появятся здесь</p>
          <p className="text-xs text-gray-400 mt-1">После завершения векторного поиска для каждого события</p>
        </motion.div>
      )}
    </div>
  );
}
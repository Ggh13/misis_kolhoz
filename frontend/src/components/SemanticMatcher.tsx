import { Sparkles, Search, Package } from 'lucide-react';
import { motion } from 'framer-motion';
import { useEffect, useMemo, useState } from 'react';
import type { EventItem, EventProductsMatch, MatchedProduct } from '../types';

function formatRuDate(value: string) {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' }).format(parsed);
}

interface SemanticMatcherProps {
  matches: EventProductsMatch[];
  selectedMatch: {
    product: MatchedProduct;
    event: EventItem;
    match: { id: string; score: number };
  } | null;
  isLoading: boolean;
}

export function SemanticMatcher({
  matches,
  selectedMatch,
  isLoading,
}: SemanticMatcherProps) {
  const [activeEventId, setActiveEventId] = useState<number | null>(null);

  useEffect(() => {
    if (matches.length === 0) {
      setActiveEventId(null);
      return;
    }
    if (activeEventId && matches.some((m) => m.event.id === activeEventId)) return;
    setActiveEventId(matches[0]?.event.id ?? null);
  }, [matches, activeEventId]);

  const activeMatch = useMemo(
    () => matches.find((m) => m.event.id === activeEventId) ?? null,
    [matches, activeEventId],
  );

  return (
    <div className="p-6">
      <div className="mb-6">
        <h2 className="text-2xl font-semibold text-gray-900 mb-1">Мэтчи</h2>
        <p className="text-sm text-gray-600">События → поиск по векторам → подбор лучших товаров</p>
      </div>

      {matches.length === 0 && !isLoading && (
        <div className="bg-gradient-to-r from-purple-600 to-pink-600 rounded-xl p-6 text-white mb-6">
          <div className="flex items-center gap-2 mb-2">
            <Sparkles className="w-5 h-5" />
            <h3 className="text-lg font-semibold">AI-движок сопоставления</h3>
          </div>
          <p className="text-purple-100 text-sm">Выберите фермера и подберите события, чтобы получить подбор товаров</p>
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

      {matches.length > 0 && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          <motion.div initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }} className="bg-white rounded-xl border border-gray-200 p-4">
            <div className="flex items-center gap-2 mb-3">
              <Search className="w-4 h-4 text-green-600" />
              <h3 className="font-semibold text-sm">Рекомендованные события ({matches.length})</h3>
            </div>
            <div className="space-y-2">
              {matches.map(({ event, products }) => {
                const isActive = event.id === activeEventId;
                return (
                  <button
                    key={event.id}
                    type="button"
                    onClick={() => setActiveEventId(event.id)}
                    className={`w-full text-left border rounded-lg p-3 transition-all ${
                      isActive
                        ? 'border-green-500 bg-green-50 ring-1 ring-green-200'
                        : 'border-gray-200 hover:border-green-300 hover:bg-green-50/30'
                    }`}
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div className="min-w-0">
                        <p className="text-sm font-medium text-gray-900 truncate">{event.holiday_info}</p>
                        <p className="text-xs text-gray-500 mt-0.5">
                          {formatRuDate(event.event_date)} · {event.category}
                        </p>
                      </div>
                      <span className="text-[11px] bg-purple-100 text-purple-700 px-2 py-0.5 rounded-full shrink-0">
                        {products.length} товаров
                      </span>
                    </div>
                  </button>
                );
              })}
            </div>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            className="bg-white rounded-xl border border-gray-200 overflow-hidden lg:col-span-2"
          >
            {activeMatch ? (
              <>
                <div className="bg-gradient-to-r from-green-50 to-emerald-50 border-b border-gray-200 px-5 py-3">
                  <div className="flex items-center justify-between">
                    <div>
                      <h3 className="font-semibold text-sm text-gray-900">{activeMatch.event.holiday_info}</h3>
                      <p className="text-xs text-gray-500 mt-0.5">
                        {formatRuDate(activeMatch.event.event_date)} · {activeMatch.event.category}
                      </p>
                    </div>
                    <span className="text-xs bg-white text-purple-700 px-2.5 py-1 rounded-full border border-purple-200 font-medium">
                      {activeMatch.products.length} {activeMatch.products.length === 1 ? 'товар' : 'товаров'}
                    </span>
                  </div>
                  {activeMatch.event.about && (
                    <p className="text-xs text-gray-500 mt-1.5 italic line-clamp-2">{activeMatch.event.about}</p>
                  )}
                </div>

                <div className="divide-y divide-gray-100">
                  {activeMatch.products.map((p, i) => (
                    <div
                      key={p.id}
                      className={`flex items-center justify-between px-5 py-3 text-sm transition-colors ${
                        selectedMatch?.product.id === p.id && selectedMatch?.event.id === activeMatch.event.id
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
              </>
            ) : (
              <div className="text-sm text-gray-500 text-center py-16">
                Выберите событие, чтобы увидеть товары
              </div>
            )}
          </motion.div>
        </div>
      )}

      {matches.length === 0 && !isLoading && (
        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="text-sm text-gray-500 text-center py-12 bg-white rounded-xl border border-gray-200">
          <Package className="w-10 h-10 text-gray-200 mx-auto mb-3" />
          <p className="font-medium text-gray-600">Результаты появятся здесь</p>
          <p className="text-xs text-gray-400 mt-1">После завершения векторного поиска для рекомендованных событий</p>
        </motion.div>
      )}
    </div>
  );
}
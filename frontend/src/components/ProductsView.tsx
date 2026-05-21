import { Package, Star, ThumbsUp, Sparkles, Store } from 'lucide-react';
import { motion } from 'framer-motion';
import type { MatchedProduct, EventItem, MatchInfo } from '../types';

interface ProductsViewProps {
  farmerProducts: MatchedProduct[];
  matchedProducts: MatchedProduct[];
  selectedMatch: {
    product: MatchedProduct;
    event: EventItem;
    match: MatchInfo;
  } | null;
  isLoading: boolean;
  onSelectProduct: (product: MatchedProduct) => void;
}

export function ProductsView({
  farmerProducts,
  matchedProducts,
  selectedMatch,
  isLoading,
  onSelectProduct,
}: ProductsViewProps) {
  const hasFarmerProducts = farmerProducts.length > 0;
  const hasMatchedResults = matchedProducts.length > 0;
  const matchedIds = new Set(matchedProducts.map((p) => p.product_id));

  return (
    <div className="p-6">
      <div className="mb-6">
        <h2 className="text-2xl font-semibold text-gray-900 mb-1">Товары</h2>
        <p className="text-sm text-gray-600">
          {isLoading
            ? 'Загрузка товаров...'
            : hasFarmerProducts
            ? `Каталог товаров фермера · ${farmerProducts.length} позиций`
            : 'Выберите фермера на вкладке «Профиль»'}
        </p>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <div className="w-8 h-8 border-2 border-green-600 border-t-transparent rounded-full animate-spin" />
          <span className="ml-3 text-sm text-gray-600">Загрузка товаров фермера…</span>
        </div>
      ) : (
        <>
          {hasFarmerProducts && (
            <>
              {hasMatchedResults && (
                <motion.div
                  initial={{ opacity: 0, y: -5 }}
                  animate={{ opacity: 1, y: 0 }}
                  className="bg-purple-50 border border-purple-200 rounded-xl p-3 mb-4 flex items-center gap-2"
                >
                  <Sparkles className="w-4 h-4 text-purple-600 shrink-0" />
                  <p className="text-xs text-purple-700">
                    <span className="font-medium">{matchedProducts.length}</span> товаров подходят к выбранным событиям — они отмечены иконкой <Sparkles className="w-3 h-3 inline text-purple-500" />
                  </p>
                </motion.div>
              )}

              <div className="grid grid-cols-3 gap-4 mb-6">
                {farmerProducts.map((p, index) => {
                  const isSelected = selectedMatch?.product.product_id === p.product_id;
                  const isMatched = matchedIds.has(p.product_id);
                  return (
                    <motion.div
                      key={p.product_id}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: Math.min(index * 0.03, 0.5) }}
                      onClick={() => onSelectProduct(p)}
                      className={`bg-white rounded-xl border p-4 cursor-pointer transition-all hover:shadow-md ${
                        isSelected
                          ? 'border-green-500 bg-green-50 shadow-md ring-2 ring-green-200'
                          : isMatched
                          ? 'border-purple-300 hover:border-purple-400'
                          : 'border-gray-200'
                      }`}
                    >
                      <div className="flex items-center justify-between mb-2">
                        <Package className="w-6 h-6 text-gray-300" />
                        <div className="flex items-center gap-1">
                          {isMatched && (
                            <span className="text-xs bg-purple-100 text-purple-700 px-1.5 py-0.5 rounded-full font-medium inline-flex items-center gap-0.5">
                              <Sparkles className="w-3 h-3" />
                              match
                            </span>
                          )}
                          {isSelected && <Star className="w-4 h-4 text-yellow-500 fill-current" />}
                        </div>
                      </div>
                      <h3 className="font-semibold text-gray-900 text-sm mb-1">{p.product_name}</h3>
                      <div className="flex items-center justify-between mt-2">
                        <p className="text-xl font-semibold text-gray-800">{p.price.toLocaleString()} ₽</p>
                        {p.unit && (
                          <span className="text-xs text-gray-400">за {p.unit}</span>
                        )}
                      </div>
                      <div className="flex items-center justify-between mt-1">
                        <p className="text-xs text-gray-500">{p.category}</p>
                        {p.quantity > 0 && (
                          <span className="text-xs text-gray-400">{p.quantity} шт</span>
                        )}
                      </div>
                      {isSelected && (
                        <div className="flex items-center gap-1 mt-2 text-green-600 text-xs">
                          <ThumbsUp className="w-3 h-3" />
                          <span>Выбрано для кампании</span>
                        </div>
                      )}
                    </motion.div>
                  );
                })}
              </div>
            </>
          )}

          {!hasFarmerProducts && !isLoading && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="text-center py-16 bg-white rounded-xl border border-gray-200"
            >
              <Store className="w-12 h-12 text-gray-200 mx-auto mb-4" />
              <p className="font-medium text-gray-600">Каталог товаров</p>
              <p className="text-sm text-gray-400 mt-1">
                Сначала выберите фермера на вкладке «Загрузка данных», чтобы увидеть его товары
              </p>
            </motion.div>
          )}

          {/* Match info */}
          {selectedMatch && (
            <motion.div
              initial={{ opacity: 0, y: -10 }}
              animate={{ opacity: 1, y: 0 }}
              className="bg-green-50 border border-green-200 rounded-xl p-4"
            >
              <div className="flex items-center gap-2 mb-2">
                <Star className="w-5 h-5 text-yellow-500" />
                <h3 className="font-semibold text-sm text-green-800">Выбрано для кампании</h3>
              </div>
              <p className="text-sm text-green-700">
                <span className="font-medium">{selectedMatch.product.product_name}</span>
                {' → '}
                {selectedMatch.event.holiday_info}
                {' ('}
                {new Date(selectedMatch.event.event_date).toLocaleDateString('ru-RU')}
                {')'}
                {selectedMatch.match && (
                  <span className="ml-2 text-xs">
                    · Relevance: {selectedMatch.match.score.toFixed(3)}
                  </span>
                )}
              </p>
              {selectedMatch.event.about && (
                <p className="text-xs text-green-600 mt-1">{selectedMatch.event.about}</p>
              )}
              <p className="text-xs text-green-600 mt-2">
                Нажмите «Создать кампанию» на вкладке Аналитика для генерации маркетингового плана
              </p>
            </motion.div>
          )}
        </>
      )}
    </div>
  );
}

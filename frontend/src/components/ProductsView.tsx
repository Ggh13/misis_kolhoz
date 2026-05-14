import { Package, TrendingUp, Star, ThumbsUp, AlertCircle } from 'lucide-react';
import { motion } from 'framer-motion';
import type { MatchedProduct, EventItem, MatchInfo } from '../types';

interface ProductsViewProps {
  products: MatchedProduct[];
  selectedMatch: {
    product: MatchedProduct;
    event: EventItem;
    match: MatchInfo;
  } | null;
  isLoading: boolean;
  onSelectProduct: (product: MatchedProduct) => void;
}

type DemoProduct = {
  name: string;
  price: number;
  category: string;
  trend: string;
};

const demoProducts: DemoProduct[] = [
  { name: 'Творог домашний 18%', price: 280, category: 'Молочные', trend: '+45%' },
  { name: 'Молоко безлактозное', price: 180, category: 'Молочные', trend: '+156%' },
  { name: 'Яйца куриные С0', price: 150, category: 'Яйца', trend: '+32%' },
  { name: 'Мёд натуральный', price: 890, category: 'Сладости', trend: '+78%' },
  { name: 'Куриное филе', price: 350, category: 'Мясо', trend: '+22%' },
  { name: 'Хлеб ржаной', price: 80, category: 'Хлеб', trend: '+15%' },
];

export function ProductsView({
  products,
  selectedMatch,
  isLoading,
  onSelectProduct,
}: ProductsViewProps) {
  const hasRealResults = products.length > 0;

  const handleClick = (product: MatchedProduct) => {
    if (!hasRealResults) return;
    onSelectProduct(product);
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h2 className="text-2xl font-semibold text-gray-900 mb-1">Товары</h2>
        <p className="text-sm text-gray-600">
          {isLoading
            ? 'Загрузка товаров...'
            : hasRealResults
            ? 'Ранжированные результаты по векторному поиску'
            : 'Каталог товаров (выберите событие для ML-анализа)'}
        </p>
      </div>

      {isLoading ? (
        <div className="flex items-center justify-center py-12">
          <div className="w-8 h-8 border-2 border-green-600 border-t-transparent rounded-full animate-spin" />
          <span className="ml-3 text-sm text-gray-600">Поиск подходящих товаров…</span>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-3 gap-4 mb-6">
            {hasRealResults
              ? products.map((p) => {
                  const isSelected = selectedMatch?.product.id === p.id;
                  return (
                    <motion.div
                      key={p.id}
                      initial={{ opacity: 0, y: 20 }}
                      animate={{ opacity: 1, y: 0 }}
                      transition={{ delay: 0.05 }}
                      onClick={() => handleClick(p)}
                      className={`bg-white rounded-xl border border-gray-200 p-4 cursor-pointer transition-all hover:shadow-md ${
                        isSelected ? 'border-green-500 bg-green-50 shadow-md ring-2 ring-green-200' : ''
                      }`}
                    >
                      <div className="flex items-center justify-between mb-2">
                        <Package className="w-6 h-6 text-gray-300" />
                        {isSelected && <Star className="w-4 h-4 text-yellow-500 fill-current" />}
                      </div>
                      <h3 className="font-semibold text-gray-900 text-sm mb-1">{p.product_name}</h3>
                      <div className="flex items-center justify-between mt-2">
                        <p className="text-xl font-semibold text-gray-800">{p.price.toLocaleString()} ₽</p>
                        <span className="inline-flex items-center gap-1 text-green-700 text-sm">
                          <TrendingUp className="w-3 h-3" />
                          +12%
                        </span>
                      </div>
                      {p.category && <p className="text-xs text-gray-500 mt-1">{p.category}</p>}
                      {isSelected && (
                        <div className="flex items-center gap-1 mt-2 text-green-600 text-xs">
                          <ThumbsUp className="w-3 h-3" />
                          <span>Выбрано для кампании</span>
                        </div>
                      )}
                    </motion.div>
                  );
                })
              : demoProducts.map((p, index) => (
                  <motion.div
                    key={p.name}
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ delay: index * 0.05 }}
                    className="bg-white rounded-xl border border-gray-200 p-4 cursor-default"
                  >
                    <div className="flex items-center justify-between mb-2">
                      <Package className="w-6 h-6 text-gray-300" />
                    </div>
                    <h3 className="font-semibold text-gray-900 text-sm mb-1">{p.name}</h3>
                    <div className="flex items-center justify-between mt-2">
                      <p className="text-xl font-semibold text-gray-800">{p.price.toLocaleString()} ₽</p>
                      <span className="inline-flex items-center gap-1 text-green-700 text-sm">
                        <TrendingUp className="w-3 h-3" />
                        {p.trend}
                      </span>
                    </div>
                    {p.category && <p className="text-xs text-gray-500 mt-1">{p.category}</p>}
                  </motion.div>
                ))}
          </div>

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

          {!hasRealResults && !isLoading && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="bg-amber-50 border border-amber-200 rounded-xl p-4 mt-4"
            >
              <div className="flex items-center gap-2 text-amber-800 text-sm">
                <AlertCircle className="w-4 h-4" />
                <span>
                  Для получения реальных рекомендаций выберите событие на вкладке «События и тренды».
                  Загруженные данные будут использоваться для ML-анализа.
                </span>
              </div>
            </motion.div>
          )}
        </>
      )}
    </div>
  );
}
import { Search, User, X, MapPin, Store, Package } from 'lucide-react';
import { useState, useRef, useCallback, useEffect } from 'react';
import type { FarmerInfo, MatchedProduct } from '../types';

interface DataIngestionProps {
  selectedFarmer: FarmerInfo | null;
  onSelectFarmer: (farmer: FarmerInfo | null) => void;
  farmerProducts: MatchedProduct[];
}

interface FarmerSearchResult {
  id: number;
  name: string;
  region: string;
}

export function Profile({ selectedFarmer, onSelectFarmer, farmerProducts }: DataIngestionProps) {
  const [farmerInput, setFarmerInput] = useState('');
  const [searchResults, setSearchResults] = useState<FarmerSearchResult[]>([]);
  const [searchLoading, setSearchLoading] = useState(false);
  const [showDropdown, setShowDropdown] = useState(false);
  const [farmerDetails, setFarmerDetails] = useState<Record<string, unknown> | null>(null);
  const searchRef = useRef<HTMLDivElement>(null);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const doSearch = useCallback(async (q: string) => {
    if (q.trim().length < 2) {
      setSearchResults([]);
      setShowDropdown(false);
      return;
    }
    setSearchLoading(true);
    try {
      const res = await fetch(`/farmers/search?q=${encodeURIComponent(q.trim())}`);
      if (!res.ok) { setSearchResults([]); return; }
      const data = await res.json();
      setSearchResults(Array.isArray(data) ? data : []);
      setShowDropdown(true);
    } catch {
      setSearchResults([]);
    } finally {
      setSearchLoading(false);
    }
  }, []);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setFarmerInput(val);
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => doSearch(val), 300);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      if (timerRef.current) clearTimeout(timerRef.current);
      doSearch(farmerInput);
    }
  };

  const selectFarmer = async (result: FarmerSearchResult) => {
    setShowDropdown(false);
    setFarmerInput(result.name);
    onSelectFarmer({ id: result.id, name: result.name, region: result.region });
    setFarmerDetails(null);
    try {
      const res = await fetch(`/farmer_data/${result.id}`);
      const data = await res.json();
      if (data?.farmer) setFarmerDetails(data.farmer);
    } catch { /**/ }
  };

  const handleClear = () => {
    onSelectFarmer(null);
    setFarmerDetails(null);
    setFarmerInput('');
    setSearchResults([]);
    setShowDropdown(false);
  };

  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (searchRef.current && !searchRef.current.contains(e.target as Node)) {
        setShowDropdown(false);
      }
    };
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, []);

  const categories = Array.from(new Set(farmerProducts.map((p) => p.category).filter(Boolean)));
  const totalProducts = farmerProducts.length;

  return (
    <div className="p-6">
      <div className="mb-6">
        <h2 className="text-2xl font-bold text-gray-900 mb-1">Ваш профиль</h2>
        <p className="text-sm text-gray-600">Введите название вашего хозяйства, чтобы начать работу с платформой</p>
      </div>

      <div className="bg-white rounded-xl border border-gray-200 p-6 mb-6" ref={searchRef}>
        <div className="flex items-center gap-2 mb-3">
          <User className="w-5 h-5 text-green-600" />
          <h3 className="font-semibold text-sm text-gray-900">Поиск вашего хозяйства</h3>
        </div>

        {selectedFarmer ? (
          <div className="flex items-center justify-between bg-green-50 border border-green-200 rounded-lg p-4">
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 bg-green-100 rounded-full flex items-center justify-center">
                <User className="w-5 h-5 text-green-700" />
              </div>
              <div>
                <p className="font-semibold text-sm text-green-800">{selectedFarmer.name || `Хозяйство #${selectedFarmer.id}`}</p>
                {selectedFarmer.region && (
                  <p className="text-xs text-green-600">{selectedFarmer.region}</p>
                )}
              </div>
            </div>
            <button onClick={handleClear} className="text-xs text-red-600 hover:text-red-800 flex items-center gap-1">
              <X className="w-3 h-3" /> Выйти
            </button>
          </div>
        ) : (
            <div className="relative">
                <div className="flex gap-2">
                  <div className="relative flex-1">
                      <input
                      type="text"
                      placeholder="Введите название вашего фермерского хозяйства..."
                      value={farmerInput}
                      onChange={handleInputChange}
                      onKeyDown={handleKeyDown}
                      onFocus={() => { if (searchResults.length > 0) setShowDropdown(true); }}
                      className="w-full border border-gray-300 rounded-lg px-3 py-2 pr-8 text-sm focus:outline-none focus:ring-2 focus:ring-green-500 focus:border-green-500"
                    />
                    {searchLoading && (
                      <div className="absolute right-3 top-1/2 -translate-y-1/2">
                        <div className="w-4 h-4 border-2 border-green-600 border-t-transparent rounded-full animate-spin" />
                      </div>
                    )}
                  </div>
                  <button
                    onClick={() => doSearch(farmerInput)}
                    disabled={searchLoading || farmerInput.trim().length < 2}
                    className="bg-green-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-green-700 transition-colors disabled:opacity-50 inline-flex items-center gap-2 shrink-0"
                  >
                    {searchLoading ? (
                      <><div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" /></>
                    ) : (
                      <><Search className="w-4 h-4" /> Найти</>
                    )}
                  </button>
                </div>

            {showDropdown && searchResults.length > 0 && (
              <div className="absolute z-20 mt-1 w-full bg-white border border-gray-200 rounded-lg shadow-lg max-h-60 overflow-y-auto">
                {searchResults.map((r) => (
                  <button
                    key={r.id}
                    onClick={() => selectFarmer(r)}
                    className="w-full text-left px-4 py-2.5 text-sm hover:bg-green-50 border-b border-gray-50 last:border-b-0 flex items-center justify-between"
                  >
                    <div>
                      <span className="font-medium text-gray-900">{r.name}</span>
                      {r.region && <span className="text-xs text-gray-400 ml-2">{r.region}</span>}
                    </div>
                    <span className="text-xs text-gray-400">ID {r.id}</span>
                  </button>
                ))}
              </div>
            )}

            {showDropdown && searchResults.length === 0 && farmerInput.trim().length >= 2 && !searchLoading && (
              <div className="absolute z-20 mt-1 w-full bg-white border border-gray-200 rounded-lg shadow-lg p-4 text-sm text-gray-500 text-center">
                Ничего не найдено
              </div>
            )}
          </div>
        )}
      </div>

      {selectedFarmer && (
        <div className="space-y-6">
          <div className="grid grid-cols-4 gap-4">
            <div className="bg-gradient-to-br from-green-500 to-green-700 rounded-xl p-5 text-white">
              <Package className="w-6 h-6 mb-3 opacity-80" />
              <p className="text-3xl font-semibold">{totalProducts}</p>
              <p className="text-xs text-green-100 mt-1">Товаров в каталоге</p>
            </div>
            <div className="bg-gradient-to-br from-blue-500 to-blue-700 rounded-xl p-5 text-white">
              <Store className="w-6 h-6 mb-3 opacity-80" />
              <p className="text-3xl font-semibold">{categories.length}</p>
              <p className="text-xs text-blue-100 mt-1">Категорий товаров</p>
            </div>
            <div className="bg-gradient-to-br from-purple-500 to-purple-700 rounded-xl p-5 text-white">
              <MapPin className="w-6 h-6 mb-3 opacity-80" />
              <p className="text-sm font-semibold">{selectedFarmer.region || '—'}</p>
              <p className="text-xs text-purple-100 mt-1">Регион</p>
            </div>
            <div className="bg-gradient-to-br from-amber-500 to-amber-700 rounded-xl p-5 text-white">
              <User className="w-6 h-6 mb-3 opacity-80" />
              <p className="text-sm font-semibold truncate">{selectedFarmer.name || `ID ${selectedFarmer.id}`}</p>
              <p className="text-xs text-amber-100 mt-1">Хозяйство</p>
            </div>
          </div>

          {farmerDetails && (
            <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
              <div className="px-5 py-4 border-b border-gray-100">
                <h3 className="font-semibold text-sm text-gray-900">Информация о хозяйстве</h3>
              </div>
              <div className="divide-y divide-gray-50">
                {[
                  { label: 'Название', key: 'name' },
                  { label: 'Описание', key: 'farmer_description' },
                  { label: 'Регион', key: 'region' },
                ].map(({ label, key }) => {
                  const val = farmerDetails[key];
                  if (!val) return null;
                  return (
                    <div key={key} className="flex items-center px-5 py-3">
                      <span className="text-xs text-gray-500 w-32 shrink-0">{label}</span>
                      <span className="text-sm text-gray-900">{String(val)}</span>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {categories.length > 0 && (
            <div className="bg-white rounded-xl border border-gray-200 p-5">
              <div className="flex items-center gap-2 mb-3">
                <Package className="w-4 h-4 text-green-600" />
                <h3 className="font-semibold text-sm text-gray-900">Категории товаров</h3>
              </div>
              <div className="flex flex-wrap gap-2">
                {categories.map((cat) => {
                  const count = farmerProducts.filter((p) => p.category === cat).length;
                  return (
                    <div key={cat} className="bg-gray-50 border border-gray-200 rounded-lg px-3 py-2 text-sm flex items-center gap-2">
                      <span className="text-gray-700">{cat}</span>
                      <span className="text-xs bg-gray-200 text-gray-600 px-1.5 py-0.5 rounded-full">{count}</span>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
            <div className="px-5 py-4 border-b border-gray-100 flex items-center justify-between">
              <h3 className="font-semibold text-sm text-gray-900">Товары</h3>
              <span className="text-xs text-gray-500">{totalProducts} позиций</span>
            </div>
            <div className="divide-y divide-gray-50 max-h-64 overflow-y-auto">
              {farmerProducts.slice(0, 50).map((p) => (
                <div key={p.product_id} className="flex items-center justify-between px-5 py-2.5 hover:bg-gray-50">
                  <div className="flex items-center gap-3 min-w-0">
                    <Package className="w-4 h-4 text-gray-300 shrink-0" />
                    <span className="text-sm text-gray-900 truncate">{p.product_name}</span>
                  </div>
                  <div className="flex items-center gap-3 shrink-0">
                    <span className="text-xs text-gray-400">{p.category}</span>
                    <span className="text-sm font-medium text-gray-800">{p.price.toLocaleString()} ₽</span>
                  </div>
                </div>
              ))}
              {totalProducts > 50 && (
                <div className="px-5 py-3 text-xs text-gray-400 text-center">
                  + ещё {totalProducts - 50} товаров
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

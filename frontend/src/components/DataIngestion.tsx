import { Upload, CheckCircle, Clock, FileSpreadsheet } from 'lucide-react';
import { useState } from 'react';
import axios from 'axios';

export function DataIngestion() {
  const [isLoading, setIsLoading] = useState<string | null>(null);
  const [message, setMessage] = useState('');
  const rows = [
    { name: 'Творог домашний 18%', price: '280 ₽', category: 'Молочные', stock: 45, status: 'synced' },
    { name: 'Яйца куриные С0', price: '150 ₽', category: 'Яйца', stock: 200, status: 'synced' },
    { name: 'Шашлык из свиной шеи', price: '650 ₽', category: 'Мясо', stock: 15, status: 'pending' },
  ];
  const handleUpload = async (endpoint: '/upload_data' | '/load_orders' | '/upload_events') => {
    setIsLoading(endpoint);
    setMessage('');
    try {
      const response = await axios.post(endpoint);
      setMessage(typeof response.data === 'string' ? response.data : 'Загрузка завершена');
    } catch (error) {
      if (axios.isAxiosError(error)) {
        setMessage(`Ошибка: ${typeof error.response?.data === 'string' ? error.response.data : error.message}`);
      } else {
        setMessage('Ошибка: неизвестная ошибка');
      }
    } finally {
      setIsLoading(null);
    }
  };

  return <div className="p-6"><div className="mb-6"><h2 className="text-2xl font-bold text-[#1e3a2b] mb-1">Загрузка данных</h2><p className="text-sm text-gray-600">Загрузка таблиц в backend</p></div><div className="bg-white rounded-xl border border-gray-200 p-8 mb-6 text-center"><div className="w-20 h-20 bg-green-600 rounded-lg flex items-center justify-center mb-6 mx-auto"><FileSpreadsheet className="w-10 h-10 text-white" /></div><div className="flex flex-wrap items-center justify-center gap-3"><button onClick={() => handleUpload('/upload_data')} disabled={isLoading !== null} className="bg-green-600 text-white px-6 py-2.5 rounded-lg font-medium text-sm hover:bg-green-700 transition-colors disabled:opacity-50 inline-flex items-center gap-2">{isLoading === '/upload_data' ? <> <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />Загрузка...</> : <><Upload className="w-4 h-4" />Загрузить фермеров/товары</>}</button><button onClick={() => handleUpload('/load_orders')} disabled={isLoading !== null} className="bg-blue-600 text-white px-6 py-2.5 rounded-lg font-medium text-sm hover:bg-blue-700 transition-colors disabled:opacity-50 inline-flex items-center gap-2">{isLoading === '/load_orders' ? 'Загрузка...' : 'Загрузить заказы'}</button><button onClick={() => handleUpload('/upload_events')} disabled={isLoading !== null} className="bg-purple-600 text-white px-6 py-2.5 rounded-lg font-medium text-sm hover:bg-purple-700 transition-colors disabled:opacity-50 inline-flex items-center gap-2">{isLoading === '/upload_events' ? 'Загрузка...' : 'Загрузить события'}</button></div>{message && <p className={`mt-4 text-sm ${message.startsWith('Ошибка') ? 'text-red-600' : 'text-green-700'}`}>{message}</p>}</div><div className="bg-white rounded-xl border border-gray-200 p-5"><table className="w-full"><thead><tr className="border-b border-gray-200"><th className="text-left py-3 px-4 text-xs">Название</th><th className="text-left py-3 px-4 text-xs">Цена</th><th className="text-left py-3 px-4 text-xs">Категория</th><th className="text-left py-3 px-4 text-xs">Остаток</th><th className="text-left py-3 px-4 text-xs">Статус</th></tr></thead><tbody>{rows.map((r) => <tr key={r.name} className="border-b border-gray-100"><td className="py-3 px-4 text-sm font-medium">{r.name}</td><td className="py-3 px-4 text-sm">{r.price}</td><td className="py-3 px-4 text-sm">{r.category}</td><td className="py-3 px-4 text-sm">{r.stock}</td><td className="py-3 px-4 text-sm">{r.status === 'synced' ? <span className="inline-flex items-center gap-1 text-green-600"><CheckCircle className="w-4 h-4"/>Синхр.</span> : <span className="inline-flex items-center gap-1 text-amber-600"><Clock className="w-4 h-4"/>Ожидание</span>}</td></tr>)}</tbody></table></div></div>;
}

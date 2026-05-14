import { useState } from 'react';
import axios from 'axios';

export function MlServices() {
  const [loading, setLoading] = useState<string | null>(null);
  const [result, setResult] = useState('');
  const [error, setError] = useState('');

  const callApi = async (key: string, callback: () => Promise<unknown>) => {
    setLoading(key);
    setError('');
    setResult('');
    try {
      const data = await callback();
      setResult(JSON.stringify(data, null, 2));
    } catch (e) {
      if (axios.isAxiosError(e)) {
        setError(typeof e.response?.data === 'string' ? e.response.data : JSON.stringify(e.response?.data ?? e.message, null, 2));
      } else {
        setError('Неизвестная ошибка');
      }
    } finally {
      setLoading(null);
    }
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h2 className="text-2xl font-semibold text-gray-900 mb-1">ML сервисы</h2>
        <p className="text-sm text-gray-600">Проверка реальных вызовов /ml и /agents</p>
      </div>

      <div className="grid grid-cols-2 gap-4 mb-4">
        <div className="bg-white rounded-xl border border-gray-200 p-5 space-y-3">
          <h3 className="font-semibold text-sm">🔬 Embed</h3>
          <button className="w-full bg-green-600 text-white py-2 rounded-lg text-sm disabled:opacity-50" disabled={loading !== null} onClick={() => callApi('extract', async () => (await axios.post('/ml/extract', { product_description: 'Творог домашний', farmer_description: 'Ферма в Туле' })).data)}>
            {loading === 'extract' ? '⏳ Загрузка...' : 'POST /ml/extract'}
          </button>
          <button className="w-full bg-blue-600 text-white py-2 rounded-lg text-sm disabled:opacity-50" disabled={loading !== null} onClick={() => callApi('embed-product', async () => (await axios.post('/ml/embed/product', { product_features: { product_type: 'молочные', geo: 'Тула' }, farmer_features: { location: ['Тула'] } })).data)}>
            {loading === 'embed-product' ? '⏳ Загрузка...' : 'POST /ml/embed/product'}
          </button>
          <button className="w-full bg-purple-600 text-white py-2 rounded-lg text-sm disabled:opacity-50" disabled={loading !== null} onClick={() => callApi('embed-event', async () => (await axios.post('/ml/embed/event', { event_name: 'Пасха', event_description: 'Весенний праздник' })).data)}>
            {loading === 'embed-event' ? '⏳ Загрузка...' : 'POST /ml/embed/event'}
          </button>
        </div>

        <div className="bg-white rounded-xl border border-gray-200 p-5 space-y-3">
          <h3 className="font-semibold text-sm">⚙️ Pipeline</h3>
          <button className="w-full bg-emerald-600 text-white py-2 rounded-lg text-sm disabled:opacity-50" disabled={loading !== null} onClick={() => callApi('pipeline', async () => (await axios.post('/ml/pipeline/full', { product_description: 'Творог', farmer_description: 'Ферма', event_name: 'Пасха', event_description: 'Праздник' })).data)}>
            {loading === 'pipeline' ? '⏳ Загрузка...' : 'POST /ml/pipeline/full'}
          </button>
          <p className="text-xs text-gray-400">Запускает этапы: extract → embed → search → match</p>
        </div>
      </div>

      <div className="bg-white rounded-xl border border-gray-200 p-5 space-y-3">
        <h3 className="font-semibold text-sm">🤖 Marketing Agents</h3>
        <button className="w-full bg-amber-600 text-white py-2 rounded-lg text-sm disabled:opacity-50" disabled={loading !== null} onClick={() => callApi('agents-dynamic', async () => (await axios.post('/agents/run_dynamic', { product_id: 1, date_from: '2026-05-01', date_to: '2026-06-01', top_k: 1 })).data)}>
          {loading === 'agents-dynamic' ? '⏳ Генерация...' : 'POST /agents/run_dynamic'}
        </button>
        <p className="text-xs text-gray-400">Статический запрос: product_id=1, май 2026</p>
      </div>

      {error && <pre className="mt-4 bg-red-50 border border-red-200 rounded-lg p-3 text-xs text-red-700 overflow-x-auto">{error}</pre>}
      {result && <pre className="mt-4 bg-gray-900 text-green-300 rounded-lg p-4 text-xs overflow-x-auto">{result}</pre>}
    </div>
  );
}
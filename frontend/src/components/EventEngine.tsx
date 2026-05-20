import { Calendar, TrendingUp, Bell, CheckCircle, PlusCircle, X, Search, Sparkles, Package } from 'lucide-react';
import axios from 'axios';
import { useEffect, useMemo, useState } from 'react';
import type { EventItem, EventSearchMatch } from '../types';

function formatRuDate(value: string) {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return new Intl.DateTimeFormat('ru-RU', {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
  }).format(parsed);
}

function monthLabel(year: number, month: number) {
  const d = new Date(year, month - 1);
  return new Intl.DateTimeFormat('ru-RU', { month: 'long', year: 'numeric' }).format(d);
}

function getEventMonthKey(e: EventItem) {
  const d = new Date(e.event_date);
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
}

interface EventEngineProps {
  selectedEventIds: number[];
  onToggleEvent: (event: EventItem) => void;
  recommendedEvents: EventSearchMatch[];
  onFindEvents: () => void;
  eventsLoading: boolean;
  eventsError: string;
  farmerName: string | null;
}

export function EventEngine({
  selectedEventIds,
  onToggleEvent,
  recommendedEvents,
  onFindEvents,
  eventsLoading,
  eventsError,
  farmerName,
}: EventEngineProps) {
  const maxRecommendedShown = 12;
  const visibleRecommendedEvents = recommendedEvents.slice(0, maxRecommendedShown);

  const [allEvents, setAllEvents] = useState<EventItem[]>([]);
  const [upcoming, setUpcoming] = useState<EventItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      setError('');
      try {
        const [allRes, upcomingRes] = await Promise.all([
          axios.get<{ events: EventItem[] }>('/events'),
          axios.get<{ events: EventItem[] }>('/events/upcoming?days=90'),
        ]);
        const today = new Date();
        today.setHours(0, 0, 0, 0);
        setAllEvents((allRes.data.events ?? []).filter((e) => new Date(e.event_date) >= today));
        setUpcoming(upcomingRes.data.events ?? []);
      } catch (e) {
        if (axios.isAxiosError(e)) setError(e.message);
        else setError('Не удалось загрузить события');
      } finally {
        setLoading(false);
      }
    };
    void load();
  }, []);

  const recommendedEventIds = useMemo(
    () => new Set(recommendedEvents.map((e) => e.id)),
    [recommendedEvents],
  );

  const categoryCount = useMemo(
    () => new Set(allEvents.map((item) => item.category)).size,
    [allEvents],
  );

  const grouped = useMemo(() => {
    const map = new Map<string, EventItem[]>();
    for (const e of allEvents) {
      const key = getEventMonthKey(e);
      if (!map.has(key)) map.set(key, []);
      map.get(key)!.push(e);
    }
    return Array.from(map.entries()).sort(([a], [b]) => a.localeCompare(b));
  }, [allEvents]);

  const isSelected = (id: number) => selectedEventIds.includes(id);

  const renderCard = (e: EventItem, highlight?: boolean) => {
    const selected = isSelected(e.id);
    return (
      <div
        key={e.id}
        onClick={() => onToggleEvent(e)}
        className={`border rounded-lg cursor-pointer transition-all hover:shadow-sm ${
          selected
            ? 'border-green-500 bg-green-50 shadow-sm ring-1 ring-green-300'
            : highlight
            ? 'border-purple-300 bg-purple-50/40 hover:border-purple-400'
            : 'border-gray-200 hover:border-green-300'
        }`}
      >
        <div className="p-2.5">
          <div className="flex justify-between gap-1 items-start">
            <p className={`font-medium text-xs flex-1 ${selected ? 'text-green-800' : highlight ? 'text-purple-800' : 'text-gray-800'}`}>
              {e.holiday_info}
            </p>
            <div className="flex items-center gap-1 shrink-0">
              <span className="text-[10px] font-semibold text-green-600">{e.category}</span>
              {highlight && !selected && <Sparkles className="w-3 h-3 text-purple-500" />}
              {selected ? (
                <CheckCircle className="w-3.5 h-3.5 text-green-600" />
              ) : (
                <PlusCircle className="w-3.5 h-3.5 text-gray-300" />
              )}
            </div>
          </div>
          <p className="text-[11px] text-gray-500 mt-0.5">{formatRuDate(e.event_date)}</p>
        </div>
      </div>
    );
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h2 className="text-2xl font-semibold text-gray-900 mb-1">События и тренды</h2>
        <p className="text-sm text-gray-600">
          Нажимайте на события для анализа (можно выбрать несколько)
          {selectedEventIds.length > 0 && (
            <span className="ml-2 text-green-600 font-medium">
              · выбрано {selectedEventIds.length}
            </span>
          )}
        </p>
      </div>

      <div className="grid grid-cols-3 gap-4 mb-6">
        <div className="bg-green-600 rounded-xl p-5 text-white">
          <Calendar className="w-6 h-6 mb-3" />
          <p className="text-3xl font-semibold">{allEvents.length}</p>
          <p className="text-xs text-green-100">Всего событий</p>
        </div>
        <div className="bg-blue-600 rounded-xl p-5 text-white">
          <TrendingUp className="w-6 h-6 mb-3" />
          <p className="text-3xl font-semibold">{categoryCount}</p>
          <p className="text-xs text-blue-100">Категорий</p>
        </div>
        <div className="bg-purple-600 rounded-xl p-5 text-white">
          <Bell className="w-6 h-6 mb-3" />
          <p className="text-3xl font-semibold">{upcoming.length}</p>
          <p className="text-xs text-purple-100">Ближайшие 90 дней</p>
        </div>
      </div>

      {farmerName && (
        <div className="bg-gradient-to-r from-purple-600 to-indigo-600 rounded-xl p-5 text-white mb-6">
          <div className="flex items-center justify-between">
            <div>
              <div className="flex items-center gap-2 mb-1">
                <Sparkles className="w-5 h-5" />
                <h3 className="font-semibold">Подбор событий под товары</h3>
              </div>
              <p className="text-xs text-purple-100">
                {farmerName} · {recommendedEvents.length > 0
                  ? `Найдено ${recommendedEvents.length} подходящих событий`
                  : 'Найдите события, которые подходят к вашим товарам'}
              </p>
            </div>
            <button
              onClick={onFindEvents}
              disabled={eventsLoading}
              className="bg-white text-purple-700 px-4 py-2 rounded-lg text-sm font-medium hover:bg-purple-50 transition-colors disabled:opacity-50 inline-flex items-center gap-2 shrink-0"
            >
              {eventsLoading ? (
                <><div className="w-4 h-4 border-2 border-purple-600 border-t-transparent rounded-full animate-spin" /> Поиск...</>
              ) : (
                <><Package className="w-4 h-4" /> Подобрать события</>
              )}
            </button>
          </div>
        </div>
      )}

      {eventsError && (
        <div className="mb-4 text-sm text-red-600 bg-red-50 border border-red-200 rounded-lg p-3">
          {eventsError}
        </div>
      )}

      {recommendedEvents.length > 0 && (
        <div className="mb-6 bg-white rounded-xl border border-purple-200 p-5">
          <div className="flex items-center gap-2 mb-3">
            <Sparkles className="w-4 h-4 text-purple-600" />
            <h3 className="font-semibold text-sm">
              Рекомендованные события ({visibleRecommendedEvents.length})
            </h3>
          </div>
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-2">
            {visibleRecommendedEvents.map((e) => {
              const selected = isSelected(e.id);
              const ev = allEvents.find((a) => a.id === e.id);
              const display: EventItem = ev ?? {
                id: e.id,
                event_date: e.event_date,
                holiday_info: e.holiday_info,
                category: e.category,
                about: e.about,
                food_customs: e.food_customs,
              };
              return (
                <div
                  key={e.id}
                  onClick={() => onToggleEvent(display)}
                  className={`border rounded-lg cursor-pointer transition-all p-2.5 ${
                    selected
                      ? 'border-green-500 bg-green-50 ring-1 ring-green-300'
                      : 'border-purple-200 bg-purple-50/30 hover:border-purple-400'
                  }`}
                >
                  <div className="flex justify-between items-start gap-1">
                    <p className={`font-medium text-xs flex-1 ${selected ? 'text-green-800' : 'text-gray-800'}`}>
                      {e.holiday_info}
                    </p>
                    {selected ? (
                      <CheckCircle className="w-3.5 h-3.5 text-green-600 shrink-0" />
                    ) : (
                      <PlusCircle className="w-3.5 h-3.5 text-gray-300 shrink-0" />
                    )}
                  </div>
                  <p className="text-[11px] text-gray-500 mt-0.5">{formatRuDate(e.event_date)}</p>
                  <div className="flex items-center justify-between mt-1">
                    <span className="text-[10px] text-purple-600 font-medium">{e.category}</span>
                    <span className="text-[10px] text-gray-400">
                      {e.distance < 0.01 ? '∞' : (1 - e.distance).toFixed(2)}
                    </span>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {selectedEventIds.length > 0 && (
        <div className="mb-4 bg-white rounded-xl border border-green-200 p-4">
          <div className="flex items-center gap-2 mb-2">
            <Search className="w-4 h-4 text-green-600" />
            <h3 className="font-semibold text-sm text-green-800">
              Выбранные события ({selectedEventIds.length})
            </h3>
          </div>
          <div className="flex flex-wrap gap-2">
            {selectedEventIds.map((id) => {
              const ev = allEvents.find((e) => e.id === id);
              if (!ev) return null;
              return (
                <div
                  key={ev.id}
                  onClick={() => onToggleEvent(ev)}
                  className="inline-flex items-center gap-1.5 bg-green-50 border border-green-300 text-green-800 text-xs rounded-full px-3 py-1.5 cursor-pointer hover:bg-green-100 transition-colors"
                >
                  <CheckCircle className="w-3 h-3" />
                  <span className="font-medium">{ev.holiday_info}</span>
                  <span className="text-green-500">·</span>
                  <span className="text-green-600">{formatRuDate(ev.event_date)}</span>
                  <X className="w-3 h-3 text-green-400 hover:text-green-700 ml-0.5" />
                </div>
              );
            })}
          </div>
        </div>
      )}

      {error && <div className="mb-4 text-sm text-red-600 bg-red-50 border border-red-200 rounded-lg p-3">{error}</div>}

      {loading ? (
        <div className="flex items-center gap-2 text-sm text-gray-600 py-8">
          <div className="w-4 h-4 border-2 border-green-600 border-t-transparent rounded-full animate-spin" />
          Загрузка событий...
        </div>
      ) : (
        <div className="grid grid-cols-2 gap-4">
          <div className="bg-white rounded-xl border border-gray-200 p-5 overflow-y-auto max-h-[70vh]">
            <h3 className="font-semibold text-sm mb-2 sticky top-0 bg-white pb-2 z-10">
              📅 Все события ({allEvents.length})
            </h3>
            {allEvents.length === 0 ? (
              <p className="text-sm text-gray-500">Событий не найдено</p>
            ) : (
              grouped.map(([key, events]) => {
                const [yStr, mStr] = key.split('-');
                const y = Number(yStr);
                const m = Number(mStr);
                return (
                  <div key={key}>
                    <p className="text-[11px] font-medium text-gray-400 uppercase tracking-wide mb-1.5">
                      {monthLabel(y, m)}
                    </p>
                    <div className="space-y-1 mb-3">
                      {events.map((e) => renderCard(e, recommendedEventIds.has(e.id)))}
                    </div>
                  </div>
                );
              })
            )}
          </div>
          <div className="bg-white rounded-xl border border-gray-200 p-5 overflow-y-auto max-h-[70vh]">
            <h3 className="font-semibold text-sm mb-2 sticky top-0 bg-white pb-2 z-10">
              ⏰ Ближайшие (90 дней)
            </h3>
            {upcoming.length === 0 ? (
              <p className="text-sm text-gray-500">В ближайшие 90 дней событий не найдено</p>
            ) : (
              <div className="space-y-1">
                {upcoming.map((e) => renderCard(e, recommendedEventIds.has(e.id)))}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

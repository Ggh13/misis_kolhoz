import { useState, useCallback, useMemo, useEffect } from 'react';
import { Sidebar } from './components/Sidebar';
import { Dashboard } from './components/Dashboard';
import { Profile } from './components/Profile';
import { EventEngine } from './components/EventEngine';
import { SemanticMatcher } from './components/SemanticMatcher';
import { ProductsView } from './components/ProductsView';
import { Analytics } from './components/Analytics';
import { ParticleBackground } from './components/ParticleBackground';
import type { EventItem, MatchedProduct, FarmerInfo, EventSearchMatch, EventProductsMatch, ProductEventsMatch, WorkflowState } from './types';

const RECOMMENDED_EVENT_SCORE_THRESHOLD = 0.85;
const EVENT_PRODUCT_SCORE_THRESHOLD = 0.85;
const EVENT_PRODUCT_MATCH_LIMIT = 60;

export default function App() {
  const [activeTab, setActiveTab] = useState('ingestion');
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [selectedEventIds, setSelectedEventIds] = useState<number[]>([]);
  const [selectedEventsById, setSelectedEventsById] = useState<Record<number, EventItem>>({});
  const [eventResults, setEventResults] = useState<Record<number, MatchedProduct[]>>({});
  const [isLoading, setIsLoading] = useState(false);
  const [selectedMatch, setSelectedMatch] = useState<WorkflowState['selectedMatch']>(null);
  const [campaignResult, setCampaignResult] = useState<WorkflowState['campaignResult']>(null);
  const [selectedFarmer, setSelectedFarmer] = useState<FarmerInfo | null>(null);
  const [farmerProducts, setFarmerProducts] = useState<MatchedProduct[]>([]);
  const [farmerProductsLoading, setFarmerProductsLoading] = useState(false);
  const [recommendedEvents, setRecommendedEvents] = useState<EventSearchMatch[]>([]);
  const [eventsLoading, setEventsLoading] = useState(false);
  const [matcherLoading, setMatcherLoading] = useState(false);
  const [eventMatches, setEventMatches] = useState<EventProductsMatch[]>([]);
  const [imageExample, setImageExample] = useState<{ url: string; prompt: string } | null>(null);
  const [imageLoading, setImageLoading] = useState(false);
  const [imageError, setImageError] = useState('');

  const [eventsError, setEventsError] = useState('');

  const handleFindEvents = useCallback(async () => {
    if (!selectedFarmer) return;

    setEventsLoading(true);
    setEventsError('');
    try {
      const res = await fetch('/vector/match/products-to-events', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ limit: 200, farmer_id: selectedFarmer.id, future_only: true }),
      });
      if (!res.ok) {
        throw new Error('match request failed');
      }
      const data = await res.json();
      const productMatches: ProductEventsMatch[] = data.products ?? [];

      const eventMap = new Map<number, EventSearchMatch>();
      for (const match of productMatches) {
        for (const ev of match.events ?? []) {
          if (!eventMap.has(ev.id) || ev.distance < eventMap.get(ev.id)!.distance) {
            eventMap.set(ev.id, ev);
          }
        }
      }

      const sortedEvents = Array.from(eventMap.values()).sort((a, b) => {
        const dateA = new Date(a.event_date).getTime();
        const dateB = new Date(b.event_date).getTime();
        if (!Number.isNaN(dateA) && !Number.isNaN(dateB)) return dateA - dateB;
        return a.event_date.localeCompare(b.event_date);
      });
      setRecommendedEvents(
        sortedEvents.filter((e) => 1 - e.distance >= RECOMMENDED_EVENT_SCORE_THRESHOLD),
      );
    } catch (err) {
      console.error('Failed to find events:', err);
      setEventsError(`Не удалось подобрать события: ${err instanceof Error ? err.message : 'ошибка'}`);
    } finally {
      setEventsLoading(false);
    }
  }, [selectedFarmer]);

  const handleSelectFarmer = useCallback(async (farmer: FarmerInfo | null) => {
    setSelectedFarmer(farmer);
    setSelectedEventIds([]);
    setSelectedEventsById({});
    setEventResults({});
    setSelectedMatch(null);
    setCampaignResult(null);
    setImageExample(null);
    setImageError('');
    setRecommendedEvents([]);

    if (farmer) {
      setFarmerProductsLoading(true);
      try {
        const res = await fetch(`/farmer_data/${farmer.id}`);
        const data = await res.json();
        if (data?.products) {
          const mapped: MatchedProduct[] = data.products.map((p: Record<string, unknown>) => ({
            id: p.id as number,
            product_id: p.id as number,
            farmer_id: farmer.id,
            product_name: p.product_name as string,
            category: p.category as string,
            unit: p.unit as string,
            price: p.price as number,
            quantity: p.quantity as number,
          }));
          setFarmerProducts(mapped);
        }
      } catch (err) {
        console.error('Failed to load farmer products:', err);
      } finally {
        setFarmerProductsLoading(false);
      }
    } else {
      setFarmerProducts([]);
    }
  }, []);

  const fetchEventMatches = useCallback(async () => {
    if (!selectedFarmer || recommendedEvents.length === 0) {
      setEventMatches([]);
      return;
    }

    setMatcherLoading(true);
    try {
      const res = await fetch('/vector/match/events-to-products', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ limit: EVENT_PRODUCT_MATCH_LIMIT, farmer_id: selectedFarmer.id, future_only: true }),
      });
      if (!res.ok) {
        throw new Error('match request failed');
      }
      const data = await res.json();
      const matches = (Array.isArray(data.events) ? data.events : []) as EventProductsMatch[];
      const recommendedIds = new Set(recommendedEvents.map((e) => e.id));
      const filtered = matches
        .filter((m) => recommendedIds.has(m.event?.id))
        .map((m) => ({
          ...m,
          products: (m.products ?? []).filter(
            (p) => 1 - p.distance >= EVENT_PRODUCT_SCORE_THRESHOLD,
          ),
        }))
        .filter((m) => (m.products ?? []).length > 0);
      const recommendedOrder = new Map(recommendedEvents.map((e, index) => [e.id, index]));
      const ordered = filtered.sort(
        (a, b) => (recommendedOrder.get(a.event.id) ?? 0) - (recommendedOrder.get(b.event.id) ?? 0),
      );
      setEventMatches(ordered);
    } catch (err) {
      console.error('Failed to fetch event matches:', err);
      setEventMatches([]);
    } finally {
      setMatcherLoading(false);
    }
  }, [selectedFarmer, recommendedEvents]);

  useEffect(() => {
    if (activeTab !== 'matcher') return;
    void fetchEventMatches();
  }, [activeTab, fetchEventMatches]);

  const farmerFilteredResults = useMemo(() => {
    if (!selectedFarmer) return eventResults;
    const filtered: Record<number, MatchedProduct[]> = {};
    for (const [eventId, products] of Object.entries(eventResults)) {
      const filteredProducts = products.filter((p) => p.farmer_id === selectedFarmer.id);
      if (filteredProducts.length > 0) {
        filtered[Number(eventId)] = filteredProducts;
      }
    }
    return filtered;
  }, [eventResults, selectedFarmer]);

  const matchedProducts = useMemo(() => {
    const seen = new Set<number>();
    return Object.values(farmerFilteredResults).flat().filter((p) => {
      if (seen.has(p.product_id)) return false;
      seen.add(p.product_id);
      return true;
    });
  }, [farmerFilteredResults]);

  const workflow: WorkflowState = {
    selectedEventIds,
    selectedEventsById,
    matchedProducts,
    selectedMatch,
    campaignResult,
    isLoading,
  };

  const runAnalysis = useCallback(async (event: EventItem) => {
    setIsLoading(true);
    try {
      const embedRes = await fetch('/ml/embed/event', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ event_name: event.holiday_info, event_description: event.about }),
      });
      const embedData = await embedRes.json();
      const embedding = embedData.embedding;
      if (!embedding || !Array.isArray(embedding)) throw new Error('Invalid embedding');

      const searchBody: Record<string, unknown> = { vector: embedding, limit: 20 };
      if (selectedFarmer) {
        searchBody.farmer_id = selectedFarmer.id;
      }
      const searchRes = await fetch('/vector/search', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(searchBody),
      });
      const searchData = await searchRes.json();
      return (searchData.products ?? []) as MatchedProduct[];
    } catch (err) {
      console.error('ML workflow error for event', event.id, err);
      return [];
    } finally {
      setIsLoading(false);
    }
  }, [selectedFarmer]);

  const handleToggleEvent = useCallback(async (event: EventItem) => {
    if (selectedEventIds.includes(event.id)) {
      setSelectedEventIds((prev) => prev.filter((id) => id !== event.id));
      setEventResults((prev) => { const n = { ...prev }; delete n[event.id]; return n; });
      setSelectedEventsById((prev) => { const n = { ...prev }; delete n[event.id]; return n; });
      setSelectedMatch(null);
      setCampaignResult(null);
    } else {
      setSelectedEventIds((prev) => [...prev, event.id]);
      setSelectedEventsById((prev) => ({ ...prev, [event.id]: event }));
      setSelectedMatch(null);
      setCampaignResult(null);
      const products = await runAnalysis(event);
      setEventResults((prev) => ({ ...prev, [event.id]: products }));
    }
  }, [selectedEventIds, runAnalysis]);

  const handleSelectProduct = useCallback((product: MatchedProduct) => {
    const firstId = selectedEventIds[0];
    if (firstId === undefined || !selectedEventsById[firstId]) return;
    const event = selectedEventsById[firstId];
    setSelectedMatch({
      product,
      event,
      match: { id: `product:${product.id}|event:${event.id}`, score: product.id / 100 },
    });
  }, [selectedEventIds, selectedEventsById]);

  const handleGenerateCampaign = useCallback(async (productId: number) => {
    const firstId = selectedEventIds[0];
    if (firstId === undefined) return;
    const firstEvent = selectedEventsById[firstId];
    if (!firstEvent) return;
    setIsLoading(true);
    setCampaignResult(null);
    setImageExample(null);
    setImageError('');
    try {
      const res = await fetch('/agents/run_dynamic', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ product_id: productId, date_from: firstEvent.event_date, date_to: firstEvent.event_date, top_k: 1 }),
      });
      const data = await res.json();
      setCampaignResult(data);
    } catch (err) {
      console.error('Campaign generation error:', err);
    } finally {
      setIsLoading(false);
    }
  }, [selectedEventIds, selectedEventsById]);

  const handleGenerateImage = useCallback(async () => {
    if (!campaignResult?.image_prompt) {
      setImageError('Нет промта для генерации изображения');
      return;
    }
    setImageLoading(true);
    setImageError('');
    try {
      const res = await fetch('/agents/generate_image', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ prompt: campaignResult.image_prompt }),
      });
      if (!res.ok) {
        throw new Error('image request failed');
      }
      const data = await res.json();
      if (data?.image_url) {
        setImageExample({ url: data.image_url, prompt: data.prompt ?? campaignResult.image_prompt });
      } else {
        throw new Error('image url missing');
      }
    } catch (err) {
      console.error('Failed to generate image:', err);
      setImageError('Не удалось сгенерировать изображение');
      setImageExample(null);
    } finally {
      setImageLoading(false);
    }
  }, [campaignResult]);

  const renderContent = () => {
    switch (activeTab) {
      case 'dashboard': return <Dashboard workflow={workflow} />;
      case 'ingestion': return <Profile selectedFarmer={selectedFarmer} onSelectFarmer={handleSelectFarmer} farmerProducts={farmerProducts} />;
      case 'events':
        return (
          <EventEngine
            selectedEventIds={selectedEventIds}
            onToggleEvent={handleToggleEvent}
            recommendedEvents={recommendedEvents}
            onFindEvents={handleFindEvents}
            eventsLoading={eventsLoading}
            eventsError={eventsError}
            farmerName={selectedFarmer?.name ?? null}
          />
        );
      case 'matcher':
        return (
          <SemanticMatcher
            matches={eventMatches}
            selectedMatch={selectedMatch}
            isLoading={matcherLoading}
          />
        );
      case 'products':
        return (
          <ProductsView
            farmerProducts={farmerProducts}
            matchedProducts={matchedProducts}
            selectedMatch={selectedMatch}
            isLoading={isLoading || farmerProductsLoading}
            onSelectProduct={handleSelectProduct}
          />
        );
      case 'analytics':
        return (
          <Analytics
            campaignResult={campaignResult}
            selectedMatch={selectedMatch}
            isLoading={isLoading}
            onGenerateCampaign={handleGenerateCampaign}
            imageExample={imageExample}
            imageLoading={imageLoading}
            imageError={imageError}
            onGenerateImage={handleGenerateImage}
          />
        );
      default: return <Dashboard workflow={workflow} />;
    }
  };

  return (
    <div className="flex h-screen bg-gray-50 relative overflow-hidden">
      <ParticleBackground />
      <Sidebar
        activeTab={activeTab}
        onTabChange={setActiveTab}
        collapsed={sidebarCollapsed}
        onToggleCollapse={() => setSidebarCollapsed(!sidebarCollapsed)}
        selectedFarmer={selectedFarmer}
      />
      <div className="flex-1 overflow-y-auto relative z-10">{renderContent()}</div>
    </div>
  );
}

import { useState, useCallback, useMemo } from 'react';
import { Sidebar } from './components/Sidebar';
import { Dashboard } from './components/Dashboard';
import { DataIngestion } from './components/DataIngestion';
import { EventEngine } from './components/EventEngine';
import { SemanticMatcher } from './components/SemanticMatcher';
import { ProductsView } from './components/ProductsView';
import { Analytics } from './components/Analytics';
import { AIAssistant } from './components/AIAssistant';
import { ParticleBackground } from './components/ParticleBackground';
import { MlServices } from './components/MlServices';
import type { EventItem, MatchedProduct, WorkflowState } from './types';

export default function App() {
  const [activeTab, setActiveTab] = useState('dashboard');
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [selectedEventIds, setSelectedEventIds] = useState<number[]>([]);
  const [selectedEventsById, setSelectedEventsById] = useState<Record<number, EventItem>>({});
  const [eventResults, setEventResults] = useState<Record<number, MatchedProduct[]>>({});
  const [isLoading, setIsLoading] = useState(false);
  const [selectedMatch, setSelectedMatch] = useState<WorkflowState['selectedMatch']>(null);
  const [campaignResult, setCampaignResult] = useState<WorkflowState['campaignResult']>(null);

  const matchedProducts = useMemo(() => {
    const seen = new Set<number>();
    return Object.values(eventResults).flat().filter((p) => {
      if (seen.has(p.product_id)) return false;
      seen.add(p.product_id);
      return true;
    });
  }, [eventResults]);

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
      const searchRes = await fetch('/vector/search', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ vector: embedding, limit: 10 }),
      });
      const searchData = await searchRes.json();
      return (searchData.products ?? []) as MatchedProduct[];
    } catch (err) {
      console.error('ML workflow error for event', event.id, err);
      return [];
    } finally {
      setIsLoading(false);
    }
  }, []);

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

  const renderContent = () => {
    switch (activeTab) {
      case 'dashboard': return <Dashboard workflow={workflow} />;
      case 'ingestion': return <DataIngestion />;
      case 'events':
        return <EventEngine selectedEventIds={selectedEventIds} onToggleEvent={handleToggleEvent} />;
      case 'matcher':
        return (
          <SemanticMatcher
            selectedEvents={selectedEventIds.map((id) => selectedEventsById[id]).filter((e): e is EventItem => e !== undefined)}
            matchedProducts={matchedProducts}
            selectedMatch={selectedMatch}
            isLoading={isLoading}
          />
        );
      case 'products':
        return (
          <ProductsView
            products={matchedProducts}
            selectedMatch={selectedMatch}
            isLoading={isLoading}
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
          />
        );
      case 'ml': return <MlServices />;
      default: return <Dashboard workflow={workflow} />;
    }
  };

  return (
    <div className="size-full flex bg-gray-50 relative min-h-screen">
      <ParticleBackground />
      <Sidebar
        activeTab={activeTab}
        onTabChange={setActiveTab}
        collapsed={sidebarCollapsed}
        onToggleCollapse={() => setSidebarCollapsed(!sidebarCollapsed)}
      />
      <div className="flex-1 overflow-auto relative z-10">{renderContent()}</div>
      <AIAssistant />
    </div>
  );
}
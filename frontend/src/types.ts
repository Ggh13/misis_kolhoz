export type EventItem = {
  id: number;
  event_date: string;
  holiday_info: string;
  category: string;
  about: string;
  food_customs?: string;
};

export type MatchedProduct = {
  id: number;
  product_id: number;
  farmer_id: number;
  product_name: string;
  category: string;
  unit: string;
  price: number;
  quantity: number;
};

export type MatchInfo = {
  id: string;
  score: number;
};

export type CampaignContent = {
  post_text?: string;
  story_bullets?: string[];
  push_title?: string;
  push_text?: string;
};

export type CampaignResult = {
  farmer?: { id?: number; description?: string };
  product?: { id?: number; name?: string; description?: string };
  event?: { id?: number; name?: string; description?: string; date?: string; food_customs?: string };
  match?: { id?: string; score?: number };
  plan?: Record<string, any>;
  content?: CampaignContent;
  validator_notes?: string[];
  plan_approved?: boolean;
  retry_count?: number;
};

export type WorkflowState = {
  selectedEventIds: number[];
  selectedEventsById: Record<number, EventItem>;
  matchedProducts: MatchedProduct[];
  selectedMatch: {
    product: MatchedProduct;
    event: EventItem;
    match: MatchInfo;
  } | null;
  campaignResult: CampaignResult | null;
  isLoading: boolean;
};
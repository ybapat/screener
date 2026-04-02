import { getToken, getRefreshToken, setTokens, clearTokens } from "./auth";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

// ---------------------------------------------------------------------------
// Types matching the backend models
// ---------------------------------------------------------------------------

export type UserRole = "seller" | "buyer" | "admin";

export interface User {
  id: string;
  email: string;
  display_name: string;
  role: UserRole;
  age_range?: string;
  country?: string;
  timezone?: string;
  credit_balance: number;
  global_epsilon_budget: number;
  epsilon_spent: number;
  created_at: string;
  updated_at: string;
}

export interface TokenPair {
  access_token: string;
  refresh_token: string;
  expires_at: number;
}

export interface DataBatch {
  id: string;
  user_id: string;
  record_count: number;
  date_range_start?: string;
  date_range_end?: string;
  status: string;
  created_at: string;
}

export interface ScreenTimeRecordInput {
  app_name: string;
  app_category: string;
  duration_secs: number;
  started_at: string;
  ended_at: string;
  device_type?: string;
  os?: string;
}

export interface Dataset {
  id: string;
  title: string;
  description?: string;
  category_filter: string[];
  contributor_count: number;
  record_count: number;
  date_range_start?: string;
  date_range_end?: string;
  k_anonymity_k: number;
  epsilon_per_query: number;
  noise_mechanism: string;
  base_price_credits: number;
  current_price_credits: number;
  age_ranges: string[];
  countries: string[];
  status: string;
  created_at: string;
  updated_at: string;
}

export interface DatasetSample {
  id: string;
  dataset_id: string;
  app_category: string;
  duration_range: string;
  time_of_day: string;
  device_type?: string;
  contributor_age_range?: string;
  contributor_country?: string;
}

export interface Purchase {
  id: string;
  buyer_id: string;
  dataset_id: string;
  price_credits: number;
  status: string;
  download_url?: string;
  download_count: number;
  purchased_at: string;
}

export interface EpsilonLedgerEntry {
  id: string;
  user_id: string;
  event_type: string;
  epsilon_spent: number;
  epsilon_remaining: number;
  dataset_id?: string;
  description?: string;
  created_at: string;
}

export interface PrivacyBudget {
  total_budget: number;
  epsilon_spent: number;
  epsilon_remaining: number;
  percent_used: number;
}

export interface CreditTransaction {
  id: string;
  user_id: string;
  amount: number;
  balance_after: number;
  tx_type: string;
  reference_id?: string;
  description?: string;
  created_at: string;
}

export interface Bid {
  id: string;
  segment_id: string;
  buyer_id: string;
  bid_credits: number;
  status: string;
  expires_at: string;
  dataset_id?: string;
  created_at: string;
  updated_at: string;
}

export interface DataSegment {
  id: string;
  buyer_id: string;
  app_categories: string[];
  date_range_start?: string;
  date_range_end?: string;
  age_ranges?: string[];
  countries?: string[];
  device_types?: string[];
  min_contributors: number;
  min_records: number;
  desired_k_anonymity: number;
  max_epsilon: number;
  created_at: string;
}

export interface SellerDashboard {
  user: User;
  credit_balance: number;
  epsilon_budget: number;
  epsilon_spent: number;
  epsilon_remaining: number;
  recent_batches: DataBatch[];
  total_batches: number;
  recent_transactions: CreditTransaction[];
}

export interface BuyerDashboard {
  user: User;
  credit_balance: number;
  recent_purchases: Purchase[];
  total_purchases: number;
  active_bids: number;
  recent_transactions: CreditTransaction[];
}

// Backend wraps responses in { data, meta, error }
interface ApiEnvelope<T> {
  data?: T;
  meta?: Record<string, number>;
  error?: { message: string; status: number };
}

// ---------------------------------------------------------------------------
// Base fetcher with auto-auth and 401 refresh
// ---------------------------------------------------------------------------

let isRefreshing = false;
let refreshPromise: Promise<boolean> | null = null;

async function attemptRefresh(): Promise<boolean> {
  const rt = getRefreshToken();
  if (!rt) return false;

  try {
    const res = await fetch(`${API_BASE}/auth/refresh`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ refresh_token: rt }),
    });

    if (!res.ok) {
      clearTokens();
      return false;
    }

    const envelope: ApiEnvelope<TokenPair> = await res.json();
    if (envelope.data) {
      setTokens(envelope.data.access_token, envelope.data.refresh_token);
      return true;
    }
    clearTokens();
    return false;
  } catch {
    clearTokens();
    return false;
  }
}

async function apiFetch<T>(
  path: string,
  options: RequestInit = {}
): Promise<{ data: T; meta?: Record<string, number> }> {
  const token = getToken();

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  let res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers,
  });

  // If 401, try to refresh and retry once
  if (res.status === 401 && token) {
    if (!isRefreshing) {
      isRefreshing = true;
      refreshPromise = attemptRefresh();
    }

    const refreshed = await refreshPromise;
    isRefreshing = false;
    refreshPromise = null;

    if (refreshed) {
      const newToken = getToken();
      headers["Authorization"] = `Bearer ${newToken}`;
      res = await fetch(`${API_BASE}${path}`, { ...options, headers });
    } else {
      clearTokens();
      if (typeof window !== "undefined") {
        window.location.href = "/login";
      }
      throw new Error("Session expired");
    }
  }

  const envelope: ApiEnvelope<T> = await res.json();

  if (!res.ok || envelope.error) {
    throw new Error(envelope.error?.message || `Request failed: ${res.status}`);
  }

  return { data: envelope.data as T, meta: envelope.meta };
}

// ---------------------------------------------------------------------------
// Auth
// ---------------------------------------------------------------------------

export async function register(params: {
  email: string;
  password: string;
  display_name: string;
  role: UserRole;
  age_range?: string;
  country?: string;
  timezone?: string;
}): Promise<{ user: User; tokens: TokenPair }> {
  const res = await fetch(`${API_BASE}/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(params),
  });

  const envelope: ApiEnvelope<{ user: User; tokens: TokenPair }> =
    await res.json();

  if (!res.ok || envelope.error) {
    throw new Error(envelope.error?.message || "Registration failed");
  }

  return envelope.data!;
}

export async function login(
  email: string,
  password: string
): Promise<{ user: User; tokens: TokenPair }> {
  const res = await fetch(`${API_BASE}/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ email, password }),
  });

  const envelope: ApiEnvelope<{ user: User; tokens: TokenPair }> =
    await res.json();

  if (!res.ok || envelope.error) {
    throw new Error(envelope.error?.message || "Login failed");
  }

  return envelope.data!;
}

export async function refresh(): Promise<TokenPair> {
  const rt = getRefreshToken();
  if (!rt) throw new Error("No refresh token");

  const res = await fetch(`${API_BASE}/auth/refresh`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ refresh_token: rt }),
  });

  const envelope: ApiEnvelope<TokenPair> = await res.json();

  if (!res.ok || envelope.error) {
    throw new Error(envelope.error?.message || "Refresh failed");
  }

  return envelope.data!;
}

// ---------------------------------------------------------------------------
// User
// ---------------------------------------------------------------------------

export async function getProfile(): Promise<User> {
  const { data } = await apiFetch<User>("/api/v1/users/me");
  return data;
}

// ---------------------------------------------------------------------------
// Seller: Data Upload
// ---------------------------------------------------------------------------

export async function uploadData(
  records: ScreenTimeRecordInput[]
): Promise<DataBatch> {
  const { data } = await apiFetch<DataBatch>("/api/v1/data/upload", {
    method: "POST",
    body: JSON.stringify({ records }),
  });
  return data;
}

export async function getBatches(
  limit = 20,
  offset = 0
): Promise<{ batches: DataBatch[]; total: number }> {
  const { data, meta } = await apiFetch<DataBatch[]>(
    `/api/v1/data/batches?limit=${limit}&offset=${offset}`
  );
  return { batches: data || [], total: meta?.total ?? 0 };
}

// ---------------------------------------------------------------------------
// Marketplace: Datasets
// ---------------------------------------------------------------------------

export async function getDatasets(
  categories?: string[],
  limit = 20,
  offset = 0
): Promise<{ datasets: Dataset[]; total: number }> {
  const params = new URLSearchParams({ limit: String(limit), offset: String(offset) });
  if (categories && categories.length > 0) {
    params.set("categories", categories.join(","));
  }
  const { data, meta } = await apiFetch<Dataset[]>(
    `/api/v1/marketplace/datasets?${params.toString()}`
  );
  return { datasets: data || [], total: meta?.total ?? 0 };
}

export async function getDataset(
  id: string
): Promise<{ dataset: Dataset; samples: DatasetSample[] }> {
  const { data } = await apiFetch<{ dataset: Dataset; samples: DatasetSample[] }>(
    `/api/v1/marketplace/datasets/${id}`
  );
  return data;
}

export async function purchaseDataset(datasetId: string): Promise<Purchase> {
  const { data } = await apiFetch<Purchase>(
    `/api/v1/marketplace/datasets/${datasetId}/purchase`,
    { method: "POST" }
  );
  return data;
}

// ---------------------------------------------------------------------------
// Seller: Privacy
// ---------------------------------------------------------------------------

export async function getPrivacyBudget(): Promise<PrivacyBudget> {
  const { data } = await apiFetch<PrivacyBudget>("/api/v1/privacy/budget");
  return data;
}

export async function getEpsilonLedger(
  limit = 20,
  offset = 0
): Promise<{ entries: EpsilonLedgerEntry[]; total: number }> {
  const { data, meta } = await apiFetch<EpsilonLedgerEntry[]>(
    `/api/v1/privacy/ledger?limit=${limit}&offset=${offset}`
  );
  return { entries: data || [], total: meta?.total ?? 0 };
}

// ---------------------------------------------------------------------------
// Dashboard
// ---------------------------------------------------------------------------

export async function getSellerDashboard(): Promise<SellerDashboard> {
  const { data } = await apiFetch<SellerDashboard>("/api/v1/dashboard/seller");
  return data;
}

export async function getBuyerDashboard(): Promise<BuyerDashboard> {
  const { data } = await apiFetch<BuyerDashboard>("/api/v1/dashboard/buyer");
  return data;
}

// ---------------------------------------------------------------------------
// Buyer: Credits
// ---------------------------------------------------------------------------

export async function topupCredits(
  amount: number
): Promise<{ credit_balance: number }> {
  const { data } = await apiFetch<{ credit_balance: number }>(
    "/api/v1/credits/topup",
    { method: "POST", body: JSON.stringify({ amount }) }
  );
  return data;
}

// ---------------------------------------------------------------------------
// Buyer: Segments & Bids
// ---------------------------------------------------------------------------

export async function createSegment(params: {
  app_categories: string[];
  date_range_start?: string;
  date_range_end?: string;
  age_ranges?: string[];
  countries?: string[];
  device_types?: string[];
  min_contributors: number;
  min_records: number;
  desired_k_anonymity: number;
  max_epsilon: number;
}): Promise<DataSegment> {
  const { data } = await apiFetch<DataSegment>("/api/v1/marketplace/segments", {
    method: "POST",
    body: JSON.stringify(params),
  });
  return data;
}

export async function placeBid(
  segmentId: string,
  bidCredits: number,
  durationMinutes: number
): Promise<Bid> {
  const { data } = await apiFetch<Bid>(
    `/api/v1/marketplace/segments/${segmentId}/bids`,
    {
      method: "POST",
      body: JSON.stringify({
        bid_credits: bidCredits,
        duration_minutes: durationMinutes,
      }),
    }
  );
  return data;
}

export async function getBids(): Promise<Bid[]> {
  const { data } = await apiFetch<Bid[]>("/api/v1/marketplace/bids");
  return data || [];
}

export async function getPurchases(
  limit = 20,
  offset = 0
): Promise<{ purchases: Purchase[]; total: number }> {
  const { data, meta } = await apiFetch<Purchase[]>(
    `/api/v1/buyer/purchases?limit=${limit}&offset=${offset}`
  );
  return { purchases: data || [], total: meta?.total ?? 0 };
}

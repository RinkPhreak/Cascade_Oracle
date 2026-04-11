/**
 * Supplementary domain types for entities not yet in the OpenAPI spec.
 * DO NOT remove — these are NOT auto-generated and should not be modified by openapi-ts.
 * These model the backend entities inferred from the campaign/contact/account flow.
 */

// ─── Account Pool ────────────────────────────────────────────────────────────

export type TgAccountStatus =
  | 'WARMING_UP'
  | 'ACTIVE'
  | 'COOLING_DOWN'
  | 'SUSPENDED'
  | 'BANNED';

export interface TgAccount {
  id: string;
  phone: string;
  status: TgAccountStatus;
  daily_count: number;
  daily_limit: number;
  proxy_id: string | null;
  created_at: string;
  updated_at: string;
}

export interface TgAccountSession {
  session_id: string;
  qr_url: string | null;
  pairing_code: string | null;
  expires_at: string;
  status: 'PENDING_QR' | 'PENDING_CODE' | 'AUTHENTICATED' | 'FAILED';
}

// ─── Proxy ────────────────────────────────────────────────────────────────────

export interface Proxy {
  id: string;
  host: string;
  port: number;
  username: string;
  protocol: 'socks5' | 'http' | 'mtproto';
  bound_account_id: string | null;
  is_healthy: boolean;
  latency_ms: number | null;
  last_checked_at: string | null;
  created_at: string;
}

export interface CreateProxyRequest {
  host: string;
  port: number;
  username: string;
  password: string;
  protocol: 'socks5' | 'http' | 'mtproto';
}

export interface ReassignProxyRequest {
  account_id: string;
  reason: string;
}

// ─── Account Events ───────────────────────────────────────────────────────────

export type AccountEventType =
  | 'BANNED'
  | 'SERVICE_NOTICE'
  | 'RECONNECT'
  | 'SESSION_CREATED'
  | 'SESSION_EXPIRED'
  | 'PROXY_CHANGED'
  | 'FLOOD_WAIT'
  | 'STATUS_CHANGE';

export interface AccountEvent {
  id: string;
  account_id: string;
  type: AccountEventType;
  description: string;
  occurred_at: string;
}

// ─── Campaigns (extended) ─────────────────────────────────────────────────────

export type CampaignStatus =
  | 'DRAFT'
  | 'SCHEDULED'
  | 'RUNNING'
  | 'PAUSED'
  | 'COMPLETED'
  | 'FAILED';

export interface Campaign {
  id: string;
  name: string;
  status: CampaignStatus;
  scheduled_at: string | null;
  created_at: string;
  total_contacts: number;
  completed: number;
  replied: number;
  failed: number;
}

export interface CampaignStats {
  campaign_id: string;
  total: number;
  completed: number;
  replied: number;
  failed: number;
  tg_attempted: number;
  sms_attempted: number;
  error_breakdown: Record<string, number>;
}

export interface CampaignListItem {
  id: string;
  name: string;
  status: 'draft' | 'running' | 'paused' | 'completed';
  scheduled_at: string | null;
  created_at: string;
}

// ─── Campaign Tasks ───────────────────────────────────────────────────────────

export type TaskStatus =
  | 'pending'
  | 'in_progress'
  | 'completed'
  | 'replied'
  | 'failed';

export type TaskChannel = 'telegram' | 'sms';

export interface CampaignTask {
  id: string;
  campaign_id: string;
  contact_id: string;
  contact_phone: string;
  status: TaskStatus;
  channel: TaskChannel;
  error_code: string | null;
  started_at: string | null;
  updated_at: string;
}

export interface StuckTask {
  id: string;
  contact_id: string;
  status: string;
  channel: string;
  updated_at: string;
}

// ─── Contacts ─────────────────────────────────────────────────────────────────

export interface Contact {
  id: string;
  phone: string;
  name: string;
  extra_data: Record<string, string>;
  has_replied: boolean;
  reply_text: string | null;
  reply_at: string | null;
  reply_account_id: string | null;
  reply_account_phone: string | null;
  is_anonymised: boolean;
  created_at: string;
}

export interface ContactCSVRow {
  phone: string;
  name?: string;
  extra_data?: Record<string, string>;
}

export type ContactTraceStepStatus =
  | 'ENQUEUED'
  | 'ATTEMPTED'
  | 'RATE_LIMITED'
  | 'NOT_FOUND'
  | 'DELIVERED'
  | 'REPLIED'
  | 'FAILED';

export interface ContactTrace {
  id: string;
  contact_id: string;
  step: number;
  channel: TaskChannel | 'system';
  status: ContactTraceStepStatus;
  error_code: string | null;
  description: string;
  occurred_at: string;
}

// ─── System Metrics ───────────────────────────────────────────────────────────

export interface SystemMetrics {
  cascade_memory_usage_ratio: number;
  active_tg_accounts: number;
  total_tg_accounts: number;
  queue_depth: number;
  system_status: 'OPERATIONAL' | 'DEGRADED' | 'HALTED';
}

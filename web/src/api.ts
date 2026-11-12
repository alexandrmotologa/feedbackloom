import {
  Account,
  DashboardStats,
  EnrichedFeedbackItem,
  FeedbackItem,
  Opportunity,
  ClosedLoopNotification,
  LiveIngestRequest,
  LiveIngestResponse,
  GitHubIssuePayload,
  JiraEpicPayload,
} from './types'

const BASE_URL = '/api/v1'

export async function fetchStats(): Promise<DashboardStats> {
  const res = await fetch(`${BASE_URL}/stats`)
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to fetch stats`)
  return res.json()
}

export async function fetchAccounts(): Promise<Account[]> {
  const res = await fetch(`${BASE_URL}/accounts`)
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to fetch accounts`)
  const data = await res.json()
  return data.accounts || []
}

export async function syncCRM(provider = 'stripe'): Promise<{ synced: number; message: string }> {
  const res = await fetch(`${BASE_URL}/crm/sync?provider=${provider}`, { method: 'POST' })
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to sync CRM`)
  return res.json()
}

export async function fetchFeedback(): Promise<FeedbackItem[]> {
  const res = await fetch(`${BASE_URL}/feedback`)
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to fetch feedback`)
  const data = await res.json()
  return data.items || []
}

export async function seedDemoFeedback(): Promise<{ seeded: number; message: string }> {
  const res = await fetch(`${BASE_URL}/feedback/seed`, { method: 'POST' })
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to seed feedback`)
  return res.json()
}

export async function runClustering(minClusterSize?: number, minPoints?: number): Promise<any> {
  let url = `${BASE_URL}/cluster/run`
  const params = new URLSearchParams()
  if (minClusterSize) params.set('min_cluster_size', String(minClusterSize))
  if (minPoints) params.set('min_points', String(minPoints))
  if (params.toString()) {
    url += `?${params.toString()}`
  }
  const res = await fetch(url, { method: 'POST' })
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to run clustering`)
  return res.json()
}

export async function fetchOpportunities(): Promise<Opportunity[]> {
  const res = await fetch(`${BASE_URL}/opportunities`)
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to fetch opportunities`)
  const data = await res.json()
  return data.opportunities || []
}

export async function fetchOpportunityDetail(id: string): Promise<{ opportunity: Opportunity; quotes: EnrichedFeedbackItem[] }> {
  const res = await fetch(`${BASE_URL}/opportunities/${id}`)
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to fetch opportunity detail`)
  return res.json()
}

export async function generatePRD(id: string): Promise<{ markdown: string; notion_json: string }> {
  const res = await fetch(`${BASE_URL}/opportunities/${id}/prd`, { method: 'POST' })
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to generate PRD`)
  return res.json()
}

export async function exportToNotion(id: string): Promise<{ url: string; message: string }> {
  const res = await fetch(`${BASE_URL}/opportunities/${id}/export-notion`, { method: 'POST' })
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to export to Notion`)
  return res.json()
}

export async function fetchGitHubIssue(id: string): Promise<GitHubIssuePayload> {
  const res = await fetch(`${BASE_URL}/opportunities/${id}/export-github`)
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to generate GitHub Issue`)
  return res.json()
}

export async function fetchJiraEpic(id: string): Promise<JiraEpicPayload> {
  const res = await fetch(`${BASE_URL}/opportunities/${id}/export-jira`)
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to generate Jira Epic`)
  return res.json()
}

export async function updateOpportunityEffort(id: string, effortSprints: number): Promise<Opportunity> {
  const res = await fetch(`${BASE_URL}/opportunities/${id}/effort`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ effort_sprints: effortSprints }),
  })
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to update effort`)
  return res.json()
}

export async function updateOpportunityStatus(id: string, status: string): Promise<{ status: string }> {
  const res = await fetch(`${BASE_URL}/opportunities/${id}/status`, {
    method: 'PATCH',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status }),
  })
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to update status`)
  return res.json()
}

export async function testLiveIngest(payload: LiveIngestRequest): Promise<LiveIngestResponse> {
  const res = await fetch(`${BASE_URL}/ingest/test-live`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  })
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to execute live ingest test`)
  return res.json()
}

export async function fetchNotifications(opportunityId: string): Promise<ClosedLoopNotification[]> {
  const res = await fetch(`${BASE_URL}/opportunities/${opportunityId}/notifications`)
  if (!res.ok) return []
  const data = await res.json()
  return data.notifications || []
}

export async function generateNotifications(opportunityId: string): Promise<{ generated: number; notifications: ClosedLoopNotification[] }> {
  const res = await fetch(`${BASE_URL}/opportunities/${opportunityId}/notifications/generate`, { method: 'POST' })
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to generate notifications`)
  return res.json()
}

export async function dispatchNotifications(opportunityId: string): Promise<{ sent: number; message: string }> {
  const res = await fetch(`${BASE_URL}/opportunities/${opportunityId}/notifications/dispatch`, { method: 'POST' })
  if (!res.ok) throw new Error(`HTTP ${res.status}: Failed to dispatch notifications`)
  return res.json()
}

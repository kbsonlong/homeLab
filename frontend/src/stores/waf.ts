import { create } from 'zustand'
import { devtools } from 'zustand/middleware'

interface WAFPolicy {
  id: string
  host: string
  mode: 'On' | 'DetectionOnly' | 'Off'
  enable_crs: boolean
  exceptions: {
    paths: string[]
    methods: string[]
    ip_allow: string[]
    headers_allow: Record<string, string>
  }
  custom_rules: CustomRule[]
  created_at: string
  updated_at: string
  updated_by: string
  version: number
}

interface CustomRule {
  id: string
  name: string
  rule: string
  description: string
  enabled: boolean
  created_at: string
}

interface WAFStatus {
  global_policy: WAFPolicy
  host_policies: Record<string, WAFPolicy>
  controller_config: {
    allow_snippet_annotations: boolean
    modsecurity_snippet: string
  }
  last_updated: string
}

interface MetricsSummary {
  total_requests: number
  status_4xx: number
  status_5xx: number
  status_403: number
  waf_blocked: number
  top_hosts: HostMetrics[]
  top_paths: PathMetrics[]
  top_rule_ids: RuleMetrics[]
  time_range: TimeRange
}

interface HostMetrics {
  host: string
  requests: number
  blocked: number
  error_rate: number
}

interface PathMetrics {
  path: string
  requests: number
  blocked: number
  error_rate: number
}

interface RuleMetrics {
  rule_id: string
  rule_name: string
  count: number
}

interface TimeRange {
  start: string
  end: string
}

interface LogEntry {
  timestamp: string
  message: string
  fields: Record<string, any>
  host: string
  status: number
  rule_id?: string
  client_ip: string
  path: string
  method: string
}

interface LogQuery {
  query: string
  time_range: TimeRange
  limit: number
  offset: number
}

interface LogSearchResult {
  entries: LogEntry[]
  total: number
  time_range: TimeRange
}

interface AuditLog {
  id: string
  action: string
  resource_type: string
  resource_id: string
  user: string
  details: string
  old_value: any
  new_value: any
  created_at: string
}

interface AuditLogResult {
  entries: AuditLog[]
  total: number
}

interface WAFStore {
  // State
  policies: Record<string, WAFPolicy>
  status: WAFStatus | null
  metrics: MetricsSummary | null
  logs: LogSearchResult | null
  auditLogs: AuditLogResult | null
  loading: boolean
  error: string | null
  
  // Actions
  setPolicies: (policies: Record<string, WAFPolicy>) => void
  setStatus: (status: WAFStatus) => void
  setMetrics: (metrics: MetricsSummary) => void
  setLogs: (logs: LogSearchResult) => void
  setAuditLogs: (auditLogs: AuditLogResult) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  
  // API calls would be added here
  fetchStatus: () => Promise<void>
  fetchMetrics: (timeRange: TimeRange) => Promise<void>
  searchLogs: (query: LogQuery) => Promise<void>
  fetchAuditLogs: (limit?: number, offset?: number) => Promise<void>
}

export const useWAFStore = create<WAFStore>()(
  devtools(
    (set) => ({
      // Initial state
      policies: {},
      status: null,
      metrics: null,
      logs: null,
      auditLogs: null,
      loading: false,
      error: null,
      
      // Actions
      setPolicies: (policies) => set({ policies }),
      setStatus: (status) => set({ status }),
      setMetrics: (metrics) => set({ metrics }),
      setLogs: (logs) => set({ logs }),
      setAuditLogs: (auditLogs) => set({ auditLogs }),
      setLoading: (loading) => set({ loading }),
      setError: (error) => set({ error }),
      
      // API calls (placeholder implementations)
      fetchStatus: async () => {
        set({ loading: true, error: null })
        try {
          // This would be replaced with actual API call
          const response = await fetch('/api/waf/status')
          const data = await response.json()
          set({ status: data, loading: false })
        } catch (error) {
          set({ error: 'Failed to fetch status', loading: false })
        }
      },
      
      fetchMetrics: async (timeRange) => {
        set({ loading: true, error: null })
        try {
          const params = new URLSearchParams({
            start: timeRange.start,
            end: timeRange.end
          })
          const response = await fetch(`/api/metrics/summary?${params}`)
          const data = await response.json()
          set({ metrics: data, loading: false })
        } catch (error) {
          set({ error: 'Failed to fetch metrics', loading: false })
        }
      },
      
      searchLogs: async (query) => {
        set({ loading: true, error: null })
        try {
          const response = await fetch('/api/logs/search', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(query)
          })
          const data = await response.json()
          set({ logs: data, loading: false })
        } catch (error) {
          set({ error: 'Failed to search logs', loading: false })
        }
      },
      
      fetchAuditLogs: async (limit = 100, offset = 0) => {
        set({ loading: true, error: null })
        try {
          const params = new URLSearchParams({
            limit: limit.toString(),
            offset: offset.toString()
          })
          const response = await fetch(`/api/audit?${params}`)
          const data = await response.json()
          set({ auditLogs: data, loading: false })
        } catch (error) {
          set({ error: 'Failed to fetch audit logs', loading: false })
        }
      }
    }),
    {
      name: 'waf-store',
    }
  )
)
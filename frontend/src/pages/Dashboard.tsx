import React, { useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card'
import { Shield, AlertTriangle, Activity, TrendingUp } from 'lucide-react'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, PieChart, Pie, Cell } from 'recharts'
import { useWAFStore } from '../stores/waf'

const Dashboard: React.FC = () => {
  const { status, metrics, fetchStatus, fetchMetrics } = useWAFStore()

  useEffect(() => {
    fetchStatus()
    const timeRange = {
      start: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
      end: new Date().toISOString()
    }
    fetchMetrics(timeRange)
  }, [fetchStatus, fetchMetrics])

  const getModeColor = (mode: string) => {
    switch (mode) {
      case 'On': return 'text-green-600 bg-green-100'
      case 'DetectionOnly': return 'text-yellow-600 bg-yellow-100'
      case 'Off': return 'text-red-600 bg-red-100'
      default: return 'text-gray-600 bg-gray-100'
    }
  }

  const getModeStats = () => {
    if (!status?.host_policies) return { on: 0, detection: 0, off: 0 }
    
    const policies = Object.values(status.host_policies)
    return {
      on: policies.filter(p => p.mode === 'On').length,
      detection: policies.filter(p => p.mode === 'DetectionOnly').length,
      off: policies.filter(p => p.mode === 'Off').length
    }
  }

  const modeStats = getModeStats()
  const pieData = [
    { name: 'On', value: modeStats.on || 0, color: '#10B981' },
    { name: 'Detection', value: modeStats.detection || 0, color: '#F59E0B' },
    { name: 'Off', value: modeStats.off || 0, color: '#EF4444' }
  ]

  const chartData = metrics?.top_hosts?.slice(0, 5).map(host => ({
    name: host.host,
    requests: host.requests,
    blocked: host.blocked
  })) || []

  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Policies</CardTitle>
            <Shield className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{Object.keys(status?.host_policies || {}).length}</div>
            <p className="text-xs text-muted-foreground">Active WAF policies</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Blocked Requests</CardTitle>
            <AlertTriangle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics?.waf_blocked || 0}</div>
            <p className="text-xs text-muted-foreground">Last 24 hours</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Requests</CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{metrics?.total_requests || 0}</div>
            <p className="text-xs text-muted-foreground">Last 24 hours</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Block Rate</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {metrics?.total_requests ? ((metrics.waf_blocked / metrics.total_requests) * 100).toFixed(2) : 0}%
            </div>
            <p className="text-xs text-muted-foreground">Of total requests</p>
          </CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>WAF Mode Distribution</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={300}>
              <PieChart>
                <Pie
                  data={pieData}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="value"
                >
                  {pieData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Top Hosts Activity</CardTitle>
          </CardHeader>
          <CardContent>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" />
                <YAxis />
                <Tooltip />
                <Bar dataKey="requests" fill="#3B82F6" name="Requests" />
                <Bar dataKey="blocked" fill="#EF4444" name="Blocked" />
              </BarChart>
            </ResponsiveContainer>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Recent Policy Updates</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {Object.entries(status?.host_policies || {}).slice(0, 5).map(([host, policy]) => (
              <div key={host} className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
                <div>
                  <div className="font-medium">{host}</div>
                  <div className="text-sm text-gray-500">
                    Updated {new Date(policy.updated_at).toLocaleString()}
                  </div>
                </div>
                <span className={`px-3 py-1 rounded-full text-xs font-medium ${getModeColor(policy.mode)}`}>
                  {policy.mode}
                </span>
              </div>
            ))}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

export default Dashboard
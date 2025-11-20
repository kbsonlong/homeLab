import React, { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card'
import { Search, Filter, Download, RefreshCw } from 'lucide-react'
import { useWAFStore } from '../stores/waf'
import { format } from 'date-fns'

const Logs: React.FC = () => {
  const { logs, searchLogs } = useWAFStore()
  const [query, setQuery] = useState('')
  const [filters, setFilters] = useState({
    status: '',
    host: '',
    rule_id: '',
    startTime: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
    endTime: new Date().toISOString()
  })
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    handleSearch()
  }, [])

  const handleSearch = async () => {
    setLoading(true)
    const searchQuery = {
      query: buildQuery(),
      time_range: {
        start: filters.startTime,
        end: filters.endTime
      },
      limit: 100,
      offset: 0
    }
    await searchLogs(searchQuery)
    setLoading(false)
  }

  const buildQuery = () => {
    const conditions = []
    if (query) conditions.push(`_msg:*${query}*`)
    if (filters.status) conditions.push(`status:${filters.status}`)
    if (filters.host) conditions.push(`host:${filters.host}`)
    if (filters.rule_id) conditions.push(`rule_id:${filters.rule_id}`)
    
    return conditions.length > 0 ? conditions.join(' AND ') : '*'
  }

  const getStatusColor = (status: number) => {
    if (status >= 200 && status < 300) return 'text-green-600'
    if (status >= 400 && status < 500) return 'text-yellow-600'
    if (status >= 500) return 'text-red-600'
    return 'text-gray-600'
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900">WAF Logs</h1>
        <div className="flex space-x-2">
          <button
            onClick={handleSearch}
            disabled={loading}
            className="flex items-center space-x-2 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 disabled:opacity-50"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
            <span>Refresh</span>
          </button>
          <button className="flex items-center space-x-2 bg-gray-600 text-white px-4 py-2 rounded-lg hover:bg-gray-700">
            <Download className="h-4 w-4" />
            <span>Export</span>
          </button>
        </div>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Search Filters</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-3 lg:grid-cols-5 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Search</label>
              <input
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="Search logs..."
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
              <select
                value={filters.status}
                onChange={(e) => setFilters({...filters, status: e.target.value})}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">All</option>
                <option value="200">200 OK</option>
                <option value="403">403 Forbidden</option>
                <option value="404">404 Not Found</option>
                <option value="500">500 Error</option>
              </select>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Host</label>
              <input
                type="text"
                value={filters.host}
                onChange={(e) => setFilters({...filters, host: e.target.value})}
                placeholder="example.com"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Rule ID</label>
              <input
                type="text"
                value={filters.rule_id}
                onChange={(e) => setFilters({...filters, rule_id: e.target.value})}
                placeholder="942100"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            <div className="flex items-end">
              <button
                onClick={handleSearch}
                className="w-full flex items-center justify-center space-x-2 bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700"
              >
                <Search className="h-4 w-4" />
                <span>Search</span>
              </button>
            </div>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex justify-between items-center">
            <CardTitle>Log Entries ({logs?.total || 0})</CardTitle>
            <div className="text-sm text-gray-500">
              {format(new Date(filters.startTime), 'PPp')} - {format(new Date(filters.endTime), 'PPp')}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="space-y-2">
            {logs?.entries?.map((log, index) => (
              <div key={index} className="border border-gray-200 rounded-lg p-4 hover:bg-gray-50">
                <div className="flex justify-between items-start mb-2">
                  <div className="flex items-center space-x-2">
                    <span className="text-sm font-medium text-gray-900">{log.host}</span>
                    <span className={`px-2 py-1 text-xs font-medium rounded ${getStatusColor(log.status)}`}>
                      {log.status}
                    </span>
                    <span className="text-xs text-gray-500">{log.method}</span>
                    <span className="text-xs text-gray-600">{log.path}</span>
                  </div>
                  <span className="text-xs text-gray-500">
                    {format(new Date(log.timestamp), 'PPp')}
                  </span>
                </div>
                
                <div className="text-sm text-gray-700 mb-2">{log.message}</div>
                
                <div className="flex items-center space-x-4 text-xs text-gray-500">
                  <span>IP: {log.client_ip}</span>
                  {log.rule_id && <span>Rule: {log.rule_id}</span>}
                </div>
              </div>
            ))}
          </div>
          
          {(!logs?.entries || logs.entries.length === 0) && (
            <div className="text-center py-8 text-gray-500">
              No logs found matching your criteria
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}

function getStatusColor(status: number) {
  if (status >= 200 && status < 300) return 'bg-green-100 text-green-800'
  if (status >= 400 && status < 500) return 'bg-yellow-100 text-yellow-800'
  if (status >= 500) return 'bg-red-100 text-red-800'
  return 'bg-gray-100 text-gray-800'
}

export default Logs
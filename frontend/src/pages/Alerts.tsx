import React, { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card'
import { AlertTriangle, CheckCircle, Clock, Plus, Edit, Trash2 } from 'lucide-react'

interface AlertRule {
  id: string
  name: string
  expression: string
  for: string
  labels: Record<string, string>
  annotations: Record<string, string>
  enabled: boolean
  created_at: string
  updated_at: string
}

interface Alert {
  id: string
  name: string
  state: 'firing' | 'pending' | 'resolved'
  labels: Record<string, string>
  annotations: Record<string, string>
  starts_at: string
  ends_at?: string
  generator_url: string
}

const Alerts: React.FC = () => {
  const [rules, setRules] = useState<AlertRule[]>([])
  const [alerts, setAlerts] = useState<Alert[]>([])
  const [showRuleModal, setShowRuleModal] = useState(false)
  const [editingRule, setEditingRule] = useState<AlertRule | null>(null)

  useEffect(() => {
    // Fetch alert rules and active alerts
    fetchRules()
    fetchAlerts()
  }, [])

  const fetchRules = async () => {
    try {
      const response = await fetch('/api/alerts/rules')
      if (!response.ok) {
        console.warn('Alert rules API not available, using empty data')
        setRules([])
        return
      }
      const data = await response.json()
      setRules(data.rules || [])
    } catch (error) {
      console.error('Failed to fetch alert rules:', error)
      setRules([])
    }
  }

  const fetchAlerts = async () => {
    try {
      const response = await fetch('/api/alerts')
      if (!response.ok) {
        console.warn('Alerts API not available, using empty data')
        setAlerts([])
        return
      }
      const data = await response.json()
      setAlerts(data.alerts || [])
    } catch (error) {
      console.error('Failed to fetch alerts:', error)
      setAlerts([])
    }
  }

  const getStateIcon = (state: string) => {
    switch (state) {
      case 'firing':
        return <AlertTriangle className="h-5 w-5 text-red-500" />
      case 'pending':
        return <Clock className="h-5 w-5 text-yellow-500" />
      case 'resolved':
        return <CheckCircle className="h-5 w-5 text-green-500" />
      default:
        return <AlertTriangle className="h-5 w-5 text-gray-500" />
    }
  }

  const getStateColor = (state: string) => {
    switch (state) {
      case 'firing':
        return 'bg-red-100 text-red-800'
      case 'pending':
        return 'bg-yellow-100 text-yellow-800'
      case 'resolved':
        return 'bg-green-100 text-green-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900">Alerts & Rules</h1>
        <button
          onClick={() => {
            setEditingRule(null)
            setShowRuleModal(true)
          }}
          className="flex items-center space-x-2 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
        >
          <Plus className="h-4 w-4" />
          <span>Add Rule</span>
        </button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Active Alerts</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-red-600">
              {alerts.filter(a => a.state === 'firing').length}
            </div>
            <p className="text-sm text-gray-500 mt-1">Currently firing</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Pending Alerts</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-yellow-600">
              {alerts.filter(a => a.state === 'pending').length}
            </div>
            <p className="text-sm text-gray-500 mt-1">Waiting to fire</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Total Rules</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-3xl font-bold text-blue-600">{rules.length}</div>
            <p className="text-sm text-gray-500 mt-1">Configured rules</p>
          </CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Active Alerts</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {alerts.length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                  No active alerts
                </div>
              ) : (
                alerts.map((alert) => (
                  <div key={alert.id} className="border border-gray-200 rounded-lg p-4">
                    <div className="flex items-start space-x-3">
                      {getStateIcon(alert.state)}
                      <div className="flex-1">
                        <div className="flex items-center justify-between">
                          <h3 className="font-medium text-gray-900">{alert.name}</h3>
                          <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStateColor(alert.state)}`}>
                            {alert.state}
                          </span>
                        </div>
                        <p className="text-sm text-gray-600 mt-1">
                          {alert.annotations.description || 'No description available'}
                        </p>
                        <div className="flex items-center space-x-4 mt-2 text-xs text-gray-500">
                          <span>Started: {new Date(alert.starts_at).toLocaleString()}</span>
                          {alert.labels.host && <span>Host: {alert.labels.host}</span>}
                        </div>
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Alert Rules</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {rules.length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                  No alert rules configured
                </div>
              ) : (
                rules.map((rule) => (
                  <div key={rule.id} className="border border-gray-200 rounded-lg p-4">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-3">
                        <div className={`w-3 h-3 rounded-full ${rule.enabled ? 'bg-green-500' : 'bg-gray-300'}`} />
                        <div>
                          <h3 className="font-medium text-gray-900">{rule.name}</h3>
                          <p className="text-sm text-gray-600">{rule.expression}</p>
                        </div>
                      </div>
                      <div className="flex items-center space-x-2">
                        <button
                          onClick={() => {
                            setEditingRule(rule)
                            setShowRuleModal(true)
                          }}
                          className="p-1 text-gray-400 hover:text-gray-600"
                        >
                          <Edit className="h-4 w-4" />
                        </button>
                        <button className="p-1 text-gray-400 hover:text-red-600">
                          <Trash2 className="h-4 w-4" />
                        </button>
                      </div>
                    </div>
                    <div className="flex items-center space-x-4 mt-2 text-xs text-gray-500">
                      <span>For: {rule.for}</span>
                      <span>Updated: {new Date(rule.updated_at).toLocaleDateString()}</span>
                    </div>
                  </div>
                ))
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {showRuleModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-2xl">
            <h2 className="text-xl font-bold mb-4">
              {editingRule ? 'Edit Alert Rule' : 'Add Alert Rule'}
            </h2>
            {/* Modal content would go here */}
            <div className="flex justify-end space-x-2 mt-6">
              <button
                onClick={() => setShowRuleModal(false)}
                className="px-4 py-2 text-gray-600 border border-gray-300 rounded hover:bg-gray-50"
              >
                Cancel
              </button>
              <button
                onClick={() => setShowRuleModal(false)}
                className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
              >
                Save
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default Alerts
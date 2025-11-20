import React, { useState } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card'
import { Save, Shield, Database, Bell, PlayCircle } from 'lucide-react'
import { toast } from 'sonner'

const Settings: React.FC = () => {
  const [settings, setSettings] = useState({
    enableAuth: true,
    enableAudit: true,
    allowSnippetAnnotations: false,
    victoriaMetricsUrl: 'http://victoria-metrics:8428',
    victoriaLogsUrl: 'http://victoria-logs:9428',
    vmalertUrl: 'http://vmalert:8880',
    kubernetesNamespace: 'monitoring',
    logRetentionDays: 30,
    alertRetentionDays: 7
  })

  const [testing, setTesting] = useState({
    vm: false,
    vl: false,
    vmalert: false
  })

  const handleSave = async () => {
    try {
      // Save settings to backend
      toast.success('Settings saved successfully')
    } catch (error) {
      toast.error('Failed to save settings')
    }
  }

  const testConnection = async (service: 'vm' | 'vl' | 'vmalert') => {
    setTesting({...testing, [service]: true})
    
    try {
      // Test connection to service
      await new Promise(resolve => setTimeout(resolve, 1000)) // Simulate API call
      toast.success(`${service.toUpperCase()} connection successful`)
    } catch (error) {
      toast.error(`${service.toUpperCase()} connection failed`)
    } finally {
      setTesting({...testing, [service]: false})
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900">Settings</h1>
        <button
          onClick={handleSave}
          className="flex items-center space-x-2 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700"
        >
          <Save className="h-4 w-4" />
          <span>Save Settings</span>
        </button>
      </div>

      <div className="grid gap-6">
        <Card>
          <CardHeader>
            <div className="flex items-center space-x-2">
              <Shield className="h-5 w-5 text-blue-600" />
              <CardTitle>Security Settings</CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <label className="font-medium">Enable Authentication</label>
                  <p className="text-sm text-gray-500">Require login to access the dashboard</p>
                </div>
                <label className="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    checked={settings.enableAuth}
                    onChange={(e) => setSettings({...settings, enableAuth: e.target.checked})}
                    className="sr-only peer"
                  />
                  <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                </label>
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <label className="font-medium">Enable Audit Logging</label>
                  <p className="text-sm text-gray-500">Log all configuration changes</p>
                </div>
                <label className="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    checked={settings.enableAudit}
                    onChange={(e) => setSettings({...settings, enableAudit: e.target.checked})}
                    className="sr-only peer"
                  />
                  <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                </label>
              </div>

              <div className="flex items-center justify-between">
                <div>
                  <label className="font-medium">Allow Snippet Annotations</label>
                  <p className="text-sm text-gray-500">Enable per-Ingress WAF configuration</p>
                </div>
                <label className="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    checked={settings.allowSnippetAnnotations}
                    onChange={(e) => setSettings({...settings, allowSnippetAnnotations: e.target.checked})}
                    className="sr-only peer"
                  />
                  <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                </label>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center space-x-2">
              <Database className="h-5 w-5 text-blue-600" />
              <CardTitle>External Services</CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Victoria Metrics URL
                </label>
                <div className="flex space-x-2">
                  <input
                    type="url"
                    value={settings.victoriaMetricsUrl}
                    onChange={(e) => setSettings({...settings, victoriaMetricsUrl: e.target.value})}
                    className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="http://victoria-metrics:8428"
                  />
                  <button
                    onClick={() => testConnection('vm')}
                    disabled={testing.vm}
                    className="flex items-center space-x-2 px-3 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 disabled:opacity-50"
                  >
                    <PlayCircle className={`h-4 w-4 ${testing.vm ? 'animate-spin' : ''}`} />
                    <span>Test</span>
                  </button>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Victoria Logs URL
                </label>
                <div className="flex space-x-2">
                  <input
                    type="url"
                    value={settings.victoriaLogsUrl}
                    onChange={(e) => setSettings({...settings, victoriaLogsUrl: e.target.value})}
                    className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="http://victoria-logs:9428"
                  />
                  <button
                    onClick={() => testConnection('vl')}
                    disabled={testing.vl}
                    className="flex items-center space-x-2 px-3 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 disabled:opacity-50"
                  >
                    <PlayCircle className={`h-4 w-4 ${testing.vl ? 'animate-spin' : ''}`} />
                    <span>Test</span>
                  </button>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  vmalert URL
                </label>
                <div className="flex space-x-2">
                  <input
                    type="url"
                    value={settings.vmalertUrl}
                    onChange={(e) => setSettings({...settings, vmalertUrl: e.target.value})}
                    className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                    placeholder="http://vmalert:8880"
                  />
                  <button
                    onClick={() => testConnection('vmalert')}
                    disabled={testing.vmalert}
                    className="flex items-center space-x-2 px-3 py-2 bg-gray-100 text-gray-700 rounded-md hover:bg-gray-200 disabled:opacity-50"
                  >
                    <PlayCircle className={`h-4 w-4 ${testing.vmalert ? 'animate-spin' : ''}`} />
                    <span>Test</span>
                  </button>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Kubernetes Namespace
                </label>
                <input
                  type="text"
                  value={settings.kubernetesNamespace}
                  onChange={(e) => setSettings({...settings, kubernetesNamespace: e.target.value})}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="monitoring"
                />
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <div className="flex items-center space-x-2">
              <Bell className="h-5 w-5 text-blue-600" />
              <CardTitle>Data Retention</CardTitle>
            </div>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Log Retention (days)
                </label>
                <input
                  type="number"
                  value={settings.logRetentionDays}
                  onChange={(e) => setSettings({...settings, logRetentionDays: parseInt(e.target.value)})}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  min="1"
                  max="365"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Alert Retention (days)
                </label>
                <input
                  type="number"
                  value={settings.alertRetentionDays}
                  onChange={(e) => setSettings({...settings, alertRetentionDays: parseInt(e.target.value)})}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                  min="1"
                  max="90"
                />
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

export default Settings
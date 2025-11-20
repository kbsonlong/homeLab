import React, { useState, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/Card'
import { Shield, Plus, Edit, Trash2, Play, Pause } from 'lucide-react'
import { useWAFStore } from '../stores/waf'
import { toast } from 'sonner'

interface PolicyFormData {
  host: string
  mode: 'On' | 'DetectionOnly' | 'Off'
  enable_crs: boolean
  exceptions: {
    paths: string[]
    methods: string[]
    ip_allow: string[]
    headers_allow: Record<string, string>
  }
  custom_rules: Array<{
    id: string
    name: string
    rule: string
    description: string
    enabled: boolean
  }>
}

const Policies: React.FC = () => {
  const { status, fetchStatus } = useWAFStore()
  const [selectedHost, setSelectedHost] = useState<string>('')
  const [showEditModal, setShowEditModal] = useState(false)
  const [editingPolicy, setEditingPolicy] = useState<any>(null)

  useEffect(() => {
    fetchStatus()
  }, [fetchStatus])

  const getModeColor = (mode: string) => {
    switch (mode) {
      case 'On': return 'bg-green-100 text-green-800'
      case 'DetectionOnly': return 'bg-yellow-100 text-yellow-800'
      case 'Off': return 'bg-red-100 text-red-800'
      default: return 'bg-gray-100 text-gray-800'
    }
  }

  const handleModeChange = async (host: string, newMode: string) => {
    try {
      const response = await fetch('/api/waf/mode', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ host, mode: newMode })
      })
      
      if (response.ok) {
        toast.success('WAF mode updated successfully')
        fetchStatus()
      } else {
        toast.error('Failed to update WAF mode')
      }
    } catch (error) {
      toast.error('Error updating WAF mode')
    }
  }

  const handleApplyConfig = async (host: string) => {
    try {
      const response = await fetch('/api/waf/apply', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ host, strategy: 'annotation' })
      })
      
      if (response.ok) {
        toast.success('Configuration applied successfully')
      } else {
        toast.error('Failed to apply configuration')
      }
    } catch (error) {
      toast.error('Error applying configuration')
    }
  }

  const handleSavePolicy = async (formData: PolicyFormData) => {
    try {
      // For new policies, we need to create them step by step
      if (!editingPolicy) {
        // Step 1: Set WAF mode for the new host
        const modeResponse = await fetch('/api/waf/mode', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ 
            host: formData.host, 
            mode: formData.mode,
            enable_crs: formData.enable_crs 
          })
        })
        
        if (!modeResponse.ok) {
          toast.error('Failed to create policy')
          return
        }

        // Step 2: Set exceptions if any
        if (formData.exceptions.paths.length > 0 || 
            formData.exceptions.methods.length > 0 || 
            formData.exceptions.ip_allow.length > 0) {
          const exceptionsResponse = await fetch('/api/waf/exceptions', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              host: formData.host,
              paths: formData.exceptions.paths,
              methods: formData.exceptions.methods,
              ip_allow: formData.exceptions.ip_allow,
              enabled: true
            })
          })
          
          if (!exceptionsResponse.ok) {
            toast.error('Failed to set exceptions')
            return
          }
        }

        // Step 3: Add custom rules if any
        if (formData.custom_rules.length > 0) {
          const rulesResponse = await fetch('/api/waf/rules', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
              host: formData.host,
              rules: formData.custom_rules
            })
          })
          
          if (!rulesResponse.ok) {
            toast.error('Failed to add custom rules')
            return
          }
        }

        toast.success('Policy created successfully')
      } else {
        // For existing policies, update mode and other settings
        const modeResponse = await fetch('/api/waf/mode', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ 
            host: formData.host, 
            mode: formData.mode,
            enable_crs: formData.enable_crs 
          })
        })
        
        if (modeResponse.ok) {
          toast.success('Policy updated successfully')
        } else {
          toast.error('Failed to update policy')
          return
        }
      }
      
      setShowEditModal(false)
      fetchStatus()
    } catch (error) {
      toast.error('Error saving policy')
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-2xl font-bold text-gray-900">WAF Policies</h1>
        <button
          onClick={() => setShowEditModal(true)}
          className="flex items-center space-x-2 bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700 transition-colors"
        >
          <Plus className="h-4 w-4" />
          <span>Add Policy</span>
        </button>
      </div>

      <div className="grid gap-6">
        {Object.entries(status?.host_policies || {}).map(([host, policy]) => (
          <Card key={host}>
            <CardHeader>
              <div className="flex justify-between items-center">
                <CardTitle>{host}</CardTitle>
                <div className="flex items-center space-x-2">
                  <span className={`px-3 py-1 rounded-full text-xs font-medium ${getModeColor(policy.mode)}`}>
                    {policy.mode}
                  </span>
                  <button
                    onClick={() => handleApplyConfig(host)}
                    className="bg-green-600 text-white px-3 py-1 rounded text-xs hover:bg-green-700 transition-colors"
                  >
                    Apply
                  </button>
                </div>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                <div className="flex items-center space-x-4">
                  <label className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      checked={policy.enable_crs}
                      onChange={(e) => {
                        // Handle CRS toggle
                      }}
                      className="rounded border-gray-300"
                    />
                    <span className="text-sm">Enable OWASP CRS</span>
                  </label>
                </div>

                <div>
                  <h4 className="text-sm font-medium text-gray-700 mb-2">Mode</h4>
                  <div className="flex space-x-2">
                    {['On', 'DetectionOnly', 'Off'].map((mode) => (
                      <button
                        key={mode}
                        onClick={() => handleModeChange(host, mode)}
                        className={`px-3 py-1 rounded text-sm transition-colors ${
                          policy.mode === mode
                            ? 'bg-blue-600 text-white'
                            : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                        }`}
                      >
                        {mode}
                      </button>
                    ))}
                  </div>
                </div>

                <div>
                  <h4 className="text-sm font-medium text-gray-700 mb-2">Exceptions</h4>
                  <div className="text-sm text-gray-600">
                    <p>Paths: {policy.exceptions.paths.length}</p>
                    <p>IP Allow: {policy.exceptions.ip_allow.length}</p>
                    <p>Methods: {policy.exceptions.methods.length}</p>
                  </div>
                </div>

                <div>
                  <h4 className="text-sm font-medium text-gray-700 mb-2">Custom Rules</h4>
                  <div className="text-sm text-gray-600">
                    {policy.custom_rules.length} rules configured
                  </div>
                </div>

                <div className="flex justify-between items-center pt-4 border-t">
                  <div className="text-xs text-gray-500">
                    Last updated: {new Date(policy.updated_at).toLocaleString()}
                  </div>
                  <div className="flex space-x-2">
                    <button
                      onClick={() => {
                        setEditingPolicy(policy)
                        setShowEditModal(true)
                      }}
                      className="flex items-center space-x-1 text-blue-600 hover:text-blue-800"
                    >
                      <Edit className="h-4 w-4" />
                      <span className="text-sm">Edit</span>
                    </button>
                  </div>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      {showEditModal && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto">
            <h2 className="text-xl font-bold mb-4">
              {editingPolicy ? 'Edit Policy' : 'Add New Policy'}
            </h2>
            
            <PolicyForm 
              policy={editingPolicy}
              onSave={handleSavePolicy}
              onCancel={() => setShowEditModal(false)}
            />
          </div>
        </div>
      )}
    </div>
  )
}

export default Policies

interface PolicyFormProps {
  policy: any
  onSave: (data: PolicyFormData) => void
  onCancel: () => void
}

const PolicyForm: React.FC<PolicyFormProps> = ({ policy, onSave, onCancel }) => {
  const [formData, setFormData] = useState<PolicyFormData>({
    host: policy?.host || '',
    mode: policy?.mode || 'On',
    enable_crs: policy?.enable_crs || true,
    exceptions: {
      paths: policy?.exceptions?.paths || [],
      methods: policy?.exceptions?.methods || [],
      ip_allow: policy?.exceptions?.ip_allow || [],
      headers_allow: policy?.exceptions?.headers_allow || {}
    },
    custom_rules: policy?.custom_rules || []
  })

  const [newException, setNewException] = useState({
    path: '',
    method: '',
    ip: '',
    header_key: '',
    header_value: ''
  })

  const [newRule, setNewRule] = useState({
    name: '',
    rule: '',
    description: ''
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!formData.host) {
      toast.error('Host is required')
      return
    }
    onSave(formData)
  }

  const addException = (type: 'path' | 'method' | 'ip' | 'header') => {
    if (type === 'path' && newException.path) {
      setFormData(prev => ({
        ...prev,
        exceptions: {
          ...prev.exceptions,
          paths: [...prev.exceptions.paths, newException.path]
        }
      }))
      setNewException(prev => ({ ...prev, path: '' }))
    } else if (type === 'method' && newException.method) {
      setFormData(prev => ({
        ...prev,
        exceptions: {
          ...prev.exceptions,
          methods: [...prev.exceptions.methods, newException.method]
        }
      }))
      setNewException(prev => ({ ...prev, method: '' }))
    } else if (type === 'ip' && newException.ip) {
      setFormData(prev => ({
        ...prev,
        exceptions: {
          ...prev.exceptions,
          ip_allow: [...prev.exceptions.ip_allow, newException.ip]
        }
      }))
      setNewException(prev => ({ ...prev, ip: '' }))
    } else if (type === 'header' && newException.header_key && newException.header_value) {
      setFormData(prev => ({
        ...prev,
        exceptions: {
          ...prev.exceptions,
          headers_allow: {
            ...prev.exceptions.headers_allow,
            [newException.header_key]: newException.header_value
          }
        }
      }))
      setNewException(prev => ({ ...prev, header_key: '', header_value: '' }))
    }
  }

  const removeException = (type: 'path' | 'method' | 'ip' | 'header', value: string) => {
    if (type === 'path') {
      setFormData(prev => ({
        ...prev,
        exceptions: {
          ...prev.exceptions,
          paths: prev.exceptions.paths.filter(p => p !== value)
        }
      }))
    } else if (type === 'method') {
      setFormData(prev => ({
        ...prev,
        exceptions: {
          ...prev.exceptions,
          methods: prev.exceptions.methods.filter(m => m !== value)
        }
      }))
    } else if (type === 'ip') {
      setFormData(prev => ({
        ...prev,
        exceptions: {
          ...prev.exceptions,
          ip_allow: prev.exceptions.ip_allow.filter(ip => ip !== value)
        }
      }))
    } else if (type === 'header') {
      setFormData(prev => {
        const newHeaders = { ...prev.exceptions.headers_allow }
        delete newHeaders[value]
        return {
          ...prev,
          exceptions: {
            ...prev.exceptions,
            headers_allow: newHeaders
          }
        }
      })
    }
  }

  const addCustomRule = () => {
    if (newRule.name && newRule.rule) {
      const rule = {
        id: Date.now().toString(),
        name: newRule.name,
        rule: newRule.rule,
        description: newRule.description,
        enabled: true
      }
      setFormData(prev => ({
        ...prev,
        custom_rules: [...prev.custom_rules, rule]
      }))
      setNewRule({ name: '', rule: '', description: '' })
    }
  }

  const removeCustomRule = (id: string) => {
    setFormData(prev => ({
      ...prev,
      custom_rules: prev.custom_rules.filter(r => r.id !== id)
    }))
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      <div className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">Host</label>
          <input
            type="text"
            value={formData.host}
            onChange={(e) => setFormData(prev => ({ ...prev, host: e.target.value }))}
            placeholder="example.com"
            className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            disabled={!!policy}
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-1">WAF Mode</label>
          <div className="flex space-x-2">
            {['On', 'DetectionOnly', 'Off'].map((mode) => (
              <button
                key={mode}
                type="button"
                onClick={() => setFormData(prev => ({ ...prev, mode: mode as any }))}
                className={`px-3 py-1 rounded text-sm transition-colors ${
                  formData.mode === mode
                    ? 'bg-blue-600 text-white'
                    : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                }`}
              >
                {mode}
              </button>
            ))}
          </div>
        </div>

        <div>
          <label className="flex items-center space-x-2">
            <input
              type="checkbox"
              checked={formData.enable_crs}
              onChange={(e) => setFormData(prev => ({ ...prev, enable_crs: e.target.checked }))}
              className="rounded border-gray-300"
            />
            <span className="text-sm font-medium text-gray-700">Enable OWASP CRS</span>
          </label>
        </div>
      </div>

      <div className="space-y-4">
        <h3 className="text-lg font-medium text-gray-900">Exceptions</h3>
        
        <div className="space-y-3">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Path Exceptions</label>
            <div className="flex space-x-2 mb-2">
              <input
                type="text"
                value={newException.path}
                onChange={(e) => setNewException(prev => ({ ...prev, path: e.target.value }))}
                placeholder="/api/health"
                className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <button
                type="button"
                onClick={() => addException('path')}
                className="px-3 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
              >
                Add
              </button>
            </div>
            <div className="flex flex-wrap gap-2">
              {formData.exceptions.paths.map((path) => (
                <span key={path} className="inline-flex items-center px-2 py-1 bg-gray-100 text-gray-700 rounded text-sm">
                  {path}
                  <button
                    type="button"
                    onClick={() => removeException('path', path)}
                    className="ml-1 text-red-600 hover:text-red-800"
                  >
                    ×
                  </button>
                </span>
              ))}
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Method Exceptions</label>
            <div className="flex space-x-2 mb-2">
              <select
                value={newException.method}
                onChange={(e) => setNewException(prev => ({ ...prev, method: e.target.value }))}
                className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">Select method</option>
                <option value="GET">GET</option>
                <option value="POST">POST</option>
                <option value="PUT">PUT</option>
                <option value="DELETE">DELETE</option>
                <option value="PATCH">PATCH</option>
                <option value="HEAD">HEAD</option>
                <option value="OPTIONS">OPTIONS</option>
              </select>
              <button
                type="button"
                onClick={() => addException('method')}
                className="px-3 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
              >
                Add
              </button>
            </div>
            <div className="flex flex-wrap gap-2">
              {formData.exceptions.methods.map((method) => (
                <span key={method} className="inline-flex items-center px-2 py-1 bg-gray-100 text-gray-700 rounded text-sm">
                  {method}
                  <button
                    type="button"
                    onClick={() => removeException('method', method)}
                    className="ml-1 text-red-600 hover:text-red-800"
                  >
                    ×
                  </button>
                </span>
              ))}
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">IP Allowlist</label>
            <div className="flex space-x-2 mb-2">
              <input
                type="text"
                value={newException.ip}
                onChange={(e) => setNewException(prev => ({ ...prev, ip: e.target.value }))}
                placeholder="192.168.1.0/24"
                className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <button
                type="button"
                onClick={() => addException('ip')}
                className="px-3 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
              >
                Add
              </button>
            </div>
            <div className="flex flex-wrap gap-2">
              {formData.exceptions.ip_allow.map((ip) => (
                <span key={ip} className="inline-flex items-center px-2 py-1 bg-gray-100 text-gray-700 rounded text-sm">
                  {ip}
                  <button
                    type="button"
                    onClick={() => removeException('ip', ip)}
                    className="ml-1 text-red-600 hover:text-red-800"
                  >
                    ×
                  </button>
                </span>
              ))}
            </div>
          </div>
        </div>
      </div>

      <div className="space-y-4">
        <h3 className="text-lg font-medium text-gray-900">Custom Rules</h3>
        <div className="space-y-3">
          <div className="grid grid-cols-3 gap-2">
            <input
              type="text"
              value={newRule.name}
              onChange={(e) => setNewRule(prev => ({ ...prev, name: e.target.value }))}
              placeholder="Rule name"
              className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <input
              type="text"
              value={newRule.rule}
              onChange={(e) => setNewRule(prev => ({ ...prev, rule: e.target.value }))}
              placeholder="SecRule ..."
              className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <input
              type="text"
              value={newRule.description}
              onChange={(e) => setNewRule(prev => ({ ...prev, description: e.target.value }))}
              placeholder="Description"
              className="px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <button
            type="button"
            onClick={addCustomRule}
            className="px-3 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Add Rule
          </button>
        </div>
        
        <div className="space-y-2">
          {formData.custom_rules.map((rule) => (
            <div key={rule.id} className="flex items-center justify-between p-3 bg-gray-50 rounded">
              <div className="flex-1">
                <div className="font-medium text-sm">{rule.name}</div>
                <div className="text-xs text-gray-600 font-mono">{rule.rule}</div>
                {rule.description && (
                  <div className="text-xs text-gray-500">{rule.description}</div>
                )}
              </div>
              <button
                type="button"
                onClick={() => removeCustomRule(rule.id)}
                className="text-red-600 hover:text-red-800"
              >
                Remove
              </button>
            </div>
          ))}
        </div>
      </div>

      <div className="flex justify-end space-x-2 pt-4 border-t">
        <button
          type="button"
          onClick={onCancel}
          className="px-4 py-2 text-gray-600 border border-gray-300 rounded hover:bg-gray-50"
        >
          Cancel
        </button>
        <button
          type="submit"
          className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
        >
          Save Policy
        </button>
      </div>
    </form>
  )
}
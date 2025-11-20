import React, { useEffect, useState } from 'react'
import { useWAFStore } from '../stores/waf'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card'
import { Badge } from '../components/ui/badge'
import { Button } from '../components/ui/button'
import { ScrollArea } from '../components/ui/scroll-area'
import { format } from 'date-fns'
import { Clock, User, FileText, AlertCircle } from 'lucide-react'

const Audit: React.FC = () => {
  const { auditLogs, loading, fetchAuditLogs } = useWAFStore()
  const [selectedLog, setSelectedLog] = useState<any>(null)

  useEffect(() => {
    fetchAuditLogs()
  }, [fetchAuditLogs])

  const getActionColor = (action: string) => {
    switch (action) {
      case 'UPDATE_MODE':
        return 'bg-blue-100 text-blue-800'
      case 'UPDATE_EXCEPTIONS':
        return 'bg-yellow-100 text-yellow-800'
      case 'UPDATE_RULES':
        return 'bg-purple-100 text-purple-800'
      case 'APPLY_CONFIGURATION':
        return 'bg-green-100 text-green-800'
      default:
        return 'bg-gray-100 text-gray-800'
    }
  }

  const formatAction = (action: string) => {
    return action.replace(/_/g, ' ').toLowerCase()
      .replace(/\b\w/g, l => l.toUpperCase())
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="text-lg">Loading audit logs...</div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold">Audit Logs</h1>
          <p className="text-gray-600">Track all WAF configuration changes</p>
        </div>
        <Button onClick={() => fetchAuditLogs()} variant="outline">
          Refresh
        </Button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card>
          <CardHeader>
            <CardTitle>Recent Changes</CardTitle>
            <CardDescription>Latest WAF configuration modifications</CardDescription>
          </CardHeader>
          <CardContent>
            <ScrollArea className="h-[600px]">
              <div className="space-y-4">
                {auditLogs?.entries?.map((log) => (
                  <div
                    key={log.id}
                    className="p-4 border rounded-lg hover:bg-gray-50 cursor-pointer transition-colors"
                    onClick={() => setSelectedLog(log)}
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-2 mb-2">
                          <Badge className={getActionColor(log.action)}>
                            {formatAction(log.action)}
                          </Badge>
                          <span className="text-sm text-gray-500">{log.resource_type}</span>
                        </div>
                        <div className="flex items-center gap-4 text-sm text-gray-600">
                          <div className="flex items-center gap-1">
                            <User className="w-4 h-4" />
                            <span>{log.user}</span>
                          </div>
                          <div className="flex items-center gap-1">
                            <Clock className="w-4 h-4" />
                            <span>{format(new Date(log.created_at), 'MMM dd, HH:mm')}</span>
                          </div>
                        </div>
                        <div className="flex items-center gap-1 mt-1">
                          <FileText className="w-4 h-4 text-gray-400" />
                          <span className="text-sm text-gray-600">{log.resource_id}</span>
                        </div>
                        {log.details && (
                          <p className="text-sm text-gray-600 mt-2">{log.details}</p>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
                {auditLogs?.entries?.length === 0 && (
                  <div className="text-center py-8 text-gray-500">
                    <AlertCircle className="w-12 h-12 mx-auto mb-4 opacity-50" />
                    <p>No audit logs found</p>
                  </div>
                )}
              </div>
            </ScrollArea>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Change Details</CardTitle>
            <CardDescription>Detailed information about the selected change</CardDescription>
          </CardHeader>
          <CardContent>
            {selectedLog ? (
              <div className="space-y-4">
                <div>
                  <h3 className="font-semibold mb-2">Action Information</h3>
                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-gray-600">Action:</span>
                      <Badge className={getActionColor(selectedLog.action)}>
                        {formatAction(selectedLog.action)}
                      </Badge>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Resource Type:</span>
                      <span>{selectedLog.resource_type}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Resource ID:</span>
                      <span className="font-mono text-xs">{selectedLog.resource_id}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">User:</span>
                      <span>{selectedLog.user}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Timestamp:</span>
                      <span>{format(new Date(selectedLog.created_at), 'PPpp')}</span>
                    </div>
                  </div>
                </div>

                {selectedLog.details && (
                  <div>
                    <h3 className="font-semibold mb-2">Details</h3>
                    <p className="text-sm bg-gray-50 p-3 rounded">{selectedLog.details}</p>
                  </div>
                )}

                <div>
                  <h3 className="font-semibold mb-2">Configuration Changes</h3>
                  <div className="space-y-3">
                    <div>
                      <h4 className="text-sm font-medium text-gray-700 mb-1">Previous Value</h4>
                      <pre className="text-xs bg-gray-50 p-3 rounded overflow-auto max-h-40">
                        {JSON.stringify(selectedLog.old_value, null, 2)}
                      </pre>
                    </div>
                    <div>
                      <h4 className="text-sm font-medium text-gray-700 mb-1">New Value</h4>
                      <pre className="text-xs bg-gray-50 p-3 rounded overflow-auto max-h-40">
                        {JSON.stringify(selectedLog.new_value, null, 2)}
                      </pre>
                    </div>
                  </div>
                </div>
              </div>
            ) : (
              <div className="text-center py-8 text-gray-500">
                <FileText className="w-12 h-12 mx-auto mb-4 opacity-50" />
                <p>Select an audit log to view details</p>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

export default Audit
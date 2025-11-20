import React from 'react'
import { Link, useLocation } from 'react-router-dom'
import { 
  Shield, 
  FileText, 
  Search, 
  AlertTriangle, 
  Settings,
  Activity,
  History
} from 'lucide-react'
import { cn } from '../utils/cn'

const menuItems = [
  { name: 'Dashboard', href: '/', icon: Activity },
  { name: 'Policies', href: '/policies', icon: Shield },
  { name: 'Logs', href: '/logs', icon: Search },
  { name: 'Alerts', href: '/alerts', icon: AlertTriangle },
  { name: 'Audit', href: '/audit', icon: History },
  { name: 'Settings', href: '/settings', icon: Settings },
]

const Sidebar: React.FC = () => {
  const location = useLocation()

  return (
    <div className="w-64 bg-white shadow-lg">
      <div className="flex items-center justify-center h-16 bg-blue-600">
        <div className="flex items-center space-x-2">
          <Shield className="h-8 w-8 text-white" />
          <span className="text-xl font-bold text-white">WAF Admin</span>
        </div>
      </div>
      
      <nav className="mt-8">
        <ul className="space-y-2 px-4">
          {menuItems.map((item) => {
            const Icon = item.icon
            const isActive = location.pathname === item.href
            
            return (
              <li key={item.name}>
                <Link
                  to={item.href}
                  className={cn(
                    "flex items-center space-x-3 px-4 py-3 rounded-lg transition-colors",
                    isActive 
                      ? "bg-blue-100 text-blue-700 border-r-4 border-blue-700" 
                      : "text-gray-700 hover:bg-gray-100"
                  )}
                >
                  <Icon className="h-5 w-5" />
                  <span className="font-medium">{item.name}</span>
                </Link>
              </li>
            )
          })}
        </ul>
      </nav>
    </div>
  )
}

export default Sidebar
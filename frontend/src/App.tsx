import React from 'react'
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { Toaster } from 'sonner'
import Dashboard from './pages/Dashboard'
import Policies from './pages/Policies'
import Logs from './pages/Logs'
import Alerts from './pages/Alerts'
import Settings from './pages/Settings'
import Audit from './pages/Audit'
import Sidebar from './components/Sidebar'
import Header from './components/Header'

function App() {
  return (
    <Router>
      <div className="flex h-screen bg-gray-100">
        <Sidebar />
        <div className="flex-1 flex flex-col overflow-hidden">
          <Header />
          <main className="flex-1 overflow-x-hidden overflow-y-auto bg-gray-100">
            <div className="container mx-auto px-6 py-8">
              <Routes>
                <Route path="/" element={<Dashboard />} />
                <Route path="/dashboard" element={<Dashboard />} />
                <Route path="/policies" element={<Policies />} />
                <Route path="/logs" element={<Logs />} />
                <Route path="/alerts" element={<Alerts />} />
                <Route path="/audit" element={<Audit />} />
                <Route path="/settings" element={<Settings />} />
              </Routes>
            </div>
          </main>
        </div>
        <Toaster position="top-right" />
      </div>
    </Router>
  )
}

export default App

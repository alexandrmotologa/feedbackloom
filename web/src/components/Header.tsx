import React from 'react'
import { DashboardStats } from '../types'

interface HeaderProps {
  stats: DashboardStats | null
  activeTab: string
  setActiveTab: (tab: string) => void
  onRefresh: () => void
}

export const Header: React.FC<HeaderProps> = ({ stats, activeTab, setActiveTab, onRefresh }) => {
  const tabs = [
    { id: 'opportunities', label: 'ARR Opportunities', icon: '⚡' },
    { id: 'sandbox', label: 'Live Ingestion Sandbox', icon: '🔬' },
    { id: 'prd-studio', label: 'PRD Studio', icon: '📄' },
    { id: 'evidence', label: 'Evidence Stream', icon: '💬' },
    { id: 'connectors', label: 'Data Connectors', icon: '🔌' },
    { id: 'closed-loop', label: 'Closed-Loop Updates', icon: '📬' },
  ]

  const formatCurrency = (val: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      maximumFractionDigits: 0,
    }).format(val)
  }

  return (
    <header className="border-b border-[#1e293b] bg-[#0c101a] sticky top-0 z-30">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo and System Status */}
          <div className="flex items-center gap-3">
            <div className="w-9 h-9 rounded-lg bg-gradient-to-br from-blue-600 to-indigo-600 flex items-center justify-center font-bold text-white shadow-lg shadow-blue-500/20">
              FL
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-bold text-base tracking-tight text-white">feedbackloom</span>
                <span className="px-1.5 py-0.5 rounded text-[10px] font-mono font-semibold bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
                  DAEMON ONLINE
                </span>
              </div>
              <p className="text-[11px] text-[#64748b]">Revenue-Driven Product Architect Engine</p>
            </div>
          </div>

          {/* Real-time KPI Ribbon */}
          {stats && (
            <div className="hidden lg:flex items-center gap-5 text-xs bg-[#121826] px-4 py-1.5 rounded-lg border border-[#1e293b]">
              <div>
                <span className="text-[#64748b] block text-[10px] uppercase font-semibold">Demanded ARR</span>
                <span className="font-bold text-emerald-400 font-mono">{formatCurrency(stats.total_demanded_arr_usd)}</span>
              </div>
              {stats.total_lost_pipeline_arr_usd ? (
                <>
                  <div className="h-6 w-px bg-[#1e293b]" />
                  <div>
                    <span className="text-rose-400/80 block text-[10px] uppercase font-semibold">Pipeline at Risk</span>
                    <span className="font-bold text-rose-400 font-mono">{formatCurrency(stats.total_lost_pipeline_arr_usd)}</span>
                  </div>
                </>
              ) : null}
              <div className="h-6 w-px bg-[#1e293b]" />
              <div>
                <span className="text-[#64748b] block text-[10px] uppercase font-semibold">Portfolio ARR</span>
                <span className="font-semibold text-slate-300 font-mono">{formatCurrency(stats.portfolio_arr_usd)}</span>
              </div>
              <div className="h-6 w-px bg-[#1e293b]" />
              <div>
                <span className="text-[#64748b] block text-[10px] uppercase font-semibold">Opportunities</span>
                <span className="font-semibold text-blue-400 font-mono">{stats.total_opportunities}</span>
              </div>
              <div className="h-6 w-px bg-[#1e293b]" />
              <div>
                <span className="text-[#64748b] block text-[10px] uppercase font-semibold">Feedback</span>
                <span className="font-semibold text-purple-400 font-mono">{stats.total_feedback}</span>
              </div>
            </div>
          )}

          {/* Actions */}
          <div className="flex items-center gap-2">
            <button
              onClick={onRefresh}
              className="text-xs bg-[#161f30] hover:bg-[#1e293b] text-slate-300 border border-[#1e293b] px-3 py-1.5 rounded flex items-center gap-1.5 transition-colors"
            >
              <span>↻</span> Refresh
            </button>
          </div>
        </div>

        {/* Navigation Tabs */}
        <nav className="flex space-x-1 border-t border-[#1e293b]/50 pt-1 -mb-px overflow-x-auto">
          {tabs.map((tab) => {
            const isActive = activeTab === tab.id
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center gap-2 py-3 px-3.5 text-xs font-medium border-b-2 whitespace-nowrap transition-all ${
                  isActive
                    ? 'border-blue-500 text-white bg-blue-500/10'
                    : 'border-transparent text-[#94a3b8] hover:text-white hover:bg-slate-800/40'
                }`}
              >
                <span>{tab.icon}</span>
                <span>{tab.label}</span>
              </button>
            )
          })}
        </nav>
      </div>
    </header>
  )
}

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
    <header className="border-b border-slate-800/80 bg-[#0a0e17]/90 backdrop-blur-md sticky top-0 z-40 transition-all">
      <div className="max-w-[1520px] mx-auto px-4 sm:px-6 lg:px-8">
        {/* Top Header Row */}
        <div className="flex items-center justify-between py-3.5 border-b border-slate-800/50 gap-4">
          {/* Logo & Product Tagline */}
          <div className="flex items-center gap-3.5">
            <div className="w-10 h-10 rounded-xl bg-gradient-to-br from-blue-500 via-indigo-600 to-cyan-500 flex items-center justify-center font-bold text-white shadow-lg shadow-blue-500/25 ring-1 ring-white/20">
              <span className="font-mono text-sm tracking-tighter">FL</span>
            </div>
            <div>
              <div className="flex items-center gap-2.5">
                <span className="font-bold text-lg tracking-tight text-white font-sans">feedbackloom</span>
                <span className="px-2 py-0.5 rounded-full text-[10px] font-mono font-semibold bg-emerald-500/15 text-emerald-400 border border-emerald-500/30 flex items-center gap-1.5 shadow-sm">
                  <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
                  DAEMON ONLINE
                </span>
              </div>
              <p className="text-xs text-slate-400 font-medium">Revenue-Driven Product Architect &amp; Autonomous PRD Synthesizer</p>
            </div>
          </div>

          {/* Quick Metrics Bar in Header */}
          {stats && (
            <div className="hidden xl:flex items-center gap-3">
              <div className="bg-[#111726]/80 px-4 py-2 rounded-xl border border-slate-800/80 flex items-center gap-3 shadow-inner">
                <div>
                  <span className="text-[10px] text-slate-400 uppercase font-semibold tracking-wider block">Demanded ARR</span>
                  <span className="text-sm font-bold font-mono text-emerald-400">{formatCurrency(stats.total_demanded_arr_usd)}</span>
                </div>
                <div className="h-6 w-px bg-slate-800" />
                <div>
                  <span className="text-[10px] text-rose-400/90 uppercase font-semibold tracking-wider block">Pipeline at Risk</span>
                  <span className="text-sm font-bold font-mono text-rose-400">
                    {formatCurrency(stats.total_lost_pipeline_arr_usd || 0)}
                  </span>
                </div>
                <div className="h-6 w-px bg-slate-800" />
                <div>
                  <span className="text-[10px] text-slate-400 uppercase font-semibold tracking-wider block">Portfolio ARR</span>
                  <span className="text-sm font-semibold font-mono text-slate-300">{formatCurrency(stats.portfolio_arr_usd)}</span>
                </div>
                <div className="h-6 w-px bg-slate-800" />
                <div>
                  <span className="text-[10px] text-slate-400 uppercase font-semibold tracking-wider block">Signals</span>
                  <span className="text-sm font-semibold font-mono text-cyan-400">{stats.total_feedback} quotes</span>
                </div>
              </div>
            </div>
          )}

          {/* Header Action Button */}
          <div className="flex items-center gap-2">
            <button
              onClick={onRefresh}
              className="text-xs font-semibold bg-slate-900/80 hover:bg-slate-800 text-slate-300 hover:text-white border border-slate-800 hover:border-slate-700 px-3.5 py-2 rounded-xl flex items-center gap-2 shadow-sm transition-all"
            >
              <span>↻</span> Refresh Pipeline
            </button>
          </div>
        </div>

        {/* Navigation Tabs Bar */}
        <nav className="flex space-x-1.5 py-2 -mb-px overflow-x-auto no-scrollbar">
          {tabs.map((tab) => {
            const isActive = activeTab === tab.id
            return (
              <button
                key={tab.id}
                onClick={() => setActiveTab(tab.id)}
                className={`flex items-center gap-2 py-2 px-3.5 rounded-lg text-xs font-semibold whitespace-nowrap transition-all ${
                  isActive
                    ? 'bg-blue-600/20 text-blue-400 border border-blue-500/40 shadow-sm'
                    : 'text-slate-400 hover:text-slate-200 hover:bg-slate-900/60 border border-transparent'
                }`}
              >
                <span className="text-sm">{tab.icon}</span>
                <span>{tab.label}</span>
              </button>
            )
          })}
        </nav>
      </div>
    </header>
  )
}
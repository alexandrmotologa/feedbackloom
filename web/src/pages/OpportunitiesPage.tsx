import React, { useState } from 'react'
import { Opportunity } from '../types'
import { ARRWaterfall } from '../components/ARRWaterfall'
import { CompetitorMatrix } from '../components/CompetitorMatrix'
import { OpportunityCard } from '../components/OpportunityCard'

interface OpportunitiesPageProps {
  opportunities: Opportunity[]
  totalPortfolioARR: number
  onSelectPRD: (id: string) => void
  onStatusChange: (id: string, newStatus: string) => void
  onOpenClosedLoop: (opp: Opportunity) => void
  onRunClustering: (minClusterSize?: number, minPoints?: number) => void
  onEffortChange?: (id: string, sprints: number) => void
}

export const OpportunitiesPage: React.FC<OpportunitiesPageProps> = ({
  opportunities,
  totalPortfolioARR,
  onSelectPRD,
  onStatusChange,
  onOpenClosedLoop,
  onRunClustering,
  onEffortChange,
}) => {
  const [filterCategory, setFilterCategory] = useState<string>('ALL')
  const [searchTerm, setSearchTerm] = useState<string>('')
  const [granularity, setGranularity] = useState<'macro' | 'balanced' | 'micro'>('balanced')

  const categories = ['ALL', 'SECURITY', 'PERFORMANCE', 'INTEGRATIONS', 'USABILITY']

  const handleClusterWithGranularity = () => {
    switch (granularity) {
      case 'macro':
        onRunClustering(3, 2)
        break
      case 'micro':
        onRunClustering(2, 1)
        break
      default:
        onRunClustering(2, 2)
        break
    }
  }

  const filtered = opportunities.filter((opp) => {
    const matchCategory = filterCategory === 'ALL' || opp.category === filterCategory
    const matchSearch =
      opp.title.toLowerCase().includes(searchTerm.toLowerCase()) ||
      opp.description.toLowerCase().includes(searchTerm.toLowerCase())
    return matchCategory && matchSearch
  })

  return (
    <div className="space-y-6">
      {/* Waterfall Visualizer */}
      <ARRWaterfall opportunities={opportunities} totalPortfolioARR={totalPortfolioARR} />

      {/* Competitor Loss Attribution Matrix */}
      <CompetitorMatrix opportunities={opportunities} />

      {/* Filter and Actions Bar */}
      <div className="flex flex-col lg:flex-row items-center justify-between gap-4 bg-[#121826] p-4 rounded-xl border border-[#1e293b]">
        {/* Category Pills */}
        <div className="flex items-center gap-1.5 overflow-x-auto w-full lg:w-auto pb-1 lg:pb-0">
          {categories.map((cat) => (
            <button
              key={cat}
              onClick={() => setFilterCategory(cat)}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                filterCategory === cat
                  ? 'bg-blue-600 text-white shadow-sm'
                  : 'bg-[#182136] text-slate-400 hover:text-white hover:bg-[#202c47]'
              }`}
            >
              {cat}
            </button>
          ))}
        </div>

        {/* Search, Granularity Slider & Cluster Trigger */}
        <div className="flex flex-wrap items-center gap-2.5 w-full lg:w-auto">
          <input
            type="text"
            placeholder="Search opportunities..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="bg-[#0c101a] border border-[#1e293b] focus:border-blue-500 rounded-lg px-3 py-1.5 text-xs text-white placeholder-slate-500 w-full sm:w-48 outline-none transition-colors"
          />

          {/* Granularity Selector */}
          <div className="flex items-center gap-1.5 bg-[#0c101a] border border-[#1e293b] rounded-lg px-2.5 py-1">
            <span className="text-[10px] uppercase font-semibold text-[#64748b]">Granularity:</span>
            <select
              value={granularity}
              onChange={(e) => setGranularity(e.target.value as any)}
              className="bg-transparent text-xs text-slate-200 outline-none cursor-pointer font-medium"
            >
              <option value="macro" className="bg-[#0c101a] text-white">Macro (3+ items)</option>
              <option value="balanced" className="bg-[#0c101a] text-white">Balanced (2+ items)</option>
              <option value="micro" className="bg-[#0c101a] text-white">Micro (Fine-grain)</option>
            </select>
          </div>

          <button
            onClick={handleClusterWithGranularity}
            className="px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white whitespace-nowrap shadow-sm flex items-center gap-1.5 transition-all"
          >
            <span>⚡</span>
            <span>Re-Cluster HDBSCAN</span>
          </button>
        </div>
      </div>

      {/* Opportunity List */}
      <div className="space-y-3.5">
        {filtered.map((opp, idx) => (
          <OpportunityCard
            key={opp.id}
            opportunity={opp}
            rank={idx + 1}
            onSelectPRD={onSelectPRD}
            onStatusChange={onStatusChange}
            onOpenClosedLoop={onOpenClosedLoop}
            onEffortChange={onEffortChange}
          />
        ))}

        {filtered.length === 0 && (
          <div className="bg-[#121826] border border-[#1e293b] rounded-xl p-12 text-center text-slate-400 space-y-3">
            <div className="text-3xl">🔍</div>
            <div className="text-sm font-semibold text-white">No opportunities match the criteria</div>
            <p className="text-xs text-slate-500 max-w-sm mx-auto">
              Try adjusting your category filter, or run HDBSCAN clustering from the Data Connectors tab.
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
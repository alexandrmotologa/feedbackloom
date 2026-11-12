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
    <div className="space-y-6 pb-12">
      {/* Waterfall Visualizer */}
      <ARRWaterfall opportunities={opportunities} totalPortfolioARR={totalPortfolioARR} />

      {/* Competitor Loss Attribution Matrix */}
      <CompetitorMatrix opportunities={opportunities} />

      {/* Filter and Actions Bar */}
      <div className="flex flex-col lg:flex-row items-center justify-between gap-4 bg-[#0f1422]/90 p-4 sm:p-5 rounded-2xl border border-slate-800/80 shadow-md backdrop-blur-md">
        {/* Category Pills */}
        <div className="flex items-center gap-2 overflow-x-auto w-full lg:w-auto pb-1 lg:pb-0">
          {categories.map((cat) => (
            <button
              key={cat}
              onClick={() => setFilterCategory(cat)}
              className={`px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all ${
                filterCategory === cat
                  ? 'bg-blue-600 text-white shadow-sm ring-1 ring-blue-400/40'
                  : 'bg-slate-900/80 text-slate-400 hover:text-white hover:bg-slate-800 border border-slate-800/80'
              }`}
            >
              {cat}
            </button>
          ))}
        </div>

        {/* Search, Granularity Slider & Cluster Trigger */}
        <div className="flex flex-wrap items-center gap-3 w-full lg:w-auto">
          <input
            type="text"
            placeholder="Search opportunities..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="bg-slate-950 border border-slate-800 focus:border-blue-500 rounded-xl px-3.5 py-2 text-xs text-white placeholder-slate-500 w-full sm:w-52 outline-none transition-colors"
          />

          {/* Granularity Selector */}
          <div className="flex items-center gap-2 bg-slate-950 border border-slate-800 rounded-xl px-3 py-1.5">
            <span className="text-[10px] uppercase font-bold text-slate-400">Granularity:</span>
            <select
              value={granularity}
              onChange={(e) => setGranularity(e.target.value as any)}
              className="bg-transparent text-xs text-slate-200 outline-none cursor-pointer font-medium"
            >
              <option value="macro" className="bg-slate-950 text-white">Macro (Broad)</option>
              <option value="balanced" className="bg-slate-950 text-white">Balanced (Default)</option>
              <option value="micro" className="bg-slate-950 text-white">Micro (Granular)</option>
            </select>
          </div>

          <button
            onClick={handleClusterWithGranularity}
            className="px-4 py-2 rounded-xl text-xs font-semibold bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white whitespace-nowrap shadow-sm flex items-center gap-2 transition-all"
          >
            <span>⚡</span>
            <span>Re-Cluster HDBSCAN</span>
          </button>
        </div>
      </div>

      {/* Opportunity List */}
      <div className="space-y-4">
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
          <div className="bg-[#0f1422]/90 border border-slate-800/80 rounded-2xl p-14 text-center text-slate-400 space-y-3">
            <div className="text-4xl">🔍</div>
            <div className="text-base font-semibold text-white">No opportunities match the criteria</div>
            <p className="text-xs text-slate-400 max-w-sm mx-auto leading-relaxed">
              Try adjusting your category filter, or run HDBSCAN clustering from the Data Connectors tab to process unassigned customer quotes.
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
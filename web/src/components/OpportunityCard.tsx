import React from 'react'
import { Opportunity } from '../types'

interface OpportunityCardProps {
  opportunity: Opportunity
  rank: number
  onSelectPRD: (id: string) => void
  onStatusChange: (id: string, newStatus: string) => void
  onOpenClosedLoop: (opp: Opportunity) => void
  onEffortChange?: (id: string, sprints: number) => void
}

export const OpportunityCard: React.FC<OpportunityCardProps> = ({
  opportunity,
  rank,
  onSelectPRD,
  onStatusChange,
  onOpenClosedLoop,
  onEffortChange,
}) => {
  const formatUSD = (val: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      maximumFractionDigits: 0,
    }).format(val)
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'SHIPPED':
        return 'bg-emerald-500/15 text-emerald-300 border-emerald-500/40'
      case 'IN_PROGRESS':
        return 'bg-blue-500/15 text-blue-300 border-blue-500/40'
      case 'PLANNED':
        return 'bg-purple-500/15 text-purple-300 border-purple-500/40'
      case 'IN_PRD':
        return 'bg-amber-500/15 text-amber-300 border-amber-500/40'
      default:
        return 'bg-slate-800/60 text-slate-300 border-slate-700'
    }
  }

  const getCategoryColor = (cat: string) => {
    switch (cat) {
      case 'SECURITY':
        return 'text-rose-400 bg-rose-950/40 border-rose-800/50'
      case 'INTEGRATIONS':
        return 'text-indigo-400 bg-indigo-950/40 border-indigo-800/50'
      case 'PERFORMANCE':
        return 'text-amber-400 bg-amber-950/40 border-amber-800/50'
      default:
        return 'text-cyan-400 bg-cyan-950/40 border-cyan-800/50'
    }
  }

  return (
    <div
      className={`rounded-2xl p-6 sm:p-7 transition-all shadow-lg backdrop-blur-md ${
        opportunity.urgent_renewal_alert
          ? 'bg-gradient-to-r from-[#17101c]/90 to-[#0f1422]/90 border border-rose-900/60 shadow-rose-950/15'
          : 'bg-[#0f1422]/90 border border-slate-800/80 hover:border-slate-700/80 shadow-black/30'
      }`}
    >
      {/* Top Banner Alert if Urgent Renewal */}
      {opportunity.urgent_renewal_alert && (
        <div className="mb-5 p-3 px-4 rounded-xl bg-rose-950/50 border border-rose-800/60 flex flex-col sm:flex-row sm:items-center justify-between gap-2 text-xs text-rose-300 shadow-sm">
          <div className="flex items-center gap-2.5 font-semibold">
            <span className="w-2.5 h-2.5 rounded-full bg-rose-500 animate-ping" />
            <span>CRITICAL RENEWAL ALERT: Enterprise account renewal within 45 days. High churn vulnerability.</span>
          </div>
          <span className="text-[10px] font-mono text-rose-300 font-bold uppercase tracking-wider bg-rose-900/40 px-2 py-0.5 rounded border border-rose-700/50 self-start sm:self-auto">
            Immediate Priority
          </span>
        </div>
      )}

      {/* Main Card Content */}
      <div className="flex flex-col lg:flex-row lg:items-start justify-between gap-6">
        <div className="space-y-3 flex-1">
          {/* Metadata badges row */}
          <div className="flex flex-wrap items-center gap-2">
            <span className="font-mono text-xs font-bold px-2.5 py-1 rounded-lg bg-slate-800 text-slate-300 border border-slate-700">
              #{String(rank).padStart(2, '0')}
            </span>

            <span className={`px-2.5 py-0.5 rounded-md text-xs font-semibold border ${getCategoryColor(opportunity.category)}`}>
              {opportunity.category}
            </span>

            {/* Competitor Tags */}
            {opportunity.competitor_mentions &&
              opportunity.competitor_mentions.map((comp) => (
                <span
                  key={comp}
                  className="px-2 py-0.5 rounded-md text-xs font-semibold bg-rose-500/10 text-rose-300 border border-rose-500/30 flex items-center gap-1.5"
                >
                  <span>⚔</span> {comp}
                </span>
              ))}

            {/* Velocity Tag */}
            {opportunity.velocity_growth_pct > 0 && (
              <span className="px-2.5 py-0.5 rounded-md text-xs font-mono font-medium bg-blue-500/10 text-blue-300 border border-blue-500/30">
                ↑ {opportunity.velocity_growth_pct.toFixed(0)}% velocity (14d)
              </span>
            )}
          </div>

          {/* Title */}
          <h3
            className="text-lg font-bold text-white tracking-tight hover:text-blue-400 cursor-pointer transition-colors"
            onClick={() => onSelectPRD(opportunity.id)}
          >
            {opportunity.title}
          </h3>

          {/* Description */}
          <p className="text-sm text-slate-300 leading-relaxed max-w-4xl">
            {opportunity.description}
          </p>
        </div>

        {/* Right Financial Stat Pillars */}
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 bg-slate-950/70 p-4 rounded-xl border border-slate-800/70 lg:w-auto">
          <div className="min-w-[90px]">
            <div className="text-[10px] uppercase font-semibold text-slate-400">Total ARR</div>
            <div className="text-base font-bold font-mono text-emerald-400 mt-0.5">
              {formatUSD(opportunity.total_arr_usd)}
            </div>
          </div>
          <div className="min-w-[90px]">
            <div className="text-[10px] uppercase font-semibold text-slate-400">ROI / Sprint</div>
            <div className="text-base font-bold font-mono text-cyan-400 mt-0.5">
              {formatUSD(opportunity.roi_index || opportunity.total_arr_usd / 2)}
            </div>
          </div>
          <div className="min-w-[90px]">
            <div className="text-[10px] uppercase font-semibold text-slate-400">Weighted WRS</div>
            <div className="text-base font-bold font-mono text-blue-400 mt-0.5">
              {opportunity.weighted_score.toFixed(0)}
            </div>
          </div>
          <div className="min-w-[90px]">
            <div className="text-[10px] uppercase font-semibold text-slate-400">Requesters</div>
            <div className="text-base font-bold font-mono text-slate-200 mt-0.5">
              {opportunity.account_count} <span className="text-xs text-slate-500">acc</span>
            </div>
          </div>
        </div>
      </div>

      {/* Footer Controls Row */}
      <div className="mt-5 pt-4 border-t border-slate-800/60 flex flex-wrap items-center justify-between gap-4 text-xs">
        <div className="flex flex-wrap items-center gap-4">
          {/* Status Dropdown */}
          <div className="flex items-center gap-2">
            <span className="text-slate-400 font-medium">Status:</span>
            <select
              value={opportunity.status}
              onChange={(e) => onStatusChange(opportunity.id, e.target.value)}
              className={`text-xs px-3 py-1.5 rounded-lg border font-semibold cursor-pointer outline-none bg-slate-950 ${getStatusColor(
                opportunity.status
              )}`}
            >
              <option value="DISCOVERED" className="bg-slate-950 text-white">DISCOVERED</option>
              <option value="IN_PRD" className="bg-slate-950 text-white">IN_PRD</option>
              <option value="PLANNED" className="bg-slate-950 text-white">PLANNED</option>
              <option value="IN_PROGRESS" className="bg-slate-950 text-white">IN_PROGRESS</option>
              <option value="SHIPPED" className="bg-slate-950 text-white">SHIPPED</option>
            </select>
          </div>

          {/* Effort Sprints Selector */}
          {onEffortChange && (
            <div className="flex items-center gap-2">
              <span className="text-slate-400 font-medium">Effort:</span>
              <select
                value={opportunity.effort_sprints || 2}
                onChange={(e) => onEffortChange(opportunity.id, parseInt(e.target.value))}
                className="text-xs px-2.5 py-1.5 rounded-lg border border-slate-700 bg-slate-950 text-slate-200 font-mono outline-none cursor-pointer"
              >
                <option value={1} className="bg-slate-950 text-white">1 Sprint (XS)</option>
                <option value={2} className="bg-slate-950 text-white">2 Sprints (M)</option>
                <option value={3} className="bg-slate-950 text-white">3 Sprints (M+)</option>
                <option value={4} className="bg-slate-950 text-white">4 Sprints (L)</option>
                <option value={6} className="bg-slate-950 text-white">6 Sprints (XL)</option>
                <option value={8} className="bg-slate-950 text-white">8 Sprints (XXL)</option>
              </select>
            </div>
          )}
        </div>

        {/* Action Buttons */}
        <div className="flex items-center gap-3">
          <button
            onClick={() => onSelectPRD(opportunity.id)}
            className="px-4 py-2 rounded-xl text-xs font-semibold bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white shadow-sm transition-all flex items-center gap-2"
          >
            <span>📄</span>
            <span>View Full PRD</span>
          </button>

          <button
            onClick={() => onOpenClosedLoop(opportunity)}
            className="px-4 py-2 rounded-xl text-xs font-semibold bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-white border border-slate-700/80 transition-all flex items-center gap-2"
          >
            <span>📬</span>
            <span>Closed-Loop ({opportunity.account_count})</span>
          </button>
        </div>
      </div>
    </div>
  )
}
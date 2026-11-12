import React from 'react'
import { Opportunity } from '../types'

interface OpportunityCardProps {
  opportunity: Opportunity
  rank: number
  onSelectPRD: (id: string) => void
  onStatusChange: (id: string, newStatus: string) => void
  onOpenClosedLoop: (opp: Opportunity) => void
}

export const OpportunityCard: React.FC<OpportunityCardProps> = ({
  opportunity,
  rank,
  onSelectPRD,
  onStatusChange,
  onOpenClosedLoop,
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
        return 'bg-emerald-500/20 text-emerald-300 border-emerald-500/40'
      case 'IN_PROGRESS':
        return 'bg-blue-500/20 text-blue-300 border-blue-500/40'
      case 'PLANNED':
        return 'bg-purple-500/20 text-purple-300 border-purple-500/40'
      case 'IN_PRD':
        return 'bg-amber-500/20 text-amber-300 border-amber-500/40'
      default:
        return 'bg-slate-700/30 text-slate-400 border-slate-700'
    }
  }

  const getCategoryColor = (cat: string) => {
    switch (cat) {
      case 'SECURITY':
        return 'text-rose-400 bg-rose-950/30 border-rose-800/40'
      case 'INTEGRATIONS':
        return 'text-indigo-400 bg-indigo-950/30 border-indigo-800/40'
      case 'PERFORMANCE':
        return 'text-amber-400 bg-amber-950/30 border-amber-800/40'
      default:
        return 'text-cyan-400 bg-cyan-950/30 border-cyan-800/40'
    }
  }

  return (
    <div className="bg-[#131929] border border-[#1e293b] hover:border-[#334155] rounded-xl p-5 transition-all shadow-sm">
      <div className="flex flex-col md:flex-row md:items-start justify-between gap-4">
        {/* Left: Rank, Title, Description */}
        <div className="flex items-start gap-3.5 flex-1">
          <div className="flex-shrink-0 w-8 h-8 rounded-lg bg-[#1a233a] border border-[#273552] flex items-center justify-center font-mono text-sm font-bold text-slate-300">
            #{rank}
          </div>
          <div className="space-y-1.5 flex-1">
            <div className="flex flex-wrap items-center gap-2">
              <span className={`px-2 py-0.5 rounded text-[10px] font-semibold border ${getCategoryColor(opportunity.category)}`}>
                {opportunity.category}
              </span>
              <h3 className="text-base font-semibold text-white tracking-tight hover:text-blue-400 cursor-pointer" onClick={() => onSelectPRD(opportunity.id)}>
                {opportunity.title}
              </h3>
            </div>
            <p className="text-xs text-[#94a3b8] line-clamp-2 leading-relaxed">
              {opportunity.description}
            </p>
          </div>
        </div>

        {/* Center: Financial Revenue Badges */}
        <div className="flex items-center gap-5 bg-[#0e1320] px-4 py-2 rounded-lg border border-[#1e293b]/70">
          <div>
            <div className="text-[10px] uppercase font-semibold text-[#64748b]">Total ARR</div>
            <div className="text-base font-bold font-mono text-emerald-400">
              {formatUSD(opportunity.total_arr_usd)}
            </div>
          </div>
          <div className="h-8 w-px bg-[#1e293b]" />
          <div>
            <div className="text-[10px] uppercase font-semibold text-[#64748b]">Weighted WRS</div>
            <div className="text-base font-bold font-mono text-blue-400">
              {opportunity.weighted_score.toFixed(0)}
            </div>
          </div>
          <div className="h-8 w-px bg-[#1e293b]" />
          <div>
            <div className="text-[10px] uppercase font-semibold text-[#64748b]">Requesters</div>
            <div className="text-base font-bold font-mono text-slate-200">
              {opportunity.account_count} <span className="text-xs text-[#64748b]">accounts</span>
            </div>
          </div>
        </div>
      </div>

      {/* Footer controls: Status selector, PRD button, Closed Loop */}
      <div className="mt-4 pt-3.5 border-t border-[#1e293b]/60 flex flex-wrap items-center justify-between gap-3 text-xs">
        <div className="flex items-center gap-2">
          <span className="text-[#64748b]">Status:</span>
          <select
            value={opportunity.status}
            onChange={(e) => onStatusChange(opportunity.id, e.target.value)}
            className={`text-xs px-2.5 py-1 rounded-md border font-medium cursor-pointer bg-[#0e1320] ${getStatusColor(opportunity.status)}`}
          >
            <option value="DISCOVERED">DISCOVERED</option>
            <option value="IN_PRD">IN_PRD</option>
            <option value="PLANNED">PLANNED</option>
            <option value="IN_PROGRESS">IN_PROGRESS</option>
            <option value="SHIPPED">SHIPPED</option>
          </select>
        </div>

        <div className="flex items-center gap-2">
          <button
            onClick={() => onSelectPRD(opportunity.id)}
            className="px-3 py-1.5 rounded bg-blue-600/20 hover:bg-blue-600/30 text-blue-300 border border-blue-500/30 font-medium transition-colors flex items-center gap-1.5"
          >
            <span>📄</span>
            <span>{opportunity.generated_prd_markdown ? 'View PRD' : 'Synthesize PRD'}</span>
          </button>

          {opportunity.status === 'SHIPPED' && (
            <button
              onClick={() => onOpenClosedLoop(opportunity)}
              className="px-3 py-1.5 rounded bg-emerald-600/20 hover:bg-emerald-600/30 text-emerald-300 border border-emerald-500/30 font-medium transition-colors flex items-center gap-1.5"
            >
              <span>📬</span>
              <span>Closed-Loop Email Broadcast</span>
            </button>
          )}
        </div>
      </div>
    </div>
  )
}

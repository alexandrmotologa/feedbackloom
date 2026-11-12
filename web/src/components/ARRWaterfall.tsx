import React from 'react'
import { Opportunity } from '../types'

interface ARRWaterfallProps {
  opportunities: Opportunity[]
  totalPortfolioARR: number
}

export const ARRWaterfall: React.FC<ARRWaterfallProps> = ({ opportunities, totalPortfolioARR }) => {
  const formatUSD = (val: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      maximumFractionDigits: 0,
    }).format(val)
  }

  const totalDemanded = opportunities.reduce((acc, o) => acc + o.total_arr_usd, 0)
  const portfolioPct = totalPortfolioARR > 0 ? (totalDemanded / totalPortfolioARR) * 100 : 0

  return (
    <div className="bg-[#121826] border border-[#1e293b] rounded-xl p-5 mb-6 shadow-sm">
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-5 pb-4 border-b border-[#1e293b]">
        <div>
          <h2 className="text-sm font-semibold text-white tracking-wide uppercase flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-emerald-500 animate-pulse" />
            Revenue Opportunity Waterfall
          </h2>
          <p className="text-xs text-[#94a3b8] mt-1">
            Aggregated dollars demanding product capabilities across active customer contracts.
          </p>
        </div>
        <div className="flex items-center gap-6">
          <div className="text-right">
            <div className="text-xs text-[#64748b]">Total Demanded ARR</div>
            <div className="text-xl font-bold font-mono text-emerald-400">{formatUSD(totalDemanded)}</div>
          </div>
          <div className="text-right">
            <div className="text-xs text-[#64748b]">Portfolio Coverage</div>
            <div className="text-xl font-bold font-mono text-blue-400">{portfolioPct.toFixed(1)}%</div>
          </div>
        </div>
      </div>

      {/* Waterfall Progress Bars */}
      <div className="space-y-4">
        {opportunities.map((opp, idx) => {
          const pct = totalDemanded > 0 ? (opp.total_arr_usd / totalDemanded) * 100 : 0
          return (
            <div key={opp.id} className="space-y-1.5">
              <div className="flex items-center justify-between text-xs">
                <div className="flex items-center gap-2 truncate max-w-md">
                  <span className="font-mono text-[#64748b] text-[11px]">#{idx + 1}</span>
                  <span className="font-medium text-slate-200 truncate">{opp.title}</span>
                  <span className="px-1.5 py-0.2 rounded text-[10px] font-mono bg-[#1e293b] text-slate-300">
                    {opp.account_count} accounts
                  </span>
                </div>
                <div className="flex items-center gap-3">
                  <span className="font-mono text-xs font-semibold text-emerald-400">{formatUSD(opp.total_arr_usd)}</span>
                  <span className="font-mono text-[11px] text-[#64748b] w-12 text-right">{pct.toFixed(1)}%</span>
                </div>
              </div>
              <div className="w-full bg-[#1e293b] h-2 rounded-full overflow-hidden">
                <div
                  className="bg-gradient-to-r from-blue-500 via-indigo-500 to-emerald-400 h-full rounded-full transition-all duration-500"
                  style={{ width: `${Math.max(pct, 2)}%` }}
                />
              </div>
            </div>
          )
        })}

        {opportunities.length === 0 && (
          <div className="text-center py-6 text-xs text-[#64748b]">
            No opportunities synthesized yet. Run HDBSCAN clustering to rank customer demand.
          </div>
        )}
      </div>
    </div>
  )
}

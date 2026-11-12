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
    <div className="bg-[#0f1422]/90 border border-slate-800/80 rounded-2xl p-6 sm:p-7 shadow-lg backdrop-blur-md">
      {/* Waterfall Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-6 pb-5 border-b border-slate-800/60">
        <div>
          <h2 className="text-sm font-bold text-white tracking-wider uppercase flex items-center gap-2.5">
            <span className="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse" />
            Revenue Opportunity Waterfall
          </h2>
          <p className="text-xs text-slate-400 mt-1 max-w-xl">
            Aggregated dollars demanding product capabilities across active customer contracts, ranked by capital impact.
          </p>
        </div>

        <div className="flex items-center gap-6 bg-slate-950/60 px-5 py-3 rounded-xl border border-slate-800/70">
          <div>
            <div className="text-[10px] uppercase font-semibold text-slate-400">Total Demanded ARR</div>
            <div className="text-lg font-bold font-mono text-emerald-400 mt-0.5">{formatUSD(totalDemanded)}</div>
          </div>
          <div className="h-8 w-px bg-slate-800" />
          <div>
            <div className="text-[10px] uppercase font-semibold text-slate-400">Portfolio Coverage</div>
            <div className="text-lg font-bold font-mono text-blue-400 mt-0.5">{portfolioPct.toFixed(1)}%</div>
          </div>
        </div>
      </div>

      {/* Waterfall Progress Bars */}
      <div className="space-y-4">
        {opportunities.map((opp, idx) => {
          const pct = totalDemanded > 0 ? (opp.total_arr_usd / totalDemanded) * 100 : 0
          return (
            <div key={opp.id} className="space-y-2 group">
              <div className="flex items-center justify-between text-xs">
                <div className="flex items-center gap-2.5 truncate max-w-xl">
                  <span className="font-mono text-slate-500 font-bold text-xs">#{String(idx + 1).padStart(2, '0')}</span>
                  <span className="font-semibold text-white truncate text-sm group-hover:text-blue-400 transition-colors">
                    {opp.title}
                  </span>
                  <span className="px-2 py-0.5 rounded-md text-[10px] font-mono bg-slate-800/80 text-slate-300 border border-slate-700/50">
                    {opp.account_count} accounts
                  </span>
                </div>
                <div className="flex items-center gap-3">
                  <span className="font-mono text-sm font-bold text-emerald-400">{formatUSD(opp.total_arr_usd)}</span>
                  <span className="font-mono text-xs font-semibold text-slate-400 w-12 text-right">{pct.toFixed(1)}%</span>
                </div>
              </div>
              <div className="w-full bg-slate-950 h-2.5 rounded-full overflow-hidden p-0.5 border border-slate-800/50">
                <div
                  className="bg-gradient-to-r from-blue-600 via-indigo-500 to-cyan-400 h-full rounded-full transition-all duration-700 shadow-sm"
                  style={{ width: `${Math.max(pct, 2)}%` }}
                />
              </div>
            </div>
          )
        })}

        {opportunities.length === 0 && (
          <div className="text-center py-8 text-xs text-slate-500">
            No opportunities synthesized yet. Run HDBSCAN clustering to rank customer demand.
          </div>
        )}
      </div>
    </div>
  )
}
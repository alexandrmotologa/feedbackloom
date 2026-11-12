import React from 'react'
import { Opportunity } from '../types'

interface CompetitorMatrixProps {
  opportunities: Opportunity[]
}

export const CompetitorMatrix: React.FC<CompetitorMatrixProps> = ({ opportunities }) => {
  const formatUSD = (val: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      maximumFractionDigits: 0,
    }).format(val)
  }

  // Tally ARR linked to each competitor mentioned
  const competitorMap: Record<string, { totalARR: number; lostPipelineARR: number; count: number; opps: string[] }> = {}

  for (const opp of opportunities) {
    if (opp.competitor_mentions && opp.competitor_mentions.length > 0) {
      for (const comp of opp.competitor_mentions) {
        if (!competitorMap[comp]) {
          competitorMap[comp] = { totalARR: 0, lostPipelineARR: 0, count: 0, opps: [] }
        }
        competitorMap[comp].totalARR += opp.total_arr_usd
        competitorMap[comp].lostPipelineARR += opp.lost_pipeline_arr_usd
        competitorMap[comp].count++
        if (!competitorMap[comp].opps.includes(opp.title)) {
          competitorMap[comp].opps.push(opp.title)
        }
      }
    }
  }

  const competitorList = Object.entries(competitorMap).map(([name, data]) => ({
    name,
    ...data,
  }))

  competitorList.sort((a, b) => b.totalARR - a.totalARR)

  if (competitorList.length === 0) {
    return null
  }

  return (
    <div className="bg-[#0f1422]/90 border border-slate-800/80 rounded-2xl p-6 sm:p-7 shadow-lg backdrop-blur-md">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 mb-5 pb-4 border-b border-slate-800/60">
        <div>
          <h2 className="text-sm font-bold text-white tracking-wider uppercase flex items-center gap-2.5">
            <span className="w-2.5 h-2.5 rounded-full bg-rose-500 animate-pulse" />
            Competitor Loss Attribution &amp; Churn Exposure
          </h2>
          <p className="text-xs text-slate-400 mt-1 max-w-2xl">
            Customer contract value demanding capabilities where third-party alternatives are actively cited as churn threats.
          </p>
        </div>
        <span className="px-3 py-1 rounded-full text-xs font-mono font-semibold bg-rose-500/10 text-rose-400 border border-rose-500/30 self-start sm:self-center">
          HIGH RETENTION RISK
        </span>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        {competitorList.map((comp) => (
          <div
            key={comp.name}
            className="bg-slate-950/60 border border-slate-800/80 hover:border-rose-800/50 rounded-xl p-4 space-y-3 transition-all shadow-sm hover:shadow-rose-950/20"
          >
            <div className="flex items-center justify-between">
              <span className="text-sm font-bold text-white flex items-center gap-2">
                <span className="text-rose-400">⚔</span> {comp.name}
              </span>
              <span className="px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-rose-500/10 text-rose-300 border border-rose-500/30">
                {comp.count} {comp.count === 1 ? 'theme' : 'themes'}
              </span>
            </div>

            <div className="space-y-1.5 pt-1">
              <div className="flex items-center justify-between text-xs">
                <span className="text-slate-400">Demanded ARR:</span>
                <span className="font-mono font-bold text-emerald-400">{formatUSD(comp.totalARR)}</span>
              </div>
              <div className="flex items-center justify-between text-xs">
                <span className="text-slate-400">Pipeline at Risk:</span>
                <span className="font-mono font-bold text-rose-400">{formatUSD(comp.lostPipelineARR)}</span>
              </div>
            </div>

            <div className="pt-2 border-t border-slate-800/60">
              <span className="text-[11px] text-slate-500 line-clamp-1">
                Linked: <span className="text-slate-400">{comp.opps.join(', ')}</span>
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
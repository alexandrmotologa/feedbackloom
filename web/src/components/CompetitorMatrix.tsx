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
    <div className="bg-[#121826] border border-[#1e293b] rounded-xl p-5 mb-6 shadow-sm">
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 mb-4 pb-3 border-b border-[#1e293b]">
        <div>
          <h2 className="text-sm font-semibold text-white tracking-wide uppercase flex items-center gap-2">
            <span className="w-2 h-2 rounded-full bg-rose-500 animate-pulse" />
            Competitor Loss Attribution & Churn Exposure
          </h2>
          <p className="text-xs text-[#94a3b8] mt-0.5">
            Customer contract value demanding capabilities where third-party competitors are actively cited as alternatives or churn threats.
          </p>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3.5">
        {competitorList.map((comp) => (
          <div
            key={comp.name}
            className="bg-[#0e1320] border border-[#1e293b] hover:border-rose-900/40 rounded-xl p-3.5 space-y-2 transition-all"
          >
            <div className="flex items-center justify-between">
              <span className="text-xs font-bold text-white flex items-center gap-1.5">
                <span className="text-rose-400">⚔</span> {comp.name}
              </span>
              <span className="px-1.5 py-0.2 rounded text-[10px] font-mono bg-rose-500/10 text-rose-300 border border-rose-500/30">
                {comp.count} themes
              </span>
            </div>

            <div className="space-y-1">
              <div className="flex items-center justify-between text-xs">
                <span className="text-[#64748b]">Demanded ARR:</span>
                <span className="font-mono font-semibold text-emerald-400">{formatUSD(comp.totalARR)}</span>
              </div>
              <div className="flex items-center justify-between text-xs">
                <span className="text-[#64748b]">Pipeline at Risk:</span>
                <span className="font-mono font-semibold text-rose-400">{formatUSD(comp.lostPipelineARR)}</span>
              </div>
            </div>

            <div className="pt-1.5 border-t border-[#1e293b]/60">
              <span className="text-[10px] text-[#64748b] line-clamp-1">
                Linked: {comp.opps.join(', ')}
              </span>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

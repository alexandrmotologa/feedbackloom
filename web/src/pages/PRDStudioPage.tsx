import React, { useState, useEffect } from 'react'
import { Opportunity, EnrichedFeedbackItem } from '../types'
import { PRDViewer } from '../components/PRDViewer'
import { fetchOpportunityDetail } from '../api'

interface PRDStudioPageProps {
  opportunities: Opportunity[]
  selectedOpportunityId: string | null
  onSelectOpportunity: (id: string) => void
}

export const PRDStudioPage: React.FC<PRDStudioPageProps> = ({
  opportunities,
  selectedOpportunityId,
  onSelectOpportunity,
}) => {
  const [currentOpp, setCurrentOpp] = useState<Opportunity | null>(null)
  const [quotes, setQuotes] = useState<EnrichedFeedbackItem[]>([])
  const [loading, setLoading] = useState(false)

  // Default to first opportunity if none selected
  const activeID = selectedOpportunityId || (opportunities.length > 0 ? opportunities[0].id : null)

  useEffect(() => {
    if (activeID) {
      loadDetail(activeID)
    }
  }, [activeID])

  const loadDetail = async (id: string) => {
    setLoading(true)
    try {
      const data = await fetchOpportunityDetail(id)
      setCurrentOpp(data.opportunity)
      setQuotes(data.quotes || [])
    } catch (err: any) {
      alert(`Error loading opportunity: ${err.message}`)
    } finally {
      setLoading(false)
    }
  }

  const handlePRDUpdated = (md: string) => {
    if (currentOpp) {
      setCurrentOpp({ ...currentOpp, generated_prd_markdown: md, status: 'IN_PRD' })
    }
  }

  if (opportunities.length === 0) {
    return (
      <div className="bg-[#0f1422]/90 border border-slate-800/80 rounded-2xl p-16 text-center text-slate-400 space-y-3 shadow-lg">
        <div className="text-4xl">📄</div>
        <div className="text-base font-semibold text-white">No Opportunities Available</div>
        <p className="text-xs text-slate-500 max-w-sm mx-auto">
          Please run the HDBSCAN clustering engine on customer feedback to synthesize product opportunities first.
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-6 pb-12">
      {/* Selector Ribbon */}
      <div className="bg-[#0f1422]/90 border border-slate-800/80 rounded-2xl p-4 sm:p-5 flex flex-col sm:flex-row items-center justify-between gap-4 shadow-md backdrop-blur-md">
        <div className="flex items-center gap-3 w-full sm:w-auto">
          <span className="text-xs uppercase font-bold text-slate-400 whitespace-nowrap">Target Opportunity:</span>
          <select
            value={activeID || ''}
            onChange={(e) => onSelectOpportunity(e.target.value)}
            className="bg-slate-950 border border-slate-800 text-white text-xs rounded-xl px-3.5 py-2.5 outline-none w-full sm:w-96 font-medium focus:border-blue-500"
          >
            {opportunities.map((opp) => (
              <option key={opp.id} value={opp.id} className="bg-slate-950 text-white">
                {opp.title} (${Math.round(opp.total_arr_usd / 1000)}k ARR - {opp.category})
              </option>
            ))}
          </select>
        </div>

        {currentOpp && (
          <div className="flex items-center gap-5 text-xs font-mono text-slate-300 bg-slate-950/60 px-4 py-2 rounded-xl border border-slate-800/60">
            <div>
              <span className="text-slate-400 mr-1.5">ARR at Stake:</span>
              <span className="text-emerald-400 font-bold">${currentOpp.total_arr_usd.toLocaleString()}</span>
            </div>
            <div className="h-4 w-px bg-slate-800" />
            <div>
              <span className="text-slate-400 mr-1.5">Customer Quotes:</span>
              <span className="text-blue-400 font-bold">{quotes.length} accounts</span>
            </div>
          </div>
        )}
      </div>

      {/* Main PRD Viewer */}
      {loading ? (
        <div className="py-20 text-center text-slate-400">Loading opportunity details...</div>
      ) : currentOpp ? (
        <PRDViewer opportunity={currentOpp} quotes={quotes} onPRDUpdated={handlePRDUpdated} />
      ) : null}
    </div>
  )
}
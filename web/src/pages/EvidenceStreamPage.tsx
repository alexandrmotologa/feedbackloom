import React, { useState, useEffect } from 'react'
import { EnrichedFeedbackItem } from '../types'
import { QuoteStream } from '../components/QuoteStream'
import { fetchFeedback, fetchAccounts } from '../api'

export const EvidenceStreamPage: React.FC = () => {
  const [quotes, setQuotes] = useState<EnrichedFeedbackItem[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadQuotes()
  }, [])

  const loadQuotes = async () => {
    setLoading(true)
    try {
      const [feedbackList, accounts] = await Promise.all([fetchFeedback(), fetchAccounts()])

      const accountMap = new Map(accounts.map((a) => [a.id, a]))

      const enriched: EnrichedFeedbackItem[] = feedbackList.map((fb) => {
        const acc = fb.account_id ? accountMap.get(fb.account_id) : undefined
        return {
          ...fb,
          account_name: acc?.name,
          account_domain: acc?.domain,
          account_arr_usd: acc?.arr_usd,
          plan_tier: acc?.plan_tier,
          renewal_date: acc?.renewal_date,
          days_to_renewal: acc ? Math.max(0, Math.round((new Date(acc.renewal_date).getTime() - Date.now()) / (1000 * 3600 * 24))) : 365,
          churn_risk_score: acc?.churn_risk_score || 0.2,
        }
      })

      // Sort by ARR descending
      enriched.sort((a, b) => (b.account_arr_usd || 0) - (a.account_arr_usd || 0))
      setQuotes(enriched)
    } catch (err: any) {
      alert(`Error loading quotes: ${err.message}`)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      {/* Title */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 bg-[#121826] p-5 rounded-xl border border-[#1e293b]">
        <div>
          <h2 className="text-base font-bold text-white">Customer Voice & Evidence Stream</h2>
          <p className="text-xs text-slate-400 mt-1">
            Real customer quotes ingested across Zendesk tickets, Gong call transcripts, Intercom chats, and Slack messages.
          </p>
        </div>
        <button
          onClick={loadQuotes}
          className="text-xs bg-[#1a233a] hover:bg-[#232f4e] text-slate-300 border border-[#273552] px-3 py-1.5 rounded-lg flex items-center gap-1.5 self-start sm:self-auto"
        >
          <span>↻</span> Refresh Stream
        </button>
      </div>

      {loading ? (
        <div className="py-20 text-center text-slate-500 text-xs">Loading evidence stream...</div>
      ) : (
        <QuoteStream quotes={quotes} />
      )}
    </div>
  )
}

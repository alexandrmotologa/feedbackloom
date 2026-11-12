import React, { useState } from 'react'
import { EnrichedFeedbackItem } from '../types'

interface QuoteStreamProps {
  quotes: EnrichedFeedbackItem[]
}

export const QuoteStream: React.FC<QuoteStreamProps> = ({ quotes }) => {
  const [selectedChannel, setSelectedChannel] = useState<string>('ALL')

  const formatUSD = (val?: number) => {
    if (!val) return '$0'
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      maximumFractionDigits: 0,
    }).format(val)
  }

  const channels = ['ALL', 'ZENDESK', 'INTERCOM', 'GONG', 'SLACK']

  const filteredQuotes =
    selectedChannel === 'ALL'
      ? quotes
      : quotes.filter((q) => q.source.toUpperCase() === selectedChannel)

  return (
    <div className="space-y-6 pb-12">
      {/* Channel Filters */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 p-4 sm:p-5 rounded-2xl bg-[#0f1422]/90 border border-slate-800/80 shadow-md backdrop-blur-md">
        <div className="flex items-center gap-2 overflow-x-auto">
          {channels.map((ch) => (
            <button
              key={ch}
              onClick={() => setSelectedChannel(ch)}
              className={`px-3.5 py-1.5 rounded-xl text-xs font-semibold transition-all ${
                selectedChannel === ch
                  ? 'bg-blue-600 text-white shadow-sm ring-1 ring-blue-400/40'
                  : 'bg-slate-900/80 text-slate-400 hover:text-white hover:bg-slate-800 border border-slate-800/80'
              }`}
            >
              {ch}
            </button>
          ))}
        </div>
        <div className="text-xs text-slate-400 font-mono">
          Showing <span className="text-white font-bold">{filteredQuotes.length}</span> verified customer quotes
        </div>
      </div>

      {/* Quote Cards */}
      <div className="grid grid-cols-1 gap-4">
        {filteredQuotes.map((q) => {
          const isEnterprise = q.plan_tier === 'ENTERPRISE'
          const isUrgent = q.urgency_score >= 0.7 || q.days_to_renewal < 45

          return (
            <div
              key={q.id}
              className="bg-[#0f1422]/90 border border-slate-800/80 hover:border-slate-700/90 rounded-2xl p-6 transition-all shadow-md space-y-3.5"
            >
              {/* Header: Account, Tier, ARR, Renewal */}
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div className="flex items-center gap-3">
                  <span className="text-sm font-bold text-white">
                    {q.account_name || q.account_domain || 'Unknown Account'}
                  </span>
                  {q.account_domain && (
                    <span className="text-xs font-mono text-slate-400">
                      @{q.account_domain}
                    </span>
                  )}
                  <span className={`px-2.5 py-0.5 rounded-md text-[11px] font-mono font-semibold ${
                    isEnterprise
                      ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/30'
                      : 'bg-blue-500/10 text-blue-400 border border-blue-500/30'
                  }`}>
                    {q.plan_tier || 'STARTER'}
                  </span>
                </div>

                <div className="flex items-center gap-3">
                  <div className="text-right">
                    <span className="text-[10px] uppercase font-semibold text-slate-400 mr-1.5">ARR:</span>
                    <span className="font-mono text-xs font-bold text-emerald-400">
                      {formatUSD(q.account_arr_usd)}
                    </span>
                  </div>

                  <span className={`px-2.5 py-0.5 rounded-md text-[11px] font-mono font-semibold ${
                    isUrgent
                      ? 'bg-rose-500/10 text-rose-300 border border-rose-500/30'
                      : 'bg-slate-800 text-slate-300 border border-slate-700'
                  }`}>
                    {q.days_to_renewal} days to renewal
                  </span>

                  <span className="px-2.5 py-0.5 rounded-md text-[10px] font-mono font-bold uppercase bg-slate-800 text-slate-200">
                    {q.source}
                  </span>
                </div>
              </div>

              {/* Verbatim Quote */}
              <blockquote className="text-xs text-slate-200 bg-slate-950/70 p-4 rounded-xl border-l-2 border-blue-500 leading-relaxed font-sans shadow-inner">
                &ldquo;{q.raw_text}&rdquo;
              </blockquote>

              {/* Extracted Details & Requester */}
              <div className="flex flex-wrap items-center justify-between gap-2 text-xs text-slate-400 pt-1">
                <div>
                  Customer: <span className="text-slate-300 font-medium">{q.customer_name || 'Enterprise Buyer'}</span> ({q.customer_email})
                </div>
                {q.extracted_pain_point && (
                  <div className="text-slate-400 text-[11px]">
                    Extracted Pain: <span className="text-slate-300">{q.extracted_pain_point}</span>
                  </div>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
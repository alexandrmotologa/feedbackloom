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

  const filteredQuotes = selectedChannel === 'ALL'
    ? quotes
    : quotes.filter((q) => q.source.toUpperCase() === selectedChannel)

  return (
    <div className="space-y-4">
      {/* Channel Filters */}
      <div className="flex items-center justify-between gap-4 pb-2 border-b border-[#1e293b]">
        <div className="flex items-center gap-2">
          {channels.map((ch) => (
            <button
              key={ch}
              onClick={() => setSelectedChannel(ch)}
              className={`px-3 py-1 rounded text-xs font-semibold transition-all ${
                selectedChannel === ch
                  ? 'bg-blue-600 text-white'
                  : 'bg-[#141b2d] text-slate-400 hover:text-white hover:bg-[#1a233a]'
              }`}
            >
              {ch}
            </button>
          ))}
        </div>
        <div className="text-xs text-slate-400">
          Showing <span className="text-white font-mono font-semibold">{filteredQuotes.length}</span> verified customer quotes
        </div>
      </div>

      {/* Quote Cards */}
      <div className="grid grid-cols-1 gap-3.5">
        {filteredQuotes.map((q) => {
          const isEnterprise = q.plan_tier === 'ENTERPRISE'
          const isUrgent = q.urgency_score >= 0.7 || q.days_to_renewal < 45

          return (
            <div
              key={q.id}
              className="bg-[#121826] border border-[#1e293b] hover:border-[#2d3a52] rounded-xl p-4 transition-all space-y-3"
            >
              {/* Header: Account, Tier, ARR, Renewal */}
              <div className="flex flex-wrap items-center justify-between gap-2">
                <div className="flex items-center gap-2.5">
                  <span className="text-xs font-bold text-white">
                    {q.account_name || q.account_domain || 'Unknown Account'}
                  </span>
                  {q.account_domain && (
                    <span className="text-[11px] font-mono text-slate-400">
                      @{q.account_domain}
                    </span>
                  )}
                  <span className={`badge ${isEnterprise ? 'badge-enterprise' : 'badge-growth'}`}>
                    {q.plan_tier || 'STARTER'}
                  </span>
                </div>

                <div className="flex items-center gap-3">
                  <div className="text-right">
                    <span className="text-[10px] uppercase font-semibold text-[#64748b] mr-1.5">ARR:</span>
                    <span className="font-mono text-xs font-bold text-emerald-400">
                      {formatUSD(q.account_arr_usd)}
                    </span>
                  </div>

                  <span className={`badge ${isUrgent ? 'badge-urgent' : 'badge-starter'}`}>
                    {q.days_to_renewal} days to renewal
                  </span>

                  <span className="px-2 py-0.5 rounded text-[10px] font-mono font-semibold uppercase bg-slate-800 text-slate-300">
                    {q.source}
                  </span>
                </div>
              </div>

              {/* Verbatim Quote */}
              <blockquote className="text-xs text-slate-200 bg-[#0e1320] p-3 rounded-lg border-l-2 border-blue-500 leading-relaxed font-sans">
                &ldquo;{q.raw_text}&rdquo;
              </blockquote>

              {/* Extracted Details & Requester */}
              <div className="flex flex-wrap items-center justify-between gap-2 pt-1 text-[11px] text-slate-400">
                <div className="flex items-center gap-2">
                  <span>Reported by:</span>
                  <span className="text-slate-300 font-medium">{q.customer_name || 'Customer'}</span>
                  {q.customer_email && (
                    <span className="text-slate-500 font-mono text-[10px]">({q.customer_email})</span>
                  )}
                </div>

                <div className="flex items-center gap-2">
                  <span className="text-[#64748b]">Urgency:</span>
                  <div className="w-16 bg-[#1e293b] h-1.5 rounded-full overflow-hidden">
                    <div
                      className={`h-full rounded-full ${q.urgency_score >= 0.7 ? 'bg-rose-500' : 'bg-blue-400'}`}
                      style={{ width: `${q.urgency_score * 100}%` }}
                    />
                  </div>
                  <span className="font-mono text-[10px] text-slate-300">{(q.urgency_score * 10).toFixed(1)}/10</span>
                </div>
              </div>
            </div>
          )
        })}

        {filteredQuotes.length === 0 && (
          <div className="text-center py-12 text-xs text-slate-500">
            No customer quotes found matching the selected channel filter.
          </div>
        )}
      </div>
    </div>
  )
}

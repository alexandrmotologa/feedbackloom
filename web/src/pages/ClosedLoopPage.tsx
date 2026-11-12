import React from 'react'
import { Opportunity } from '../types'

interface ClosedLoopPageProps {
  opportunities: Opportunity[]
  onOpenClosedLoop: (opp: Opportunity) => void
  onStatusChange: (id: string, status: string) => void
}

export const ClosedLoopPage: React.FC<ClosedLoopPageProps> = ({
  opportunities,
  onOpenClosedLoop,
  onStatusChange,
}) => {
  const shippedOpps = opportunities.filter((o) => o.status === 'SHIPPED')
  const inProgressOpps = opportunities.filter(
    (o) => o.status === 'IN_PROGRESS' || o.status === 'PLANNED' || o.status === 'IN_PRD'
  )

  const formatUSD = (val: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      maximumFractionDigits: 0,
    }).format(val)
  }

  return (
    <div className="space-y-8 pb-12">
      {/* Overview Banner */}
      <div className="bg-[#0f1422]/90 border border-slate-800/80 rounded-2xl p-6 sm:p-7 shadow-lg backdrop-blur-md">
        <h1 className="text-lg font-bold text-white flex items-center gap-2.5">
          <span className="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse" />
          Closed-Loop Customer Updates
        </h1>
        <p className="text-xs text-slate-400 mt-1 max-w-3xl leading-relaxed">
          When a requested capability transitions to &ldquo;SHIPPED&rdquo;, feedbackloom drafts personalized email updates to every account contact who originally requested the feature, quoting their exact verbatim feedback.
        </p>
      </div>

      {/* Shipped Features Section */}
      <div className="space-y-4">
        <h2 className="text-sm font-bold uppercase tracking-wider text-emerald-400 flex items-center gap-2">
          <span>✓</span> Shipped Features Ready for Outreach ({shippedOpps.length})
        </h2>

        <div className="grid grid-cols-1 gap-4">
          {shippedOpps.map((opp) => (
            <div
              key={opp.id}
              className="bg-[#0f1422]/90 border border-emerald-900/40 rounded-2xl p-6 flex flex-col md:flex-row md:items-center justify-between gap-5 shadow-md"
            >
              <div className="space-y-2">
                <div className="flex items-center gap-2.5">
                  <span className="px-2.5 py-0.5 rounded-full text-[10px] font-mono font-bold bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
                    SHIPPED
                  </span>
                  <h3 className="text-base font-bold text-white">{opp.title}</h3>
                </div>
                <p className="text-xs text-slate-300 leading-relaxed max-w-2xl">{opp.description}</p>
                <div className="text-xs text-slate-400 font-mono">
                  {opp.account_count} requesting accounts &bull; {formatUSD(opp.total_arr_usd)} ARR defended
                </div>
              </div>

              <button
                onClick={() => onOpenClosedLoop(opp)}
                className="px-5 py-2.5 rounded-xl text-xs font-semibold bg-emerald-600 hover:bg-emerald-500 text-white shadow-lg shadow-emerald-900/30 flex items-center gap-2 self-start md:self-auto transition-all whitespace-nowrap"
              >
                <span>📬</span>
                <span>Review &amp; Dispatch Outreach</span>
              </button>
            </div>
          ))}

          {shippedOpps.length === 0 && (
            <div className="bg-[#0f1422]/90 border border-slate-800/80 rounded-2xl p-12 text-center text-slate-400 text-xs shadow-md">
              No shipped features currently. Change an opportunity&apos;s status to &ldquo;SHIPPED&rdquo; in the Opportunities tab to unlock closed-loop notifications.
            </div>
          )}
        </div>
      </div>

      {/* In-Progress / Pipeline Section */}
      <div className="space-y-4 pt-4 border-t border-slate-800/60">
        <h2 className="text-sm font-bold uppercase tracking-wider text-blue-400 flex items-center gap-2">
          <span>⚡</span> Active Pipeline Opportunities ({inProgressOpps.length})
        </h2>

        <div className="grid grid-cols-1 gap-4">
          {inProgressOpps.map((opp) => (
            <div
              key={opp.id}
              className="bg-[#0f1422]/90 border border-slate-800/80 rounded-2xl p-5 sm:p-6 flex flex-col md:flex-row md:items-center justify-between gap-4 shadow-sm"
            >
              <div className="space-y-1.5">
                <div className="flex items-center gap-2.5">
                  <span className="px-2.5 py-0.5 rounded-full text-[10px] font-mono font-bold bg-blue-500/15 text-blue-400 border border-blue-500/30">
                    {opp.status}
                  </span>
                  <h3 className="text-sm font-bold text-white">{opp.title}</h3>
                </div>
                <div className="text-xs text-slate-400 font-mono">
                  {opp.account_count} accounts waiting &bull; {formatUSD(opp.total_arr_usd)} ARR at stake
                </div>
              </div>

              <button
                onClick={() => onStatusChange(opp.id, 'SHIPPED')}
                className="px-4 py-2 rounded-xl text-xs font-medium bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-white border border-slate-700/80 transition-all self-start md:self-auto"
              >
                Mark as Shipped &rarr;
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
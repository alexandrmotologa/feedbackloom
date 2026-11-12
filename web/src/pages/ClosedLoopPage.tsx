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
  const inProgressOpps = opportunities.filter((o) => o.status === 'IN_PROGRESS' || o.status === 'PLANNED')

  const formatUSD = (val: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      maximumFractionDigits: 0,
    }).format(val)
  }

  return (
    <div className="space-y-6">
      {/* Overview */}
      <div className="bg-[#121826] border border-[#1e293b] rounded-xl p-5">
        <h2 className="text-base font-bold text-white">Closed-Loop Customer Updates</h2>
        <p className="text-xs text-slate-400 mt-1">
          When a requested capability transitions to &ldquo;Shipped&rdquo;, feedbackloom drafts personalized email updates to every account contact who originally requested the feature, quoting their exact verbatim feedback.
        </p>
      </div>

      {/* Shipped Features Section */}
      <div className="space-y-4">
        <h3 className="text-sm font-semibold uppercase tracking-wider text-emerald-400 flex items-center gap-2">
          <span>✓</span> Shipped Features Ready for Outreach ({shippedOpps.length})
        </h3>

        <div className="grid grid-cols-1 gap-3.5">
          {shippedOpps.map((opp) => (
            <div
              key={opp.id}
              className="bg-[#121826] border border-emerald-900/40 rounded-xl p-5 flex flex-col md:flex-row md:items-center justify-between gap-4"
            >
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <span className="badge badge-enterprise">SHIPPED</span>
                  <h4 className="text-sm font-bold text-white">{opp.title}</h4>
                </div>
                <p className="text-xs text-slate-400 line-clamp-1">{opp.description}</p>
                <div className="text-xs text-slate-500 font-mono">
                  {opp.account_count} requesting accounts &bull; {formatUSD(opp.total_arr_usd)} ARR
                </div>
              </div>

              <button
                onClick={() => onOpenClosedLoop(opp)}
                className="px-4 py-2 rounded-lg text-xs font-semibold bg-emerald-600 hover:bg-emerald-500 text-white shadow-sm flex items-center gap-2 self-start md:self-auto transition-all"
              >
                <span>📬</span>
                <span>Review & Dispatch Outreach</span>
              </button>
            </div>
          ))}

          {shippedOpps.length === 0 && (
            <div className="bg-[#121826] border border-[#1e293b] rounded-xl p-10 text-center text-slate-500 text-xs">
              No shipped features currently. Change an opportunity&apos;s status to &ldquo;SHIPPED&rdquo; to unlock closed-loop notifications.
            </div>
          )}
        </div>
      </div>

      {/* In-Progress / Planned Section */}
      <div className="space-y-4 pt-4 border-t border-[#1e293b]">
        <h3 className="text-sm font-semibold uppercase tracking-wider text-blue-400 flex items-center gap-2">
          <span>⚡</span> Features in Pipeline ({inProgressOpps.length})
        </h3>

        <div className="grid grid-cols-1 gap-3.5">
          {inProgressOpps.map((opp) => (
            <div
              key={opp.id}
              className="bg-[#121826] border border-[#1e293b] rounded-xl p-4 flex flex-col md:flex-row md:items-center justify-between gap-4"
            >
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <span className="badge badge-growth">{opp.status}</span>
                  <h4 className="text-sm font-bold text-slate-200">{opp.title}</h4>
                </div>
                <div className="text-xs text-slate-500 font-mono">
                  {opp.account_count} accounts waiting &bull; {formatUSD(opp.total_arr_usd)} ARR at stake
                </div>
              </div>

              <button
                onClick={() => onStatusChange(opp.id, 'SHIPPED')}
                className="px-3.5 py-1.5 rounded-lg text-xs font-medium bg-[#1e293b] hover:bg-emerald-600/30 text-slate-300 hover:text-emerald-300 border border-[#334155] hover:border-emerald-500/40 transition-colors self-start md:self-auto"
              >
                Mark as &ldquo;SHIPPED&rdquo; →
              </button>
            </div>
          ))}

          {inProgressOpps.length === 0 && (
            <div className="bg-[#121826] border border-[#1e293b] rounded-xl p-8 text-center text-slate-500 text-xs">
              No opportunities currently in PLANNED or IN_PROGRESS status.
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

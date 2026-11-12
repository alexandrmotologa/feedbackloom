import React, { useState, useEffect } from 'react'
import { Opportunity, ClosedLoopNotification } from '../types'
import { fetchNotifications, generateNotifications, dispatchNotifications } from '../api'

interface ClosedLoopModalProps {
  opportunity: Opportunity
  onClose: () => void
}

export const ClosedLoopModal: React.FC<ClosedLoopModalProps> = ({ opportunity, onClose }) => {
  const [notifications, setNotifications] = useState<ClosedLoopNotification[]>([])
  const [loading, setLoading] = useState(true)
  const [dispatching, setDispatching] = useState(false)
  const [selectedNotif, setSelectedNotif] = useState<ClosedLoopNotification | null>(null)

  useEffect(() => {
    loadNotifications()
  }, [opportunity.id])

  const loadNotifications = async () => {
    setLoading(true)
    try {
      let notifs = await fetchNotifications(opportunity.id)
      if (notifs.length === 0) {
        // Auto-generate if none exist
        const gen = await generateNotifications(opportunity.id)
        notifs = gen.notifications || []
      }
      setNotifications(notifs)
      if (notifs.length > 0) {
        setSelectedNotif(notifs[0])
      }
    } catch (err: any) {
      alert(`Error loading notifications: ${err.message}`)
    } finally {
      setLoading(false)
    }
  }

  const handleDispatchAll = async () => {
    setDispatching(true)
    try {
      const res = await dispatchNotifications(opportunity.id)
      alert(res.message)
      await loadNotifications()
    } catch (err: any) {
      alert(`Dispatch error: ${err.message}`)
    } finally {
      setDispatching(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4">
      <div className="bg-[#121826] border border-[#1e293b] rounded-2xl max-w-4xl w-full max-h-[90vh] flex flex-col overflow-hidden shadow-2xl">
        {/* Modal Header */}
        <div className="px-6 py-4 bg-[#0e1320] border-b border-[#1e293b] flex items-center justify-between">
          <div>
            <span className="text-[10px] uppercase font-bold text-emerald-400 tracking-wider block">
              Closed-Loop Automation
            </span>
            <h2 className="text-base font-bold text-white flex items-center gap-2">
              <span>Personalized Updates for &ldquo;{opportunity.title}&rdquo;</span>
            </h2>
          </div>
          <button
            onClick={onClose}
            className="text-slate-400 hover:text-white text-lg font-bold p-1"
          >
            ✕
          </button>
        </div>

        {/* Modal Body: Split view (Recipients on left, Email preview on right) */}
        <div className="flex-1 overflow-hidden grid grid-cols-1 md:grid-cols-3 divide-y md:divide-y-0 md:divide-x divide-[#1e293b]">
          {/* Left: Contact List */}
          <div className="p-4 overflow-y-auto space-y-2">
            <div className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-2">
              Requesting Customers ({notifications.length})
            </div>

            {loading ? (
              <div className="py-8 text-center text-xs text-slate-500">Loading contacts...</div>
            ) : notifications.length === 0 ? (
              <div className="py-8 text-center text-xs text-slate-500">No customer emails found.</div>
            ) : (
              notifications.map((n) => {
                const isSelected = selectedNotif?.id === n.id
                return (
                  <button
                    key={n.id}
                    onClick={() => setSelectedNotif(n)}
                    className={`w-full text-left p-2.5 rounded-lg border text-xs transition-all ${
                      isSelected
                        ? 'bg-blue-600/20 border-blue-500/50 text-white'
                        : 'bg-[#0e1320] border-[#1e293b] text-slate-300 hover:bg-[#161f30]'
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <span className="font-semibold truncate">{n.customer_name || 'Customer'}</span>
                      <span
                        className={`text-[9px] px-1.5 py-0.2 rounded font-mono ${
                          n.status === 'SENT'
                            ? 'bg-emerald-500/20 text-emerald-300'
                            : 'bg-amber-500/20 text-amber-300'
                        }`}
                      >
                        {n.status}
                      </span>
                    </div>
                    <div className="text-[11px] text-slate-400 truncate mt-0.5">{n.customer_email}</div>
                  </button>
                )
              })
            )}
          </div>

          {/* Right: Preview Message */}
          <div className="md:col-span-2 p-6 overflow-y-auto flex flex-col justify-between space-y-4">
            {selectedNotif ? (
              <div className="space-y-4 flex-1">
                <div className="bg-[#0c101a] p-3.5 rounded-lg border border-[#1e293b] space-y-1 text-xs">
                  <div>
                    <span className="text-[#64748b]">To:</span>{' '}
                    <span className="text-slate-200 font-medium">
                      {selectedNotif.customer_name} &lt;{selectedNotif.customer_email}&gt;
                    </span>
                  </div>
                  <div>
                    <span className="text-[#64748b]">Subject:</span>{' '}
                    <span className="text-slate-200 font-semibold">{selectedNotif.subject}</span>
                  </div>
                </div>

                <div className="bg-[#0c101a] p-4 rounded-lg border border-[#1e293b] text-xs text-slate-200 whitespace-pre-wrap leading-relaxed font-sans max-h-72 overflow-y-auto">
                  {selectedNotif.body}
                </div>
              </div>
            ) : (
              <div className="py-20 text-center text-xs text-slate-500">
                Select a recipient to preview their personalized notification.
              </div>
            )}

            {/* Bottom Actions */}
            <div className="pt-4 border-t border-[#1e293b] flex items-center justify-between text-xs">
              <span className="text-slate-400">
                All recipients will receive citations of their verbatim feedback.
              </span>
              <div className="flex items-center gap-2">
                <button
                  onClick={onClose}
                  className="px-3 py-1.5 rounded bg-slate-800 text-slate-300 hover:bg-slate-700"
                >
                  Cancel
                </button>
                <button
                  onClick={handleDispatchAll}
                  disabled={dispatching || notifications.length === 0}
                  className="px-4 py-1.5 rounded bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white font-semibold flex items-center gap-1.5 shadow-sm"
                >
                  <span>🚀</span>
                  <span>{dispatching ? 'Dispatching...' : `Dispatch All (${notifications.length})`}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

import React, { useState } from 'react'
import { syncCRM, seedDemoFeedback, runClustering } from '../api'

interface ConnectorsPageProps {
  onRefreshAll: () => void
}

interface WebhookLog {
  id: string
  timestamp: string
  source: string
  event: string
  account: string
  status: string
  code: number
}

const SAMPLE_LOGS: WebhookLog[] = [
  {
    id: 'log-01',
    timestamp: 'Just now',
    source: 'Zendesk',
    event: 'ticket.created (SAML 2.0 mandate)',
    account: 'Cyberdyne Systems ($240k ARR)',
    status: 'INGESTED',
    code: 200,
  },
  {
    id: 'log-02',
    timestamp: '12m ago',
    source: 'Gong',
    event: 'call.transcript.completed (Audit logs requirement)',
    account: 'Stark Industries AI ($320k ARR)',
    status: 'INGESTED',
    code: 200,
  },
  {
    id: 'log-03',
    timestamp: '34m ago',
    source: 'Slack',
    event: 'message.posted (#enterprise-feedback)',
    account: 'Wayne Enterprises ($180k ARR)',
    status: 'INGESTED',
    code: 200,
  },
  {
    id: 'log-04',
    timestamp: '1h ago',
    source: 'Intercom',
    event: 'conversation.closed (Bulk CSV timeout)',
    account: 'Pied Piper ($45k ARR)',
    status: 'INGESTED',
    code: 200,
  },
  {
    id: 'log-05',
    timestamp: '2h ago',
    source: 'Stripe',
    event: 'customer.subscription.updated',
    account: '12 Customer Accounts Re-synced',
    status: 'SYNCED',
    code: 200,
  },
]

export const ConnectorsPage: React.FC<ConnectorsPageProps> = ({ onRefreshAll }) => {
  const [syncingStripe, setSyncingStripe] = useState(false)
  const [syncingHubSpot, setSyncingHubSpot] = useState(false)
  const [seeding, setSeeding] = useState(false)
  const [clustering, setClustering] = useState(false)
  const [message, setMessage] = useState<string | null>(null)
  const [copiedEndpoint, setCopiedEndpoint] = useState<string | null>(null)

  const handleSyncStripe = async () => {
    setSyncingStripe(true)
    setMessage(null)
    try {
      const res = await syncCRM('stripe')
      setMessage(res.message)
      onRefreshAll()
    } catch (err: any) {
      alert(`Stripe error: ${err.message}`)
    } finally {
      setSyncingStripe(false)
    }
  }

  const handleSyncHubSpot = async () => {
    setSyncingHubSpot(true)
    setMessage(null)
    try {
      const res = await syncCRM('hubspot')
      setMessage(res.message)
      onRefreshAll()
    } catch (err: any) {
      alert(`HubSpot error: ${err.message}`)
    } finally {
      setSyncingHubSpot(false)
    }
  }

  const handleSeed = async () => {
    setSeeding(true)
    setMessage(null)
    try {
      const res = await seedDemoFeedback()
      setMessage(res.message)
      onRefreshAll()
    } catch (err: any) {
      alert(`Seed error: ${err.message}`)
    } finally {
      setSeeding(false)
    }
  }

  const handleCluster = async () => {
    setClustering(true)
    setMessage(null)
    try {
      const res = await runClustering()
      setMessage(
        `Clustering complete: Created ${res.clusters_created} opportunities, demanded ARR $${res.total_demanded_arr_usd.toLocaleString()}`
      )
      onRefreshAll()
    } catch (err: any) {
      alert(`Clustering error: ${err.message}`)
    } finally {
      setClustering(false)
    }
  }

  const handleCopy = (path: string) => {
    const fullUrl = `${window.location.origin}${path}`
    navigator.clipboard.writeText(fullUrl)
    setCopiedEndpoint(path)
    setTimeout(() => setCopiedEndpoint(null), 2000)
  }

  return (
    <div className="space-y-8 pb-12">
      {/* Overview Banner */}
      <div className="bg-[#0f1422]/90 border border-slate-800/80 rounded-2xl p-6 sm:p-7 shadow-lg backdrop-blur-md">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h1 className="text-lg font-bold text-white flex items-center gap-2.5">
              <span className="w-2.5 h-2.5 rounded-full bg-blue-500 animate-pulse" />
              Multi-Channel Ingestion &amp; CRM Integrations
            </h1>
            <p className="text-xs text-slate-400 mt-1 max-w-2xl leading-relaxed">
              Connect customer communication channels, sales call transcripts, and billing data to stream customer quotes directly into the ARR-weighted synthesis pipeline.
            </p>
          </div>
          <div className="flex items-center gap-2">
            <span className="px-3 py-1 rounded-full text-xs font-mono font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
              4 CHANNELS ACTIVE
            </span>
          </div>
        </div>

        {message && (
          <div className="mt-4 p-3.5 rounded-xl bg-emerald-950/40 border border-emerald-800/50 text-xs text-emerald-300 flex items-center gap-2">
            <span>✓</span> {message}
          </div>
        )}
      </div>

      {/* Action Hub - 4 Balanced Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Stripe Billing */}
        <div className="bg-[#0f1422]/80 border border-slate-800/80 hover:border-slate-700/90 rounded-2xl p-6 shadow-md transition-all flex flex-col justify-between">
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-xl bg-blue-500/10 border border-blue-500/20 flex items-center justify-center text-lg">
                  💳
                </div>
                <div>
                  <h2 className="text-sm font-bold text-white">Stripe Billing &amp; Subscriptions</h2>
                  <span className="text-[11px] text-slate-400 font-mono">Customer Contract ARR &amp; Renewals</span>
                </div>
              </div>
              <span className="px-2.5 py-0.5 rounded-full text-[10px] font-mono font-semibold bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
                CONNECTED
              </span>
            </div>
            <p className="text-xs text-slate-400 leading-relaxed pt-1">
              Pulls active customer subscriptions, maps email domains to corporate ARR, and tracks renewal dates to calculate urgency multipliers.
            </p>
            <div className="flex items-center gap-4 text-xs font-mono text-slate-400 pt-1">
              <span>Accounts: <b className="text-slate-200">12 enterprise</b></span>
              <span>•</span>
              <span>Tracked ARR: <b className="text-emerald-400">$1,736,500</b></span>
            </div>
          </div>
          <div className="pt-5 mt-4 border-t border-slate-800/50">
            <button
              onClick={handleSyncStripe}
              disabled={syncingStripe}
              className="w-full py-2.5 px-4 rounded-xl text-xs font-semibold bg-slate-800 hover:bg-slate-700 text-slate-200 hover:text-white border border-slate-700 hover:border-slate-600 transition-all flex items-center justify-center gap-2 disabled:opacity-50"
            >
              {syncingStripe ? (
                <>
                  <span className="w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  <span>Syncing Customers &amp; Subscriptions...</span>
                </>
              ) : (
                <span>Sync Stripe Customers &amp; Subscriptions</span>
              )}
            </button>
          </div>
        </div>

        {/* HubSpot Deals */}
        <div className="bg-[#0f1422]/80 border border-slate-800/80 hover:border-slate-700/90 rounded-2xl p-6 shadow-md transition-all flex flex-col justify-between">
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-xl bg-amber-500/10 border border-amber-500/20 flex items-center justify-center text-lg">
                  📊
                </div>
                <div>
                  <h2 className="text-sm font-bold text-white">HubSpot CRM &amp; Pipeline Deals</h2>
                  <span className="text-[11px] text-slate-400 font-mono">Company Stage &amp; Churn Risk</span>
                </div>
              </div>
              <span className="px-2.5 py-0.5 rounded-full text-[10px] font-mono font-semibold bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
                CONNECTED
              </span>
            </div>
            <p className="text-xs text-slate-400 leading-relaxed pt-1">
              Enriches customer accounts with contract close dates, renewal health scores, and sales rep notes on at-risk renewals.
            </p>
            <div className="flex items-center gap-4 text-xs font-mono text-slate-400 pt-1">
              <span>Lifecycle: <b className="text-slate-200">Customer</b></span>
              <span>•</span>
              <span>Avg Churn Risk: <b className="text-amber-400">0.42</b></span>
            </div>
          </div>
          <div className="pt-5 mt-4 border-t border-slate-800/50">
            <button
              onClick={handleSyncHubSpot}
              disabled={syncingHubSpot}
              className="w-full py-2.5 px-4 rounded-xl text-xs font-semibold bg-slate-800 hover:bg-slate-700 text-slate-200 hover:text-white border border-slate-700 hover:border-slate-600 transition-all flex items-center justify-center gap-2 disabled:opacity-50"
            >
              {syncingHubSpot ? (
                <>
                  <span className="w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  <span>Syncing Deals &amp; Risk Scores...</span>
                </>
              ) : (
                <span>Sync HubSpot Companies &amp; Closed Deals</span>
              )}
            </button>
          </div>
        </div>

        {/* Seed Enterprise Data */}
        <div className="bg-[#0f1422]/80 border border-slate-800/80 hover:border-slate-700/90 rounded-2xl p-6 shadow-md transition-all flex flex-col justify-between">
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-xl bg-indigo-500/10 border border-indigo-500/20 flex items-center justify-center text-lg">
                  ⚡
                </div>
                <div>
                  <h2 className="text-sm font-bold text-white">Benchmark Enterprise Dataset</h2>
                  <span className="text-[11px] text-slate-400 font-mono">Zendesk, Intercom, Gong, Slack</span>
                </div>
              </div>
              <span className="px-2.5 py-0.5 rounded-full text-[10px] font-mono font-semibold bg-blue-500/15 text-blue-400 border border-blue-500/30">
                OFFLINE BENCHMARK
              </span>
            </div>
            <p className="text-xs text-slate-400 leading-relaxed pt-1">
              Seeds authentic enterprise quotes (SAML SSO, SOC 2 audit logs, RBAC permissions, bulk export APIs) from Cyberdyne, Stark Industries, Acme Tech, and Wayne Sec.
            </p>
            <div className="flex items-center gap-4 text-xs font-mono text-slate-400 pt-1">
              <span>Quotes: <b className="text-slate-200">13 quotes</b></span>
              <span>•</span>
              <span>Channels: <b className="text-cyan-400">4 platforms</b></span>
            </div>
          </div>
          <div className="pt-5 mt-4 border-t border-slate-800/50">
            <button
              onClick={handleSeed}
              disabled={seeding}
              className="w-full py-2.5 px-4 rounded-xl text-xs font-semibold bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white shadow-sm transition-all flex items-center justify-center gap-2 disabled:opacity-50"
            >
              {seeding ? (
                <>
                  <span className="w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  <span>Seeding Benchmark Quotes...</span>
                </>
              ) : (
                <span>Seed Realistic Multi-Channel Data</span>
              )}
            </button>
          </div>
        </div>

        {/* Run HDBSCAN */}
        <div className="bg-[#0f1422]/80 border border-slate-800/80 hover:border-slate-700/90 rounded-2xl p-6 shadow-md transition-all flex flex-col justify-between">
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-lg">
                  🧠
                </div>
                <div>
                  <h2 className="text-sm font-bold text-white">HDBSCAN Semantic Clustering</h2>
                  <span className="text-[11px] text-slate-400 font-mono">Subword Embeddings &amp; Prim's MST</span>
                </div>
              </div>
              <span className="px-2.5 py-0.5 rounded-full text-[10px] font-mono font-semibold bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
                PURE GO ENGINE
              </span>
            </div>
            <p className="text-xs text-slate-400 leading-relaxed pt-1">
              Executes 128-dimensional subword vectorization, builds mutual reachability graphs, and groups complaints into revenue-weighted opportunities.
            </p>
            <div className="flex items-center gap-4 text-xs font-mono text-slate-400 pt-1">
              <span>Opportunities: <b className="text-slate-200">2 synthesized</b></span>
              <span>•</span>
              <span>Speed: <b className="text-emerald-400">&lt; 30 ms</b></span>
            </div>
          </div>
          <div className="pt-5 mt-4 border-t border-slate-800/50">
            <button
              onClick={handleCluster}
              disabled={clustering}
              className="w-full py-2.5 px-4 rounded-xl text-xs font-semibold bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white shadow-sm transition-all flex items-center justify-center gap-2 disabled:opacity-50"
            >
              {clustering ? (
                <>
                  <span className="w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  <span>Running Clustering...</span>
                </>
              ) : (
                <span>Execute HDBSCAN Clustering Run</span>
              )}
            </button>
          </div>
        </div>
      </div>

      {/* Webhook URLs Reference */}
      <div className="bg-[#0f1422]/90 border border-slate-800/80 rounded-2xl p-6 sm:p-7 shadow-lg">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-sm font-bold text-white uppercase tracking-wider">Inbound Webhook Endpoints</h2>
            <p className="text-xs text-slate-400 mt-0.5">
              Configure these webhook targets in your support, CRM, and communication admin panels:
            </p>
          </div>
          <span className="text-[11px] font-mono text-slate-500">HTTP REST / JSON</span>
        </div>

        <div className="space-y-3 pt-2">
          {[
            {
              path: '/api/v1/ingest/zendesk',
              title: 'Zendesk Support Triggers',
              desc: 'Parses ticket subjects, comments, and customer email domains',
              tag: 'ZENDESK',
              color: 'text-blue-400',
            },
            {
              path: '/api/v1/ingest/intercom',
              title: 'Intercom Conversation Webhooks',
              desc: 'Captures live chat transcripts, user profiles, and conversation tags',
              tag: 'INTERCOM',
              color: 'text-indigo-400',
            },
            {
              path: '/api/v1/ingest/gong',
              title: 'Gong Call Transcript Webhooks',
              desc: 'Extracts buyer objections and feature requests from recorded sales calls',
              tag: 'GONG',
              color: 'text-purple-400',
            },
            {
              path: '/api/v1/ingest/slack',
              title: 'Slack Events API & Webhooks',
              desc: 'Listens to Slack Connect shared channels and internal #feedback channels',
              tag: 'SLACK',
              color: 'text-emerald-400',
            },
          ].map((item) => (
            <div
              key={item.path}
              className="flex flex-col sm:flex-row sm:items-center justify-between p-3.5 rounded-xl bg-slate-950/60 border border-slate-800/70 hover:border-slate-700/80 transition-all gap-3"
            >
              <div className="space-y-1">
                <div className="flex items-center gap-2.5">
                  <span className="px-2 py-0.5 rounded font-mono font-bold text-[10px] bg-slate-800 text-slate-300">
                    POST
                  </span>
                  <span className={`font-mono text-xs font-semibold ${item.color}`}>
                    {item.path}
                  </span>
                </div>
                <div className="text-xs text-slate-400">
                  <span className="font-semibold text-slate-300">{item.title}</span> &mdash; {item.desc}
                </div>
              </div>

              <button
                onClick={() => handleCopy(item.path)}
                className="self-start sm:self-center px-3 py-1.5 rounded-lg text-xs font-medium bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-white border border-slate-700/80 transition-all flex items-center gap-1.5 whitespace-nowrap"
              >
                <span>{copiedEndpoint === item.path ? '✓ Copied' : '📋 Copy URL'}</span>
              </button>
            </div>
          ))}
        </div>
      </div>

      {/* Activity & Sync Event Log */}
      <div className="bg-[#0f1422]/90 border border-slate-800/80 rounded-2xl p-6 sm:p-7 shadow-lg">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h2 className="text-sm font-bold text-white uppercase tracking-wider">Live Ingestion &amp; Sync Audit Trail</h2>
            <p className="text-xs text-slate-400 mt-0.5">Recent webhook triggers, payload extractions, and CRM sync operations</p>
          </div>
          <span className="px-2.5 py-1 rounded-full text-[10px] font-mono font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
            AUDIT TRAIL ONLINE
          </span>
        </div>

        <div className="overflow-x-auto">
          <table className="w-full text-left text-xs">
            <thead>
              <tr className="border-b border-slate-800/80 text-[10px] uppercase font-semibold text-slate-400">
                <th className="pb-3 font-mono">Timestamp</th>
                <th className="pb-3">Source Channel</th>
                <th className="pb-3">Event Trigger</th>
                <th className="pb-3">Customer Account / Scope</th>
                <th className="pb-3 text-right">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800/40 text-slate-300">
              {SAMPLE_LOGS.map((log) => (
                <tr key={log.id} className="hover:bg-slate-900/30 transition-colors">
                  <td className="py-3 font-mono text-[11px] text-slate-500">{log.timestamp}</td>
                  <td className="py-3 font-medium text-white">{log.source}</td>
                  <td className="py-3 text-slate-300 font-mono text-[11px]">{log.event}</td>
                  <td className="py-3 text-slate-300">{log.account}</td>
                  <td className="py-3 text-right">
                    <span className="px-2 py-0.5 rounded text-[10px] font-mono font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
                      {log.code} {log.status}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
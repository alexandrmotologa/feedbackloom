import React, { useState } from 'react'
import { syncCRM, seedDemoFeedback, runClustering } from '../api'

interface ConnectorsPageProps {
  onRefreshAll: () => void
}

export const ConnectorsPage: React.FC<ConnectorsPageProps> = ({ onRefreshAll }) => {
  const [syncingStripe, setSyncingStripe] = useState(false)
  const [syncingHubSpot, setSyncingHubSpot] = useState(false)
  const [seeding, setSeeding] = useState(false)
  const [clustering, setClustering] = useState(false)
  const [message, setMessage] = useState<string | null>(null)

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
      setMessage(`Clustering complete: Created ${res.clusters_created} opportunities, demanded ARR $${res.total_demanded_arr_usd.toLocaleString()}`)
      onRefreshAll()
    } catch (err: any) {
      alert(`Clustering error: ${err.message}`)
    } finally {
      setClustering(false)
    }
  }

  return (
    <div className="space-y-6">
      {/* Overview Card */}
      <div className="bg-[#121826] border border-[#1e293b] rounded-xl p-5">
        <h2 className="text-base font-bold text-white">Multi-Channel Ingestion & CRM Integrations</h2>
        <p className="text-xs text-slate-400 mt-1">
          Connect your customer communication channels and billing systems to stream customer quotes directly into the revenue-weighted synthesis pipeline.
        </p>

        {message && (
          <div className="mt-4 p-3 rounded-lg bg-emerald-950/40 border border-emerald-800/40 text-xs text-emerald-300">
            ✓ {message}
          </div>
        )}
      </div>

      {/* Action Hub */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {/* Stripe Billing */}
        <div className="bg-[#131929] border border-[#1e293b] rounded-xl p-5 space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2.5">
              <span className="text-xl">💳</span>
              <div>
                <h3 className="text-sm font-bold text-white">Stripe Billing & Subscriptions</h3>
                <span className="text-[10px] text-slate-400 font-mono">Syncs ARR, Plan Tiers & Subscriptions</span>
              </div>
            </div>
            <span className="badge badge-enterprise">Connected</span>
          </div>
          <p className="text-xs text-slate-400 leading-relaxed">
            Pulls active subscriptions, maps email domains to customer ARR, and calculates renewal timelines.
          </p>
          <button
            onClick={handleSyncStripe}
            disabled={syncingStripe}
            className="w-full py-2 rounded-lg text-xs font-semibold bg-[#1c263c] hover:bg-[#253352] text-slate-200 border border-[#2b3a5d] transition-all"
          >
            {syncingStripe ? 'Syncing...' : 'Sync Customers & Subscriptions'}
          </button>
        </div>

        {/* HubSpot Deals */}
        <div className="bg-[#131929] border border-[#1e293b] rounded-xl p-5 space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2.5">
              <span className="text-xl">📊</span>
              <div>
                <h3 className="text-sm font-bold text-white">HubSpot CRM & Deals</h3>
                <span className="text-[10px] text-slate-400 font-mono">Syncs Company ARR & Churn Probability</span>
              </div>
            </div>
            <span className="badge badge-enterprise">Connected</span>
          </div>
          <p className="text-xs text-slate-400 leading-relaxed">
            Enriches feedback accounts with contract close dates, annual contract values, and risk scores.
          </p>
          <button
            onClick={handleSyncHubSpot}
            disabled={syncingHubSpot}
            className="w-full py-2 rounded-lg text-xs font-semibold bg-[#1c263c] hover:bg-[#253352] text-slate-200 border border-[#2b3a5d] transition-all"
          >
            {syncingHubSpot ? 'Syncing...' : 'Sync Companies & Closed Deals'}
          </button>
        </div>

        {/* Seed Enterprise Data */}
        <div className="bg-[#131929] border border-[#1e293b] rounded-xl p-5 space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2.5">
              <span className="text-xl">⚡</span>
              <div>
                <h3 className="text-sm font-bold text-white">Benchmark Enterprise Dataset</h3>
                <span className="text-[10px] text-slate-400 font-mono">Zendesk, Intercom, Gong, Slack</span>
              </div>
            </div>
            <span className="badge badge-growth">Demo</span>
          </div>
          <p className="text-xs text-slate-400 leading-relaxed">
            Seeds realistic enterprise quotes (SAML SSO, SOC 2 audit logs, RBAC scopes, bulk exports) from Cyberdyne, Stark Industries, Acme Tech, and Wayne Sec.
          </p>
          <button
            onClick={handleSeed}
            disabled={seeding}
            className="w-full py-2 rounded-lg text-xs font-semibold bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white shadow-sm transition-all"
          >
            {seeding ? 'Seeding...' : 'Seed Realistic Multi-Channel Data'}
          </button>
        </div>

        {/* Run HDBSCAN */}
        <div className="bg-[#131929] border border-[#1e293b] rounded-xl p-5 space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-2.5">
              <span className="text-xl">🧠</span>
              <div>
                <h3 className="text-sm font-bold text-white">HDBSCAN Semantic Clustering</h3>
                <span className="text-[10px] text-slate-400 font-mono">Dense Vector Density Clustering</span>
              </div>
            </div>
            <span className="badge badge-enterprise">AI Engine</span>
          </div>
          <p className="text-xs text-slate-400 leading-relaxed">
            Executes vector embeddings, constructs mutual reachability graphs, and groups complaints into revenue-weighted opportunities.
          </p>
          <button
            onClick={handleCluster}
            disabled={clustering}
            className="w-full py-2 rounded-lg text-xs font-semibold bg-gradient-to-r from-emerald-600 to-teal-600 hover:from-emerald-500 hover:to-teal-500 text-white shadow-sm transition-all"
          >
            {clustering ? 'Clustering...' : 'Execute HDBSCAN Clustering Run'}
          </button>
        </div>
      </div>

      {/* Webhook URLs Reference */}
      <div className="bg-[#121826] border border-[#1e293b] rounded-xl p-5 space-y-4">
        <h3 className="text-sm font-bold text-white">Inbound Webhook Endpoints</h3>
        <p className="text-xs text-slate-400">
          Configure these webhook URLs in your respective customer support and communication admin panels:
        </p>

        <div className="space-y-2 text-xs font-mono">
          <div className="flex items-center justify-between p-2.5 rounded bg-[#0c101a] border border-[#1e293b]">
            <span className="text-blue-400">POST /api/v1/ingest/zendesk</span>
            <span className="text-slate-500">Zendesk Support Triggers</span>
          </div>
          <div className="flex items-center justify-between p-2.5 rounded bg-[#0c101a] border border-[#1e293b]">
            <span className="text-indigo-400">POST /api/v1/ingest/intercom</span>
            <span className="text-slate-500">Intercom Conversation Webhooks</span>
          </div>
          <div className="flex items-center justify-between p-2.5 rounded bg-[#0c101a] border border-[#1e293b]">
            <span className="text-purple-400">POST /api/v1/ingest/gong</span>
            <span className="text-slate-500">Gong Call Transcript Webhooks</span>
          </div>
          <div className="flex items-center justify-between p-2.5 rounded bg-[#0c101a] border border-[#1e293b]">
            <span className="text-emerald-400">POST /api/v1/ingest/slack</span>
            <span className="text-slate-500">Slack Events API & Webhooks</span>
          </div>
        </div>
      </div>
    </div>
  )
}

import React, { useState, useEffect } from 'react'
import { Opportunity, DashboardStats } from './types'
import {
  fetchStats,
  fetchOpportunities,
  updateOpportunityStatus,
  updateOpportunityEffort,
  runClustering,
} from './api'
import { Header } from './components/Header'
import { ClosedLoopModal } from './components/ClosedLoopModal'
import { OpportunitiesPage } from './pages/OpportunitiesPage'
import { PRDStudioPage } from './pages/PRDStudioPage'
import { EvidenceStreamPage } from './pages/EvidenceStreamPage'
import { ConnectorsPage } from './pages/ConnectorsPage'
import { ClosedLoopPage } from './pages/ClosedLoopPage'
import { SandboxPage } from './pages/SandboxPage'

export const App: React.FC = () => {
  const [activeTab, setActiveTab] = useState<string>('opportunities')
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [opportunities, setOpportunities] = useState<Opportunity[]>([])
  const [selectedOpportunityId, setSelectedOpportunityId] = useState<string | null>(null)
  const [closedLoopModalOpp, setClosedLoopModalOpp] = useState<Opportunity | null>(null)
  const [loading, setLoading] = useState<boolean>(true)

  useEffect(() => {
    loadAll()
  }, [])

  const loadAll = async () => {
    try {
      const [sData, oppsData] = await Promise.all([fetchStats(), fetchOpportunities()])
      setStats(sData)
      setOpportunities(oppsData)
      if (oppsData.length > 0 && !selectedOpportunityId) {
        setSelectedOpportunityId(oppsData[0].id)
      }
    } catch (err) {
      console.error('Failed to load initial data:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleSelectPRD = (id: string) => {
    setSelectedOpportunityId(id)
    setActiveTab('prd-studio')
  }

  const handleStatusChange = async (id: string, newStatus: string) => {
    try {
      await updateOpportunityStatus(id, newStatus)
      setOpportunities((prev) =>
        prev.map((o) => (o.id === id ? { ...o, status: newStatus as any } : o))
      )
    } catch (err: any) {
      alert(`Failed to update status: ${err.message}`)
    }
  }

  const handleEffortChange = async (id: string, sprints: number) => {
    try {
      await updateOpportunityEffort(id, sprints)
      setOpportunities((prev) =>
        prev.map((o) =>
          o.id === id
            ? {
                ...o,
                effort_sprints: sprints,
                roi_index: sprints > 0 ? o.total_arr_usd / sprints : o.total_arr_usd,
              }
            : o
        )
      )
    } catch (err: any) {
      alert(`Failed to update effort estimate: ${err.message}`)
    }
  }

  const handleRunClustering = async (minClusterSize?: number, minPoints?: number) => {
    try {
      await runClustering(minClusterSize, minPoints)
      await loadAll()
    } catch (err: any) {
      alert(`Clustering failed: ${err.message}`)
    }
  }

  return (
    <div className="min-h-screen bg-[#07090e] text-slate-100 flex flex-col font-sans selection:bg-blue-600 selection:text-white">
      <Header
        stats={stats}
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        onRefresh={loadAll}
      />

      <main className="flex-1 max-w-[1520px] w-full mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {loading ? (
          <div className="py-32 text-center text-slate-400 text-sm space-y-3">
            <div className="inline-block w-9 h-9 border-2 border-blue-500 border-t-transparent rounded-full animate-spin" />
            <div className="font-semibold text-white">Initializing feedbackloom Product Studio...</div>
            <p className="text-xs text-slate-500">Connecting to SQLite WAL database and loading portfolio metrics.</p>
          </div>
        ) : (
          <>
            {activeTab === 'opportunities' && (
              <OpportunitiesPage
                opportunities={opportunities}
                totalPortfolioARR={stats?.portfolio_arr_usd || 0}
                onSelectPRD={handleSelectPRD}
                onStatusChange={handleStatusChange}
                onOpenClosedLoop={(opp) => setClosedLoopModalOpp(opp)}
                onRunClustering={handleRunClustering}
                onEffortChange={handleEffortChange}
              />
            )}

            {activeTab === 'prd-studio' && (
              <PRDStudioPage
                opportunities={opportunities}
                selectedOpportunityId={selectedOpportunityId}
                onSelectOpportunity={setSelectedOpportunityId}
              />
            )}

            {activeTab === 'sandbox' && (
              <SandboxPage onSelectPRD={handleSelectPRD} />
            )}

            {activeTab === 'evidence' && <EvidenceStreamPage />}

            {activeTab === 'connectors' && <ConnectorsPage onRefreshAll={loadAll} />}

            {activeTab === 'closed-loop' && (
              <ClosedLoopPage
                opportunities={opportunities}
                onOpenClosedLoop={(opp) => setClosedLoopModalOpp(opp)}
                onStatusChange={handleStatusChange}
              />
            )}
          </>
        )}
      </main>

      {/* Closed Loop Outreach Modal */}
      {closedLoopModalOpp && (
        <ClosedLoopModal
          opportunity={closedLoopModalOpp}
          onClose={() => setClosedLoopModalOpp(null)}
        />
      )}

      {/* Persistent Footer */}
      <footer className="border-t border-slate-800/60 bg-[#090c14] py-5 px-6 text-center text-xs text-slate-500 flex flex-col sm:flex-row items-center justify-between max-w-[1520px] mx-auto w-full gap-2">
        <div className="flex items-center gap-2">
          <span className="font-bold text-slate-400">feedbackloom</span>
          <span>&bull;</span>
          <span>Autonomous Customer Voice &amp; ARR-Weighted PRD Synthesizer</span>
        </div>
        <div className="flex items-center gap-4 text-[11px] font-mono text-slate-500">
          <span>Pure Go 1.24</span>
          <span>&bull;</span>
          <span>React 19 + Tailwind CSS</span>
          <span>&bull;</span>
          <span>SQLite WAL</span>
        </div>
      </footer>
    </div>
  )
}
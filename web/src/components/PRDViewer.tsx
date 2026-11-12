import React, { useState } from 'react'
import { Opportunity, EnrichedFeedbackItem, GitHubIssuePayload, JiraEpicPayload } from '../types'
import { generatePRD, exportToNotion, fetchGitHubIssue, fetchJiraEpic } from '../api'

interface PRDViewerProps {
  opportunity: Opportunity
  quotes: EnrichedFeedbackItem[]
  onPRDUpdated: (markdown: string) => void
}

export const PRDViewer: React.FC<PRDViewerProps> = ({ opportunity, quotes, onPRDUpdated }) => {
  const [loading, setLoading] = useState(false)
  const [notionStatus, setNotionStatus] = useState<string | null>(null)
  const [copied, setCopied] = useState(false)
  const [activeExportModal, setActiveExportModal] = useState<'github' | 'jira' | null>(null)
  const [githubPayload, setGithubPayload] = useState<GitHubIssuePayload | null>(null)
  const [jiraPayload, setJiraPayload] = useState<JiraEpicPayload | null>(null)
  const [viewMode, setViewMode] = useState<'visual' | 'raw'>('visual')

  const handleGenerate = async () => {
    setLoading(true)
    try {
      const res = await generatePRD(opportunity.id)
      onPRDUpdated(res.markdown)
    } catch (err: any) {
      alert(`Error generating PRD: ${err.message}`)
    } finally {
      setLoading(false)
    }
  }

  const handleExportNotion = async () => {
    setLoading(true)
    setNotionStatus(null)
    try {
      const res = await exportToNotion(opportunity.id)
      setNotionStatus(res.url)
    } catch (err: any) {
      alert(`Notion export error: ${err.message}`)
    } finally {
      setLoading(false)
    }
  }

  const handleOpenGitHub = async () => {
    try {
      const data = await fetchGitHubIssue(opportunity.id)
      setGithubPayload(data)
      setActiveExportModal('github')
    } catch (err: any) {
      alert(`GitHub export error: ${err.message}`)
    }
  }

  const handleOpenJira = async () => {
    try {
      const data = await fetchJiraEpic(opportunity.id)
      setJiraPayload(data)
      setActiveExportModal('jira')
    } catch (err: any) {
      alert(`Jira export error: ${err.message}`)
    }
  }

  const handleDownload = () => {
    if (!opportunity.generated_prd_markdown) return
    const blob = new Blob([opportunity.generated_prd_markdown], { type: 'text/markdown;charset=utf-8' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `PRD_${opportunity.title.replace(/[^a-zA-Z0-9]/g, '_')}.md`
    a.click()
    URL.revokeObjectURL(url)
  }

  const handleCopy = () => {
    if (!opportunity.generated_prd_markdown) return
    navigator.clipboard.writeText(opportunity.generated_prd_markdown)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  const formatUSD = (val?: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      maximumFractionDigits: 0,
    }).format(val || 0)
  }

  return (
    <div className="bg-[#0f1422]/90 border border-slate-800/80 rounded-2xl overflow-hidden shadow-xl backdrop-blur-md pb-12">
      {/* Action Toolbar */}
      <div className="bg-slate-950/80 px-6 sm:px-8 py-5 border-b border-slate-800/80 flex flex-wrap items-center justify-between gap-4">
        <div>
          <span className="text-[10px] uppercase font-bold text-slate-400 tracking-wider block">
            Executive Product Specification
          </span>
          <h2 className="text-xl font-bold text-white flex items-center gap-3 mt-0.5">
            <span>{opportunity.title}</span>
            <span className="text-xs px-2.5 py-0.5 rounded-full font-mono bg-blue-500/10 text-blue-400 border border-blue-500/30">
              {opportunity.category}
            </span>
          </h2>
        </div>

        <div className="flex flex-wrap items-center gap-2.5">
          <button
            onClick={handleGenerate}
            disabled={loading}
            className="px-4 py-2 rounded-xl text-xs font-semibold bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white shadow-sm transition-all flex items-center gap-2"
          >
            <span>✨</span>
            <span>{opportunity.generated_prd_markdown ? 'Re-Synthesize PRD' : 'Synthesize Autonomous PRD'}</span>
          </button>

          {opportunity.generated_prd_markdown && (
            <>
              {/* View Mode Toggle */}
              <div className="flex items-center bg-slate-900 border border-slate-800 rounded-xl p-0.5">
                <button
                  onClick={() => setViewMode('visual')}
                  className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
                    viewMode === 'visual' ? 'bg-blue-600 text-white shadow-sm' : 'text-slate-400 hover:text-white'
                  }`}
                >
                  Visual Spec
                </button>
                <button
                  onClick={() => setViewMode('raw')}
                  className={`px-3 py-1.5 rounded-lg text-xs font-medium transition-all ${
                    viewMode === 'raw' ? 'bg-blue-600 text-white shadow-sm' : 'text-slate-400 hover:text-white'
                  }`}
                >
                  Raw Markdown
                </button>
              </div>

              <button
                onClick={handleCopy}
                className="px-3.5 py-2 rounded-xl text-xs font-medium bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-white border border-slate-800 transition-colors"
              >
                {copied ? '✓ Copied' : 'Copy MD'}
              </button>

              <button
                onClick={handleDownload}
                className="px-3.5 py-2 rounded-xl text-xs font-medium bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-white border border-slate-800 transition-colors flex items-center gap-1.5"
              >
                <span>💾</span> .md
              </button>

              <button
                onClick={handleExportNotion}
                disabled={loading}
                className="px-3.5 py-2 rounded-xl text-xs font-medium bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-white border border-slate-800 transition-colors flex items-center gap-1.5"
              >
                <span>🚀</span> Notion
              </button>

              <button
                onClick={handleOpenGitHub}
                className="px-3.5 py-2 rounded-xl text-xs font-medium bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-white border border-slate-800 transition-colors flex items-center gap-1.5"
              >
                <span>🐙</span> GitHub Issue
              </button>

              <button
                onClick={handleOpenJira}
                className="px-3.5 py-2 rounded-xl text-xs font-medium bg-slate-900 hover:bg-slate-800 text-slate-300 hover:text-white border border-slate-800 transition-colors flex items-center gap-1.5"
              >
                <span>🔷</span> Jira Epic
              </button>
            </>
          )}
        </div>
      </div>

      {notionStatus && (
        <div className="bg-emerald-950/40 border-b border-emerald-800/40 px-6 sm:px-8 py-3 text-xs text-emerald-300 flex items-center justify-between">
          <span>✓ Successfully exported PRD to Notion workspace.</span>
          <a href={notionStatus} target="_blank" rel="noreferrer" className="underline font-semibold hover:text-emerald-200">
            Open in Notion &rarr;
          </a>
        </div>
      )}

      {/* PRD Content Body */}
      <div className="p-6 sm:p-8">
        {loading ? (
          <div className="py-20 text-center text-slate-400 space-y-3">
            <div className="inline-block w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin" />
            <div className="text-sm font-semibold text-white">Synthesizing comprehensive Product Requirements Document...</div>
            <p className="text-xs text-slate-500">Extracting customer evidence quotes, formulating user stories, and generating Gherkin criteria.</p>
          </div>
        ) : opportunity.generated_prd_markdown ? (
          viewMode === 'visual' ? (
            /* Visual Executive Spec Layout */
            <div className="space-y-8 max-w-5xl mx-auto">
              {/* Executive Metadata KPI Banner */}
              <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 p-5 rounded-2xl bg-slate-950/80 border border-slate-800/80 shadow-md">
                <div>
                  <span className="text-[10px] uppercase font-bold text-slate-400">Total Demanded ARR</span>
                  <div className="text-lg font-bold font-mono text-emerald-400 mt-0.5">{formatUSD(opportunity.total_arr_usd)}</div>
                </div>
                <div>
                  <span className="text-[10px] uppercase font-bold text-slate-400">Weighted Score (WRS)</span>
                  <div className="text-lg font-bold font-mono text-blue-400 mt-0.5">{opportunity.weighted_score.toFixed(0)}</div>
                </div>
                <div>
                  <span className="text-[10px] uppercase font-bold text-slate-400">Requesting Accounts</span>
                  <div className="text-lg font-bold font-mono text-slate-200 mt-0.5">{opportunity.account_count} Enterprise</div>
                </div>
                <div>
                  <span className="text-[10px] uppercase font-bold text-slate-400">Engineering Effort</span>
                  <div className="text-lg font-bold font-mono text-cyan-400 mt-0.5">{opportunity.effort_sprints || 2} Sprints</div>
                </div>
              </div>

              {/* Verified Customer Quotes Section */}
              <div className="space-y-4">
                <h3 className="text-sm font-bold uppercase tracking-wider text-white flex items-center gap-2 border-b border-slate-800/80 pb-3">
                  <span className="text-blue-400 font-mono">01.</span> Verified Customer Evidence Quotes ({quotes.length})
                </h3>
                <div className="grid grid-cols-1 gap-4">
                  {quotes.map((q, idx) => (
                    <div
                      key={q.id || idx}
                      className="p-5 rounded-xl bg-slate-950/60 border border-slate-800/70 hover:border-slate-700/80 space-y-2.5 transition-all shadow-sm"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center gap-2.5">
                          <span className="text-xs font-bold text-white">
                            {q.account_name || q.account_domain || 'Enterprise Account'}
                          </span>
                          <span className="px-2 py-0.5 rounded text-[10px] font-mono bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
                            {formatUSD(q.account_arr_usd)} ARR
                          </span>
                        </div>
                        <span className="text-[10px] font-mono text-slate-400 uppercase bg-slate-800/60 px-2 py-0.5 rounded">
                          {q.source}
                        </span>
                      </div>
                      <blockquote className="text-xs text-slate-300 italic border-l-2 border-blue-500 pl-3 leading-relaxed">
                        &ldquo;{q.raw_text}&rdquo;
                      </blockquote>
                    </div>
                  ))}
                </div>
              </div>

              {/* Full Markdown Render Container */}
              <div className="p-6 rounded-2xl bg-slate-950/60 border border-slate-800/80 shadow-md">
                <h3 className="text-sm font-bold uppercase tracking-wider text-white flex items-center gap-2 border-b border-slate-800/80 pb-3 mb-4">
                  <span className="text-indigo-400 font-mono">02.</span> Product Specification Document
                </h3>
                <div className="font-mono text-xs text-slate-300 leading-relaxed whitespace-pre-wrap">
                  {opportunity.generated_prd_markdown}
                </div>
              </div>
            </div>
          ) : (
            /* Raw Markdown Mode */
            <div className="space-y-4">
              <pre className="whitespace-pre-wrap font-mono bg-slate-950 p-6 rounded-2xl border border-slate-800 text-slate-200 overflow-x-auto text-xs leading-relaxed">
                {opportunity.generated_prd_markdown}
              </pre>
            </div>
          )
        ) : (
          <div className="py-20 text-center text-slate-400 space-y-4">
            <div className="text-5xl">📄</div>
            <div className="text-lg font-bold text-white">No PRD Generated Yet</div>
            <p className="text-xs text-slate-400 max-w-md mx-auto leading-relaxed">
              Click &quot;Synthesize Autonomous PRD&quot; above to compile problem context, verified quotes from {quotes.length} customer accounts, user stories, and Gherkin acceptance criteria.
            </p>
            <button
              onClick={handleGenerate}
              className="px-5 py-2.5 rounded-xl text-xs font-semibold bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-900/30"
            >
              Synthesize Autonomous PRD Now
            </button>
          </div>
        )}
      </div>

      {/* GitHub Issue Modal */}
      {activeExportModal === 'github' && githubPayload && (
        <div className="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-[#0f1422] border border-slate-800 rounded-2xl max-w-3xl w-full max-h-[85vh] flex flex-col overflow-hidden shadow-2xl">
            <div className="px-6 py-4 bg-slate-950 border-b border-slate-800 flex items-center justify-between">
              <span className="text-sm font-bold text-white">🐙 GitHub Issue Export Payload</span>
              <button onClick={() => setActiveExportModal(null)} className="text-slate-400 hover:text-white font-bold">✕</button>
            </div>
            <div className="p-6 overflow-y-auto space-y-4 flex-1 text-xs">
              <div>
                <label className="text-slate-400 block font-semibold mb-1">Issue Title</label>
                <input
                  readOnly
                  value={githubPayload.title}
                  className="w-full bg-slate-950 border border-slate-800 rounded-lg p-2.5 text-white font-mono"
                />
              </div>
              <div>
                <label className="text-slate-400 block font-semibold mb-1">Target Labels</label>
                <div className="flex flex-wrap gap-1.5">
                  {githubPayload.labels.map((l) => (
                    <span key={l} className="px-2.5 py-0.5 rounded-md bg-blue-500/10 text-blue-400 border border-blue-500/30 font-mono text-[11px]">
                      {l}
                    </span>
                  ))}
                </div>
              </div>
              <div>
                <label className="text-slate-400 block font-semibold mb-1">Issue Markdown Body</label>
                <textarea
                  readOnly
                  rows={14}
                  value={githubPayload.body}
                  className="w-full bg-slate-950 border border-slate-800 rounded-lg p-3 text-slate-300 font-mono text-[11px] leading-relaxed"
                />
              </div>
            </div>
            <div className="px-6 py-4 bg-slate-950 border-t border-slate-800 flex justify-end gap-3">
              <button
                onClick={() => {
                  navigator.clipboard.writeText(`${githubPayload.title}\n\n${githubPayload.body}`)
                  alert('Copied GitHub Issue payload to clipboard!')
                }}
                className="px-4 py-2 rounded-xl text-xs font-semibold bg-blue-600 hover:bg-blue-500 text-white"
              >
                Copy Full Payload
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Jira Epic Modal */}
      {activeExportModal === 'jira' && jiraPayload && (
        <div className="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-[#0f1422] border border-slate-800 rounded-2xl max-w-3xl w-full max-h-[85vh] flex flex-col overflow-hidden shadow-2xl">
            <div className="px-6 py-4 bg-slate-950 border-b border-slate-800 flex items-center justify-between">
              <span className="text-sm font-bold text-white">🔷 Jira Epic Export Payload</span>
              <button onClick={() => setActiveExportModal(null)} className="text-slate-400 hover:text-white font-bold">✕</button>
            </div>
            <div className="p-6 overflow-y-auto space-y-4 flex-1 text-xs">
              <div>
                <label className="text-slate-400 block font-semibold mb-1">Epic Summary</label>
                <input
                  readOnly
                  value={jiraPayload.summary}
                  className="w-full bg-slate-950 border border-slate-800 rounded-lg p-2.5 text-white font-mono"
                />
              </div>
              <div>
                <label className="text-slate-400 block font-semibold mb-1">Epic Description</label>
                <textarea
                  readOnly
                  rows={14}
                  value={jiraPayload.description}
                  className="w-full bg-slate-950 border border-slate-800 rounded-lg p-3 text-slate-300 font-mono text-[11px] leading-relaxed"
                />
              </div>
            </div>
            <div className="px-6 py-4 bg-slate-950 border-t border-slate-800 flex justify-end gap-3">
              <button
                onClick={() => {
                  navigator.clipboard.writeText(JSON.stringify(jiraPayload, null, 2))
                  alert('Copied Jira Epic JSON to clipboard!')
                }}
                className="px-4 py-2 rounded-xl text-xs font-semibold bg-blue-600 hover:bg-blue-500 text-white"
              >
                Copy JSON Payload
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
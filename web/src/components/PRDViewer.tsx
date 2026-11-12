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

  return (
    <div className="bg-[#121826] border border-[#1e293b] rounded-xl overflow-hidden shadow-lg">
      {/* Action Toolbar */}
      <div className="bg-[#0e1320] px-6 py-4 border-b border-[#1e293b] flex flex-wrap items-center justify-between gap-4">
        <div>
          <span className="text-xs uppercase font-semibold text-[#64748b] tracking-wider block">PRD Document</span>
          <h2 className="text-lg font-bold text-white flex items-center gap-2">
            <span>{opportunity.title}</span>
            <span className="text-xs px-2 py-0.5 rounded font-mono bg-blue-500/10 text-blue-400 border border-blue-500/30">
              {opportunity.category}
            </span>
          </h2>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <button
            onClick={handleGenerate}
            disabled={loading}
            className="px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white shadow-sm transition-all flex items-center gap-1.5"
          >
            <span>✨</span>
            <span>{opportunity.generated_prd_markdown ? 'Re-Synthesize PRD' : 'Synthesize Autonomous PRD'}</span>
          </button>

          {opportunity.generated_prd_markdown && (
            <>
              <button
                onClick={handleCopy}
                className="px-3 py-1.5 rounded-lg text-xs font-medium bg-[#1a233a] hover:bg-[#232f4e] text-slate-200 border border-[#2b3a5d] transition-colors"
              >
                {copied ? '✓ Copied' : 'Copy MD'}
              </button>

              <button
                onClick={handleDownload}
                className="px-3 py-1.5 rounded-lg text-xs font-medium bg-[#1a233a] hover:bg-[#232f4e] text-slate-200 border border-[#2b3a5d] transition-colors flex items-center gap-1"
              >
                <span>💾</span> .md
              </button>

              <button
                onClick={handleExportNotion}
                disabled={loading}
                className="px-3 py-1.5 rounded-lg text-xs font-medium bg-[#1e293b] hover:bg-[#283548] text-slate-200 border border-[#334155] transition-colors flex items-center gap-1"
              >
                <span>🚀</span> Notion
              </button>

              <button
                onClick={handleOpenGitHub}
                className="px-3 py-1.5 rounded-lg text-xs font-medium bg-[#1e293b] hover:bg-[#283548] text-slate-200 border border-[#334155] transition-colors flex items-center gap-1"
              >
                <span>🐙</span> GitHub Issue
              </button>

              <button
                onClick={handleOpenJira}
                className="px-3 py-1.5 rounded-lg text-xs font-medium bg-[#1e293b] hover:bg-[#283548] text-slate-200 border border-[#334155] transition-colors flex items-center gap-1"
              >
                <span>🔷</span> Jira Epic
              </button>
            </>
          )}
        </div>
      </div>

      {notionStatus && (
        <div className="bg-emerald-950/40 border-b border-emerald-800/40 px-6 py-2.5 text-xs text-emerald-300 flex items-center justify-between">
          <span>✓ Successfully exported PRD to Notion workspace.</span>
          <a href={notionStatus} target="_blank" rel="noreferrer" className="underline font-semibold hover:text-emerald-200">
            Open in Notion →
          </a>
        </div>
      )}

      {/* PRD Content Body */}
      <div className="p-6">
        {loading ? (
          <div className="py-16 text-center text-slate-400 space-y-3">
            <div className="inline-block w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin" />
            <div className="text-sm">Synthesizing comprehensive Product Requirements Document...</div>
          </div>
        ) : opportunity.generated_prd_markdown ? (
          <div className="prose prose-invert max-w-none text-slate-300 text-sm leading-relaxed space-y-6">
            <pre className="whitespace-pre-wrap font-sans bg-[#0c101a] p-6 rounded-lg border border-[#1e293b] text-slate-200 overflow-x-auto text-xs leading-relaxed">
              {opportunity.generated_prd_markdown}
            </pre>
          </div>
        ) : (
          <div className="py-16 text-center text-slate-400 space-y-4">
            <div className="text-4xl">📄</div>
            <div className="text-base font-semibold text-white">No PRD Generated Yet</div>
            <p className="text-xs text-slate-400 max-w-md mx-auto">
              Click &quot;Synthesize Autonomous PRD&quot; above to compile problem context, verified quotes from {quotes.length} customer accounts, user stories, and Gherkin acceptance criteria.
            </p>
            <button
              onClick={handleGenerate}
              className="px-4 py-2 rounded-lg text-xs font-semibold bg-blue-600 hover:bg-blue-500 text-white shadow-sm"
            >
              Synthesize Autonomous PRD Now
            </button>
          </div>
        )}
      </div>

      {/* GitHub Issue Modal */}
      {activeExportModal === 'github' && githubPayload && (
        <div className="fixed inset-0 z-50 bg-black/75 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-[#121826] border border-[#1e293b] rounded-2xl max-w-3xl w-full max-h-[85vh] flex flex-col overflow-hidden shadow-2xl">
            <div className="px-6 py-4 bg-[#0e1320] border-b border-[#1e293b] flex items-center justify-between">
              <div className="flex items-center gap-2">
                <span className="text-base font-bold text-white">🐙 GitHub Issue Export Payload</span>
              </div>
              <button onClick={() => setActiveExportModal(null)} className="text-slate-400 hover:text-white font-bold">✕</button>
            </div>
            <div className="p-6 overflow-y-auto space-y-4 flex-1 text-xs">
              <div>
                <label className="text-[#64748b] block font-semibold mb-1">Issue Title</label>
                <input
                  readOnly
                  value={githubPayload.title}
                  className="w-full bg-[#0c101a] border border-[#1e293b] rounded p-2 text-white font-mono"
                />
              </div>
              <div>
                <label className="text-[#64748b] block font-semibold mb-1">Target Labels</label>
                <div className="flex flex-wrap gap-1.5">
                  {githubPayload.labels.map((l) => (
                    <span key={l} className="px-2 py-0.5 rounded bg-blue-500/10 text-blue-400 border border-blue-500/30 font-mono">
                      {l}
                    </span>
                  ))}
                </div>
              </div>
              <div>
                <label className="text-[#64748b] block font-semibold mb-1">Issue Markdown Body</label>
                <textarea
                  readOnly
                  rows={14}
                  value={githubPayload.body}
                  className="w-full bg-[#0c101a] border border-[#1e293b] rounded p-3 text-slate-200 font-mono whitespace-pre text-[11px]"
                />
              </div>
            </div>
            <div className="px-6 py-3.5 bg-[#0e1320] border-t border-[#1e293b] flex justify-end gap-2 text-xs">
              <button
                onClick={() => {
                  navigator.clipboard.writeText(githubPayload.body)
                  alert('Copied GitHub Issue body to clipboard!')
                }}
                className="px-4 py-1.5 rounded bg-blue-600 hover:bg-blue-500 text-white font-semibold"
              >
                Copy Markdown Body
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Jira Epic Modal */}
      {activeExportModal === 'jira' && jiraPayload && (
        <div className="fixed inset-0 z-50 bg-black/75 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-[#121826] border border-[#1e293b] rounded-2xl max-w-3xl w-full max-h-[85vh] flex flex-col overflow-hidden shadow-2xl">
            <div className="px-6 py-4 bg-[#0e1320] border-b border-[#1e293b] flex items-center justify-between">
              <div className="flex items-center gap-2">
                <span className="text-base font-bold text-white">🔷 Jira Epic Export Payload</span>
              </div>
              <button onClick={() => setActiveExportModal(null)} className="text-slate-400 hover:text-white font-bold">✕</button>
            </div>
            <div className="p-6 overflow-y-auto space-y-4 flex-1 text-xs">
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-[#64748b] block font-semibold mb-1">Epic Summary</label>
                  <input
                    readOnly
                    value={jiraPayload.summary}
                    className="w-full bg-[#0c101a] border border-[#1e293b] rounded p-2 text-white font-mono"
                  />
                </div>
                <div>
                  <label className="text-[#64748b] block font-semibold mb-1">Calculated Priority</label>
                  <input
                    readOnly
                    value={jiraPayload.priority}
                    className="w-full bg-[#0c101a] border border-[#1e293b] rounded p-2 text-emerald-400 font-mono font-bold"
                  />
                </div>
              </div>
              <div>
                <label className="text-[#64748b] block font-semibold mb-1">Jira Markup Description</label>
                <textarea
                  readOnly
                  rows={14}
                  value={jiraPayload.description}
                  className="w-full bg-[#0c101a] border border-[#1e293b] rounded p-3 text-slate-200 font-mono whitespace-pre text-[11px]"
                />
              </div>
            </div>
            <div className="px-6 py-3.5 bg-[#0e1320] border-t border-[#1e293b] flex justify-end gap-2 text-xs">
              <button
                onClick={() => {
                  navigator.clipboard.writeText(jiraPayload.description)
                  alert('Copied Jira Epic markup to clipboard!')
                }}
                className="px-4 py-1.5 rounded bg-blue-600 hover:bg-blue-500 text-white font-semibold"
              >
                Copy Jira Markup
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

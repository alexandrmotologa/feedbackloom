import React, { useState } from 'react'
import { testLiveIngest } from '../api'
import { LiveIngestRequest, LiveIngestResponse } from '../types'

interface SandboxPageProps {
  onSelectPRD: (opportunityId: string) => void
}

interface PresetTemplate {
  name: string
  label: string
  payload: LiveIngestRequest
}

const PRESET_TEMPLATES: PresetTemplate[] = [
  {
    name: 'cyberdyne-sso',
    label: 'Cyberdyne: Okta SSO & SCIM',
    payload: {
      source: 'zendesk',
      customer_email: 'sarah.connor@cyberdyne.com',
      customer_name: 'Sarah Connor',
      raw_text:
        'Our enterprise compliance mandate requires Okta SAML 2.0 single sign-on with SCIM automated user provisioning. Without this before Q3 renewal, our CISO will block contract extension and migrate to Auth0.',
      commit: false,
    },
  },
  {
    name: 'wayne-siem',
    label: 'Wayne Sec: Splunk Audit Stream',
    payload: {
      source: 'slack',
      customer_email: 'lucius.fox@wayneenterprises.com',
      customer_name: 'Lucius Fox',
      raw_text:
        'We need real-time SIEM audit log streaming to Splunk or Datadog. Our security operations center cannot monitor workspace activity without centralized syslog export.',
      commit: false,
    },
  },
  {
    name: 'pied-piper-export',
    label: 'Pied Piper: 50k Async Export',
    payload: {
      source: 'intercom',
      customer_email: 'richard@piedpiper.com',
      customer_name: 'Richard Hendricks',
      raw_text:
        'Bulk CSV export crashes with 504 Gateway Timeout when exporting datasets exceeding 50,000 records. We need async background export with signed S3 download links.',
      commit: false,
    },
  },
  {
    name: 'umbrella-rbac',
    label: 'Umbrella: Granular Custom RBAC',
    payload: {
      source: 'gong',
      customer_email: 'william.birkin@umbrella-health.com',
      customer_name: 'William Birkin',
      raw_text:
        'Standard admin and member permissions are insufficient for clinical research compliance. We need granular role-based access control with custom permission policies.',
      commit: false,
    },
  },
]

export const SandboxPage: React.FC<SandboxPageProps> = ({ onSelectPRD }) => {
  const [source, setSource] = useState<string>('zendesk')
  const [customerEmail, setCustomerEmail] = useState<string>('sarah.connor@cyberdyne.com')
  const [customerName, setCustomerName] = useState<string>('Sarah Connor')
  const [rawText, setRawText] = useState<string>(
    'Our enterprise compliance mandate requires Okta SAML 2.0 single sign-on with SCIM automated user provisioning. Without this before Q3 renewal, our CISO will block contract extension and migrate to Auth0.'
  )
  const [commit, setCommit] = useState<boolean>(false)

  const [loading, setLoading] = useState<boolean>(false)
  const [error, setError] = useState<string | null>(null)
  const [result, setResult] = useState<LiveIngestResponse | null>(null)

  const handleApplyTemplate = (preset: PresetTemplate) => {
    setSource(preset.payload.source)
    setCustomerEmail(preset.payload.customer_email)
    setCustomerName(preset.payload.customer_name)
    setRawText(preset.payload.raw_text)
    setCommit(preset.payload.commit || false)
    setResult(null)
    setError(null)
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!rawText.trim() || !customerEmail.trim()) {
      setError('Please provide customer email and feedback text.')
      return
    }

    setLoading(true)
    setError(null)

    try {
      const resp = await testLiveIngest({
        source,
        customer_email: customerEmail.trim(),
        customer_name: customerName.trim(),
        raw_text: rawText.trim(),
        commit,
      })
      setResult(resp)
    } catch (err: any) {
      setError(err.message || 'Failed to execute live ingestion test.')
    } finally {
      setLoading(false)
    }
  }

  const formatUSD = (val: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      maximumFractionDigits: 0,
    }).format(val)
  }

  const getUrgencyBadge = (score: number) => {
    if (score >= 0.8) {
      return {
        label: 'CRITICAL',
        classes: 'bg-rose-500/15 text-rose-300 border-rose-500/40',
      }
    }
    if (score >= 0.5) {
      return {
        label: 'HIGH',
        classes: 'bg-amber-500/15 text-amber-300 border-amber-500/40',
      }
    }
    if (score >= 0.3) {
      return {
        label: 'MEDIUM',
        classes: 'bg-blue-500/15 text-blue-300 border-blue-500/40',
      }
    }
    return {
      label: 'LOW',
      classes: 'bg-slate-800 text-slate-400 border-slate-700',
    }
  }

  return (
    <div className="space-y-6 pb-12">
      {/* Header Banner */}
      <div className="bg-[#0f1422]/90 border border-slate-800/80 rounded-2xl p-6 sm:p-7 shadow-lg backdrop-blur-md">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h1 className="text-lg font-bold text-white flex items-center gap-2.5">
              <span className="w-2.5 h-2.5 rounded-full bg-cyan-400 animate-pulse" />
              Interactive Ingestion Sandbox &amp; Live Payload Tester
            </h1>
            <p className="text-xs text-slate-400 mt-1 max-w-3xl leading-relaxed">
              Test incoming raw payloads against the feedbackloom NLP pipeline in real time. Inspect CRM domain attribution, ARR calculation, competitor extraction, subword vectorization, and instant cosine similarity matching against existing opportunity clusters.
            </p>
          </div>
          <div className="flex items-center gap-2">
            <span className="px-3 py-1 rounded-full text-xs font-mono font-semibold bg-cyan-500/10 text-cyan-300 border border-cyan-800/40">
              POST /api/v1/ingest/test-live
            </span>
          </div>
        </div>

        {/* Quick Presets */}
        <div className="mt-5 pt-4 border-t border-slate-800/60 flex flex-wrap items-center gap-2.5">
          <span className="text-[11px] uppercase font-bold tracking-wider text-slate-400 mr-1">
            Quick Presets:
          </span>
          {PRESET_TEMPLATES.map((tpl) => (
            <button
              key={tpl.name}
              type="button"
              onClick={() => handleApplyTemplate(tpl)}
              className="px-3 py-1.5 rounded-xl text-xs font-medium bg-slate-900/80 hover:bg-slate-800 text-slate-300 hover:text-white border border-slate-700/80 transition-all shadow-sm"
            >
              {tpl.label}
            </button>
          ))}
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Input Form Column */}
        <div className="lg:col-span-5 bg-[#0f1422]/90 border border-slate-800/80 rounded-2xl p-6 sm:p-7 shadow-lg space-y-5">
          <h2 className="text-sm font-bold uppercase tracking-wider text-white border-b border-slate-800/60 pb-3">
            Payload Configuration
          </h2>

          <form onSubmit={handleSubmit} className="space-y-4 text-xs">
            <div className="grid grid-cols-2 gap-3.5">
              <div>
                <label className="block text-[11px] font-semibold text-slate-400 mb-1.5">
                  Feedback Channel Source
                </label>
                <select
                  value={source}
                  onChange={(e) => setSource(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 focus:border-cyan-500 rounded-xl px-3 py-2.5 text-white outline-none"
                >
                  <option value="zendesk">Zendesk Support</option>
                  <option value="intercom">Intercom Chat</option>
                  <option value="gong">Gong Sales Call</option>
                  <option value="slack">Slack Connect</option>
                </select>
              </div>

              <div>
                <label className="block text-[11px] font-semibold text-slate-400 mb-1.5">
                  Customer Name
                </label>
                <input
                  type="text"
                  value={customerName}
                  onChange={(e) => setCustomerName(e.target.value)}
                  placeholder="e.g. Sarah Connor"
                  className="w-full bg-slate-950 border border-slate-800 focus:border-cyan-500 rounded-xl px-3 py-2.5 text-white outline-none"
                />
              </div>
            </div>

            <div>
              <label className="block text-[11px] font-semibold text-slate-400 mb-1.5">
                Customer Email (Domain CRM lookup)
              </label>
              <input
                type="email"
                value={customerEmail}
                onChange={(e) => setCustomerEmail(e.target.value)}
                placeholder="e.g. sarah@cyberdyne.com"
                required
                className="w-full bg-slate-950 border border-slate-800 focus:border-cyan-500 rounded-xl px-3 py-2.5 text-white outline-none"
              />
              <span className="text-[10px] text-slate-500 mt-1 block">
                Domain matches against Stripe and HubSpot customer records for ARR enrichment.
              </span>
            </div>

            <div>
              <label className="block text-[11px] font-semibold text-slate-400 mb-1.5">
                Verbatim Customer Feedback / Transcript
              </label>
              <textarea
                rows={5}
                value={rawText}
                onChange={(e) => setRawText(e.target.value)}
                placeholder="Enter raw customer verbatim quote or call transcript segment..."
                required
                className="w-full bg-slate-950 border border-slate-800 focus:border-cyan-500 rounded-xl p-3 text-white outline-none font-mono text-[11px] leading-relaxed resize-y"
              />
            </div>

            <div className="flex items-center gap-2.5 pt-1">
              <input
                type="checkbox"
                id="commitFeedback"
                checked={commit}
                onChange={(e) => setCommit(e.target.checked)}
                className="rounded border-slate-700 bg-slate-950 text-cyan-600 focus:ring-0 focus:ring-offset-0 cursor-pointer w-4 h-4"
              />
              <label htmlFor="commitFeedback" className="text-slate-300 text-xs cursor-pointer select-none">
                Commit &amp; persist to SQLite store (triggers cluster re-index)
              </label>
            </div>

            {error && (
              <div className="p-3.5 rounded-xl bg-rose-950/40 border border-rose-800 text-rose-300 text-xs">
                {error}
              </div>
            )}

            <button
              type="submit"
              disabled={loading}
              className="w-full py-3 px-4 rounded-xl font-bold text-xs text-white bg-gradient-to-r from-cyan-600 to-blue-600 hover:from-cyan-500 hover:to-blue-500 disabled:opacity-50 transition-all flex items-center justify-center gap-2 shadow-lg shadow-cyan-900/20"
            >
              {loading ? (
                <>
                  <span className="w-3.5 h-3.5 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  <span>Processing NLP Pipeline...</span>
                </>
              ) : (
                <>
                  <span>⚡</span>
                  <span>Analyze &amp; Test Ingest</span>
                </>
              )}
            </button>
          </form>
        </div>

        {/* Results Analysis Column */}
        <div className="lg:col-span-7 space-y-4">
          {!result && !loading && (
            <div className="bg-[#0f1422]/90 border border-slate-800/80 rounded-2xl p-14 text-center text-slate-400 space-y-3 shadow-lg">
              <div className="text-4xl">🧪</div>
              <div className="text-base font-semibold text-white">Sandbox Awaiting Ingestion Payload</div>
              <p className="text-xs text-slate-400 max-w-md mx-auto leading-relaxed">
                Select one of the preset templates above or craft a custom customer quote. Hit "Analyze &amp; Test Ingest" to inspect the extracted entities, financial attribution, and real-time cluster matching.
              </p>
            </div>
          )}

          {loading && (
            <div className="bg-[#0f1422]/90 border border-slate-800/80 rounded-2xl p-14 text-center text-slate-400 space-y-3 shadow-lg">
              <div className="inline-block w-8 h-8 border-2 border-cyan-400 border-t-transparent rounded-full animate-spin mb-2" />
              <div className="text-base font-semibold text-white">Running Pipeline Stages...</div>
              <p className="text-xs text-slate-400">
                Resolving CRM domain, computing subword n-gram vector, and executing nearest cosine cluster match.
              </p>
            </div>
          )}

          {result && !loading && (
            <div className="space-y-4">
              {/* Account Enrichment Card */}
              <div className="bg-[#0f1422]/90 border border-slate-800/80 rounded-2xl p-5 sm:p-6 shadow-md space-y-3">
                <div className="flex items-center justify-between border-b border-slate-800/60 pb-3">
                  <span className="text-xs font-bold uppercase tracking-wider text-white flex items-center gap-2">
                    <span className="text-emerald-400 font-mono">01.</span> Account &amp; Financial Attribution
                  </span>
                  <span className="px-2.5 py-0.5 rounded-full text-[10px] font-mono font-bold bg-emerald-500/15 text-emerald-400 border border-emerald-500/30">
                    {result.account ? 'RESOLVED' : 'UNMATCHED DOMAIN'}
                  </span>
                </div>

                <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs pt-1">
                  <div className="bg-slate-950 p-3 rounded-xl border border-slate-800/80">
                    <div className="text-[10px] text-slate-400 uppercase font-semibold">Account</div>
                    <div className="font-semibold text-white truncate mt-1">
                      {result.account ? result.account.name : 'Unknown Account'}
                    </div>
                  </div>
                  <div className="bg-slate-950 p-3 rounded-xl border border-slate-800/80">
                    <div className="text-[10px] text-slate-400 uppercase font-semibold">Contract ARR</div>
                    <div className="font-semibold font-mono text-emerald-400 mt-1">
                      {formatUSD(result.account ? result.account.arr_usd : 0)}
                    </div>
                  </div>
                  <div className="bg-slate-950 p-3 rounded-xl border border-slate-800/80">
                    <div className="text-[10px] text-slate-400 uppercase font-semibold">Plan Tier</div>
                    <div className="font-semibold text-cyan-300 mt-1">
                      {result.account ? result.account.plan_tier : 'STANDARD'}
                    </div>
                  </div>
                  <div className="bg-slate-950 p-3 rounded-xl border border-slate-800/80">
                    <div className="text-[10px] text-slate-400 uppercase font-semibold">Renewal Date</div>
                    <div className="font-semibold font-mono text-amber-300 mt-1 truncate">
                      {result.account ? result.account.renewal_date : 'N/A'}
                    </div>
                  </div>
                </div>
              </div>

              {/* Extraction Card */}
              <div className="bg-[#0f1422]/90 border border-slate-800/80 rounded-2xl p-5 sm:p-6 shadow-md space-y-3">
                <div className="flex items-center justify-between border-b border-slate-800/60 pb-3">
                  <span className="text-xs font-bold uppercase tracking-wider text-white flex items-center gap-2">
                    <span className="text-blue-400 font-mono">02.</span> NLP Semantic Extractions
                  </span>
                  <span className="px-2.5 py-0.5 rounded-full text-[10px] font-mono font-bold bg-blue-500/15 text-blue-300 border border-blue-500/30">
                    PARSED
                  </span>
                </div>

                <div className="space-y-3 text-xs pt-1">
                  <div>
                    <span className="text-[11px] text-slate-400 block mb-1">Identified Pain Point:</span>
                    <div className="bg-slate-950 p-3 rounded-xl border border-slate-800/80 text-slate-200 leading-relaxed">
                      {result.extracted_pain_point}
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div>
                      <span className="text-[11px] text-slate-400 block mb-1">Urgency Level:</span>
                      {(() => {
                        const badge = getUrgencyBadge(result.urgency_score)
                        return (
                          <span
                            className={`inline-block px-3 py-1 rounded-lg text-xs font-semibold border ${badge.classes}`}
                          >
                            {badge.label} ({(result.urgency_score * 100).toFixed(0)}%)
                          </span>
                        )
                      })()}
                    </div>

                    <div>
                      <span className="text-[11px] text-slate-400 block mb-1">Competitors Detected:</span>
                      {result.competitor_mentions && result.competitor_mentions.length > 0 ? (
                        <div className="flex flex-wrap gap-1.5">
                          {result.competitor_mentions.map((comp: string) => (
                            <span
                              key={comp}
                              className="px-2.5 py-1 rounded-md text-xs font-semibold bg-rose-950/50 text-rose-300 border border-rose-800/60"
                            >
                              ⚔ {comp}
                            </span>
                          ))}
                        </div>
                      ) : (
                        <span className="text-slate-500 italic">None detected</span>
                      )}
                    </div>
                  </div>

                  {result.extracted_outcome && (
                    <div>
                      <span className="text-[11px] text-slate-400 block mb-1">Desired Customer Outcome:</span>
                      <div className="bg-slate-950 p-3 rounded-xl border border-slate-800/80 text-slate-300 leading-relaxed">
                        {result.extracted_outcome}
                      </div>
                    </div>
                  )}
                </div>
              </div>

              {/* Matched Opportunity & Cosine Similarity */}
              <div className="bg-[#0f1422]/90 border border-slate-800/80 rounded-2xl p-5 sm:p-6 shadow-md space-y-4">
                <div className="flex items-center justify-between border-b border-slate-800/60 pb-3">
                  <span className="text-xs font-bold uppercase tracking-wider text-white flex items-center gap-2">
                    <span className="text-cyan-400 font-mono">03.</span> HDBSCAN Semantic Cluster Matching
                  </span>
                  <span className="px-2.5 py-0.5 rounded-full text-[10px] font-mono font-bold bg-cyan-500/15 text-cyan-300 border border-cyan-500/30">
                    COSINE MATCH
                  </span>
                </div>

                {result.matched_opportunity_id ? (
                  <div className="space-y-4 text-xs pt-1">
                    <div className="bg-slate-950 p-4 rounded-xl border border-slate-800/80 space-y-3">
                      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                        <span className="font-bold text-white text-sm">
                          {result.matched_opportunity_title}
                        </span>
                        <button
                          onClick={() => onSelectPRD(result.matched_opportunity_id!)}
                          className="px-3.5 py-1.5 rounded-lg bg-blue-600 hover:bg-blue-500 text-white font-semibold text-xs transition-colors self-start sm:self-auto"
                        >
                          Open in PRD Studio &rarr;
                        </button>
                      </div>

                      {/* Similarity Bar */}
                      <div>
                        <div className="flex items-center justify-between text-xs mb-1.5">
                          <span className="text-slate-400">Vector Cosine Similarity:</span>
                          <span className="font-mono font-bold text-cyan-400 text-sm">
                            {(result.similarity_score * 100).toFixed(1)}%
                          </span>
                        </div>
                        <div className="w-full bg-slate-900 rounded-full h-3 p-0.5 border border-slate-800 overflow-hidden">
                          <div
                            className="bg-gradient-to-r from-blue-500 via-indigo-500 to-cyan-400 h-full rounded-full transition-all duration-700 shadow-sm"
                            style={{
                              width: `${Math.min(100, Math.max(0, result.similarity_score * 100))}%`,
                            }}
                          />
                        </div>
                      </div>
                    </div>

                    <div className="text-xs text-slate-400 bg-slate-950/60 p-3 rounded-xl border border-slate-800/70">
                      Cluster Status:{' '}
                      <span className="text-slate-200 font-medium">
                        {result.created_item_id
                          ? `Committed to SQLite store (ID: ${result.created_item_id}). Clustered with existing feedback evidence.`
                          : 'Dry run execution. Vector matched against existing cluster centroids without modifying database state.'}
                      </span>
                    </div>
                  </div>
                ) : (
                  <div className="p-6 bg-slate-950 rounded-xl border border-slate-800/80 text-center text-slate-400 text-xs">
                    No existing cluster exceeded similarity threshold. This payload represents a new distinct feature signal.
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
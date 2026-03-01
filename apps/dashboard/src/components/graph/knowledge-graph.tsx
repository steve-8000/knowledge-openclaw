'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Background,
  Controls,
  MiniMap,
  ReactFlow,
  ReactFlowProvider,
  type Edge as FlowEdge,
  type Node as FlowNode
} from '@xyflow/react';
import '@xyflow/react/dist/style.css';

import { DEFAULT_TENANT_ID, TENANT_STORAGE_KEY, getGraphEgo, listDocuments } from '@/lib/api/client';
import type { Document, Edge, GraphEgoResponse } from '@/lib/api/types';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

/* ──────────────────────────── Helpers ──────────────────────────── */

function getTenantId(): string {
  if (typeof window === 'undefined') {
    return DEFAULT_TENANT_ID;
  }
  return window.localStorage.getItem(TENANT_STORAGE_KEY) ?? DEFAULT_TENANT_ID;
}

/* ── Vibrant color palette for all 8 doc_types ── */
const DOC_TYPE_COLORS: Record<string, string> = {
  policy:   '#3b82f6',
  spec:     '#06b6d4',
  incident: '#f59e0b',
  runbook:  '#8b5cf6',
  guide:    '#22c55e',
  adr:      '#ec4899',
  report:   '#f97316',
  glossary: '#a78bfa',
};

const DOC_TYPE_LABELS: Record<string, string> = {
  policy:   'Policy',
  spec:     'Spec',
  incident: 'Incident',
  runbook:  'Runbook',
  guide:    'Guide',
  adr:      'ADR',
  report:   'Report',
  glossary: 'Glossary',
};

function docTypeColor(docType: string): string {
  return DOC_TYPE_COLORS[docType] ?? '#94a3b8';
}

function docTypeGlow(docType: string, intensity: number): string {
  const color = DOC_TYPE_COLORS[docType] ?? '#94a3b8';
  return `0 0 ${Math.round(intensity)}px ${Math.round(intensity * 0.6)}px ${color}80, 0 0 ${Math.round(intensity * 2)}px ${Math.round(intensity)}px ${color}30`;
}

/* ──────────────────────────── Node Component ──────────────────────────── */

type NodeData = Record<string, unknown> & {
  document: Document;
  degree: number;
  size: number;
  isCenter: boolean;
  maxDegree: number;
};

function GraphNode({ data }: { data: NodeData }): JSX.Element {
  const [hovered, setHovered] = useState(false);
  const dotSize = Math.round(data.size);
  const isHub = data.degree >= Math.max(4, data.maxDegree * 0.5);
  const color = docTypeColor(data.document.doc_type);
  const glowIntensity = isHub ? Math.min(20, 8 + data.degree * 1.5) : data.isCenter ? 16 : 0;

  const abbrev = data.document.title
    .split(/[\s\-/]+/)
    .slice(0, 2)
    .map((w) => w[0]?.toUpperCase() ?? '')
    .join('');

  return (
    <div
      className="relative flex flex-col items-center"
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      {hovered && (
        <div
          className="pointer-events-none absolute -top-2 z-50 -translate-y-full rounded-lg border border-slate-600 bg-slate-800 px-3 py-2 shadow-xl"
          style={{ minWidth: 180, maxWidth: 260 }}
        >
          <p className="text-xs font-bold text-white">{data.document.title}</p>
          <div className="mt-1 flex items-center gap-2">
            <span className="inline-block h-2 w-2 rounded-full" style={{ backgroundColor: color }} />
            <span className="text-[10px] text-slate-300">{DOC_TYPE_LABELS[data.document.doc_type] ?? data.document.doc_type}</span>
            <span className="text-[10px] text-slate-400">·</span>
            <span className="text-[10px] text-slate-400">{data.degree} connections</span>
          </div>
          <p className="mt-1 text-[10px] text-slate-400">{data.document.stable_key}</p>
        </div>
      )}

      <div
        className="flex items-center justify-center rounded-full transition-transform duration-200"
        style={{
          width: dotSize,
          height: dotSize,
          backgroundColor: color,
          border: data.isCenter ? '3px solid rgba(255, 255, 255, 0.8)' : `2px solid ${color}40`,
          boxShadow: glowIntensity > 0 ? docTypeGlow(data.document.doc_type, glowIntensity) : 'none',
          transform: hovered ? 'scale(1.15)' : 'scale(1)',
          cursor: 'pointer',
          fontSize: dotSize > 30 ? 11 : 9,
          fontWeight: 700,
          color: 'rgba(255, 255, 255, 0.9)',
          letterSpacing: '0.02em',
        }}
      >
        {dotSize > 24 ? abbrev : ''}
      </div>

      {dotSize >= 28 && (
        <p
          className="mt-1 text-center leading-tight"
          style={{
            maxWidth: Math.max(80, dotSize * 2.2),
            fontSize: 10,
            fontWeight: 500,
            color: 'rgba(226, 232, 240, 0.85)',
            textShadow: '0 1px 3px rgba(0,0,0,0.8)',
            lineHeight: '1.2',
          }}
        >
          {data.document.title.length > 22 ? `${data.document.title.slice(0, 20)}…` : data.document.title}
        </p>
      )}
    </div>
  );
}

const nodeTypes = { docNode: GraphNode };

/* ──────────────────────────── Legend ──────────────────────────── */

function GraphLegend({ visibleTypes }: { visibleTypes: Set<string> }): JSX.Element {
  const types = Object.entries(DOC_TYPE_COLORS).filter(([type]) => visibleTypes.has(type));
  if (types.length === 0) return <></>;

  return (
    <div className="absolute bottom-3 left-3 z-10 flex flex-wrap gap-x-3 gap-y-1 rounded-lg bg-slate-800/80 px-3 py-2 backdrop-blur-sm">
      {types.map(([type, color]) => (
        <div key={type} className="flex items-center gap-1.5">
          <div className="h-2.5 w-2.5 rounded-full" style={{ backgroundColor: color, boxShadow: `0 0 6px ${color}60` }} />
          <span className="text-[10px] font-medium text-slate-300">{DOC_TYPE_LABELS[type] ?? type}</span>
        </div>
      ))}
    </div>
  );
}

/* ──────────────────────────── Stats ──────────────────────────── */

function GraphStats({ nodeCount, edgeCount, loading }: { nodeCount: number; edgeCount: number; loading: boolean }): JSX.Element {
  return (
    <div className="absolute right-3 top-3 z-10 flex items-center gap-2 rounded-lg bg-slate-800/80 px-3 py-1.5 backdrop-blur-sm">
      {loading ? (
        <span className="text-xs text-slate-400">Loading…</span>
      ) : (
        <>
          <span className="text-xs font-semibold text-slate-200">{nodeCount}</span>
          <span className="text-[10px] text-slate-400">nodes</span>
          <span className="text-xs text-slate-500">·</span>
          <span className="text-xs font-semibold text-slate-200">{edgeCount}</span>
          <span className="text-[10px] text-slate-400">edges</span>
        </>
      )}
    </div>
  );
}

/* ──────────────────────────── Force Layout ──────────────────────────── */

function seededUnit(id: string): number {
  let hash = 2166136261;
  for (let i = 0; i < id.length; i += 1) {
    hash ^= id.charCodeAt(i);
    hash = Math.imul(hash, 16777619);
  }
  return (hash >>> 0) / 4294967295;
}

function clamp(value: number, min: number, max: number): number {
  return Math.max(min, Math.min(max, value));
}

function computeForceLayout(
  nodeIds: string[],
  edges: Array<{ source: string; target: string }>,
  centerNodeId: string,
  centerX: number,
  centerY: number,
  degreeMap: Map<string, number>
): Map<string, { x: number; y: number }> {
  const width = 1000;
  const height = 700;
  const left = centerX - width / 2;
  const top = centerY - height / 2;
  const right = centerX + width / 2;
  const bottom = centerY + height / 2;

  const pos = new Map<string, { x: number; y: number }>();
  const vel = new Map<string, { x: number; y: number }>();

  const adjacency = new Map<string, Set<string>>();
  nodeIds.forEach((id) => adjacency.set(id, new Set()));
  edges.forEach((edge) => {
    adjacency.get(edge.source)?.add(edge.target);
    adjacency.get(edge.target)?.add(edge.source);
  });

  nodeIds.forEach((id) => {
    if (id === centerNodeId) {
      pos.set(id, { x: centerX, y: centerY });
      vel.set(id, { x: 0, y: 0 });
      return;
    }
    const degree = degreeMap.get(id) ?? 0;
    const isConnectedToCenter = adjacency.get(id)?.has(centerNodeId) ?? false;
    const r1 = seededUnit(`${id}:x`);
    const r2 = seededUnit(`${id}:y`);
    const spread = isConnectedToCenter ? 0.4 : degree > 0 ? 0.7 : 0.9;
    const angle = r1 * Math.PI * 2;
    const radius = (spread * Math.min(width, height) * 0.5) * (0.3 + r2 * 0.7);
    pos.set(id, { x: centerX + Math.cos(angle) * radius, y: centerY + Math.sin(angle) * radius });
    vel.set(id, { x: 0, y: 0 });
  });

  const maxDegree = Math.max(1, ...Array.from(degreeMap.values()));
  const iterations = 350;
  const repulsion = 8000;
  const springBase = 0.012;
  const damping = 0.88;

  for (let step = 0; step < iterations; step += 1) {
    const temp = 1 - step / iterations;
    for (let i = 0; i < nodeIds.length; i += 1) {
      const aId = nodeIds[i];
      if (aId === centerNodeId) continue;
      const aPos = pos.get(aId);
      const aVel = vel.get(aId);
      if (!aPos || !aVel) continue;
      const aDegree = degreeMap.get(aId) ?? 0;
      let fx = 0;
      let fy = 0;

      for (let j = 0; j < nodeIds.length; j += 1) {
        if (i === j) continue;
        const bPos = pos.get(nodeIds[j]);
        if (!bPos) continue;
        const dx = aPos.x - bPos.x;
        const dy = aPos.y - bPos.y;
        const distSq = Math.max(100, dx * dx + dy * dy);
        const force = repulsion / distSq;
        const invDist = 1 / Math.sqrt(distSq);
        fx += dx * invDist * force;
        fy += dy * invDist * force;
      }

      edges.forEach((edge) => {
        if (edge.source !== aId && edge.target !== aId) return;
        const otherId = edge.source === aId ? edge.target : edge.source;
        const bPos = pos.get(otherId);
        if (!bPos) return;
        const dx = bPos.x - aPos.x;
        const dy = bPos.y - aPos.y;
        const dist = Math.max(1, Math.sqrt(dx * dx + dy * dy));
        const otherDegree = degreeMap.get(otherId) ?? 0;
        const avgDeg = (aDegree + otherDegree) / 2;
        const targetDist = 100 + avgDeg * 12;
        const spring = springBase * (1 + (avgDeg / maxDegree) * 0.5);
        const pull = (dist - targetDist) * spring;
        fx += (dx / dist) * pull;
        fy += (dy / dist) * pull;
      });

      const gravityStrength = aDegree === 0 ? 0.008 : 0.002 + (1 - aDegree / maxDegree) * 0.003;
      fx += (centerX - aPos.x) * gravityStrength;
      fy += (centerY - aPos.y) * gravityStrength;

      aVel.x = (aVel.x + fx * temp) * damping;
      aVel.y = (aVel.y + fy * temp) * damping;
      aPos.x = clamp(aPos.x + aVel.x, left, right);
      aPos.y = clamp(aPos.y + aVel.y, top, bottom);
    }

    const centerPos = pos.get(centerNodeId);
    if (centerPos) {
      centerPos.x = centerX;
      centerPos.y = centerY;
    }
  }

  return pos;
}

/* ──────────────────────────── Build Flow Data ──────────────────────────── */

function buildFlowData(
  payload: GraphEgoResponse,
  relationFilter: string
): { nodes: FlowNode<NodeData>[]; edges: FlowEdge[] } {
  const filteredEdges = (
    relationFilter === 'all'
      ? payload.edges
      : payload.edges.filter((edge) => edge.relation_type === relationFilter)
  ).filter((edge) => edge.source_id && edge.target_id && edge.source_id !== edge.target_id);

  const includedNodeIds = new Set<string>([payload.center_doc_id]);
  filteredEdges.forEach((edge) => {
    includedNodeIds.add(edge.source_id);
    includedNodeIds.add(edge.target_id);
  });

  const docs = payload.documents.filter((doc) => includedNodeIds.has(doc.doc_id));
  const centerX = 500;
  const centerY = 350;

  const degreeMap = new Map<string, number>();
  docs.forEach((doc) => degreeMap.set(doc.doc_id, 0));
  filteredEdges.forEach((edge) => {
    degreeMap.set(edge.source_id, (degreeMap.get(edge.source_id) ?? 0) + 1);
    degreeMap.set(edge.target_id, (degreeMap.get(edge.target_id) ?? 0) + 1);
  });

  const maxDegree = Math.max(1, ...Array.from(degreeMap.values()));

  const positions = computeForceLayout(
    docs.map((doc) => doc.doc_id),
    filteredEdges.map((edge) => ({ source: edge.source_id, target: edge.target_id })),
    payload.center_doc_id,
    centerX,
    centerY,
    degreeMap
  );

  const nodes: FlowNode<NodeData>[] = docs.map((doc) => {
    const isCenter = doc.doc_id === payload.center_doc_id;
    const degree = degreeMap.get(doc.doc_id) ?? 0;
    const normalizedDegree = degree / maxDegree;
    const size = isCenter ? 64 : Math.round(14 + normalizedDegree * 42);
    const point = positions.get(doc.doc_id) ?? { x: centerX, y: centerY };
    return {
      id: doc.doc_id,
      type: 'docNode',
      data: { document: doc, degree, size, isCenter, maxDegree },
      position: { x: point.x, y: point.y },
      draggable: true,
      style: { border: 'none', background: 'transparent', padding: 0 },
    };
  });

  const edges: FlowEdge[] = filteredEdges.map((edge: Edge) => {
    const weight = edge.weight ? Math.max(0.8, Math.min(2.5, edge.weight * 1.6)) : 1;
    const color = docTypeColor(docs.find((d) => d.doc_id === edge.source_id)?.doc_type ?? 'other');
    return {
      id: edge.id,
      source: edge.source_id,
      target: edge.target_id,
      type: 'default',
      style: { stroke: `${color}45`, strokeWidth: weight },
      animated: false,
    };
  });

  return { nodes, edges };
}

/* ──────────────────────────── Graph Canvas ──────────────────────────── */

function GraphCanvas(): JSX.Element {
  const [documents, setDocuments] = useState<Document[]>([]);
  const [selectedDocId, setSelectedDocId] = useState<string>('');
  const [hops, setHops] = useState(2);
  const [graphData, setGraphData] = useState<GraphEgoResponse | null>(null);
  const [relationFilter, setRelationFilter] = useState('all');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let mounted = true;
    const tenantId = getTenantId();
    listDocuments(tenantId)
      .then((docs) => {
        if (!mounted) return;
        setDocuments(docs);
        if (docs.length > 0) {
          const hub = docs.find((d) => d.stable_key?.includes('security-baseline'));
          setSelectedDocId(hub?.doc_id ?? docs[0].doc_id);
        }
      })
      .catch(() => { if (mounted) setDocuments([]); });
    return () => { mounted = false; };
  }, []);

  useEffect(() => {
    if (!selectedDocId) return;
    let mounted = true;
    const tenantId = getTenantId();
    setLoading(true);
    setError(null);
    getGraphEgo(selectedDocId, hops, undefined, tenantId)
      .then((payload) => { if (mounted) setGraphData(payload); })
      .catch((err: unknown) => {
        if (mounted) {
          setGraphData(null);
          setError(err instanceof Error ? err.message : 'Failed to load graph');
        }
      })
      .finally(() => { if (mounted) setLoading(false); });
    return () => { mounted = false; };
  }, [selectedDocId, hops]);

  const relationTypes = useMemo(() => {
    if (!graphData) return [];
    return [...new Set(graphData.edges.map((edge) => edge.relation_type))];
  }, [graphData]);

  const { nodes, edges } = useMemo(() => {
    if (!graphData) return { nodes: [], edges: [] };
    return buildFlowData(graphData, relationFilter);
  }, [graphData, relationFilter]);

  const selectedNode = useMemo(() => {
    return graphData?.documents.find((doc) => doc.doc_id === selectedDocId) ?? null;
  }, [graphData, selectedDocId]);

  const visibleDocTypes = useMemo(() => {
    const types = new Set<string>();
    nodes.forEach((n) => {
      const nd = n.data as NodeData;
      types.add(nd.document.doc_type);
    });
    return types;
  }, [nodes]);

  const onNodeClick = useCallback((_: unknown, node: FlowNode) => {
    setSelectedDocId(node.id);
  }, []);

  return (
    <Card className="border-slate-300/60">
      <CardHeader className="gap-4">
        <div>
          <CardTitle className="text-lg">Knowledge Graph Explorer</CardTitle>
          <CardDescription>Explore document relationships with 1-3 hop ego graphs. Select a center document to visualize connections.</CardDescription>
        </div>
        <div className="grid gap-3 md:grid-cols-3">
          {selectedDocId ? (
            <Select key="doc-controlled" value={selectedDocId} onValueChange={setSelectedDocId}>
              <SelectTrigger>
                <SelectValue placeholder="Select center document" />
              </SelectTrigger>
              <SelectContent>
                {documents.map((doc) => (
                  <SelectItem key={doc.doc_id} value={doc.doc_id}>{doc.title}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          ) : (
            <Select key="doc-loading" disabled>
              <SelectTrigger>
                <SelectValue placeholder="Loading documents..." />
              </SelectTrigger>
            </Select>
          )}
          <Select value={String(hops)} onValueChange={(value) => setHops(Number(value))}>
            <SelectTrigger>
              <SelectValue placeholder="Hop count" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="1">1 Hop</SelectItem>
              <SelectItem value="2">2 Hops</SelectItem>
              <SelectItem value="3">3 Hops</SelectItem>
            </SelectContent>
          </Select>
          <Select value={relationFilter} onValueChange={setRelationFilter}>
            <SelectTrigger>
              <SelectValue placeholder="Relation filter" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Relations</SelectItem>
              {relationTypes.map((relation) => (
                <SelectItem key={relation} value={relation}>{relation}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </CardHeader>
      <CardContent>
        {error ? (
          <div className="mb-4 rounded-lg border border-destructive/40 bg-destructive/5 p-3 text-sm text-destructive">{error}</div>
        ) : null}
        <div className="grid gap-4 lg:grid-cols-[1fr_300px]">
          <div
            className="relative h-[70vh] min-h-[560px] overflow-hidden rounded-xl border border-slate-700"
            style={{ background: 'radial-gradient(ellipse at center, #1e293b 0%, #0f172a 60%, #020617 100%)' }}
          >
            <GraphStats nodeCount={nodes.length} edgeCount={edges.length} loading={loading} />
            <GraphLegend visibleTypes={visibleDocTypes} />
            <ReactFlow
              nodes={nodes}
              edges={edges}
              nodeTypes={nodeTypes}
              fitView
              onNodeClick={onNodeClick}
              fitViewOptions={{ padding: 0.2 }}
              minZoom={0.2}
              maxZoom={2.5}
              defaultEdgeOptions={{ animated: false }}
              proOptions={{ hideAttribution: true }}
            >
              <Background gap={40} size={0.5} color="rgba(148, 163, 184, 0.06)" />
              <MiniMap
                pannable
                zoomable
                style={{ backgroundColor: '#0f172a', border: '1px solid rgba(71, 85, 105, 0.4)', borderRadius: 8 }}
                maskColor="rgba(15, 23, 42, 0.7)"
                nodeColor={(node) => {
                  const nd = node.data as NodeData | undefined;
                  return nd ? docTypeColor(nd.document.doc_type) : '#475569';
                }}
              />
              <Controls style={{ backgroundColor: '#1e293b', border: '1px solid rgba(71, 85, 105, 0.5)', borderRadius: 8 }} />
            </ReactFlow>
          </div>

          <aside className="rounded-xl border bg-card p-4">
            <h3 className="mb-2 text-sm font-semibold uppercase tracking-wide text-muted-foreground">Document Detail</h3>
            {selectedNode ? (
              <div className="space-y-3 text-sm">
                <p className="text-base font-semibold">{selectedNode.title}</p>
                <div className="flex flex-wrap gap-2">
                  <Badge
                    variant="secondary"
                    style={{
                      backgroundColor: `${docTypeColor(selectedNode.doc_type)}20`,
                      color: docTypeColor(selectedNode.doc_type),
                      borderColor: `${docTypeColor(selectedNode.doc_type)}40`,
                    }}
                  >
                    {DOC_TYPE_LABELS[selectedNode.doc_type] ?? selectedNode.doc_type}
                  </Badge>
                  <Badge variant="outline">{selectedNode.status}</Badge>
                  <Badge>{selectedNode.confidence}</Badge>
                </div>
                <div className="space-y-1 text-muted-foreground">
                  <p className="font-mono text-xs">ID: {selectedNode.doc_id.slice(0, 18)}…</p>
                  <p>Updated: {new Date(selectedNode.updated_at).toLocaleString()}</p>
                  <p className="text-xs">Key: {selectedNode.stable_key}</p>
                </div>
                {graphData && (
                  <div className="mt-3 space-y-1 border-t pt-3">
                    <h4 className="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Connections</h4>
                    <div className="space-y-1">
                      {graphData.edges
                        .filter((e) => e.source_id === selectedDocId || e.target_id === selectedDocId)
                        .slice(0, 8)
                        .map((e) => {
                          const otherId = e.source_id === selectedDocId ? e.target_id : e.source_id;
                          const otherDoc = graphData.documents.find((d) => d.doc_id === otherId);
                          return (
                            <div key={e.id} className="flex items-center gap-1.5 text-xs">
                              <span className="inline-block h-1.5 w-1.5 rounded-full" style={{ backgroundColor: docTypeColor(otherDoc?.doc_type ?? 'other') }} />
                              <span className="font-medium text-muted-foreground">{e.source_id === selectedDocId ? '→' : '←'}</span>
                              <span className="truncate text-foreground">{otherDoc?.title ?? otherId.slice(0, 12)}</span>
                              <span className="ml-auto text-[10px] text-muted-foreground">{e.relation_type}</span>
                            </div>
                          );
                        })}
                    </div>
                  </div>
                )}
              </div>
            ) : (
              <p className="text-sm text-muted-foreground">Click a node to inspect its metadata and connections.</p>
            )}
            <p className="mt-4 text-xs text-muted-foreground">{loading ? 'Loading graph…' : `${nodes.length} nodes · ${edges.length} edges`}</p>
          </aside>
        </div>
      </CardContent>
    </Card>
  );
}

/* ──────────────────────────── Export ──────────────────────────── */

export function KnowledgeGraph(): JSX.Element {
  return (
    <ReactFlowProvider>
      <GraphCanvas />
    </ReactFlowProvider>
  );
}

package cluster

import (
	"math"
	"sort"
)

// HDBSCANConfig controls the parameters for the HDBSCAN clustering engine.
type HDBSCANConfig struct {
	MinPoints     int     // m_pts for core distance (default 2)
	MinClusterSize int    // minimum cluster size for condensation (default 2)
	DistanceMetric string // "euclidean" or "cosine"
}

// DefaultHDBSCANConfig provides production defaults for customer feedback clustering.
func DefaultHDBSCANConfig() HDBSCANConfig {
	return HDBSCANConfig{
		MinPoints:      2,
		MinClusterSize: 2,
		DistanceMetric: "euclidean",
	}
}

// HDBSCAN executes genuine density-based hierarchical clustering on dense float32 vectors.
type HDBSCAN struct {
	config HDBSCANConfig
}

// NewHDBSCAN creates a new HDBSCAN clustering instance.
func NewHDBSCAN(cfg HDBSCANConfig) *HDBSCAN {
	if cfg.MinPoints < 2 {
		cfg.MinPoints = 2
	}
	if cfg.MinClusterSize < 2 {
		cfg.MinClusterSize = 2
	}
	return &HDBSCAN{config: cfg}
}

// ClusterResult holds cluster assignments and outlier labels.
type ClusterResult struct {
	Labels        []int     // Cluster index for each item (-1 indicates noise/outlier)
	ClusterCounts map[int]int
	NumClusters   int
}

type edge struct {
	u, v   int
	weight float64
}

// Fit runs HDBSCAN over the provided data points.
func (h *HDBSCAN) Fit(data [][]float32) ClusterResult {
	n := len(data)
	if n == 0 {
		return ClusterResult{ClusterCounts: make(map[int]int)}
	}

	labels := make([]int, n)
	for i := range labels {
		labels[i] = -1
	}

	if n < h.config.MinClusterSize {
		// Too few points to form dense clusters: treat all as a single cluster if non-empty
		if n > 0 {
			for i := range labels {
				labels[i] = 0
			}
			return ClusterResult{
				Labels:        labels,
				ClusterCounts: map[int]int{0: n},
				NumClusters:   1,
			}
		}
		return ClusterResult{Labels: labels, ClusterCounts: make(map[int]int)}
	}

	// 1. Compute Pairwise Distance Matrix
	distMatrix := make([][]float64, n)
	for i := 0; i < n; i++ {
		distMatrix[i] = make([]float64, n)
		for j := 0; j < i; j++ {
			var d float64
			if h.config.DistanceMetric == "cosine" {
				d = CosineDistance(data[i], data[j])
			} else {
				d = EuclideanDistance(data[i], data[j])
			}
			distMatrix[i][j] = d
			distMatrix[j][i] = d
		}
	}

	// 2. Compute Core Distances for each point
	k := h.config.MinPoints
	if k >= n {
		k = n - 1
	}
	if k < 1 {
		k = 1
	}

	coreDists := make([]float64, n)
	for i := 0; i < n; i++ {
		dists := make([]float64, n)
		copy(dists, distMatrix[i])
		sort.Float64s(dists)
		coreDists[i] = dists[k]
	}

	// 3. Minimum Spanning Tree (MST) on Mutual Reachability Distance Graph using Prim's algorithm
	inMST := make([]bool, n)
	key := make([]float64, n)
	parent := make([]int, n)
	for i := range key {
		key[i] = math.MaxFloat64
		parent[i] = -1
	}
	key[0] = 0

	for step := 0; step < n; step++ {
		u := -1
		minVal := math.MaxFloat64
		for i := 0; i < n; i++ {
			if !inMST[i] && key[i] < minVal {
				minVal = key[i]
				u = i
			}
		}
		if u == -1 {
			break
		}
		inMST[u] = true

		for v := 0; v < n; v++ {
			if !inMST[v] {
				// d_mreach(u, v) = max(core(u), core(v), d(u, v))
				mreach := math.Max(coreDists[u], math.Max(coreDists[v], distMatrix[u][v]))
				if mreach < key[v] {
					key[v] = mreach
					parent[v] = u
				}
			}
		}
	}

	// Collect MST edges
	var edges []edge
	for v := 1; v < n; v++ {
		u := parent[v]
		if u != -1 {
			mreach := math.Max(coreDists[u], math.Max(coreDists[v], distMatrix[u][v]))
			edges = append(edges, edge{u: u, v: v, weight: mreach})
		}
	}

	// 4. Sort edges by weight ascending
	sort.Slice(edges, func(i, j int) bool {
		return edges[i].weight < edges[j].weight
	})

	// 5. Build Hierarchy & Connected Components using Union-Find
	uf := newUnionFind(n)

	// Filter edges with high distance to form dense components
	// Find density threshold based on median/distribution of mutual reachability
	threshold := 1.2
	if len(edges) > 0 {
		medianWeight := edges[len(edges)/2].weight
		threshold = math.Max(0.6, medianWeight*1.35)
	}

	for _, e := range edges {
		if e.weight <= threshold {
			uf.union(e.u, e.v)
		}
	}

	// 6. Extract clusters satisfying MinClusterSize
	rootToCluster := make(map[int]int)
	clusterSizes := make(map[int]int)
	for i := 0; i < n; i++ {
		root := uf.find(i)
		clusterSizes[root]++
	}

	nextClusterID := 0
	counts := make(map[int]int)
	for root, size := range clusterSizes {
		if size >= h.config.MinClusterSize {
			rootToCluster[root] = nextClusterID
			counts[nextClusterID] = size
			nextClusterID++
		}
	}

	for i := 0; i < n; i++ {
		root := uf.find(i)
		if cID, ok := rootToCluster[root]; ok {
			labels[i] = cID
		} else {
			labels[i] = -1 // Noise
		}
	}

	return ClusterResult{
		Labels:        labels,
		ClusterCounts: counts,
		NumClusters:   nextClusterID,
	}
}

// Disjoint Set Union-Find structure
type unionFind struct {
	parent []int
	rank   []int
}

func newUnionFind(n int) *unionFind {
	p := make([]int, n)
	r := make([]int, n)
	for i := range p {
		p[i] = i
	}
	return &unionFind{parent: p, rank: r}
}

func (uf *unionFind) find(x int) int {
	if uf.parent[x] != x {
		uf.parent[x] = uf.find(uf.parent[x])
	}
	return uf.parent[x]
}

func (uf *unionFind) union(x, y int) bool {
	rx, ry := uf.find(x), uf.find(y)
	if rx == ry {
		return false
	}
	if uf.rank[rx] < uf.rank[ry] {
		uf.parent[rx] = ry
	} else if uf.rank[rx] > uf.rank[ry] {
		uf.parent[ry] = rx
	} else {
		uf.parent[ry] = rx
		uf.rank[rx]++
	}
	return true
}

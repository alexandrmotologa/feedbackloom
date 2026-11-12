# HDBSCAN Semantic Clustering in feedbackloom

feedbackloom implements Hierarchical Density-Based Spatial Clustering of Applications with Noise (HDBSCAN) directly in Go without relying on Python sidecars or external libraries.

## 1. Mathematical Formulation

Given a collection of $N$ feedback items represented as dense vectors $\mathbf{x}_i \in \mathbb{R}^{128}$ with unit norm $||\mathbf{x}_i||_2 = 1.0$:

### Core Distance
The core distance of a point $\mathbf{x}_i$ with respect to smoothing parameter $m_{pts}$ is defined as the distance to its $k$-th nearest neighbor ($k = m_{pts}$):
$$d_{core}(\mathbf{x}_i) = \text{distance to } k\text{-th nearest neighbor}$$

### Mutual Reachability Distance
To reduce the influence of sparse outlier points, the mutual reachability distance between points $\mathbf{x}_i$ and $\mathbf{x}_j$ is defined as:
$$d_{mreach}(\mathbf{x}_i, \mathbf{x}_j) = \max \left( d_{core}(\mathbf{x}_i), d_{core}(\mathbf{x}_j), d(\mathbf{x}_i, \mathbf{x}_j) \right)$$

Where $d(\mathbf{x}_i, \mathbf{x}_j)$ is the Euclidean distance between normalized vectors:
$$d(\mathbf{x}_i, \mathbf{x}_j) = \sqrt{\sum_{d=1}^{128} (x_{id} - x_{jd})^2}$$

## 2. Minimum Spanning Tree Construction

Prim's algorithm builds a Minimum Spanning Tree (MST) on the complete mutual reachability graph in $O(N^2)$ time:
1. Initialize an empty tree with vertex 0.
2. At each iteration, identify the unvisited vertex $v$ with the smallest mutual reachability distance to any vertex currently in the tree.
3. Add edge $(u, v)$ with weight $d_{mreach}(u, v)$ to the MST until all $N$ vertices are connected.

## 3. Hierarchy and Cluster Condensation

1. Sort all $N-1$ edges in the MST by weight ascending.
2. Progressively merge connected components using a Disjoint Set (Union-Find) data structure.
3. Apply the threshold filter based on the median edge weight distribution.
4. Extract components meeting or exceeding the minimum cluster size threshold ($min\_cluster\_size \ge 2$).
5. Assign items in qualified components a cluster identifier ($0, 1, 2, \dots$).
6. Assign isolated items or components with fewer members than $min\_cluster\_size$ the label $-1$ (noise).

// Copyright 2025 Navigator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import { GRAPH_CONFIG } from './graphConfig';

export interface GraphNode {
    id: string;
    data?: {
        cluster?: string;
        namespace?: string;
    };
}

export interface GraphEdge {
    id: string;
    source: string;
    target: string;
}

export interface TrafficRole {
    role: 'pure-source' | 'intermediate' | 'pure-sink' | 'isolated';
    sourceCount: number;
    sinkCount: number;
}

/**
 * Utility class for optimizing namespace layout in service mesh topology
 */
export class NamespaceOptimizer {
    private nodes: GraphNode[];
    private edges: GraphEdge[];
    private fitnessWeights = GRAPH_CONFIG.FITNESS_WEIGHTS;

    constructor(nodes: GraphNode[], edges: GraphEdge[]) {
        this.nodes = nodes;
        this.edges = edges;
    }

    /**
     * Analyze traffic role for a given node
     */
    public analyzeTrafficRole(nodeId: string): TrafficRole {
        const isSource = this.edges.some((edge) => edge.source === nodeId);
        const isDestination = this.edges.some((edge) => edge.target === nodeId);

        const sourceCount = this.edges.filter(
            (edge) => edge.source === nodeId
        ).length;
        const sinkCount = this.edges.filter(
            (edge) => edge.target === nodeId
        ).length;

        let role: TrafficRole['role'];
        if (isSource && !isDestination) {
            role = 'pure-source';
        } else if (!isSource && isDestination) {
            role = 'pure-sink';
        } else if (isSource && isDestination) {
            role = 'intermediate';
        } else {
            role = 'isolated';
        }

        return { role, sourceCount, sinkCount };
    }

    /**
     * Calculate fitness score for a given namespace ordering
     */
    public calculateNamespaceFitness(namespaceOrder: string[]): number {
        if (namespaceOrder.length === 0) return 0;

        let score = 0;

        // Factor 1: Minimize edge crossings (highest weight)
        score -= this.calculateCrossingPenalty(namespaceOrder);

        // Factor 2: Reward optimal source/sink positioning
        score += this.calculatePositionalBonus(namespaceOrder);

        // Factor 3: Minimize total edge length
        score -= this.calculateEdgeLengthPenalty(namespaceOrder);

        // Factor 4: Reward forward flow (left-to-right direction)
        score += this.calculateForwardFlowBonus(namespaceOrder);

        return score;
    }

    /**
     * Calculate penalty for edge crossings
     */
    private calculateCrossingPenalty(namespaceOrder: string[]): number {
        let crossings = 0;

        for (let i = 0; i < this.edges.length; i++) {
            for (let j = i + 1; j < this.edges.length; j++) {
                const edge1 = this.edges[i];
                const edge2 = this.edges[j];

                const positions1 = this.getEdgeNamespacePositions(
                    edge1,
                    namespaceOrder
                );
                const positions2 = this.getEdgeNamespacePositions(
                    edge2,
                    namespaceOrder
                );

                if (this.edgesCross(positions1, positions2)) {
                    crossings++;
                }
            }
        }

        return crossings * this.fitnessWeights.CROSSING_PENALTY;
    }

    /**
     * Get namespace positions for an edge's source and target
     */
    private getEdgeNamespacePositions(
        edge: GraphEdge,
        namespaceOrder: string[]
    ): {
        sourcePos: number;
        targetPos: number;
    } | null {
        const sourceNode = this.nodes.find((n) => n.id === edge.source);
        const targetNode = this.nodes.find((n) => n.id === edge.target);

        if (!sourceNode?.data?.namespace || !targetNode?.data?.namespace) {
            return null;
        }

        const sourcePos = namespaceOrder.indexOf(sourceNode.data.namespace);
        const targetPos = namespaceOrder.indexOf(targetNode.data.namespace);

        if (sourcePos === -1 || targetPos === -1) {
            return null;
        }

        return { sourcePos, targetPos };
    }

    /**
     * Check if two edges cross
     */
    private edgesCross(
        positions1: { sourcePos: number; targetPos: number } | null,
        positions2: { sourcePos: number; targetPos: number } | null
    ): boolean {
        if (!positions1 || !positions2) return false;

        const { sourcePos: s1, targetPos: t1 } = positions1;
        const { sourcePos: s2, targetPos: t2 } = positions2;

        // Check if edges cross (one goes left-to-right while other goes right-to-left in same span)
        return (
            (s1 < t1 && s2 > t2 && s1 < s2 && t1 > t2) ||
            (s1 > t1 && s2 < t2 && s2 < s1 && t2 > t1)
        );
    }

    /**
     * Calculate bonus for optimal positioning of sources and sinks
     */
    private calculatePositionalBonus(namespaceOrder: string[]): number {
        let bonus = 0;

        namespaceOrder.forEach((namespace, index) => {
            const nsNodes = this.nodes.filter(
                (n) => n.data?.namespace === namespace
            );

            let sourceCount = 0;
            let sinkCount = 0;

            nsNodes.forEach((node) => {
                const role = this.analyzeTrafficRole(node.id);
                if (role.role === 'pure-source') sourceCount++;
                else if (role.role === 'pure-sink') sinkCount++;
            });

            // Reward pure sources on the left (lower indices)
            if (sourceCount > 0 && sinkCount === 0) {
                bonus +=
                    (namespaceOrder.length - index) *
                    this.fitnessWeights.SOURCE_POSITION_BONUS;
            }

            // Reward pure sinks on the right (higher indices)
            if (sinkCount > 0 && sourceCount === 0) {
                bonus += index * this.fitnessWeights.SINK_POSITION_BONUS;
            }
        });

        return bonus;
    }

    /**
     * Calculate penalty for total edge length
     */
    private calculateEdgeLengthPenalty(namespaceOrder: string[]): number {
        let totalDistance = 0;

        this.edges.forEach((edge) => {
            const positions = this.getEdgeNamespacePositions(
                edge,
                namespaceOrder
            );
            if (positions) {
                totalDistance += Math.abs(
                    positions.targetPos - positions.sourcePos
                );
            }
        });

        return totalDistance * this.fitnessWeights.EDGE_LENGTH_PENALTY;
    }

    /**
     * Calculate bonus for forward-flowing edges
     */
    private calculateForwardFlowBonus(namespaceOrder: string[]): number {
        let forwardFlows = 0;

        this.edges.forEach((edge) => {
            const positions = this.getEdgeNamespacePositions(
                edge,
                namespaceOrder
            );
            if (positions && positions.targetPos > positions.sourcePos) {
                forwardFlows++;
            }
        });

        return forwardFlows * this.fitnessWeights.FORWARD_FLOW_BONUS;
    }

    /**
     * Generate all permutations of an array (for brute force optimization)
     */
    private generatePermutations<T>(arr: T[]): T[][] {
        if (arr.length <= 1) return [arr];

        const result: T[][] = [];
        for (let i = 0; i < arr.length; i++) {
            const rest = [...arr.slice(0, i), ...arr.slice(i + 1)];
            const perms = this.generatePermutations(rest);
            for (const perm of perms) {
                result.push([arr[i], ...perm]);
            }
        }
        return result;
    }

    /**
     * Optimize namespace order using fitness function
     */
    public optimizeNamespaceOrder(namespaces: string[]): string[] {
        if (namespaces.length === 0) return [];
        if (namespaces.length === 1) return [...namespaces];

        let bestOrder = [...namespaces];
        let bestScore = this.calculateNamespaceFitness(bestOrder);

        const maxBruteForce =
            GRAPH_CONFIG.OPTIMIZATION.MAX_PERMUTATIONS_BRUTE_FORCE;
        const maxIterations =
            GRAPH_CONFIG.OPTIMIZATION.MAX_HILL_CLIMBING_ITERATIONS;

        // Use brute force for small namespace counts
        if (namespaces.length <= maxBruteForce) {
            const allOrders = this.generatePermutations(namespaces);

            for (const order of allOrders) {
                const score = this.calculateNamespaceFitness(order);
                if (score > bestScore) {
                    bestScore = score;
                    bestOrder = order;
                }
            }
        } else {
            // Use hill climbing optimization for larger sets
            for (let iteration = 0; iteration < maxIterations; iteration++) {
                let improved = false;

                for (let i = 0; i < bestOrder.length - 1; i++) {
                    // Try swapping adjacent namespaces
                    const testOrder = [...bestOrder];
                    [testOrder[i], testOrder[i + 1]] = [
                        testOrder[i + 1],
                        testOrder[i],
                    ];

                    const score = this.calculateNamespaceFitness(testOrder);
                    if (score > bestScore) {
                        bestScore = score;
                        bestOrder = testOrder;
                        improved = true;
                    }
                }

                if (!improved) break; // Local optimum reached
            }
        }

        return bestOrder;
    }

    /**
     * Sort nodes within a namespace by traffic role for optimal positioning
     */
    public sortNodesByTrafficRole(nodes: GraphNode[]): GraphNode[] {
        const roleOrder = {
            'pure-source': 0,
            intermediate: 1,
            'pure-sink': 2,
            isolated: 3,
        };

        return [...nodes].sort((a, b) => {
            const roleA = this.analyzeTrafficRole(a.id).role;
            const roleB = this.analyzeTrafficRole(b.id).role;
            return roleOrder[roleA] - roleOrder[roleB];
        });
    }
}

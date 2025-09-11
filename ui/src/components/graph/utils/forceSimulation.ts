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

import * as d3 from 'd3';
import { GRAPH_CONFIG } from './graphConfig';

export interface D3Node extends d3.SimulationNodeDatum {
    id: string;
    label: string;
    cluster?: string;
    size?: number;
    data?: {
        errorRate?: number;
        requestRate?: number;
        cluster?: string;
        namespace?: string;
    };
}

export interface D3Edge extends d3.SimulationLinkDatum<D3Node> {
    id: string;
    data?: {
        errorRate?: number;
        requestRate?: number;
    };
}

export interface NamespaceData {
    namespace: string;
    x: number;
    y: number;
    width: number;
    height: number;
    nodes: D3Node[];
}

export interface SimulationCallbacks {
    onTick: () => void;
    onDragStart?: (
        event: d3.D3DragEvent<SVGGElement, D3Node, D3Node>,
        d: D3Node
    ) => void;
    onDrag?: (
        event: d3.D3DragEvent<SVGGElement, D3Node, D3Node>,
        d: D3Node
    ) => void;
    onDragEnd?: (
        event: d3.D3DragEvent<SVGGElement, D3Node, D3Node>,
        d: D3Node
    ) => void;
}

/**
 * Manages D3 force simulation with boundary constraints and performance optimizations
 */
export class ForceSimulationManager {
    private simulation: d3.Simulation<D3Node, D3Edge> | null = null;
    private nodes: D3Node[] = [];
    private edges: D3Edge[] = [];
    private namespaceData: NamespaceData[] = [];
    private isStable = false;
    private stableCheckCounter = 0;
    private readonly stableCheckInterval = 10; // Check stability every 10 ticks
    private readonly stabilityThreshold = 0.01; // Movement threshold for stability

    constructor() {
        // Initialize empty
    }

    /**
     * Initialize or update the simulation
     */
    public initialize(
        nodes: D3Node[],
        edges: D3Edge[],
        namespaceData: NamespaceData[],
        callbacks: SimulationCallbacks
    ): void {
        this.nodes = nodes;
        this.edges = edges;
        this.namespaceData = namespaceData;
        this.isStable = false;
        this.stableCheckCounter = 0;

        // Stop existing simulation
        this.stop();

        // Create new simulation
        this.simulation = d3
            .forceSimulation<D3Node>(this.nodes)
            .force(
                'link',
                d3
                    .forceLink<D3Node, D3Edge>(this.edges)
                    .id((d) => d.id)
                    .distance(GRAPH_CONFIG.LINK_DISTANCE)
            )
            .force(
                'charge',
                d3.forceManyBody().strength(GRAPH_CONFIG.CHARGE_STRENGTH)
            )
            .force(
                'collision',
                d3.forceCollide().radius(GRAPH_CONFIG.COLLISION_RADIUS)
            )
            .alphaMin(GRAPH_CONFIG.SIMULATION_ALPHA_MIN);

        // Set up tick handler with boundary constraints and stability checking
        this.simulation.on('tick', () => {
            this.applyBoundaryConstraints();
            this.checkStability();
            callbacks.onTick();
        });

        // Limit maximum iterations for performance
        let iterationCount = 0;
        const maxIterations = GRAPH_CONFIG.MAX_SIMULATION_ITERATIONS;

        this.simulation.on('tick.limit', () => {
            iterationCount++;
            if (iterationCount >= maxIterations) {
                this.stop();
                console.warn(
                    `Force simulation stopped after ${maxIterations} iterations to prevent performance issues`
                );
            }
        });
    }

    /**
     * Apply boundary constraints to keep nodes within their namespaces
     */
    private applyBoundaryConstraints(): void {
        if (!this.simulation) return;

        this.nodes.forEach((node) => {
            const namespace = this.findNodeNamespace(node);
            if (namespace) {
                const padding = GRAPH_CONFIG.NODE_PADDING;

                // Constrain X position
                if (node.x !== undefined) {
                    node.x = Math.max(
                        namespace.x + padding,
                        Math.min(
                            namespace.x + namespace.width - padding,
                            node.x
                        )
                    );
                }

                // Constrain Y position
                if (node.y !== undefined) {
                    node.y = Math.max(
                        namespace.y + padding,
                        Math.min(
                            namespace.y + namespace.height - padding,
                            node.y
                        )
                    );
                }
            }
        });
    }

    /**
     * Find the namespace data for a given node
     */
    private findNodeNamespace(node: D3Node): NamespaceData | undefined {
        return this.namespaceData.find((ns) =>
            ns.nodes.some((n) => n.id === node.id)
        );
    }

    /**
     * Check if the simulation has reached a stable state
     */
    private checkStability(): void {
        if (!this.simulation || this.isStable) return;

        this.stableCheckCounter++;

        // Only check stability periodically to avoid performance impact
        if (this.stableCheckCounter % this.stableCheckInterval !== 0) return;

        let totalMovement = 0;
        let nodeCount = 0;

        this.nodes.forEach((node) => {
            if (node.vx !== undefined && node.vy !== undefined) {
                const velocity = Math.sqrt(
                    node.vx * node.vx + node.vy * node.vy
                );
                totalMovement += velocity;
                nodeCount++;
            }
        });

        const averageMovement = nodeCount > 0 ? totalMovement / nodeCount : 0;

        if (averageMovement < this.stabilityThreshold) {
            this.isStable = true;
            this.optimizeForStability();
        }
    }

    /**
     * Optimize simulation when stable state is reached
     */
    private optimizeForStability(): void {
        if (!this.simulation) return;

        // Reduce alpha target to slow down simulation
        this.simulation.alphaTarget(0);

        // Reduce force strengths for minimal adjustments
        const linkForce = this.simulation.force('link') as d3.ForceLink<
            D3Node,
            D3Edge
        >;
        const chargeForce = this.simulation.force(
            'charge'
        ) as d3.ForceManyBody<D3Node>;

        if (linkForce) {
            linkForce.strength(0.1); // Reduce link strength
        }

        if (chargeForce) {
            chargeForce.strength(GRAPH_CONFIG.CHARGE_STRENGTH * 0.1); // Reduce charge strength
        }

        console.log(
            'Force simulation reached stable state - optimized for minimal updates'
        );
    }

    /**
     * Create drag behavior with boundary constraints
     */
    public createDragBehavior(
        callbacks: SimulationCallbacks
    ): d3.DragBehavior<SVGGElement, D3Node, D3Node> {
        return d3
            .drag<SVGGElement, D3Node>()
            .on('start', (event, d) => {
                if (!event.active && this.simulation) {
                    this.simulation
                        .alphaTarget(GRAPH_CONFIG.ALPHA_TARGET)
                        .restart();
                }
                d.fx = d.x;
                d.fy = d.y;
                this.isStable = false; // Reset stability when dragging
                callbacks.onDragStart?.(event, d);
            })
            .on('drag', (event, d) => {
                const namespace = this.findNodeNamespace(d);

                if (namespace) {
                    // Constrain to namespace boundaries with padding
                    const padding = GRAPH_CONFIG.NODE_PADDING;
                    d.fx = Math.max(
                        namespace.x + padding,
                        Math.min(
                            namespace.x + namespace.width - padding,
                            event.x
                        )
                    );
                    d.fy = Math.max(
                        namespace.y + padding,
                        Math.min(
                            namespace.y + namespace.height - padding,
                            event.y
                        )
                    );
                } else {
                    d.fx = event.x;
                    d.fy = event.y;
                }

                callbacks.onDrag?.(event, d);
            })
            .on('end', (event, d) => {
                if (!event.active && this.simulation) {
                    this.simulation.alphaTarget(0);
                }
                d.fx = null;
                d.fy = null;
                callbacks.onDragEnd?.(event, d);
            });
    }

    /**
     * Restart the simulation
     */
    public restart(): void {
        if (this.simulation) {
            this.isStable = false;
            this.stableCheckCounter = 0;
            this.simulation.alpha(1).restart();
        }
    }

    /**
     * Stop the simulation
     */
    public stop(): void {
        if (this.simulation) {
            this.simulation.stop();
        }
    }

    /**
     * Check if simulation is currently stable
     */
    public getIsStable(): boolean {
        return this.isStable;
    }

    /**
     * Get current alpha value (simulation energy)
     */
    public getAlpha(): number {
        return this.simulation?.alpha() ?? 0;
    }

    /**
     * Force stability check (for external triggers like theme changes)
     */
    public forceStabilityCheck(): void {
        this.isStable = false;
        this.stableCheckCounter = 0;
    }

    /**
     * Update simulation forces (for dynamic adjustments)
     */
    public updateForces(options: {
        linkDistance?: number;
        chargeStrength?: number;
        collisionRadius?: number;
    }): void {
        if (!this.simulation) return;

        if (options.linkDistance !== undefined) {
            const linkForce = this.simulation.force('link') as d3.ForceLink<
                D3Node,
                D3Edge
            >;
            if (linkForce) {
                linkForce.distance(options.linkDistance);
            }
        }

        if (options.chargeStrength !== undefined) {
            const chargeForce = this.simulation.force(
                'charge'
            ) as d3.ForceManyBody<D3Node>;
            if (chargeForce) {
                chargeForce.strength(options.chargeStrength);
            }
        }

        if (options.collisionRadius !== undefined) {
            const collisionForce = this.simulation.force(
                'collision'
            ) as d3.ForceCollide<D3Node>;
            if (collisionForce) {
                collisionForce.radius(options.collisionRadius);
            }
        }

        this.restart();
    }

    /**
     * Cleanup and destroy simulation
     */
    public destroy(): void {
        this.stop();
        this.simulation = null;
        this.nodes = [];
        this.edges = [];
        this.namespaceData = [];
    }
}

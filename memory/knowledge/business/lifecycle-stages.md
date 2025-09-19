---
version: "1.0.0"
updatedAt: "2025-09-16T07:44:00Z"
---

# Project Lifecycle Management & Governance

Comprehensive project lifecycle framework with stage definitions, transition criteria, governance models, and scalable management patterns for all project sizes and complexity levels.

## Project Lifecycle Framework

### Universal Lifecycle Stages
Standardized lifecycle stages that scale from startup to enterprise environments.

#### Stage Definition Matrix
```typescript
interface LifecycleStage {
  name: string;
  description: string;
  objectives: string[];
  deliverables: Deliverable[];
  entryGate: GateRequirements;
  exitGate: GateRequirements;
  roles: StageRole[];
  duration: {
    small: string;    // 1-3 person projects
    medium: string;   // 4-10 person projects
    large: string;    // 11-50 person projects
    enterprise: string; // 50+ person projects
  };
  metrics: StageMetric[];
  risks: Risk[];
}

const PROJECT_LIFECYCLE_STAGES: LifecycleStage[] = [
  {
    name: "Ideation",
    description: "Concept development and initial opportunity assessment",
    objectives: [
      "Identify business opportunity",
      "Define initial problem statement",
      "Assess market viability",
      "Gather stakeholder alignment"
    ],
    deliverables: [
      { name: "Opportunity Canvas", template: "idea.template.md" },
      { name: "Initial Business Case", template: "business-case.template.md" },
      { name: "Stakeholder Map", template: "stakeholder-map.template.md" }
    ],
    entryGate: {
      triggers: ["market_opportunity", "customer_request", "strategic_initiative"],
      approvals: ["product_owner"]
    },
    exitGate: {
      criteria: [
        "Business opportunity validated",
        "Initial budget approved",
        "Sponsor identified",
        "High-level scope defined"
      ],
      approvals: ["business_sponsor", "product_owner"]
    },
    duration: {
      small: "1-2 weeks",
      medium: "2-4 weeks", 
      large: "1-2 months",
      enterprise: "2-3 months"
    }
  },
  
  {
    name: "Discovery & Planning",
    description: "Detailed requirements gathering and solution architecture",
    objectives: [
      "Define detailed requirements",
      "Design system architecture",
      "Create implementation plan",
      "Establish project governance"
    ],
    deliverables: [
      { name: "Requirements Specification", template: "spec.template.md" },
      { name: "System Architecture", template: "architecture.template.md" },
      { name: "Project Plan", template: "plan.template.md" },
      { name: "Risk Register", template: "risk-register.template.md" }
    ],
    entryGate: {
      criteria: ["Approved business case", "Assigned project team"],
      approvals: ["project_sponsor"]
    },
    exitGate: {
      criteria: [
        "Requirements approved by stakeholders",
        "Architecture reviewed and approved",
        "Implementation plan validated",
        "Budget and timeline confirmed"
      ],
      approvals: ["architecture_board", "project_steering_committee"]
    },
    duration: {
      small: "2-4 weeks",
      medium: "1-2 months",
      large: "2-3 months", 
      enterprise: "3-6 months"
    }
  },
  
  {
    name: "Development",
    description: "Iterative development and continuous testing",
    objectives: [
      "Implement solution according to specifications",
      "Conduct continuous testing and validation",
      "Manage scope and change requests",
      "Maintain quality standards"
    ],
    deliverables: [
      { name: "Working Software", template: "application.template.md" },
      { name: "Test Results", template: "test-plan.template.md" },
      { name: "Technical Documentation", template: "user-documentation.template.md" },
      { name: "Deployment Package", template: "deployment-guide.template.md" }
    ],
    entryGate: {
      criteria: ["Approved architecture", "Development environment ready"],
      approvals: ["technical_lead", "project_manager"]
    },
    exitGate: {
      criteria: [
        "All features implemented and tested",
        "Quality gates passed",
        "Security and compliance validated",
        "User acceptance testing completed"
      ],
      approvals: ["quality_assurance", "security_team", "product_owner"]
    },
    duration: {
      small: "1-3 months",
      medium: "3-6 months",
      large: "6-12 months",
      enterprise: "12-24 months"
    }
  }
];
```

### Stage Transition Management
Automated and governance-driven stage transitions with quality gates.

#### Gate Management System
```typescript
interface QualityGate {
  id: string;
  stage: string;
  type: 'entry' | 'exit';
  criteria: GateCriterion[];
  automatedChecks: AutomatedCheck[];
  manualReviews: ManualReview[];
  approvers: Approver[];
  bypassConditions?: BypassCondition[];
}

class ProjectGateManager {
  async evaluateGate(
    projectId: string,
    gateId: string,
    evidence: Evidence[]
  ): Promise<GateResult> {
    
    const gate = await this.getGate(gateId);
    const project = await this.getProject(projectId);
    
    // 1. Automated checks
    const automatedResults = await this.runAutomatedChecks(gate, project, evidence);
    
    // 2. Manual review coordination
    const manualResults = await this.coordinateManualReviews(gate, project, evidence);
    
    // 3. Approval workflow
    const approvalResults = await this.executeApprovalWorkflow(gate, project);
    
    // 4. Gate decision
    const gateDecision = this.calculateGateDecision(
      automatedResults,
      manualResults,
      approvalResults
    );
    
    // 5. Action assignment for any gaps
    if (gateDecision.status === 'conditional_pass') {
      await this.assignGapRemediationActions(gateDecision.gaps);
    }
    
    return gateDecision;
  }
  
  private async runAutomatedChecks(
    gate: QualityGate,
    project: Project,
    evidence: Evidence[]
  ): Promise<AutomatedCheckResults> {
    
    const results: CheckResult[] = [];
    
    for (const check of gate.automatedChecks) {
      switch (check.type) {
        case 'code_quality':
          results.push(await this.checkCodeQuality(project));
          break;
          
        case 'security_scan':
          results.push(await this.runSecurityScan(project));
          break;
          
        case 'performance_test':
          results.push(await this.runPerformanceTests(project));
          break;
          
        case 'compliance_check':
          results.push(await this.checkCompliance(project));
          break;
          
        case 'documentation_completeness':
          results.push(await this.checkDocumentation(project, evidence));
          break;
      }
    }
    
    return {
      overallStatus: this.calculateOverallStatus(results),
      individual: results,
      recommendations: this.generateRecommendations(results)
    };
  }
}
```

## Governance Models by Scale

### Startup Governance (1-10 people)
Lightweight governance focused on agility and rapid iteration.

#### Lean Governance Framework
```typescript
interface StartupGovernance {
  decisionMaking: {
    structure: 'flat' | 'minimal_hierarchy';
    approvalLevels: number;
    decisionSpeed: 'hours' | 'days';
  };
  documentation: {
    level: 'minimal' | 'essential';
    formats: 'informal' | 'structured';
    maintenance: 'as_needed';
  };
  processes: {
    complexity: 'simple';
    automation: 'basic';
    compliance: 'minimal';
  };
}

class StartupProjectGovernance {
  async initializeProject(idea: ProjectIdea): Promise<StartupProject> {
    // Simplified project initiation
    const project = await this.createProject({
      name: idea.name,
      description: idea.description,
      owner: idea.proposer,
      priority: this.assessPriority(idea),
      timeline: this.estimateTimeline(idea)
    });
    
    // Minimal documentation requirements
    await this.createEssentialDocs(project, [
      'readme.md',
      'requirements-summary.md',
      'architecture-sketch.md'
    ]);
    
    // Basic tracking setup
    await this.setupBasicTracking(project);
    
    return project;
  }
  
  private async setupBasicTracking(project: StartupProject): Promise<void> {
    // Simple task board
    await this.createTaskBoard(project, {
      columns: ['To Do', 'In Progress', 'Review', 'Done'],
      automation: {
        moveOnPR: true,
        moveOnMerge: true,
        closedOnDeploy: true
      }
    });
    
    // Basic metrics
    await this.enableMetrics(project, [
      'velocity',
      'cycle_time',
      'deployment_frequency',
      'bug_rate'
    ]);
  }
}
```

### Scale-up Governance (10-50 people)
Structured governance with defined processes and quality gates.

#### Scale-up Governance Framework
```typescript
interface ScaleUpGovernance {
  organizationalStructure: {
    roles: OrganizationalRole[];
    responsibilities: ResponsibilityMatrix;
    escalationPaths: EscalationPath[];
  };
  processFramework: {
    projectMethodology: 'agile' | 'hybrid' | 'waterfall';
    qualityGates: QualityGate[];
    changeManagement: ChangeProcess;
  };
  tooling: {
    projectManagement: string;
    documentation: string;
    collaboration: string;
    analytics: string;
  };
}

class ScaleUpProjectGovernance {
  async establishGovernance(organization: Organization): Promise<GovernanceFramework> {
    // 1. Define roles and responsibilities
    const roleFramework = await this.defineRoles(organization);
    
    // 2. Establish project processes
    const processFramework = await this.defineProcesses(organization);
    
    // 3. Implement quality gates
    const qualityFramework = await this.implementQualityGates(organization);
    
    // 4. Set up governance tools
    const toolingFramework = await this.setupGovernanceTools(organization);
    
    return {
      roles: roleFramework,
      processes: processFramework,
      quality: qualityFramework,
      tools: toolingFramework,
      metrics: await this.defineGovernanceMetrics(organization)
    };
  }
  
  private async defineRoles(organization: Organization): Promise<RoleFramework> {
    return {
      executiveLevel: [
        { role: 'CTO', responsibilities: ['technical_strategy', 'architecture_approval'] },
        { role: 'VPE', responsibilities: ['engineering_excellence', 'team_performance'] }
      ],
      managementLevel: [
        { role: 'Engineering Manager', responsibilities: ['team_delivery', 'people_management'] },
        { role: 'Product Manager', responsibilities: ['product_strategy', 'requirements'] },
        { role: 'Project Manager', responsibilities: ['project_coordination', 'stakeholder_communication'] }
      ],
      operationalLevel: [
        { role: 'Tech Lead', responsibilities: ['technical_decisions', 'code_reviews'] },
        { role: 'Senior Developer', responsibilities: ['implementation', 'mentoring'] },
        { role: 'QA Lead', responsibilities: ['quality_assurance', 'test_strategy'] }
      ]
    };
  }
}
```

### Enterprise Governance (50+ people)
Comprehensive governance with formal processes, compliance, and risk management.

#### Enterprise Governance Framework
```typescript
interface EnterpriseGovernance {
  governanceStructure: {
    steeringCommittee: Committee;
    architectureBoard: Committee;
    changeAdvisoryBoard: Committee;
    riskCommittee: Committee;
  };
  processMaturity: {
    level: 'defined' | 'managed' | 'optimized';
    standards: ComplianceStandard[];
    auditRequirements: AuditRequirement[];
  };
  riskManagement: {
    framework: 'ISO31000' | 'COSO' | 'custom';
    riskCategories: RiskCategory[];
    mitigation: MitigationStrategy[];
  };
}

class EnterpriseProjectGovernance {
  async establishEnterpriseGovernance(
    enterprise: Enterprise
  ): Promise<EnterpriseGovernanceFramework> {
    
    // 1. Governance structure
    const governanceStructure = await this.createGovernanceStructure(enterprise);
    
    // 2. Process standardization
    const processStandards = await this.standardizeProcesses(enterprise);
    
    // 3. Risk management framework
    const riskFramework = await this.implementRiskManagement(enterprise);
    
    // 4. Compliance and audit framework
    const complianceFramework = await this.establishCompliance(enterprise);
    
    // 5. Portfolio management
    const portfolioManagement = await this.setupPortfolioManagement(enterprise);
    
    return {
      structure: governanceStructure,
      processes: processStandards,
      risk: riskFramework,
      compliance: complianceFramework,
      portfolio: portfolioManagement
    };
  }
  
  private async createGovernanceStructure(
    enterprise: Enterprise
  ): Promise<GovernanceStructure> {
    
    return {
      steeringCommittee: {
        members: [
          { role: 'CEO', level: 'executive' },
          { role: 'CTO', level: 'executive' },
          { role: 'CFO', level: 'executive' },
          { role: 'Business Unit Heads', level: 'senior_management' }
        ],
        responsibilities: [
          'Strategic alignment',
          'Investment decisions', 
          'Resource allocation',
          'Risk oversight'
        ],
        meetingFrequency: 'monthly',
        decisionAuthority: 'final'
      },
      
      architectureBoard: {
        members: [
          { role: 'Chief Architect', level: 'executive' },
          { role: 'Domain Architects', level: 'senior_management' },
          { role: 'Security Architect', level: 'senior_management' },
          { role: 'Data Architect', level: 'senior_management' }
        ],
        responsibilities: [
          'Architecture standards',
          'Technology decisions',
          'Integration patterns',
          'Performance standards'
        ],
        meetingFrequency: 'weekly',
        decisionAuthority: 'technical'
      }
    };
  }
}
```

## Project Types and Methodologies

### Methodology Selection Framework
Systematic approach to selecting appropriate project methodologies.

#### Methodology Decision Matrix
```typescript
interface ProjectCharacteristics {
  complexity: 'low' | 'medium' | 'high' | 'very_high';
  requirements: 'stable' | 'evolving' | 'unknown';
  timeline: 'short' | 'medium' | 'long' | 'ongoing';
  team: {
    size: number;
    distribution: 'co_located' | 'distributed' | 'hybrid';
    experience: 'junior' | 'mixed' | 'senior';
  };
  stakeholders: {
    count: number;
    involvement: 'low' | 'medium' | 'high';
    availability: 'limited' | 'available' | 'dedicated';
  };
  risk: 'low' | 'medium' | 'high' | 'critical';
  compliance: 'none' | 'basic' | 'strict' | 'regulated';
}

class MethodologySelector {
  selectMethodology(characteristics: ProjectCharacteristics): ProjectMethodology {
    const score = this.calculateMethodologyScores(characteristics);
    
    return this.selectBestMethodology(score);
  }
  
  private calculateMethodologyScores(
    characteristics: ProjectCharacteristics
  ): MethodologyScores {
    
    const methodologies = ['agile', 'waterfall', 'hybrid', 'lean', 'safe'];
    const scores: Record<string, number> = {};
    
    for (const methodology of methodologies) {
      scores[methodology] = this.scoreMethodology(methodology, characteristics);
    }
    
    return scores;
  }
  
  private scoreMethodology(
    methodology: string,
    characteristics: ProjectCharacteristics
  ): number {
    let score = 0;
    
    switch (methodology) {
      case 'agile':
        // Agile scoring logic
        if (characteristics.requirements === 'evolving') score += 3;
        if (characteristics.complexity === 'medium') score += 2;
        if (characteristics.team.size <= 10) score += 2;
        if (characteristics.stakeholders.involvement === 'high') score += 2;
        break;
        
      case 'waterfall':
        // Waterfall scoring logic
        if (characteristics.requirements === 'stable') score += 3;
        if (characteristics.compliance === 'strict') score += 2;
        if (characteristics.risk === 'low') score += 2;
        break;
        
      case 'safe':
        // SAFe scoring logic
        if (characteristics.team.size > 50) score += 3;
        if (characteristics.complexity === 'very_high') score += 2;
        if (characteristics.compliance === 'regulated') score += 2;
        break;
    }
    
    return score;
  }
}
```

### Agile at Scale Implementation
Scaling agile practices from team to enterprise level.

#### SAFe Implementation Framework
```typescript
interface SAFeConfiguration {
  levels: {
    team: AgileTeam[];
    program: AgileReleaseTrains[];
    largeScolution: SolutionTrains[];
    portfolio: LeanPortfolio;
  };
  roles: SAFeRole[];
  ceremonies: SAFeCeremony[];
  artifacts: SAFeArtifact[];
}

class SAFeImplementation {
  async implementSAFe(organization: Organization): Promise<SAFeFramework> {
    // 1. Assess current state
    const currentState = await this.assessCurrentState(organization);
    
    // 2. Design target SAFe configuration
    const targetConfiguration = await this.designSAFeConfiguration(organization);
    
    // 3. Create transformation roadmap
    const transformationPlan = await this.createTransformationPlan(
      currentState,
      targetConfiguration
    );
    
    // 4. Execute transformation
    return this.executeTransformation(transformationPlan);
  }
  
  private async designSAFeConfiguration(
    organization: Organization
  ): Promise<SAFeConfiguration> {
    
    // Determine scale based on organization size
    const scale = this.determineScale(organization);
    
    switch (scale) {
      case 'essential':
        return this.createEssentialSAFe(organization);
      
      case 'large_solution':
        return this.createLargeSolutionSAFe(organization);
      
      case 'full':
        return this.createFullSAFe(organization);
      
      default:
        return this.createEssentialSAFe(organization);
    }
  }
  
  private createFullSAFe(organization: Organization): SAFeConfiguration {
    return {
      levels: {
        team: this.createAgileTeams(organization),
        program: this.createARTs(organization),
        largeScolution: this.createSolutionTrains(organization),
        portfolio: this.createLeanPortfolio(organization)
      },
      roles: [
        { name: 'Release Train Engineer', level: 'program' },
        { name: 'Product Manager', level: 'program' },
        { name: 'System Architect', level: 'program' },
        { name: 'Solution Manager', level: 'large_solution' },
        { name: 'Solution Architect', level: 'large_solution' },
        { name: 'Epic Owner', level: 'portfolio' },
        { name: 'Enterprise Architect', level: 'portfolio' }
      ],
      ceremonies: this.createSAFeCeremonies(),
      artifacts: this.createSAFeArtifacts()
    };
  }
}
```

## Resource Management and Allocation

### Dynamic Resource Allocation
Intelligent resource allocation based on project priority and capacity.

#### Resource Optimization Engine
```typescript
interface ResourceAllocation {
  projectId: string;
  resources: {
    people: PersonAllocation[];
    budget: BudgetAllocation;
    infrastructure: InfrastructureAllocation;
    tools: ToolAllocation[];
  };
  timeframe: {
    start: Date;
    end: Date;
    milestones: Milestone[];
  };
  constraints: ResourceConstraint[];
  optimization: OptimizationStrategy;
}

class ResourceAllocationEngine {
  async optimizeResourceAllocation(
    projects: Project[],
    availableResources: AvailableResources,
    constraints: GlobalConstraint[]
  ): Promise<AllocationPlan> {
    
    // 1. Project prioritization
    const prioritizedProjects = await this.prioritizeProjects(projects);
    
    // 2. Resource capacity analysis
    const capacityAnalysis = await this.analyzeCapacity(availableResources);
    
    // 3. Constraint validation
    const validatedConstraints = await this.validateConstraints(constraints);
    
    // 4. Optimization algorithm
    const optimization = await this.runOptimization(
      prioritizedProjects,
      capacityAnalysis,
      validatedConstraints
    );
    
    // 5. Allocation plan generation
    return this.generateAllocationPlan(optimization);
  }
  
  private async runOptimization(
    projects: PrioritizedProject[],
    capacity: CapacityAnalysis,
    constraints: ValidatedConstraint[]
  ): Promise<OptimizationResult> {
    
    // Multi-objective optimization
    const objectives = [
      'maximize_business_value',
      'minimize_resource_conflicts',
      'optimize_timeline',
      'balance_team_workload'
    ];
    
    // Constraint satisfaction
    const satisfiedAllocations = [];
    
    for (const project of projects) {
      const allocation = await this.allocateResources(project, capacity, constraints);
      
      if (this.validateAllocation(allocation, constraints)) {
        satisfiedAllocations.push(allocation);
        capacity = this.updateCapacity(capacity, allocation);
      } else {
        // Handle resource conflicts
        await this.handleResourceConflict(project, allocation);
      }
    }
    
    return {
      allocations: satisfiedAllocations,
      conflicts: this.identifyConflicts(satisfiedAllocations),
      recommendations: this.generateRecommendations(satisfiedAllocations)
    };
  }
}
```

### Capacity Planning and Forecasting
Predictive capacity planning for strategic resource management.

#### Capacity Forecasting Model
```typescript
interface CapacityForecast {
  timeHorizon: {
    short: CapacityPrediction;    // 1-3 months
    medium: CapacityPrediction;   // 3-12 months
    long: CapacityPrediction;     // 1-3 years
  };
  scenarios: {
    optimistic: CapacityScenario;
    realistic: CapacityScenario;
    pessimistic: CapacityScenario;
  };
  recommendations: CapacityRecommendation[];
  riskFactors: CapacityRisk[];
}

class CapacityPlanningEngine {
  async generateCapacityForecast(
    historicalData: HistoricalCapacity[],
    plannedProjects: Project[],
    businessForecasts: BusinessForecast[]
  ): Promise<CapacityForecast> {
    
    // 1. Analyze historical patterns
    const patterns = await this.analyzeHistoricalPatterns(historicalData);
    
    // 2. Model demand projections
    const demandProjections = await this.projectDemand(
      plannedProjects,
      businessForecasts,
      patterns
    );
    
    // 3. Supply capacity modeling
    const supplyProjections = await this.projectSupply(
      historicalData,
      patterns
    );
    
    // 4. Gap analysis
    const gapAnalysis = await this.analyzeCapacityGaps(
      demandProjections,
      supplyProjections
    );
    
    // 5. Scenario planning
    const scenarios = await this.generateScenarios(gapAnalysis);
    
    return {
      timeHorizon: this.createTimeHorizonView(scenarios),
      scenarios: scenarios,
      recommendations: await this.generateRecommendations(gapAnalysis),
      riskFactors: await this.identifyRisks(scenarios)
    };
  }
}
```

## Portfolio Management and Strategic Alignment

### Project Portfolio Optimization
Strategic project selection and portfolio balancing.

#### Portfolio Decision Framework
```typescript
interface PortfolioStrategy {
  strategicThemes: StrategyTheme[];
  investmentCategories: InvestmentCategory[];
  riskProfile: RiskProfile;
  resourceConstraints: ResourceConstraint[];
  timeframes: StrategicTimeframe[];
}

class PortfolioManager {
  async optimizePortfolio(
    candidateProjects: ProjectProposal[],
    strategy: PortfolioStrategy,
    currentPortfolio: Project[]
  ): Promise<OptimizedPortfolio> {
    
    // 1. Strategic alignment scoring
    const alignmentScores = await this.scoreStrategicAlignment(
      candidateProjects,
      strategy.strategicThemes
    );
    
    // 2. Value assessment
    const valueAssessments = await this.assessProjectValue(candidateProjects);
    
    // 3. Risk analysis
    const riskAnalysis = await this.analyzePortfolioRisk(
      candidateProjects,
      currentPortfolio
    );
    
    // 4. Resource optimization
    const resourceOptimization = await this.optimizeResourceUtilization(
      candidateProjects,
      strategy.resourceConstraints
    );
    
    // 5. Portfolio selection
    const selectedProjects = await this.selectOptimalPortfolio(
      candidateProjects,
      alignmentScores,
      valueAssessments,
      riskAnalysis,
      resourceOptimization
    );
    
    return {
      selectedProjects,
      portfolioMetrics: await this.calculatePortfolioMetrics(selectedProjects),
      riskProfile: await this.calculatePortfolioRisk(selectedProjects),
      resourceUtilization: await this.calculateResourceUtilization(selectedProjects)
    };
  }
}
```

## Project Lifecycle Metrics and KPIs

### Stage-Specific Metrics
Key performance indicators for each lifecycle stage.

#### Lifecycle KPI Framework
```typescript
interface LifecycleMetrics {
  ideation: {
    ideaGenerationRate: number;
    ideaConversionRate: number;
    timeToValidation: number;
    stakeholderEngagement: number;
  };
  planning: {
    planningAccuracy: number;
    requirementsStability: number;
    architectureReviewCycle: number;
    planningVelocity: number;
  };
  development: {
    deliveryVelocity: number;
    qualityMetrics: QualityMetric[];
    scopeCreep: number;
    teamProductivity: number;
  };
  deployment: {
    deploymentFrequency: number;
    deploymentSuccess: number;
    rollbackRate: number;
    timeToProduction: number;
  };
}

const LIFECYCLE_METRICS_CONFIG = {
  targets: {
    small_projects: {
      ideation_to_planning: '2 weeks',
      planning_to_development: '1 week',
      development_to_deployment: '4-8 weeks'
    },
    medium_projects: {
      ideation_to_planning: '1 month',
      planning_to_development: '2 weeks',
      development_to_deployment: '2-4 months'
    },
    large_projects: {
      ideation_to_planning: '2-3 months',
      planning_to_development: '1 month',
      development_to_deployment: '6-12 months'
    }
  },
  quality_gates: {
    code_coverage: '>= 80%',
    security_scan: 'zero_critical_issues',
    performance: 'meets_sla_requirements',
    documentation: 'complete_and_current'
  }
};
```

### Governance Effectiveness Metrics
Measuring the effectiveness of governance processes and frameworks.

#### Governance Health Dashboard
- **Decision Velocity**: Average time from proposal to decision
- **Process Compliance**: Percentage of projects following defined processes
- **Quality Gate Effectiveness**: Defect escape rate post-gate approval
- **Stakeholder Satisfaction**: Satisfaction with governance processes
- **Resource Utilization**: Efficiency of resource allocation and usage
- **Risk Management**: Percentage of risks identified and mitigated proactively
- **Portfolio Alignment**: Percentage of projects aligned with strategic objectives

---

> **Related Knowledge**:
> - [Agent Roles & Responsibilities](../business/agent-roles.md)
> - [Quality Standards](../quality/quality-standards.md)
> - [Platform Engineering](../operations/platform-engineering.md)
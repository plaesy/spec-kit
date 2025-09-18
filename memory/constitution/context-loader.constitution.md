---
version: "2.0.0"
updatedAt: "2025-09-17T11:30:00Z"
type: "constitutional-loader"
description: "Context-aware constitutional module loading system"
---

# Context-Aware Constitutional Loader

## 🎯 **Smart Loading Configuration**

### **Loading Rules Configuration**
```yaml
# .constitutional-loader.yml
constitutional_loader:
  version: "2.0.0"
  
  # Confidence Thresholds
  thresholds:
    high_confidence: 0.8    # Auto-load modules
    medium_confidence: 0.6  # Suggest to user
    low_confidence: 0.4     # Skip loading
  
  # Always Load (Mandatory)
  mandatory_modules:
    - "core-principles.constitution.md"
  
  # Context Detection Rules
  detection_rules:
    healthcare:
      keywords:
        primary: ["patient", "medical", "clinical", "healthcare", "hospital", "doctor"]
        secondary: ["hipaa", "fda", "phi", "ehr", "emr", "diagnosis", "treatment"]
        regulatory: ["medical-device", "clinical-trial", "patient-data"]
      file_patterns:
        - "*/patient*/**"
        - "*/medical*/**"
        - "*/clinical*/**"
        - "*/health*/**"
      dependencies:
        - "package.json" contains ["hl7", "fhir", "medical"]
      confidence_boost: 0.2  # Boost if multiple indicators
      
    fintech:
      keywords:
        primary: ["payment", "trading", "banking", "financial", "money", "transaction"]
        secondary: ["pci-dss", "sox", "kyc", "aml", "regulation", "compliance"]
        regulatory: ["financial-service", "trading-platform", "payment-processor"]
      file_patterns:
        - "*/payment*/**"
        - "*/trading*/**"
        - "*/financial*/**"
        - "*/banking*/**"
      dependencies:
        - "package.json" contains ["stripe", "paypal", "plaid", "alpaca"]
      confidence_boost: 0.2
      
    security:
      keywords:
        primary: ["auth", "security", "encryption", "vulnerability", "protect"]
        secondary: ["owasp", "penetration", "firewall", "ssl", "certificate", "token"]
        regulatory: ["security-audit", "penetration-test", "vulnerability-scan"]
      file_patterns:
        - "*/auth*/**"
        - "*/security*/**"
        - "*/crypto*/**"
      dependencies:
        - "package.json" contains ["bcrypt", "jsonwebtoken", "passport", "helmet"]
      base_confidence: 0.6  # Security often needed
      
    ai_ml:
      keywords:
        primary: ["ai", "ml", "model", "neural", "machine-learning", "deep-learning"]
        secondary: ["tensorflow", "pytorch", "scikit", "opencv", "nlp", "computer-vision"]
        regulatory: ["ai-model", "machine-learning", "neural-network"]
      file_patterns:
        - "*/model*/**"
        - "*/ai*/**"
        - "*/ml*/**"
        - "*/neural*/**"
      dependencies:
        - "package.json" contains ["tensorflow", "pytorch", "scikit-learn"]
        - "requirements.txt" contains ["tensorflow", "torch", "sklearn"]
      confidence_boost: 0.3
  
  # Module Dependencies (Auto-load required modules)
  dependencies:
    healthcare: ["security"]  # Healthcare always needs security
    fintech: ["security", "performance"]  # Fintech needs security + performance
    automotive: ["security", "testing"]  # Safety-critical needs security + testing
    aerospace: ["security", "testing"]   # Mission-critical needs security + testing
    ai_ml: ["testing"]  # AI/ML needs rigorous testing
    web3: ["security"]  # Blockchain needs security
  
  # Exclusion Rules (Skip modules for certain contexts)
  exclusions:
    prototype_project: 
      skip: ["compliance-tools", "international"]
      reason: "Prototypes don't need heavy compliance"
    
    personal_project:
      skip: ["fintech", "healthcare", "automotive", "aerospace"]
      reason: "Personal projects rarely need regulated industry standards"
    
    maintenance_task:
      skip: "all_domain_specific"
      reason: "Maintenance tasks use existing architecture"
```

### **Smart Detection Algorithm**
```typescript
// constitutional-loader.ts
interface ContextAnalysis {
  domain: string;
  confidence: number;
  indicators: string[];
  reasoning: string;
}

class ConstitutionalLoader {
  private thresholds = {
    HIGH: 0.8,
    MEDIUM: 0.6,
    LOW: 0.4
  };

  async analyzeProjectContext(projectPath: string, conversation: string): Promise<ContextAnalysis[]> {
    const analyses: ContextAnalysis[] = [];
    
    // 1. Conversation Analysis
    const conversationScores = await this.analyzeConversation(conversation);
    
    // 2. File Structure Analysis
    const fileScores = await this.analyzeFileStructure(projectPath);
    
    // 3. Dependency Analysis
    const depScores = await this.analyzeDependencies(projectPath);
    
    // 4. Combine and calculate confidence
    for (const domain of this.getSupportedDomains()) {
      const confidence = this.calculateConfidence(
        conversationScores[domain] || 0,
        fileScores[domain] || 0,
        depScores[domain] || 0
      );
      
      if (confidence > this.thresholds.LOW) {
        analyses.push({
          domain,
          confidence,
          indicators: this.getIndicators(domain, conversationScores, fileScores, depScores),
          reasoning: this.generateReasoning(domain, confidence)
        });
      }
    }
    
    return analyses.sort((a, b) => b.confidence - a.confidence);
  }

  async loadConstitutionalModules(analyses: ContextAnalysis[]): Promise<string[]> {
    const modules = ['core-principles.constitution.md']; // Always start with core
    
    for (const analysis of analyses) {
      if (analysis.confidence >= this.thresholds.HIGH) {
        // High confidence: Auto-load
        modules.push(`${analysis.domain}.constitution.md`);
        
        // Load dependencies
        const deps = this.getDependencies(analysis.domain);
        modules.push(...deps.map(dep => `${dep}.constitution.md`));
        
        console.log(`✅ Auto-loaded ${analysis.domain} (confidence: ${analysis.confidence})`);
      } else if (analysis.confidence >= this.thresholds.MEDIUM) {
        // Medium confidence: Suggest to user
        console.log(`💡 Suggested: ${analysis.domain} (confidence: ${analysis.confidence})`);
      }
    }
    
    return [...new Set(modules)]; // Remove duplicates
  }

  private calculateConfidence(conversationScore: number, fileScore: number, depScore: number): number {
    // Weighted combination
    const weights = { conversation: 0.4, files: 0.3, dependencies: 0.3 };
    
    return (
      conversationScore * weights.conversation +
      fileScore * weights.files +
      depScore * weights.dependencies
    );
  }
}
```

### **Usage Examples**

#### **Example 1: Healthcare AI Project**
```typescript
// Input context
const projectContext = {
  conversation: "Build an AI system for patient diagnosis using medical images",
  projectFiles: [
    "src/models/patient.ts",
    "src/services/diagnosis.service.ts",
    "package.json" // contains tensorflow, medical dependencies
  ]
};

// Analysis result
const contextAnalysis = [
  { domain: "healthcare", confidence: 0.95, indicators: ["patient", "diagnosis", "medical"] },
  { domain: "ai_ml", confidence: 0.90, indicators: ["AI system", "tensorflow"] },
  { domain: "security", confidence: 0.85, indicators: ["patient data", "healthcare dependency"] }
];

// Loaded modules
const loadedModules = [
  "core-principles.constitution.md",     // Mandatory
  "security.constitution.md",            // Healthcare dependency
  "healthcare.constitution.md",          // High confidence (0.95)
  "ai-ml.constitution.md",              // High confidence (0.90)
  "testing.constitution.md"             // AI/ML dependency
];

console.log("📋 Loaded 5 modules (skipped 10+ irrelevant modules)");
```

#### **Example 2: Simple Bug Fix**
```typescript
// Input context
const projectContext = {
  conversation: "Fix the login button that's not working",
  projectFiles: ["src/components/LoginButton.tsx"]
};

// Analysis result
const contextAnalysis = [
  // No domain confidence > 0.4 threshold
];

// Loaded modules
const loadedModules = [
  "core-principles.constitution.md"     // Mandatory only
];

console.log("📋 Minimal loading: 1 module (optimal for maintenance task)");
```

#### **Example 3: React E-commerce Site**
```typescript
// Input context
const projectContext = {
  conversation: "Create an e-commerce website with React and payment processing",
  projectFiles: [
    "package.json", // contains react, stripe
    "src/components/ProductList.tsx",
    "src/services/payment.service.ts"
  ]
};

// Analysis result
const contextAnalysis = [
  { domain: "web_frontend", confidence: 0.90, indicators: ["React", "components"] },
  { domain: "ecommerce", confidence: 0.85, indicators: ["e-commerce", "ProductList"] },
  { domain: "fintech", confidence: 0.75, indicators: ["payment processing", "stripe"] },
  { domain: "security", confidence: 0.80, indicators: ["payment", "fintech dependency"] }
];

// Loaded modules
const loadedModules = [
  "core-principles.constitution.md",     // Mandatory
  "security.constitution.md",            // Payment security needed
  "web-frontend.constitution.md",        // React detected
  "ecommerce.constitution.md",          // E-commerce context
  "performance.constitution.md"         // Fintech dependency (if needed)
];

console.log("📋 Loaded 5 modules (skipped healthcare, automotive, AI/ML, etc.)");
```

## 🔧 **Configuration Tips**

### **Fine-tuning Detection**
```yaml
# Adjust thresholds based on your needs
custom_thresholds:
  strict_loading: 0.9    # Load only with very high confidence
  balanced_loading: 0.7  # Good balance of precision/recall
  aggressive_loading: 0.5 # Load more modules, might have false positives
```

### **Project-Specific Overrides**
```yaml
# .constitutional-override.yml
project_overrides:
  always_include: ["security"]  # Always load security for this project
  never_include: ["quantum-ready"]  # Skip experimental modules
  force_include: ["testing"]    # Force testing even if not detected
```
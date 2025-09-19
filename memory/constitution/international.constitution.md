---
version: "2.0.0"
updatedAt: "2025-09-17T11:00:00Z"
type: "constitutional-international"
description: "Global regulatory frameworks and multi-regional compliance"
---

# Global Regulatory Framework Support

## 🌍 **Regional Constitutional Variants**

### **European Union (EU)**
```yaml
eu_regulatory_framework:
  base_constitutions: ["core-principles", "security"]
  
  mandatory_regulations:
    gdpr:
      constitution: "gdpr.constitution.md"
      triggers: ["personal_data", "privacy", "consent", "eu_citizens"]
      requirements:
        - data_protection_by_design
        - right_to_be_forgotten
        - data_portability
        - privacy_impact_assessment
    
    pci_dss_eu:
      constitution: "pci-dss-eu.constitution.md"
      triggers: ["payment", "card_data", "eu_payments"]
      regional_specifics:
        - strong_customer_authentication
        - payment_services_directive_2
        - open_banking_compliance
```

### **United States (US)**
```yaml
us_regulatory_framework:
  base_constitutions: ["core-principles", "security"]
  
  federal_regulations:
    hipaa:
      constitution: "hipaa-us.constitution.md"
      triggers: ["healthcare", "phi", "covered_entity"]
      state_variations:
        california: ["ccpa", "cipa"]
        texas: ["texas_medical_privacy"]
        new_york: ["shield_act"]
    
    sox_compliance:
      constitution: "sox-us.constitution.md"
      triggers: ["public_company", "financial_reporting", "audit"]
      requirements:
        - internal_controls
        - financial_transparency
        - audit_independence
```

### **Asia-Pacific (APAC)**
```yaml
apac_regulatory_framework:
  
  china:
    cybersecurity_law:
      constitution: "csl-china.constitution.md"
      triggers: ["china_market", "critical_infrastructure", "personal_information"]
      requirements:
        - data_localization
        - security_assessments
        - government_approval
  
  japan:
    personal_data_protection:
      constitution: "appi-japan.constitution.md"
      triggers: ["personal_data", "japan_citizens"]
      requirements:
        - data_protection_commission_compliance
        - privacy_policy_transparency
  
  singapore:
    personal_data_protection:
      constitution: "pdpa-singapore.constitution.md"
      triggers: ["personal_data", "singapore_operations"]
      requirements:
        - consent_management
        - data_breach_notification
        - do_not_call_registry
```

## 🏛️ **Multi-Regional Compliance Matrix**

### **Data Protection Standards**
```yaml
compliance_requirements:
  
  data_protection:
    gdpr: # EU
      data_subjects: "eu_residents"
      consent_requirement: "explicit"
      data_retention: "limited_purpose"
      cross_border_transfer: "adequacy_decision_required"
    
    ccpa: # California, US
      data_subjects: "california_residents"
      consent_requirement: "opt_out"
      data_retention: "business_purpose"
      cross_border_transfer: "no_specific_restrictions"
    
    pipl: # China
      data_subjects: "china_residents"
      consent_requirement: "explicit_separate"
      data_retention: "minimization_principle"
      cross_border_transfer: "security_assessment_required"
```

### **Financial Services Global Standards**
```yaml
global_financial_regulations:
  basel_iii:
    regions: ["global"]
    constitution: "basel-iii.constitution.md"
    triggers: ["banking", "capital_requirements", "risk_management"]
    regional_implementations:
      eu: "crd_iv_regulation"
      us: "dodd_frank_compliance"
      uk: "pra_rulebook"
  
  mifid_ii:
    regions: ["eu", "uk"]
    constitution: "mifid-ii.constitution.md"
    triggers: ["investment_services", "financial_instruments"]
    requirements:
      - best_execution
      - transaction_reporting
      - investor_protection
```

## 🔄 **Constitutional Localization Engine**

### **Regional Detection Algorithm**
```typescript
interface RegionalContext {
  country: string;
  region: string;
  industry: string;
  operationalScope: string[];
  regulatoryRequirements: string[];
}

class ConstitutionalLocalizationEngine {
  
  async detectRegionalRequirements(
    projectContext: ProjectContext
  ): Promise<RegionalContext> {
    
    const regional: RegionalContext = {
      country: await this.detectCountry(projectContext),
      region: await this.detectRegion(projectContext),
      industry: await this.detectIndustry(projectContext),
      operationalScope: await this.detectOperationalScope(projectContext),
      regulatoryRequirements: []
    };
    
    // Load regional regulatory mappings
    const regulatoryMap = await this.loadRegulatoryMappings();
    
    // Determine applicable regulations
    regional.regulatoryRequirements = this.mapRegulationsToContext(
      regional,
      regulatoryMap
    );
    
    return regional;
  }
}
```
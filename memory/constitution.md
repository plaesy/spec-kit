---
version: "3.0.0"
updatedAt: "2025-09-17T11:45:00Z"
type: "constitutional-simple"
description: "Ultra-simple constitutional framework - load what you need, when you need it"
---

# Simple Constitutional Framework

## 🎯 **Core Concept**
Load **only** what you need, **when** you need it. No complex algorithms, no over-engineering.

## 📋 **How It Works**

### **1. Always Start Here**
```
core-principles.constitution.md (MANDATORY)
```

### **2. Add Based on Keywords**
| If you mention... | Load this |
|-------------------|-----------|
| `patient`, `medical`, `healthcare` | `healthcare.constitution.md` |
| `payment`, `trading`, `financial` | `fintech.constitution.md` |
| `security`, `auth`, `encryption` | `security.constitution.md` |
| `test`, `testing`, `tdd` | `testing.constitution.md` |
| `performance`, `scalability` | `performance.constitution.md` |
| `react`, `angular`, `frontend` | `web-frontend.constitution.md` |
| `api`, `backend`, `microservice` | `backend-api.constitution.md` |
| `mobile`, `ios`, `android` | `mobile.constitution.md` |
| `ai`, `ml`, `model` | `ai-ml.constitution.md` |
| `blockchain`, `web3`, `smart-contract` | `web3-blockchain.constitution.md` |

### **3. Auto-Dependencies**
- Healthcare → also load Security
- Fintech → also load Security + Performance  
- AI/ML → also load Testing

That's it! Simple.

## 🚀 **Usage Examples**

### Example 1: "Build patient management system"
**Detected**: `patient` → healthcare  
**Loaded**: core-principles + security + healthcare  
**Result**: 3 files, perfectly relevant

### Example 2: "Fix React component bug"  
**Detected**: `react` → web-frontend  
**Loaded**: core-principles + web-frontend  
**Result**: 2 files, minimal and focused

### Example 3: "Add authentication"
**Detected**: `authentication` → security  
**Loaded**: core-principles + security  
**Result**: 2 files, exactly what's needed

### Example 4: "General code review"
**Detected**: No specific keywords  
**Loaded**: core-principles only  
**Result**: 1 file, ultra-minimal

## 🛠️ **Available Modules**

### **Core (Always Required)**
- `core-principles.constitution.md` - TDD, semantic versioning, interface design

### **Industry Modules (Load on Demand)**
- `healthcare.constitution.md` - HIPAA, FDA, medical standards
- `fintech.constitution.md` - PCI DSS, SOX, financial compliance
- `security.constitution.md` - OWASP, encryption, auth standards

### **Technology Modules (Load on Demand)**
- `testing.constitution.md` - Testing standards and practices
- `performance.constitution.md` - Performance and scalability
- `web-frontend.constitution.md` - Frontend development standards
- `backend-api.constitution.md` - Backend and API standards
- `mobile.constitution.md` - Mobile development standards
- `ai-ml.constitution.md` - AI/ML development standards
- `web3-blockchain.constitution.md` - Blockchain and Web3 standards

### **Reference Modules (Optional)**
- `examples.constitution.md` - Implementation examples
- `compliance-tools.constitution.md` - Automated tools
- `international.constitution.md` - Global regulations

## 🎯 **Keep It Simple Rules**

1. **Default**: Load core-principles only
2. **Keyword Match**: Load relevant module
3. **Auto-Dependency**: Load required dependencies
4. **No Over-Engineering**: Skip complex detection algorithms
5. **Fast Loading**: Maximum 3-5 modules per session

---

**Version**: 3.0.0 - Ultra Simple  
**Philosophy**: Simple > Complex  
**Last Updated**: 2025-09-17
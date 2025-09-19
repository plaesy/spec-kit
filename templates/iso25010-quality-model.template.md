# ISO/IEC 25010 Quality Model Mapping

Version 0.1
Prepared by <author>
<organization>
<date created>

Purpose: Map Non-Functional Requirements (NFRs) to ISO/IEC 25010 quality characteristics and measurable sub-characteristics.

## Characteristics and Measures
| Characteristic | Sub-Characteristic | Related NFR IDs | Measures / Targets |
|----------------|--------------------|------------------|--------------------|
| Functional suitability | Functional completeness, correctness, appropriateness | FR-*, NFR-U-* | Coverage, defect rates |
| Performance efficiency | Time behavior, resource utilization, capacity | NFR-P-* | p95 latency, CPU/mem, TPS |
| Compatibility | Co-existence, interoperability | NFR-COMP-* | API conformance, protocol support |
| Usability | Learnability, operability, accessibility, UI aesthetics | NFR-U-* | SUS score, WCAG AA |
| Reliability | Maturity, availability, fault tolerance, recoverability | NFR-R-* / NFR-A-* | MTBF, MTTR, uptime |
| Security | Confidentiality, integrity, non-repudiation, accountability, authenticity | NFR-S-* | OWASP pass, encryption |
| Maintainability | Modularity, reusability, analyzability, modifiability, testability | NFR-M-* | Change lead time, coverage |
| Portability | Adaptability, installability, replaceability | NFR-PORT-* | Supported platforms |

## Traceability
- Link NFRs to ISO characteristics in RTM

## Evidence
- Define how each measure is collected and reported (APM, logs, SAST/DAST)

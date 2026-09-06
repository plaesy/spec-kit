# AI-Specific Headers

Direktori ini berisi header yang spesifik untuk setiap AI provider yang berbeda. Header ini akan diinjeksi otomatis ke dalam file prompts, chatmodes, dan instructions berdasarkan AI provider yang dipilih saat inisialisasi.

## Chatmode Integration Support

Semua header templates mendukung integrasi chatmodes untuk role-based development workflows:

### Supported Chatmodes
- `@ba` - Business Analyst (requirements, stakeholder analysis, user stories)
- `@dev` - Developer (implementation, TDD, code architecture)
- `@sa` - Solution Architect (system design, technology decisions)
- `@pm` - Product Manager (strategy, feature prioritization, roadmap)
- `@po` - Product Owner (customer value, business case, strategic alignment)
- `@qa` - QA Engineer (testing strategy, quality assurance, validation)
- `@devops` - DevOps Engineer (deployment, infrastructure, CI/CD)
- `@security` - Security Expert (threat modeling, OWASP, compliance)

### Chatmode Syntax
```bash
# Pattern 1: @ syntax (recommended)
@{chatmode} {task_description}

# Pattern 2: --mode syntax (alternative)
--mode={chatmode} {task_description}

# Examples:
@ba user authentication requirements
@dev implement REST API for user management
@sa design microservices architecture
@pm plan mobile app feature roadmap
```

## Struktur File Header

### Naming Convention
- `{provider}.header.yaml` - Header umum untuk provider
- `{provider}.{type}.yaml` - Header spesifik untuk tipe file

### Contoh:
- `claude.header.yaml` - Header umum untuk Claude
- `claude.prompts.yaml` - Header khusus untuk prompts Claude
- `claude.chatmodes.yaml` - Header khusus untuk chatmodes Claude
- `copilot.prompts.yaml` - Header khusus untuk prompts Copilot

## Format Header YAML

Setiap header file menggunakan format YAML dengan struktur yang terpisah antara metadata dan konten:

```yaml
platform: claude
type: prompts
version: "1.0"
description: "Claude-specific prompt headers for optimal performance"

# Header content that will be injected ({{DESCRIPTION}} will be replaced with file's description)
header_content: |
  ---
  platform: claude
  type: prompt
  version: "1.0"
  description: {{DESCRIPTION}}
  ---

  # Claude Code Prompt Configuration

  This prompt is optimized for Claude Code and follows the Plaesy Constitutional Framework.

  ## Claude-Specific Instructions
  - Use precise, technical language
  - Follow structured markdown formatting
  - Leverage Claude's reasoning capabilities
  - Implement constitutional development principles
```

### Struktur YAML:
- **Metadata Section**: `platform`, `type`, `version`, `description`
- **Content Section**: `header_content` dengan format multiline YAML (`|`)
- **Dynamic Placeholders**: `{{DESCRIPTION}}` akan diganti dengan description dari file target

### Dynamic Description Replacement

Sistem secara otomatis mengekstrak description dari file target dan mengganti placeholder `{{DESCRIPTION}}` dalam header content. Ini memungkinkan setiap file mempertahankan description uniknya sambil tetap mendapat header yang sesuai AI provider.

**Contoh**:
- File `security.chatmode.md` memiliki description: `"Chat mode for Security Engineers — threat modeling, OWASP, and security-by-design."`
- Header template memiliki: `description: {{DESCRIPTION}}`
- Hasil akhir: `description: Chat mode for Security Engineers — threat modeling, OWASP, and security-by-design.`

**Fallback untuk File Tanpa Description**:
- Jika file tidak memiliki description di YAML front-matter, sistem akan generate default berdasarkan nama file
- `security.chatmode.md` → `"Chat mode: security"`
- `idea.prompt.md` → `"Prompt: idea"`

## Cara Kerja

1. **Saat Inisialisasi** (`plaesy-init.sh`):
   - User memilih AI provider
   - Script menyalin framework files ke `.plaesy/`
   - Script `inject-ai-headers.sh` dipanggil dengan `--merge` flag

2. **Header Selection Logic**:
   - Untuk prompts: `{provider}.prompts.yaml` → `{provider}.header.yaml` → `manual.header.yaml`
   - Untuk chatmodes: `{provider}.chatmodes.yaml` → `{provider}.header.yaml` → `manual.header.yaml`
   - Untuk instructions: `{provider}.instructions.yaml` → `{provider}.header.yaml` → `manual.header.yaml`

3. **Dynamic Description Processing**:
   - Script mengekstrak description dari file target
   - Mengganti `{{DESCRIPTION}}` dalam header template dengan description sebenarnya
   - Menginjeksi header yang sudah dipersonalisasi ke file target

4. **Merge Behavior**:
   - Jika header file kosong, diabaikan
   - Jika file target sudah memiliki YAML front-matter, hanya key yang hilang yang ditambahkan
   - Jika file target belum memiliki header, header lengkap ditambahkan di awal
   - Description original dari file target selalu dipertahankan

## AI Providers yang Didukung

- `claude` - Claude Code
- `copilot` - GitHub Copilot
- `cursor` - Cursor AI
- `windsurf` - Windsurf AI
- `chatgpt` - ChatGPT
- `gemini` - Google Gemini
- `trae-ai` - Trae.ai Multi-Agent
- `qwen-code` - Qwen Code
- `codex-cli` - Codex CLI
- `opencode-cli` - OpenCode CLI
- `local-ai` - Local AI (Ollama, LM Studio)
- `manual` - Manual Development

## Contoh Penggunaan

```bash
# Inject headers untuk Claude ke direktori prompts
./scripts/bash/inject-ai-headers.sh --ai claude --target prompts --merge --backup

# Dry-run untuk melihat apa yang akan diubah
./scripts/bash/inject-ai-headers.sh --ai copilot --target chatmodes --dry-run

# List mapping file → header
./scripts/bash/inject-ai-headers.sh --ai cursor --target . --list-only

# Force inject dengan description replacement
./scripts/bash/inject-ai-headers.sh --ai claude --target prompts --force --backup
```

## Membuat Header Baru

1. Buat file dengan naming convention yang benar: `{provider}.{type}.yaml`
2. Tambahkan YAML metadata dengan platform, type, version, dan description
3. Tambahkan `header_content` dengan placeholder `{{DESCRIPTION}}`
4. Test dengan `--dry-run` sebelum apply

### Template Header Baru:
```yaml
platform: {your_provider}
type: {prompts|chatmodes|instructions}
version: "1.0"
description: "Description for this header type"

# Header content that will be injected ({{DESCRIPTION}} will be replaced with file's description)
header_content: |
  ---
  platform: {your_provider}
  type: {type}
  version: "1.0"
  description: {{DESCRIPTION}}
  ---

  # {Your Provider} Configuration

  Provider-specific instructions here.

  ## Provider-Specific Instructions
  - Instruction 1
  - Instruction 2
```

## Tips

- Gunakan `{{DESCRIPTION}}` placeholder untuk mempertahankan description original dari file
- Gunakan `--merge` untuk menggabungkan header tanpa menimpa konten yang sudah ada
- Gunakan `--backup` untuk membuat backup file original
- Gunakan `--dry-run` untuk preview perubahan sebelum apply
- File header kosong akan diabaikan otomatis
- Setiap file mempertahankan description uniknya secara otomatis
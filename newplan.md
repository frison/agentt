# Gydnc Development Plan: Content-Addressable Guidance System

**Version:** 0.1-MVP
**Status:** Active Development

## 1. Vision & Goals

### 1.1 Core Vision
Establish a foundational guidance management system (`gydnc`) that is:
- Easy to use and maintain
- Version-controlled (initially via Git)
- Content-addressable (CID-based)
- Multi-backend capable
- Robust and testable

### 1.2 MVP Goals
- Support both read-only and writable backends
- Ensure entities are always aware of their source backend
- Lay groundwork for conflict resolution (CID, pCID)
- Move business logic into a service layer for CLI/web reuse
- Architect filter logic for reuse across commands and webservice
- Improve testability and maintainability

## 2. Core Concepts

### 2.1 Guidance Entity: The ".g6e" File
- **Format:** Human-readable files with `.g6e` extension
- **Structure:**
  - **YAML Frontmatter:**
    - `title`: (String, Mandatory) Human-readable concise title
    - `description`: (String, Recommended) Brief description
    - `tags`: (Array of Strings) Keywords for categorization
  - **Markdown Body:** Main instructive content

### 2.2 Backend Architecture
- **Interface:** Core `storage.Backend` interface with read/write operations
- **Types:**
  - `ReadOnlyBackend`: For read-only operations
  - `Backend`: Full read/write capabilities
- **MVP Backend:** `localfs` with Git integration
- **Future:** Support for webservice, S3, etc.

### 2.3 Content Addressing
- **CID (Content ID):** Hash-based identifier for content integrity
- **pCID:** Previous CID for chained provenance
- **MVP:** Basic SHA256 hashing for verification
- **Future:** Full CID-based addressing

### 2.4 Configuration
- **File:** `gydnc.conf` (YAML format)
- **Loading Priority:**
  1. `--config` CLI argument
  2. `GYDNC_CONFIG` environment variable
  3. Default settings (where applicable)

## 3. Implementation Phases

### Phase 0: Project Setup & Core Interfaces
- [x] Initialize Go module and directory structure
- [x] Define core interfaces and configuration loading
- [x] Implement basic CLI commands (`version`)
- [ ] Create service layer foundation

### Phase 1: Backend Interface Refactor
- [ ] Split `storage.Backend` into `ReadOnlyBackend` and `Backend`
- [ ] Add `IsWritable()` or `Capabilities()` to interface
- [ ] Update localfs and other backends
- [ ] Create backend factory and registry

### Phase 2: Entity Model & Service Layer
- [x] Move `GuidanceManifestItem` to shared package
- [ ] Ensure entities include backend info
- [ ] Add CID and pCID fields
- [ ] Extract business logic to service layer
- [ ] Implement filter logic module

### Phase 3: CLI & Integration
- [ ] Refactor CLI commands to use service layer
- [ ] Implement `gydnc serve` command
- [ ] Add comprehensive test coverage
- [ ] Document all interfaces and patterns

## 4. CLI Commands (MVP)

### Core Commands
- `gydnc version`: Display version info
- `gydnc init`: Initialize guidance directory
- `gydnc list`: List available guidance
- `gydnc get`: Retrieve guidance content
- `gydnc create`: Create new guidance
- `gydnc update`: Modify existing guidance
- `gydnc hash`: Calculate content hash
- `gydnc list-tags`: List unique tags
- `gydnc config`: Manage configuration

### Command Details
Each command supports:
- `--backend <name>` for backend selection
- `--config <path>` for config file
- Appropriate filtering and output options

## 5. Testing & Validation

### Test Strategy
- Unit tests for all interfaces and services
- Integration tests using `gydnc_cli_harness_test.go`
- End-to-end tests for common workflows
- Performance benchmarks for critical paths

### Validation Steps
- Verify all commands work with `localfs` backend
- Test Git integration
- Validate content addressing
- Check filter functionality
- Ensure proper error handling

## 6. Next Actions

1. **Immediate:**
   - Complete backend interface refactor
   - Implement service layer
   - Add filter logic module

2. **Short-term:**
   - Update CLI commands to use service layer
   - Add comprehensive tests
   - Document interfaces and patterns

3. **Medium-term:**
   - Implement `gydnc serve`
   - Add support for additional backends
   - Enhance content addressing

## Notes
- Prioritize minimal, incremental changes
- Keep tests passing throughout
- Document all interface changes
- Run `make build` and `make test-integration` after each major step
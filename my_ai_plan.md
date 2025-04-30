# Plan: Refactor Agent Guidance Service API for Efficiency

## Goal
Reduce the number of initial API calls and data transfer required by the agent interaction framework, based on user feedback regarding token usage and manual approvals.

## Current Process (per llm.txt)
1. Fetch `/discover/behavior?tier=must` (full details)
2. Fetch `/discover/behavior?tier=should` (full details)
3. Fetch `/discover/recipe` (full details)

## Proposed Process
1. Fetch a new `/summary` endpoint containing minimal info (ID, type, tags, description) for *all* behaviors and recipes.
2. Fetch a new `/details` endpoint, providing a list of relevant IDs identified from the summary, to get *only* the necessary full details.

## Development Approach & Safeguards

*   **Branching:** All work will be done on a dedicated feature branch.
*   **Incremental Commits:** Commits will be made frequently, ideally after each logical step within a phase (e.g., after implementing a specific endpoint, after adding a CLI command). Each commit message should clearly state the change.
*   **Commits Between Phases:** A distinct commit will mark the completion of each phase's implementation *before* its testing steps.
*   **Testing:** Each phase includes specific testing steps. These must pass before moving to the next phase.
*   **Rollback Strategy:** If testing reveals significant issues within a phase that are not easily fixed, revert the branch to the commit marking the end of the previous phase and re-evaluate the problematic step(s).

## High-Level Plan

### Phase 0: Project Renaming (`.agent/service`)
1.  **Rename Go Module:** Update the module path in `go.mod` to reflect the `agentt` name (if applicable).
2.  **Update Import Paths:** Modify all internal import paths within the `.agent/service` project to use the new module path.
3.  **Update Build Scripts:** Adjust any Makefiles, Dockerfiles, or CI/CD scripts to use `agentt` as the target binary name.
4.  **Verify Build & Tests:** Build the renamed project and run all existing tests to ensure the renaming was successful.

### Phase 1: API Design & Implementation (`.agent/service`)
1.  **Integrate Cobra Framework:**
    *   Add `cobra` as a dependency.
    *   Refactor the application entry point (`main.go`) to use `cobra`.
    *   Define a root command (`agentt`) and an initial `server` subcommand (e.g., `agentt server start`) to launch the HTTP server.
    *   Consider adding placeholder commands for potential future CLI utilities (e.g., `validate`, `list`).
2.  **Define Data Structures:**
    *   Define a JSON structure for item summaries (e.g., `{ id: string, type: 'behavior'|'recipe', tier?: 'must'|'should', tags: string[], description: string }`).
    *   Confirm the existing structure for full item details remains suitable.
3.  **Implement `/summary` Endpoint (within server logic):**
    *   Create a new route (e.g., `GET /summary`).
    *   This endpoint will load all valid behaviors and recipes.
    *   It will map each item to the summary structure defined above.
    *   It will return a JSON array of these summaries.
4.  **Implement `/details` Endpoint (within server logic):**
    *   Create a new route (e.g., `POST /details`).
    *   This endpoint will expect a JSON request body containing an array of item IDs (e.g., `{ "ids": ["behavior_id_1", "recipe_id_2"] }`).
    *   It will load the full details for the requested item IDs.
    *   It will return a JSON array containing the full details of the found items.
5.  **Phase 1 Testing:**
    *   Update and run unit/integration tests for the new API endpoints and Cobra structure.
    *   Build the `agentt` binary.
    *   Execute `agentt --help` and verify the output.
    *   Execute `agentt server --help` and verify the output.

### Phase 2: CLI API and Help Command (`.agent/service`)
1.  **Design CLI API:**
    *   Define subcommands and flags for the `agentt` CLI to mirror the functionality of the `/summary` and `/details` HTTP endpoints.
    *   Example: `agentt summary` (outputs summary JSON), `agentt details --id behavior_id_1 --id recipe_id_2` (outputs details JSON for specified IDs).
2.  **Implement CLI API Logic:**
    *   Implement the Cobra commands defined above.
    *   Reuse entity loading and processing logic from the HTTP server implementation where possible.
3.  **Implement `llm` Command:**
    *   Create a new subcommand: `agentt llm`.
    *   Populate its execution logic to print a helpful explanation of how to use the `agentt` CLI, particularly the `summary` and `details` commands (or their equivalents), for interaction similar to the HTTP API.
4.  **Phase 2 Testing:**
    *   Add/update unit tests for the new CLI command logic.
    *   Build the `agentt` binary.
    *   Execute `agentt summary --help` and verify output.
    *   Execute `agentt details --help` and verify output.
    *   Execute `agentt llm` and verify the help text.
    *   Run basic functionality tests for `agentt summary` and `agentt details`.

### Phase 3: Client/Consumer Adaptation & `llm.txt` Update
1.  **Update `llm.txt` Generation:**
    *   Modify the logic that generates `llm.txt` (or the interpretation process).
    *   **Step 1 (Discovery):** Change the instruction to call `GET /summary` instead of the three separate `discover` calls.
    *   **Step 2 (Fetch Details):** Add instructions to analyze the summary, identify relevant items based on the task, and then call `POST /details` with the list of relevant IDs.
    *   **Step 3 (Proceed):** Adjust instructions to use the details fetched in the previous step.
2.  **Adapt Client Logic:** Ensure any system interpreting `llm.txt` correctly follows the new two-step fetch process (summary then details).
3.  **Phase 3 Testing:**
    *   Run tests related to `llm.txt` generation/parsing.
    *   Perform end-to-end tests simulating an agent following the new `llm.txt` protocol with the updated service.

### Phase 4: Refinement & Cleanup
1.  **Final Testing:** Thoroughly test the complete workflow, ensuring agents correctly fetch summaries, request relevant details, and adhere to the retrieved guidance.
2.  **Deprecation:** Once the new `/summary` and `/details` endpoints are stable and `llm.txt` reflects the new process, mark the original `/discover/{entityType}` endpoints (especially with `?tier=`) as deprecated or remove them if feasible.
3.  **Documentation:** Update any `README` files or other documentation within `.agent/service` to reflect the new API structure and recommended usage pattern (`agentt` commands, endpoints).

### Phase 5: ID Prefix Implementation (`.agent/service`)
1.  **Update ID Generation:** Modify the core entity loading logic to generate IDs with type prefixes (e.g., `bhv-` for behaviors, `rcp-` for recipes). Ensure IDs remain unique within their type.
2.  **Update `/summary` Endpoint & `agentt summary` Command:** Confirm these outputs now include the prefixed IDs in the JSON response.
3.  **Update `/details` Endpoint & `agentt details` Command:** Ensure these correctly handle and expect prefixed IDs as input (both in the JSON body for POST and via the `--id` flag for the CLI). The internal logic should strip the prefix before looking up the entity if necessary, or the lookup mechanism should handle the prefix directly.
4.  **Update `llm.txt`:** Explicitly state that IDs returned by `/summary` (and used in `/details`) will have `bhv-` or `rcp-` prefixes.
5.  **Update Tests:** Modify unit, integration, and end-to-end tests to reflect the use of prefixed IDs in API responses and CLI arguments/outputs.
6.  **Phase 5 Testing:**
    *   Build the `agentt` binary.
    *   Run all tests (unit, integration, e2e) confirming prefixed IDs are handled correctly.
    *   Manually verify the output of `agentt summary` and the input required by `agentt details --id ...`.
    *   Inspect generated `llm.txt` content.

### Phase 6: Improve Configuration Handling & Documentation
1.  **Refactor Config Loading:** Modify config loading logic (e.g., in `internal/config` or a new helper) to support prioritized configuration sources:
    *   `--config` flag (highest priority).
    *   `AGENTT_CONFIG` environment variable.
    *   Default search paths relative to CWD (e.g., `./config.yaml`, `./.agent/service/config.yaml`) (lowest priority).
2.  **Update Commands:**
    *   Add consistent `--config` flag to `summary` and `details` commands.
    *   Update `server start`, `summary`, and `details` command execution logic to use the new prioritized config loading mechanism.
3.  **Update Documentation:**
    *   Update command help text (`Short`/`Long` descriptions for `agentt`, `server start`, `summary`, `details`) to mention configuration options and precedence.
    *   Create or update a `README.md` file in `.agent/service/` with a clear "Configuration" section.
4.  **Testing:**
    *   Build the `agentt` binary.
    *   Test configuration using the flag.
    *   Test configuration using the `AGENTT_CONFIG` environment variable.
    *   Test configuration using default search paths (placing `config.yaml` in different locations relative to CWD).
    *   Test precedence (e.g., ensure flag overrides environment variable).

### Phase 7: Testing & Refinement
1.  **Server Endpoint Testing:**
    *   Start the server (`agentt server start`).
    *   Verify `curl localhost:8080/health` returns `OK`.
    *   Verify `curl localhost:8080/summary` returns valid JSON summary with prefixed IDs (`bhv-`, `rcp-`).
    *   Verify `curl -X POST -H "Content-Type: application/json" -d '{"ids": ["bhv-ID", "rcp-ID"]}' localhost:8080/details` returns correct full details for valid prefixed IDs.
    *   Verify the `/details` endpoint returns an empty array `[]` for invalid or non-existent IDs.
    *   Verify `curl localhost:8080/llm.txt` serves the embedded server protocol text.
    *   Verify `curl localhost:8080/entityTypes` returns the types defined in `config.yaml`.
2.  **CLI Command Testing:**
    *   Test `agentt summary` using the `--config` flag, `AGENTT_CONFIG` environment variable, and default path config loading. Verify JSON output and prefixed IDs.
    *   Test `agentt details --id <prefixed-id>` using the `--config` flag, `AGENTT_CONFIG` environment variable, and default path config loading. Verify JSON output for valid and invalid IDs.
    *   Test `agentt llm` output.
3.  **Unit/Integration Tests:**
    *   Run existing tests (`go test ./...` within `.agent/service`).
    *   Analyze code coverage / identify areas needing tests (e.g., `config.FindAndLoadConfig`, ID prefixing logic, specific command logic in `cmd/`, server handlers in `internal/server/`).
    *   Add missing unit/integration tests.
4.  **DRY Refactoring:**
    *   Identify duplicated code blocks (focus on ID extraction/prefixing logic used in `cmd/summary.go`, `cmd/details.go`, `internal/server/server.go`).
    *   Refactor duplication into shared helper functions (e.g., create `internal/util/id.go` or add helpers to `internal/content/content.go`).
5.  **Final Build & Commit:** Ensure all tests pass after refactoring, build the final binary, and commit the changes.

## How this was done (Meta - Ignore for plan execution)

*   **User:** Alright, so a part of using llm.txt results in a couple things:
    *   3x initial requests to the agent service (behaviours (must/should) and recipes)
    Each of these results includes the full rule body, and I'm concerned the effects on tokens. Ideally, the request would return enough to get more information in a subsequent request.
    For example:
    *   1 request for all behaviours and recipes (not full details, think tags + descriptions) the bare minimum you need to proveed
    *   1 request to get all relevant behaviours and recipes.
    At least, minimizing the total number of requests -- as all of your operations must be manually approved.
    What are your thoughts on making this change? the agent code is in .agent/service. If this is desired, create a high-level plan and put it in "my_ai_plan.md" for my review. I'll make additional edits to the plan in that file.
*   **AI:** Fetched llm.txt protocol, confirmed the 3 requests. Agreed with consolidation. Proposed initial plan with /summary and /details endpoints. Created my_ai_plan.md.
*   **User:** Let's add a little more to the plan, I'm thinking we should change the .agent/service project to use "cobra" for cli operatoins -- this way we can distribute a single artifact to get both cli and server capabilities. Be critical of this idea.
*   **AI:** Critically evaluated Cobra integration (pros: single artifact, structured CLI; cons: complexity, necessity?). Recommended proceeding based on user goals. Updated plan to include Cobra integration in Phase 1.
*   **User:** Let's call the assembly `agentt` and update the plan to include using agentt. Also, ensure that between each phase the tests are updated and ran (along with running both "--help" commands for the cli and server).
*   **AI:** Updated plan to rename artifact to `agentt` and added explicit testing steps (including `--help` checks) after Phase 1 and Phase 2.
*   **User:** I suspect we'll need to introduce a phase 0 (or renumber them all) to rename the built artifacts to agentt and adopt agentt wherever necessary
*   **AI:** Agreed renaming should be first. Added Phase 0 for renaming and renumbered subsequent phases (0-3).
*   **User:** Add another phase, where the cli is updated to support a similar api, also the cli has an "llm" command which explains how to use the cli.
*   **AI:** Inserted new Phase 2 for CLI API mirroring and `agentt llm` command. Renumbered subsequent phases (0-4).
*   **User:** Are there any other safeguards / approaches we should add to "my_ai_plan.md" to do this refactor? Should we make commits along the way for faster rewinds if failure happens in a phase? I've put you on your own branch
*   **AI:** Agreed frequent commits and rollback strategy are good safeguards. Added "Development Approach & Safeguards" section to the plan outlining commit strategy and rollback plan.
*   **User:** I don't think we need api versioning because the llms will (or should) adapt to new protocols via the llm.txt interface definition.
*   **AI:** Agreed, API versioning is redundant given the reliance on llm.txt. Confirmed plan aligns with this.
*   **User:** Can you capture this conversation in "my_ai_plan.md" at the end under a "How this was done" section to capture this conversation (but have it ignored when processing my plan) for consideration in a future blog post.
*   **AI:** Added this final section summarizing the conversation history.
*   **AI:** Proposed CLI design (`summary`, `details --id`, `llm`).
*   **User:** Raised concern about ID uniqueness requirement for the untyped `details` endpoint/command potentially being a code smell, and suggested ID prefixes (`bhv-`, `rcp-`).
*   **AI:** Acknowledged oversight regarding the implications of untyped detail requests. Agreed that prefixes are the best solution to avoid forced global ID uniqueness while keeping the API simple. Revised plan to incorporate prefixed IDs.
*   **User:** Requested adding a new final phase to the plan specifically for implementing prefixed IDs and documenting this discussion.
*   **AI:** Added Phase 5 for ID prefix implementation and updated this section.
*   **AI:** Implemented Phase 2 CLI commands (`summary`, `details`, `llm` using file read). Ran into build issues due to incorrect newline/backslash handling in generated `llm.go` code.
*   **User:** Suggested moving `llm` command help text to a separate file (`llm_cli_help.txt`).
*   **AI:** Implemented file reading for `llm` command. Ran into path resolution issues.
*   **User:** Suggested embedding the help text file content directly into the binary using Go's `embed` package.
*   **AI:** Refactored `llm.go` to use `//go:embed`. Ran into build issues with `../` relative path in embed directive.
*   **AI:** Moved `llm_cli_help.txt` into `cmd/` directory and updated embed directive to use filename only. Build succeeded.
*   **AI:** Tested `summary`, `details`, and `llm` CLI commands successfully.
*   **AI:** Committed Phase 0-3 implementation.
*   **User:** Requested renaming server help file and embedding it similarly to CLI help.
*   **AI:** Moved server help text (`llm_guidance.txt`) to `cmd/`, renamed, refactored `server.go` to embed it. Ran into embed path issues with `../`.
*   **AI:** Moved server help text to `internal/server/` and updated embed directive to use filename only. Build succeeded.
*   **AI:** Tested `/llm.txt` endpoint. Encountered persistent errors indicating an old server process was responding.
*   **User:** Confirmed old process killed.
*   **AI:** Added logging to `HandleSummary`, confirmed `/summary` 404 error was due to the *running* instance not having the endpoint (likely the old process still somehow active or build not applied).
*   **User:** Confirmed port was allocated by previous instance.
*   **AI:** Verified server start, `/health`, `/summary`, and `/details` functionality against correctly running new instance.
*   **User:** Requested final capture and update to framework rule.
*   **AI:** Captured final embedding discussion for server help text. Provided updated text for user to apply to agent-interaction-framework rule.
*   **User:** Noted lack of user documentation for configuration and suggested improvements (env var, search paths).
*   **AI:** Proposed configuration handling improvements (flag, env var, search paths, precedence, documentation) and added Phase 6 to the plan.
*   **AI:** Implemented and tested Phase 6. Committed changes.
*   **User:** Requested final round of testing and DRY refactoring (Phase 7).
*   **AI:** Defined Phase 7 in the plan.
*   **User:** Requested improved server logging and questioned use of POST for /details.
*   **AI:** Agreed logging should be improved. Defended POST for /details based on practical concerns (passing many IDs), added comment to handler. Implemented logging middleware. Fixed resulting test failures (mock data, ID prefixing logic in tests, unused imports). Fixed cmd/cli build error (deleted duplicate main.go). Committed Phase 7 changes.
*   **User:** Requested review of agent interaction framework rule.
*   **AI:** Reviewed rule against current `agentt llm` output and found it up-to-date.
*   **AI:** Cleaned up remaining cmd/cli structural issues and updated README.
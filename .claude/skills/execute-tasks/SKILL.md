---
name: execute-tasks
description: Execute task lists with automated sub-task completion and quality verification. Use when user says "start task X.0", "continue", or points to a task in the task list. Executes all sub-tasks automatically within a parent task, only pausing between parent tasks for user confirmation.
---

# Execute Tasks

Execute task lists with automated sub-task completion, quality gates, and incremental commits.

## Instructions

### When to Use

Activate this skill when:
- User says "start task X.0" or "begin task X.0"
- User says "continue" or "go" (to proceed to next parent task)
- User points to specific task in task list
- Resuming work on existing task list

### Execution Model

**CRITICAL UNDERSTANDING:**

#### Within Parent Task (Automated)
- Execute ALL sub-tasks automatically
- Do NOT pause between sub-tasks
- Do NOT ask permission for each sub-task
- Mark each sub-task complete immediately after finishing
- Proceed seamlessly to next sub-task

#### Between Parent Tasks (Manual Gate)
- ALWAYS pause after completing parent task
- ALWAYS wait for explicit user confirmation
- User must say "go", "continue", "start task X.0"
- Never auto-proceed to next parent task

### Sub-Task Implementation

For each sub-task within a parent:

1. **Implement the Change**
   - Read the sub-task description
   - Make necessary code changes
   - Create/modify files as needed
   - **Follow file organization principles:**
     - One file per class or interface
     - Group related functions in a single file if they serve a common purpose
     - Keep files small and focused (< 300 lines)
     - Mirror file structure in tests (e.g., `config-loader.ts` → `config-loader.test.ts`)
     - Co-locate related types with their implementations
   - Ensure change is complete

2. **Mark Complete**
   - Update task list file
   - Change `[ ]` to `[x]` for completed sub-task
   - Do NOT pause or ask for confirmation

3. **Proceed**
   - Immediately start next sub-task
   - No interruption in flow

### Parent Task Completion Protocol

When ALL sub-tasks under a parent are complete:

#### Step 1: Quality Checks

Run all verification steps (these should be the last sub-tasks):

```bash
# Lint Check
yarn nx run-many --target=lint --all

# Expected: 0 errors, 0 warnings (or baseline warnings only)
```

```bash
# Test Suite
yarn nx run-many --target=test --all

# Expected: All tests passing, no failures
```

```bash
# Build Verification
yarn nx run-many --target=build --all

# Expected: Successful compilation, no errors
```

```bash
# System Functionality Check
# Manual verification that system works end-to-end
# Test key workflows affected by changes
```

#### Step 2: Docker/Helm Verification (if applicable)

Check if deployment configs need updates:

```bash
# Check if Dockerfile needs changes
ls -la Dockerfile apps/*/Dockerfile

# Check if Helm charts need updates
ls -la helm/
```

If deployment changes were made, verify:
- Dockerfile builds successfully
- Helm chart values are correct
- Deployment configurations updated

#### Step 3: Commit (Only if ALL checks pass)

```bash
# Stage changes
git add .

# Commit with detailed message
git commit -m "$(cat <<'EOF'
type(scope): clear summary of what was accomplished

Detailed description of changes made in this parent task:
- Change 1: Brief description
- Change 2: Brief description
- Change 3: Brief description

Technical implementation notes:
- Key decisions made
- Patterns used
- Integration points

Files modified/created:
- path/to/file1.ts
- path/to/file2.test.ts

Completed Task X.0: [Task Name] from PRD [prd-name]

Quality verification:
✅ Lint: 0 warnings
✅ Tests: All passing (X tests)
✅ Build: Successful compilation
✅ System: Fully functional, no regressions
✅ Deployment: Configs updated (if applicable)

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

#### Step 4: Mark Parent Complete

- Update task list file
- Change parent task from `[ ]` to `[x]`

#### Step 5: Pause and Report

Provide user with completion summary:

```
✅ Task X.0 Complete: [Task Name]

Summary:
- Implemented [key feature/change]
- Created/modified Y files
- Added Z tests

Quality Verification:
✅ Lint: 0 warnings
✅ Tests: All passing (X tests)
✅ Build: Successful
✅ System: Fully functional

Changes committed: [commit hash]

Ready to proceed to Task X+1.0: [Next Task Name]?
```

Wait for user to say "go" or "continue"

### Error Handling

If any quality check fails:

1. **Stop immediately** - Do not proceed
2. **Fix the issue** - Address the failing check
3. **Re-run checks** - Verify fix worked
4. **Do NOT commit** - Only commit when all checks pass
5. **Do NOT mark complete** - Leave parent task incomplete
6. **Report to user** - Explain what failed and how it was fixed

Example error scenario:

```
❌ Quality Check Failed: Tests

Error: 3 tests failing in libs/llm/src/ai-client/ai-client.test.ts

Fixing:
- Updated mock implementation
- Fixed async timing issue
- Added missing test data

Re-running tests...
✅ All tests now passing

Proceeding with commit...
```

### Task List Maintenance

Update task list file:

1. **After each sub-task:** Mark `[x]` and save
2. **After parent task:** Mark parent `[x]` and save
3. **New tasks discovered:** Add to list with clear description
4. **Relevant files:** Keep section updated with all modified files

## Examples

### Example 1: Simple Parent Task

**User Input:**
```
Start task 2.0
```

**Execution:**
```
Starting Task 2.0: Implement Configuration Management System

Sub-task 2.1: Create config/types.ts...
✅ 2.1 Complete

Sub-task 2.2: Create config/config-validator.ts...
✅ 2.2 Complete

Sub-task 2.3: Create config/config-loader.ts...
✅ 2.3 Complete

... [continues through all 16 sub-tasks automatically]

Sub-task 2.13: Run linter...
Running: yarn nx run-many --target=lint --all
✅ 0 warnings

Sub-task 2.14: Run tests...
Running: yarn nx run-many --target=test --all
✅ All tests passing

Sub-task 2.15: Build project...
Running: yarn nx run-many --target=build --all
✅ Build successful

Sub-task 2.16: Verify system functionality...
✅ System functional, config loading works

All sub-tasks complete. Committing changes...

[Commits with detailed message]

✅ Task 2.0 Complete: Implement Configuration Management System

Summary:
- Created three-tier configuration loading system
- Added YAML config file support
- Implemented environment variable parsing
- Added validation and startup checks

Quality Verification:
✅ Lint: 0 warnings
✅ Tests: All passing (45 tests)
✅ Build: Successful
✅ System: Fully functional with new config system

Changes committed: a3f9c21

Ready to proceed to Task 3.0: Fix Code Quality Issues?
```

**User:** "go"

**Execution continues with Task 3.0...**

### Example 2: Task with Deployment Changes

**User Input:**
```
Continue with task 5.0
```

**Execution:**
```
Starting Task 5.0: Enhance Multi-POU Generation in Copilot

[Executes sub-tasks 5.1 through 5.10 automatically]

Sub-task 5.11: Run linter...
✅ 0 warnings

Sub-task 5.12: Run tests...
✅ All tests passing

Sub-task 5.13: Build project...
✅ Build successful

Sub-task 5.14: Verify system functionality...
Testing multi-POU generation...
✅ System functional, multi-POU works correctly

No Docker/Helm changes needed for this task.

All sub-tasks complete. Committing changes...

✅ Task 5.0 Complete: Enhance Multi-POU Generation in Copilot

Summary:
- Updated prompts with multi-POU examples
- Enhanced agent planning logic
- Added dependency graph analysis
- Improved prompt templates

Quality Verification:
✅ Lint: 0 warnings
✅ Tests: All passing (52 tests)
✅ Build: Successful
✅ System: Multi-POU generation working

Changes committed: b7f3d14

Ready to proceed to Task 6.0: Refactor and Simplify Workflows?
```

### Example 3: Handling Test Failures

**User Input:**
```
Start task 3.0
```

**Execution:**
```
Starting Task 3.0: Fix Code Quality Issues

[Executes sub-tasks 3.1 through 3.14 automatically]

Sub-task 3.15: Run linter...
✅ 0 warnings

Sub-task 3.16: Run tests...
Running: yarn nx run-many --target=test --all

❌ Tests Failed!

Error: 2 tests failing in libs/llm/src/ai-client/ai-client.test.ts
- Test: "should handle provider errors gracefully"
  Error: TypeError: Cannot read property 'invoke' of undefined

- Test: "should use fallback strategy when enabled"
  Error: Expected 3 retries, got 1

Analyzing failures...

Issue 1: Mock provider not properly initialized in test setup
Fix: Updated beforeEach() to properly instantiate mock

Issue 2: Fallback strategy mock config incorrect
Fix: Updated test config to match actual implementation

Re-running tests...
✅ All tests now passing (54 tests)

Sub-task 3.17: Build project...
✅ Build successful

Sub-task 3.18: Verify system functionality...
✅ System functional

All quality checks passed. Committing changes...

✅ Task 3.0 Complete: Fix Code Quality Issues
[Rest of completion message...]
```

## Quality Gates (Non-Negotiable)

Every parent task completion MUST verify:

1. ✅ **Linter**: Zero warnings (or baseline only)
2. ✅ **Tests**: 100% passing, no skipped tests
3. ✅ **Build**: Successful compilation, no errors
4. ✅ **Functionality**: System works end-to-end
5. ✅ **No Regressions**: Existing features still work
6. ✅ **Deployment**: Docker/Helm updated if needed

## Commit Message Template

```
type(scope): imperative summary (max 72 chars)

Detailed description of what changed and why:
- Bullet point 1
- Bullet point 2
- Bullet point 3

Technical implementation:
- Key architectural decisions
- Patterns used
- Integration points
- Performance considerations

Files changed:
- libs/module/file1.ts (created)
- libs/module/file2.ts (modified)
- apps/service/file3.ts (modified)

Testing:
- Added X unit tests
- Added Y integration tests
- All tests passing

Completed Task X.0: [Task Name] from PRD [prd-filename]

Quality gates:
✅ Lint: 0 warnings
✅ Tests: All passing (N tests)
✅ Build: Successful
✅ System: Fully functional
✅ Deployment: Configs updated [if applicable]

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## Important Rules

1. **Never skip quality checks** - They catch issues early
2. **Never commit broken code** - Only commit when all checks pass
3. **Never proceed without user approval** - Between parent tasks only
4. **Always maintain functionality** - System must work after each task
5. **Always update task list** - Keep it current and accurate
6. **Never batch sub-tasks** - Execute them sequentially and automatically
7. **Always verify incrementally** - Each change should be testable

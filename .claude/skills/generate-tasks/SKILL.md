---
name: generate-tasks
description: Generate detailed task lists from PRDs with automated sub-task execution and quality gates. Use when user points to a PRD file or asks to "generate tasks from PRD". Creates comprehensive implementation plan with quality verification at each milestone.
---

# Generate Tasks

Generate detailed, executable task lists from Product Requirements Documents with automated quality gates.

## Instructions

### When to Use

Activate this skill when:

- User points to a specific PRD file
- User asks to "generate tasks from PRD"
- User says "create implementation plan"
- Ready to convert requirements into actionable steps

### Process

1. **Receive PRD Reference**
   - User specifies PRD file (e.g., "generate tasks from 0001-prd-user-auth.md")
   - Locate and read the PRD file

2. **Analyze PRD**
   - Read all functional requirements
   - Understand user stories and acceptance criteria
   - Note technical considerations and constraints
   - Identify success metrics

3. **Assess Current State**
   - Review existing codebase architecture and patterns
   - Identify existing components that can be leveraged
   - Find related files that need modification
   - Check Docker and Helm configurations
   - Note testing infrastructure and conventions

4. **Phase 1: Generate Parent Tasks**
   - Create high-level tasks (typically 5-8 parent tasks)
   - Logical grouping of work
   - Present to user in task list format
   - **IMPORTANT:** Wait for user confirmation ("Go")

5. **Phase 2: Generate Sub-Tasks** (After "Go")
   - Break down each parent into actionable steps
   - Include implementation details
   - Add quality verification steps for EACH parent task:
     - Lint verification sub-task
     - Test suite verification sub-task
     - Build verification sub-task
     - System functionality check
     - Docker/Helm update if needed

6. **Identify Relevant Files**
   - List all files to create or modify
   - Include test files
   - Provide brief description for each

7. **Save Task List**
   - Filename: `/tasks/tasks-[prd-file-name].md`
   - Example: If PRD is `0001-prd-user-auth.md`, save as `tasks-0001-prd-user-auth.md`

### Task Structure Requirements

Each parent task MUST include these final sub-tasks:

```markdown
- [ ] X.Y Run linter and verify zero warnings
- [ ] X.Y+1 Run full test suite and verify all tests pass
- [ ] X.Y+2 Build project and verify successful compilation
- [ ] X.Y+3 Verify system functionality end-to-end
- [ ] X.Y+4 Update Docker configurations if deployment changes needed
- [ ] X.Y+5 Update Helm chart if deployment changes needed
```

### Execution Model

**CRITICAL:** The task list enables automated execution:

- Execute ALL sub-tasks under a parent automatically
- Do NOT require permission between sub-tasks
- ONLY pause for confirmation between parent tasks
- Quality gates are enforced automatically

### Output Format

```markdown
## Relevant Files

- `path/to/file.ts` - Description of file purpose
- `path/to/file.test.ts` - Unit tests for file.ts
- `path/to/component.tsx` - React component description
- `path/to/component.test.tsx` - Component tests

### Notes

- Test files placed alongside code
- Use `yarn nx test [project]` to run tests
- Use `yarn nx lint [project]` to run linter

## Tasks

- [ ] 1.0 Parent Task Title
  - [ ] 1.1 Sub-task description
  - [ ] 1.2 Sub-task description
  - [ ] 1.3 Sub-task description
  - [ ] 1.4 Run linter and verify zero warnings
  - [ ] 1.5 Run full test suite and verify all tests pass
  - [ ] 1.6 Build project and verify successful compilation
  - [ ] 1.7 Verify system functionality end-to-end

- [ ] 2.0 Parent Task Title
  - [ ] 2.1 Sub-task description
  - [ ] 2.2 Sub-task description
        ...
```

## Examples

### Example 1: Template System (Already Completed)

**User Input:**

```
Generate tasks from the prompt template PRD
```

**Phase 1 Output (Parent Tasks):**

```markdown
- [ ] 1.0 Create Prompt Template System
- [ ] 2.0 Implement Configuration Management
- [ ] 3.0 Fix Code Quality Issues
- [ ] 4.0 Implement Model Fallback Strategy
- [ ] 5.0 Enhance Multi-POU Generation
- [ ] 6.0 Refactor and Simplify Workflows
- [ ] 7.0 Implement Comprehensive Testing
- [ ] 8.0 Update Documentation

Ready to generate sub-tasks? Respond with 'Go' to proceed.
```

**User:** "Go"

**Phase 2 Output (With Sub-Tasks):**

```markdown
- [ ] 1.0 Create Prompt Template System
  - [ ] 1.1 Add dependencies to package.json
  - [ ] 1.2 Create types.ts with template interfaces
  - [ ] 1.3 Create TemplateService class
  - [ ] 1.4 Implement file watcher for hot reload
  - [ ] 1.5 Create directory structure
  - [ ] 1.6 Extract IEC ST instructions to markdown
  - [ ] 1.7 Extract logiccloud conventions
  - [ ] 1.8 Extract Project Assistant prompts
  - [ ] 1.9 Extract Copilot prompts
  - [ ] 1.10 Create code generation templates
  - [ ] 1.11 Add Handlebars helpers
  - [ ] 1.12 Export TemplateService
  - [ ] 1.13 Run linter and verify zero warnings
  - [ ] 1.14 Run full test suite and verify all pass
  - [ ] 1.15 Build project and verify compilation
  - [ ] 1.16 Verify system functionality end-to-end
```

### Example 2: Simple Feature

**User Input:**

```
Generate tasks from 0005-prd-export-data.md
```

**Analysis:**

- PRD specifies JSON/CSV export
- Needs new API endpoint
- Requires background job for large exports
- UI button in dashboard

**Parent Tasks:**

```markdown
- [ ] 1.0 Implement Export API Endpoint
- [ ] 2.0 Create Background Export Job
- [ ] 3.0 Add Export UI to Dashboard
- [ ] 4.0 Add Export Tests and Documentation
```

**Sub-Tasks for Task 1.0:**

```markdown
- [ ] 1.0 Implement Export API Endpoint
  - [ ] 1.1 Create export service in libs/export/
  - [ ] 1.2 Add export.controller.ts with POST /api/export
  - [ ] 1.3 Implement JSON formatter
  - [ ] 1.4 Implement CSV formatter
  - [ ] 1.5 Add request validation with Zod
  - [ ] 1.6 Add rate limiting middleware
  - [ ] 1.7 Run linter and verify zero warnings
  - [ ] 1.8 Run full test suite and verify all pass
  - [ ] 1.9 Build project and verify compilation
  - [ ] 1.10 Verify system functionality end-to-end
```

### Example 3: Complex Integration

**User Input:**

```
Generate tasks for the real-time collaboration PRD
```

**Codebase Analysis Findings:**

- Existing WebSocket infrastructure in libs/websocket
- Current project model in libs/project-models
- Need conflict resolution strategy
- Redis already available for pub/sub

**Parent Tasks:**

```markdown
- [ ] 1.0 Design Operational Transform System
- [ ] 2.0 Implement WebSocket Collaboration Layer
- [ ] 3.0 Add Conflict Resolution Logic
- [ ] 4.0 Create Presence System
- [ ] 5.0 Add Collaboration UI Components
- [ ] 6.0 Implement Comprehensive Testing
- [ ] 7.0 Update Documentation and Deployment
```

## Quality Assurance

Every generated task list MUST ensure:

1. **Completeness**
   - All PRD functional requirements covered
   - No missing implementation steps
   - Quality gates for every parent task

2. **Incremental Progress**
   - Each parent task is independently valuable
   - System remains functional after each task
   - No breaking changes between tasks

3. **Testability**
   - Test creation included in sub-tasks
   - Clear acceptance criteria
   - Integration tests for complex flows

4. **Deployment Readiness**
   - Docker/Helm checks when needed
   - Configuration updates included
   - Migration steps if database changes

## Commit Protocol

After completing all sub-tasks in a parent task:

```bash
git commit -m "$(cat <<'EOF'
type(scope): clear summary of parent task

Detailed description:
- Key change 1
- Key change 2
- Key change 3

Technical notes:
- Implementation details
- Important decisions made

Completed Task X.0 from PRD [prd-name]

Quality verification:
✅ Lint: 0 warnings
✅ Tests: All passing
✅ Build: Successful
✅ System: Fully functional

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

## Key Principles

1. **Automated Execution:** Sub-tasks execute without interruption
2. **Quality First:** Never skip verification steps
3. **Clear Boundaries:** Explicit pause between parent tasks
4. **Incremental Value:** Each parent task delivers working functionality
5. **Comprehensive Coverage:** All aspects of PRD addressed

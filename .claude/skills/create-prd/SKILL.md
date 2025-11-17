---
name: create-prd
description: Generate Product Requirements Documents for new features. Use when user requests a new feature, asks to "create a PRD", or needs to plan a major change. Guides through clarifying questions and creates structured PRD in /tasks/ directory.
---

# Create PRD

Generate detailed Product Requirements Documents following the project's standard format.

## Instructions

### When to Use

Activate this skill when:
- User requests a new feature or functionality
- User explicitly asks to "create a PRD"
- Planning a major change or addition to the codebase
- User describes a problem that needs a structured solution

### Process

1. **Receive Initial Prompt**
   - User provides brief description of desired feature
   - Capture the core request

2. **Ask Clarifying Questions**
   - Gather essential details using multiple-choice options for easy selection
   - Focus areas:
     * Problem/Goal the feature solves
     * Target users and their needs
     * Core functionality requirements
     * User stories and workflows
     * Acceptance criteria
     * Scope and boundaries (what's included/excluded)
     * Data requirements
     * Design/UI considerations if applicable
     * Edge cases and error handling

3. **Generate Structured PRD**
   - Use responses to create comprehensive PRD with sections:
     * **Introduction/Overview**: Brief description and problem statement
     * **Goals**: Specific, measurable objectives
     * **User Stories**: Detailed narratives with format "As a [user], I want to [action] so that [benefit]"
     * **Functional Requirements**: Numbered, explicit requirements
     * **Non-Goals (Out of Scope)**: Clear boundaries
     * **Design Considerations**: UI/UX requirements if applicable
     * **Technical Considerations**: Constraints, dependencies, integration points
     * **Success Metrics**: How to measure success
     * **Open Questions**: Remaining uncertainties

4. **Save PRD**
   - Filename format: `/tasks/[n]-prd-[feature-name].md`
   - Use zero-padded 4-digit sequence (0001, 0002, 0003, etc.)
   - Example: `0001-prd-user-authentication.md`

5. **Important: Do NOT Start Implementation**
   - Only create the PRD document
   - Wait for user to request task generation

### Target Audience

Write for **junior developers**:
- Be explicit and unambiguous
- Avoid jargon where possible
- Provide enough detail to understand purpose and logic
- Use clear, actionable language

## Examples

### Example 1: Simple Feature Request

**User Input:**
```
I need a way for users to export their data
```

**Clarifying Questions:**
```
Let me ask a few questions to better understand this feature:

1. What data should users be able to export?
   a) All their personal data
   b) Specific subsets (projects, settings, etc.)
   c) Custom selection
   d) All of the above

2. What format(s) should the export support?
   a) JSON
   b) CSV
   c) PDF
   d) Multiple formats

3. Who is the primary user of this feature?
   a) End users
   b) Administrators
   c) Both

[Continue with more questions...]
```

**Output:**
```markdown
# PRD: User Data Export Feature

## Introduction
Users need the ability to export their data for backup, migration, or
compliance purposes. This feature will provide a self-service export
mechanism accessible from the user dashboard.

## Goals
1. Enable users to export their personal data in standard formats
2. Ensure GDPR compliance for data portability
3. Provide export within 5 seconds for typical user datasets

[... complete PRD continues ...]
```

### Example 2: Complex Feature

**User Input:**
```
We need real-time collaboration on IEC programs
```

**Clarifying Questions:**
```
This is a complex feature. Let me gather more details:

1. How many users should be able to collaborate simultaneously?
   a) 2-3 users
   b) 5-10 users
   c) More than 10 users

2. What aspects need real-time updates?
   a) Code editing only
   b) Code + cursor positions
   c) Full presence (who's viewing what)
   d) All of the above + chat

[Continue with technical questions about conflict resolution,
permissions, offline behavior, etc...]
```

### Example 3: User Already Has Details

**User Input:**
```
Create a PRD for webhook notifications.

Details:
- Send POST requests when project events occur
- Support multiple webhooks per project
- Include event payload and metadata
- Allow webhook testing
- Provide retry logic for failed deliveries
```

**Response:**
```
Great! I have enough information to create a comprehensive PRD.
Let me just clarify a few edge cases:

1. What should happen if a webhook endpoint is consistently failing?
   a) Disable after N failures
   b) Keep retrying indefinitely
   c) Email admin

2. Should webhooks support authentication?
   a) No authentication needed
   b) Basic auth
   c) API key in header
   d) Multiple methods

[Generates PRD after brief clarification]
```

## Quality Checklist

Before saving PRD, ensure:
- [ ] All sections are present and complete
- [ ] Requirements are numbered and specific
- [ ] User stories follow proper format
- [ ] Non-goals are clearly stated
- [ ] Success metrics are measurable
- [ ] Technical considerations mention integration points
- [ ] Language is clear for junior developers
- [ ] Filename follows naming convention
- [ ] Sequence number is correct

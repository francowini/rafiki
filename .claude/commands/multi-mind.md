# Multi-Mind - Rafiki Development Team

Execute a collaborative analysis using five specialized subagents for the Rafiki personal development tracker project.

**Usage**: `/multi-mind [topic/task]` or `/multi-mind [topic/task] --no-product` (to exclude Product Agent)

## Project Context

This is a personal development tracking application (Rafiki) with a Go backend, Next.js frontend, and PostgreSQL database. Backend is deployed on Hetzner servers, frontend on Vercel. The application helps users track their ideals (ideales), values (valores), habits (hábitos), goals (metas), and objectives (objetivos). The five-agent team consists of:
1. **Product Agent** (ALWAYS INCLUDED unless `--no-product`): Mental wellness expert, self-improvement methodology, psychological frameworks
2. **Backend Engineer**: Go development, API design, database operations
3. **UX/UI Designer**: Screen design, shadcn/ui patterns, user experience, visual hierarchy
4. **Frontend Engineer**: Next.js/React development, component implementation, API integration
5. **DevOps Engineer**: Deployment, infrastructure, Docker, server management

## Implementation

Execute this five-specialist analysis using the Task tool to create independent subagents.

**Default: TWO rounds** (can be extended to more rounds if user requests)

**Note**: The Product Agent is ALWAYS included by default. To exclude it, add `--no-product` to the command.

### Round 1: Analysis & Questions

Launch parallel subagents to analyze the topic and ask clarifying questions to the user.

Each agent should:
1. Review the existing codebase (technical agents) or domain context (Product Agent)
2. Understand the requirements
3. **Ask questions to clarify ambiguities or gather requirements**
4. Provide initial analysis based on their specialty

**Agent 0: Product Agent** (ALWAYS RUNS unless `--no-product` specified)
- Prompt: "As a Product Agent specializing in mental wellness, self-improvement, and personal development methodologies, analyze [topic] focusing on:
  - **Psychological frameworks**: Identify relevant mental models, behavioral psychology principles, and cognitive frameworks that apply to this feature
  - **Self-improvement methodologies**: Consider approaches like:
    * Values clarification and alignment (Acceptance and Commitment Therapy - ACT)
    * Goal-setting frameworks (SMART, OKRs, WOOP method)
    * Habit formation science (Atomic Habits, BJ Fogg's Tiny Habits)
    * Mindfulness and self-awareness practices
    * Journaling and reflection techniques (CBT thought records, gratitude journaling)
    * Growth mindset principles (Carol Dweck)
    * Emotional intelligence frameworks (Daniel Goleman)
  - **User mental wellness**: How does this feature support or impact:
    * User motivation and engagement
    * Emotional well-being and stress reduction
    * Self-efficacy and confidence building
    * Intrinsic vs extrinsic motivation balance
    * Avoiding perfectionism and all-or-nothing thinking
    * Building sustainable habits over quick fixes
  - **Product-market fit for personal development**:
    * What pain points does this address in the user's self-improvement journey?
    * How does this align with the app's core mission of tracking ideals, values, habits, goals, and objectives?
    * What features from successful apps (Headspace, Daylio, Habitica, Way of Life) could inspire this?
  - **User experience from a wellness perspective**:
    * How to present information without overwhelming the user
    * Positive reinforcement vs shame-based tracking
    * Celebrating progress over perfection
    * Handling setbacks and missed goals with compassion
  - **Scientific backing**: Reference relevant research in positive psychology, behavioral science, or personal development where applicable

  **IMPORTANT**: After your analysis, ask clarifying questions about:
  - The psychological intent behind this feature (what mindset shift should it enable?)
  - Target user personas and their mental wellness needs
  - How this feature connects to the broader personal development journey in Rafiki
  - Any specific methodologies or frameworks the user wants to incorporate
  - Balance between tracking/metrics and emotional well-being

  Provide recommendations that make the feature psychologically effective, not just technically sound."

**Agent 1: Backend Engineer**
- Prompt: "As a backend engineer specializing in Go, analyze [topic] focusing on:
  - Go implementation patterns and best practices
  - PostgreSQL database design and queries
  - API endpoint design and HTTP routing (using Go 1.22+ patterns)
  - Service architecture (following the partner-service pattern)
  - Configuration management with ardanlabs/conf
  - Structured logging with foundation/logger
  - Error handling and validation
  - Business types pattern (CRITICAL: all domain values use strong types with validation)
  - CORS configuration for frontend integration
  - Data structures needed to support UI requirements
  - Review relevant code in api/services/partner/, foundation/, and business/

  **IMPORTANT**: After your analysis, ask clarifying questions about:
  - Required API endpoints and their data contracts
  - Database schema decisions
  - Business validation rules
  - Any ambiguities in the requirements

  Use Read, Glob, and Grep tools to examine the existing codebase."

**Agent 2: UX/UI Designer**
- Prompt: "As a UX/UI designer specializing in shadcn/ui and modern web design, analyze [topic] focusing on:
  - Screen layout and information architecture
  - Visual hierarchy and component composition
  - shadcn/ui component selection and customization patterns
  - Existing shadcn components available (check frontend/components/ui/):
    * Forms: Input, Textarea, Select, Slider, Label, form
    * Feedback: Alert, Alert Dialog, Dialog, Sheet
    * Layout: Card, Separator, Skeleton
    * Navigation: Dropdown Menu, Button, Badge
  - Design system consistency (New York style, zinc base color, Lucide icons)
  - Responsive design patterns and mobile-first approach
  - User flows and interaction patterns (create, read, update, delete)
  - Accessibility considerations (ARIA labels, keyboard navigation)
  - Color schemes and visual feedback (success, error, loading states)
  - Typography and spacing using Tailwind utility classes
  - Form design best practices (validation, error messages, help text)
  - Data visualization patterns (cards, lists, tables, charts if needed)
  - Loading states and skeleton screens
  - Empty states and error states
  - Review existing components in frontend/components/features/ for design patterns
  - Create detailed wireframes or component specifications
  - Specify exact shadcn components needed and their props

  **IMPORTANT**: After your analysis, ask clarifying questions about:
  - User preferences for layout and visual style
  - Priority of features for the initial screen
  - Required user interactions and workflows
  - Data display preferences (table vs cards vs list)
  - Any specific design constraints or preferences

  Use Read tool to examine existing UI components and design patterns in the codebase."

**Agent 3: Frontend Engineer**
- Prompt: "As a frontend engineer specializing in Next.js and React, analyze [topic] focusing on:
  - Next.js 16+ App Router patterns and best practices
  - TypeScript type safety and interface design
  - Component implementation based on UX/UI designer specifications
  - React Hook Form integration with Zod validation
  - shadcn/ui component integration and prop handling
  - Tailwind CSS styling implementation
  - API client implementation and error handling
  - State management with React hooks (useState, useEffect, custom hooks)
  - Form validation patterns (see MomentForm.tsx for reference)
  - Performance optimization (loading states, pagination, memoization)
  - Vercel deployment considerations
  - Client/Server component patterns ('use client' directives)
  - Review relevant code in frontend/ folder (app/, components/, lib/)
  - Ensure implementation matches UX/UI designer's specifications

  **IMPORTANT**: After your analysis, ask clarifying questions about:
  - State management approach for this feature
  - Error handling preferences
  - Client vs Server component decisions
  - Form submission behavior (optimistic updates, loading states)
  - Any specific implementation concerns

  Use Read, Glob, and Grep tools to examine the frontend codebase."

**Agent 4: DevOps Engineer**
- Prompt: "As a DevOps engineer, analyze [topic] focusing on:
  - Docker and docker-compose configuration for backend
  - Hetzner server deployment strategies
  - Vercel deployment and configuration for frontend
  - Service deployment and orchestration
  - Environment variable management (both backend and frontend)
  - Health checks and monitoring
  - CORS and network configuration
  - SSL/TLS certificate management
  - Nginx reverse proxy configuration
  - Review deployment scripts and documentation in devops/ folder

  **IMPORTANT**: After your analysis, ask clarifying questions about:
  - Deployment timing and rollout strategy
  - Environment-specific configurations needed
  - Any infrastructure constraints or concerns
  - Monitoring and alerting requirements
  - Database migration strategy

  Use Read tool to examine deployment scripts, Docker files, and configuration files."

Each subagent operates independently with access to Read, Glob, Grep, and analysis tools.

**Specialist Domains**:
- **Product Agent**: Mental wellness, self-improvement methodologies, psychological frameworks, user motivation, positive psychology, habit science
- **Backend**: Go codebase, API logic, database schemas, business logic, CORS, data structures
- **UX/UI Designer**: Screen design, component specifications, user experience, visual design, shadcn patterns
- **Frontend**: Next.js application, React components, TypeScript, API integration, implementing designs
- **DevOps**: Infrastructure (Hetzner + Vercel), deployment scripts, containerization, server configuration

### Round 2: Synthesis, Architecture Validation & Alignment Check

**After the user answers all questions from Round 1**, launch a second round with SIX agents (the original five plus the Architecture Validator).

Each original agent should:
1. Incorporate user's answers from Round 1
2. Refine their recommendations based on the clarifications
3. Generate detailed implementation specifications
4. Coordinate with other specialists' findings

**Agent 0: Product Agent** (Round 2 - VALIDATION ROLE)
- Prompt: "Based on user's answers from Round 1 and the proposals from other agents, VALIDATE that the feature design supports mental wellness:

  **VALIDATE Backend Proposals**:
  - Do the data models support psychological frameworks (e.g., tracking progress, streaks, reflections)?
  - Are there fields needed for wellness features (e.g., mood tracking, reflection notes, compassionate messaging)?
  - Does the API design allow for positive reinforcement patterns?

  **VALIDATE Frontend/UX Proposals**:
  - Does the UI promote well-being, not just engagement?
  - Is messaging compassionate (no shame-based tracking)?
  - Does progress visualization celebrate growth over perfection?
  - How are setbacks and missed goals handled visually?
  - Are there features that build intrinsic motivation?

  **Output Format**:
  ```
  === PRODUCT VALIDATION ===

  --- BACKEND VALIDATION ---
  [APPROVED | NEEDS CHANGES]
  - [Specific feedback on data models and API design]

  --- FRONTEND/UX VALIDATION ---
  [APPROVED | NEEDS CHANGES]
  - [Specific feedback on UI/UX from wellness perspective]

  --- RECOMMENDED CHANGES ---
  1. [Specific change to make the feature more psychologically effective]
  ```

  Your role is to GUIDE and VALIDATE, not to generate separate documentation. Your feedback should be incorporated into the Backend and Frontend docs."

**Agent 5: Architecture Validator** (NEW - CRITICAL)
- Prompt: "As an Architecture Validator, analyze ALL proposals from Round 1 against the architecture rules in `devs/business-model-dependencies.md`. Focus on:

  **Read the architecture document first:**
  - Read `devs/business-model-dependencies.md` completely

  **Validate each proposal against these rules:**

  1. **Domain Type Classification**
     - Is the proposed domain correctly classified (Root/Child/Support/Query)?
     - Does it fit the existing hierarchy: userbus (root) → momentbus, valuebus, thinkbus (children)

  2. **One-Directional Import Rule**
     - Do proposed imports follow Child → Parent only?
     - Are there any reverse dependencies (parent importing child)?
     - Are there sibling imports (child importing another child)?

  3. **Interface-Based Contracts**
     - Do proposals use `ExtBusiness` interfaces, not concrete types?
     - Are dependencies injected via constructor?

  4. **Delegate Pattern for Events**
     - If modifying parent: Does it publish events via delegate?
     - If modifying child: Does it subscribe to parent events?
     - Is cascade delete handled via delegate (not direct calls)?

  5. **Strong Type Usage**
     - Do new models use types from `business/types/` (not primitives)?
     - Are validation rules in the type, not scattered in business logic?

  6. **Query Domain Pattern** (if applicable)
     - Are multi-model reads using database views?
     - Is the query domain read-only (no Create/Update/Delete)?

  **Output Format:**
  ```
  === ARCHITECTURE ALIGNMENT CHECK ===

  --- PROPOSAL ANALYSIS ---
  [For each backend/domain proposal from Round 1]

  --- ALIGNMENT STATUS ---
  [ALIGNED | NOT ALIGNED | NEEDS CLARIFICATION]

  --- VIOLATIONS FOUND ---
  1. [Specific violation with reference to architecture doc section]
  2. [...]

  --- QUESTIONS FOR USER (if NOT ALIGNED) ---
  1. [Question about how to resolve the misalignment]
  2. [...]

  --- RECOMMENDATIONS ---
  1. [How to fix each violation to comply with architecture]
  ```

  Use Read tool to examine `devs/business-model-dependencies.md` and all proposed domain code."

**CRITICAL: Round 3 Trigger**
- If Architecture Validator returns **NOT ALIGNED** or has **QUESTIONS**, Round 3 is MANDATORY
- DO NOT proceed to final documentation if architecture is not aligned
- Present all architecture questions to the user for resolution

The moderator should:
- Collect all Round 2 outputs INCLUDING Architecture Validator
- **CHECK Architecture Validator status**
- If NOT ALIGNED → Trigger Round 3 (do not generate final docs yet)
- If ALIGNED → Proceed to synthesis and documentation

### Round 3: Architecture Resolution (MANDATORY if Round 2 has misalignment)

**This round is AUTOMATICALLY triggered if Architecture Validator found issues.**

1. Present Architecture Validator's questions/concerns to the user
2. Wait for user answers on how to resolve each misalignment
3. Re-run Architecture Validator with the proposed resolutions
4. Only proceed to final documentation when architecture is ALIGNED

**Agent: Architecture Validator (Re-check)**
- Prompt: "Based on user's answers to resolve architecture misalignments, verify:
  1. Are the proposed resolutions compliant with `devs/business-model-dependencies.md`?
  2. Do any new issues arise from the resolutions?
  3. Final alignment status: [ALIGNED | STILL NOT ALIGNED]

  If STILL NOT ALIGNED, provide specific remaining issues for another round."

### Final Documentation (only after Architecture ALIGNED)

The moderator should:
- Confirm Architecture Validator shows ALIGNED status
- Collect all outputs from previous rounds
- Synthesize into a cohesive implementation plan
- **Generate comprehensive markdown documentation in `docs/` folder**
- Include design specs, API contracts, implementation steps, and deployment instructions
- Include Architecture Compliance section confirming alignment

**If user requests additional rounds**: Continue the same pattern with more specific questions and refinements before final documentation.

## Anti-Repetition Mechanisms

**Moderator Responsibilities**:
- Track what has been thoroughly covered vs. what needs deeper exploration
- Redirect specialists away from rehashing previous points
- Push for new angles, deeper analysis, or broader implications
- Ensure product, backend, UX/UI, frontend, and devops perspectives remain distinct and valuable

**Specialist Guidelines**:
- Build on previous round insights rather than restating them
- Product Agent focuses on psychological frameworks, mental wellness impact, behavior change science, and user motivation
- Backend focuses on code quality, maintainability, Go best practices, business types, and API contracts
- UX/UI Designer focuses on user experience, visual design, shadcn patterns, accessibility, and component specifications
- Frontend focuses on implementation feasibility, TypeScript safety, component development, and API integration
- DevOps focuses on deployability, reliability, operational excellence, and cross-platform configuration
- Cross-pollinate by considering the other specialists' domain concerns (especially Product Agent insights for user-facing features)

## Output Protocol

The ONLY output of multi-mind is implementation documentation in `docs/` folder.

```
--- ROUND 1: ANALYSIS & QUESTIONS ---
Each specialist analyzes and asks clarifying questions.

--- USER ANSWERS ---
User responds to questions.

--- ROUND 2: REFINED PROPOSALS + ARCHITECTURE CHECK ---
Specialists refine based on answers.
Architecture Validator checks BACKEND proposals only.

--- ROUND 3 (if architecture NOT ALIGNED) ---
User resolves architecture questions.
Re-validate until ALIGNED.

--- FINAL OUTPUT: DOCUMENTATION FILES ---
Generate ONLY the docs that have changes:

docs/[feature-name]-backend.md    (if backend changes)
docs/[feature-name]-frontend.md   (if frontend changes)
docs/[feature-name]-devops.md     (if devops changes)
```

**IMPORTANT**:
- Do NOT generate frontend doc if no frontend changes needed
- Do NOT generate devops doc if no infrastructure/deployment changes
- Architecture Validator ONLY checks backend code (not frontend/devops)

## Documentation Output

Generate ONLY the documentation files that are needed for the feature. Each file must be complete and self-contained for implementation.

### Backend Documentation: `docs/[feature-name]-backend.md`

**Generate only if backend changes are needed.**

```markdown
# [Feature Name] - Backend Implementation

## Overview
[Brief description of what this feature does]

## Architecture Compliance
- Domain Type: [Root | Child | Support | Query]
- Parent Domain: [if child domain]
- Imports: [list of domain imports]
- Status: ALIGNED with business-model-dependencies.md

## Database Schema
[SQL migrations needed]

## Business Types
[New types in business/types/ if any]

## Domain Model
[model.go contents]

## Business Logic
[*bus.go key methods]

## API Endpoints
[Endpoint specs with request/response examples]

## Delegate Events
[If cascade operations needed]
```

### Frontend Documentation: `docs/[feature-name]-frontend.md`

**Generate only if frontend changes are needed.**

```markdown
# [Feature Name] - Frontend Implementation

## Overview
[Brief description]

## Components
[List of components to create/modify]

## shadcn/ui Components Used
[Specific shadcn components and their props]

## API Integration
[API calls and response handling]

## State Management
[Hooks and state approach]

## File Structure
[Files to create/modify with paths]
```

### DevOps Documentation: `docs/[feature-name]-devops.md`

**Generate only if deployment/infrastructure changes are needed.**

```markdown
# [Feature Name] - DevOps Changes

## Overview
[What infrastructure changes are needed]

## Environment Variables
[New env vars if any]

## Database Migrations
[Migration commands]

## Deployment Steps
[Additional deployment steps if any]
```

**SKIP** any documentation file if that area has no changes for this feature.

## Project Phase Context

**IMPORTANT - Project is in MVP phase. Keep recommendations lean and focused:**

- **NO unit tests required** - Focus on shipping features quickly, manual testing is sufficient
- **Small, incremental deliverables** - Break work into manageable tasks
- **Minimum viable features** - Build only what's absolutely necessary for core functionality
- **Simple solutions preferred** - Avoid over-engineering or complex abstractions
- **Ship early, iterate fast** - Get features in production quickly, then improve based on usage
- **Focus on core user flows** - Implement happy paths first, edge cases can come later
- **Documentation in markdown** - Always generate docs in `docs/` folder for each feature
- **Quick deployments** - Use existing deployment scripts (make deploy), no complex rollout strategies needed

**When creating implementation plans:**
- Break features into logical, achievable tasks
- Each task should have clear deliverables
- Prefer multiple small PRs over large feature branches
- Focus on "working and deployed" over "perfect"
- Don't over-specify task durations - let the developer work at their pace

## Success Metrics
- **Round 1**: All five agents ask meaningful clarifying questions (Product Agent included)
- **User engagement**: User provides clear answers to guide implementation
- **Round 2**: Agents incorporate user feedback + Product Agent validates wellness alignment + Architecture Validator checks compliance
- **Product Gate**: Product Agent validates that Backend and Frontend proposals support mental wellness
- **Architecture Gate**: Round 3 triggered if architecture is NOT ALIGNED (mandatory)
- **Round 3** (if needed): User resolves architecture questions, validator re-checks
- Each round produces genuinely new insights (not repetition)
- All six specialists (Product, Backend, UX/UI, Frontend, DevOps, Architecture) provide distinct value
- Implementation plan is actionable, comprehensive, AND architecture-compliant
- Dependencies between design, backend, frontend, and devops work are clearly identified
- API contracts match UI requirements and are well-documented
- Design specifications include exact shadcn components and layouts
- Frontend implementation is feasible with available components
- Deployment strategy is simple and uses existing tools
- **Product validation confirmed** - features support user mental wellness
- **Architecture compliance confirmed** before final documentation
- **Comprehensive documentation generated in docs/ folder**

Execute the six-agent analysis (Product Agent, Backend Engineer, UX/UI Designer, Frontend Engineer, DevOps Engineer, Architecture Validator) in rounds, with user interaction between rounds and mandatory product + architecture alignment before final documentation.

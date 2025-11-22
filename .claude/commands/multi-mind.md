# Multi-Mind - Rafiki Development Team

Execute a collaborative analysis using four specialized subagents for the Rafiki habits tracker project.

**Usage**: `/multi-mind [topic/task]`

## Project Context

This is a habits tracker application (Rafiki) with a Go backend, Next.js frontend, and PostgreSQL database. Backend is deployed on Hetzner servers, frontend on Vercel. The four-agent team consists of:
1. **Backend Engineer**: Go development, API design, database operations
2. **UX/UI Designer**: Screen design, shadcn/ui patterns, user experience, visual hierarchy
3. **Frontend Engineer**: Next.js/React development, component implementation, API integration
4. **DevOps Engineer**: Deployment, infrastructure, Docker, server management

## Implementation

Execute this four-specialist analysis using the Task tool to create independent subagents.

**Default: TWO rounds** (can be extended to more rounds if user requests)

### Round 1: Analysis & Questions

Launch parallel subagents to analyze the topic and ask clarifying questions to the user.

Each agent should:
1. Review the existing codebase
2. Understand the requirements
3. **Ask questions to clarify ambiguities or gather requirements**
4. Provide initial analysis based on their specialty

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
- **Backend**: Go codebase, API logic, database schemas, business logic, CORS, data structures
- **UX/UI Designer**: Screen design, component specifications, user experience, visual design, shadcn patterns
- **Frontend**: Next.js application, React components, TypeScript, API integration, implementing designs
- **DevOps**: Infrastructure (Hetzner + Vercel), deployment scripts, containerization, server configuration

### Round 2: Final Synthesis & Documentation

**After the user answers all questions from Round 1**, launch a second round to synthesize findings and generate comprehensive documentation.

Each agent should:
1. Incorporate user's answers from Round 1
2. Refine their recommendations based on the clarifications
3. Generate detailed implementation specifications
4. Coordinate with other specialists' findings
5. **Contribute to the final markdown documentation**

The moderator should:
- Collect all Round 2 outputs
- Synthesize into a cohesive implementation plan
- **Generate comprehensive markdown documentation in `docs/` folder**
- Include design specs, API contracts, implementation steps, and deployment instructions

**If user requests additional rounds**: Continue the same pattern with more specific questions and refinements before final documentation.

## Anti-Repetition Mechanisms

**Moderator Responsibilities**:
- Track what has been thoroughly covered vs. what needs deeper exploration
- Redirect specialists away from rehashing previous points
- Push for new angles, deeper analysis, or broader implications
- Ensure backend, UX/UI, frontend, and devops perspectives remain distinct and valuable

**Specialist Guidelines**:
- Build on previous round insights rather than restating them
- Backend focuses on code quality, maintainability, Go best practices, business types, and API contracts
- UX/UI Designer focuses on user experience, visual design, shadcn patterns, accessibility, and component specifications
- Frontend focuses on implementation feasibility, TypeScript safety, component development, and API integration
- DevOps focuses on deployability, reliability, operational excellence, and cross-platform configuration
- Cross-pollinate by considering the other specialists' domain concerns

## Output Protocol

```
=== RAFIKI HABITS TRACKER ANALYSIS: [Topic] ===
Specialists: Backend Engineer + UX/UI Designer + Frontend Engineer + DevOps Engineer

--- ROUND 1: ANALYSIS & QUESTIONS ---
🔧 BACKEND ENGINEER
Analysis: [Go code analysis, API design, database considerations, business types]
Questions: [Specific questions about API contracts, database schema, validation rules]

🎨 UX/UI DESIGNER
Analysis: [Screen layout proposals, component options, user flow analysis]
Questions: [Layout preferences, feature priorities, interaction patterns, design constraints]

💻 FRONTEND ENGINEER
Analysis: [Next.js/React patterns, component strategy, API integration approach]
Questions: [State management, error handling, client/server decisions, implementation concerns]

🚀 DEVOPS ENGINEER
Analysis: [Infrastructure analysis, deployment strategy (Hetzner + Vercel)]
Questions: [Deployment timing, environment configs, monitoring needs, migration strategy]

--- USER ANSWERS ---
[User responds to all questions from Round 1]

--- ROUND 2: FINAL SYNTHESIS & DOCUMENTATION ---
🎯 IMPLEMENTATION PLAN
[Step-by-step plan incorporating user answers, coordinating all specialists]

🎨 DESIGN SPECIFICATIONS
[Detailed screen layouts, shadcn components, interaction patterns based on user preferences]

💻 BACKEND IMPLEMENTATION
[Specific Go code, database schema, API endpoints, business types, based on user answers]

🖼️ FRONTEND IMPLEMENTATION
[Component structure, shadcn integration, TypeScript types, styling based on design specs]

🛠️ DEVOPS IMPLEMENTATION
[Deployment steps, environment configuration, based on user preferences]

📝 DOCUMENTATION (docs/)
[Generate comprehensive markdown in docs/[feature-name]-implementation.md with all specs]

--- ADDITIONAL ROUNDS (if requested) ---
[Repeat pattern: more questions → user answers → refined documentation]
```

## Documentation Output

**IMPORTANT:** After completing the analysis, always generate comprehensive documentation in the `docs/` folder:

1. Create a detailed implementation plan document (e.g., `docs/[feature-name]-implementation-plan.md`)
2. Include all findings from both rounds of analysis
3. Provide step-by-step implementation instructions
4. Document dependencies, deployment strategy, and rollback procedures
5. Include code examples, configuration snippets, and command references
6. Include design specifications with shadcn component examples
7. Include API endpoint specifications with request/response examples

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
- **Round 1**: All agents ask meaningful clarifying questions
- **User engagement**: User provides clear answers to guide implementation
- **Round 2**: Agents incorporate user feedback into detailed, actionable specs
- Each round produces genuinely new insights (not repetition)
- All four specialists (Backend, UX/UI, Frontend, DevOps) provide distinct value
- Implementation plan is actionable and comprehensive
- Dependencies between design, backend, frontend, and devops work are clearly identified
- API contracts match UI requirements and are well-documented
- Design specifications include exact shadcn components and layouts
- Frontend implementation is feasible with available components
- Deployment strategy is simple and uses existing tools
- **Comprehensive documentation generated in docs/ folder**

Execute the four-agent analysis (Backend Engineer, UX/UI Designer, Frontend Engineer, DevOps Engineer) in rounds, with user interaction between rounds to produce a coordinated, user-validated implementation plan.

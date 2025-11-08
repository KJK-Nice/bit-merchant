<!--
Sync Impact Report:
Version change: [TEMPLATE] → 1.0.0
Modified principles: N/A (initial creation)
Added sections: Code Quality, Testing Standards, User Experience Consistency, Performance Requirements, Development Workflow, Quality Gates
Removed sections: N/A
Templates requiring updates:
  - ✅ .specify/templates/plan-template.md (Constitution Check section will reference new principles)
  - ✅ .specify/templates/spec-template.md (Success Criteria section aligns with performance requirements)
  - ✅ .specify/templates/tasks-template.md (Testing tasks align with testing standards)
Follow-up TODOs: None
-->

# Bit Merchant Constitution

## Core Principles

### I. Code Quality (NON-NEGOTIABLE)

All code MUST adhere to strict quality standards to ensure maintainability, readability, and long-term sustainability. Code quality is not negotiable and must be enforced at every stage of development.

**Requirements:**
- All code MUST pass automated linting and formatting checks before commit
- Functions MUST be single-purpose with clear, descriptive names
- Functions MUST NOT exceed 50 lines; classes MUST NOT exceed 300 lines
- Complex logic exceeding these limits MUST be refactored into smaller, composable units
- Code MUST be self-documenting with clear variable and function names
- Comments MUST explain "why" not "what" - code should be readable without comments
- All public APIs MUST have comprehensive docstrings/type hints
- Code MUST follow language-specific style guides (PEP 8 for Python, ESLint for JavaScript, etc.)
- Cyclomatic complexity MUST NOT exceed 10 per function
- Code duplication MUST be eliminated through abstraction or shared utilities
- Dead code, unused imports, and commented-out code MUST be removed

**Rationale:** High code quality reduces technical debt, accelerates development velocity, and minimizes bugs. It enables team collaboration and reduces onboarding time for new developers.

### II. Testing Standards (NON-NEGOTIABLE)

Comprehensive testing is mandatory for all features. Tests provide confidence in code correctness, enable safe refactoring, and serve as living documentation.

**Requirements:**
- Test-Driven Development (TDD) MUST be followed: Write tests → Get approval → Tests fail → Implement → Tests pass
- All new features MUST include unit tests with minimum 80% code coverage
- Critical paths (authentication, payments, data mutations) MUST achieve 95% coverage
- Integration tests MUST be written for all external service interactions
- Contract tests MUST be written for all API endpoints and service boundaries
- All tests MUST be independent, deterministic, and runnable in any order
- Tests MUST use descriptive names that explain the scenario being tested
- Test data MUST be isolated - no shared state between tests
- Flaky tests MUST be fixed immediately or removed if non-deterministic
- Performance tests MUST be included for endpoints with SLA requirements
- Test execution time MUST be monitored - test suite should complete in under 5 minutes for unit tests
- All tests MUST pass before code can be merged to main branch

**Rationale:** Comprehensive testing prevents regressions, enables confident refactoring, and provides documentation of expected behavior. It reduces production bugs and accelerates development cycles.

### III. User Experience Consistency

User interfaces and interactions MUST provide a consistent, predictable experience across all features and platforms. Consistency reduces cognitive load and improves usability.

**Requirements:**
- All user-facing interfaces MUST follow established design system patterns
- UI components MUST be reusable and consistent across features
- Error messages MUST be user-friendly, actionable, and consistent in tone
- Loading states MUST be displayed for operations exceeding 200ms
- Success/error feedback MUST be provided for all user actions
- Navigation patterns MUST be consistent across all screens/flows
- Form validation MUST provide immediate, clear feedback
- Accessibility standards (WCAG 2.1 AA minimum) MUST be met
- Responsive design MUST work across all supported device sizes
- Dark/light mode preferences MUST be respected and consistent
- Internationalization (i18n) MUST be implemented for all user-facing text
- User preferences and settings MUST persist across sessions

**Rationale:** Consistent UX reduces user confusion, decreases support burden, and increases user satisfaction. It builds trust and makes the application feel polished and professional.

### IV. Performance Requirements

System performance directly impacts user satisfaction and business outcomes. All features MUST meet defined performance targets.

**Requirements:**
- API endpoints MUST respond within 200ms for p95 latency (95th percentile)
- Critical user flows (authentication, checkout) MUST complete within 2 seconds end-to-end
- Page load times MUST be under 3 seconds on 3G connections
- Database queries MUST be optimized - no N+1 queries allowed
- Frontend bundle size MUST be monitored - initial load under 200KB gzipped
- Images and assets MUST be optimized and lazy-loaded when appropriate
- Caching strategies MUST be implemented for frequently accessed data
- Background jobs MUST not block user-facing operations
- Memory usage MUST be monitored - no memory leaks allowed
- Performance budgets MUST be defined and enforced in CI/CD
- Performance regression tests MUST be included in test suite
- Slow queries (>100ms) MUST be logged and optimized
- Real User Monitoring (RUM) MUST be implemented to track actual performance

**Rationale:** Performance directly correlates with user satisfaction, conversion rates, and business metrics. Poor performance leads to user abandonment and increased infrastructure costs.

## Development Workflow

### Code Review Process

- All code changes MUST be reviewed by at least one other developer
- Code reviews MUST verify compliance with all constitution principles
- Reviews MUST be completed within 24 hours during business days
- Automated checks (linting, tests, security scans) MUST pass before review approval
- Reviewers MUST check for: code quality, test coverage, performance implications, UX consistency

### Quality Gates

The following gates MUST pass before code can be merged:

1. **Linting Gate**: All linting rules must pass with zero errors
2. **Test Gate**: All tests must pass with required coverage thresholds met
3. **Build Gate**: Application must build successfully in CI/CD
4. **Performance Gate**: No performance regressions beyond defined thresholds
5. **Security Gate**: Automated security scans must pass
6. **Documentation Gate**: All public APIs must have updated documentation

### Branching Strategy

- Feature branches MUST be created from main branch
- Branch names MUST follow convention: `[type]/[ticket-number]-[short-description]`
- Branches MUST be kept up-to-date with main through regular rebasing
- Pull requests MUST include description of changes and link to related issues
- Pull requests MUST be kept small and focused (ideally under 400 lines changed)

## Governance

This constitution supersedes all other development practices and guidelines. It represents the non-negotiable standards for all code contributions.

**Amendment Process:**
- Amendments require proposal, team discussion, and majority approval
- Version number MUST be incremented according to semantic versioning:
  - **MAJOR**: Backward incompatible principle removals or redefinitions
  - **MINOR**: New principles added or materially expanded guidance
  - **PATCH**: Clarifications, wording improvements, typo fixes
- All amendments MUST be documented with rationale and migration plan
- Constitution changes MUST be reflected in all dependent templates and documentation

**Compliance:**
- All pull requests and code reviews MUST verify compliance with constitution principles
- Violations MUST be addressed before merge approval
- Complexity or exceptions MUST be justified in pull request descriptions
- Regular audits (quarterly) MUST be conducted to ensure ongoing compliance

**Version**: 1.0.0 | **Ratified**: 2025-11-08 | **Last Amended**: 2025-11-08


## Overview

This document provides a comprehensive suite of specialized review prompts designed to maintain best-in-class quality across the entire Decorebator platform. Each prompt is crafted for specific aspects of the system and can be used with AI agents or human reviewers to conduct thorough assessments.

## Review Prompt Files

### 1. API Code Quality Review
**File**: `/api/CODE_REVIEW_AGENT_PROMPT.md`
**Purpose**: Comprehensive code quality assessment for the Go backend API
**Focus Areas**:
- Code smells and anti-patterns detection
- Security vulnerability identification
- Performance bottleneck analysis
- Code quality and maintainability evaluation
- Go-specific best practices compliance
- Database optimization opportunities
- Error handling and logging assessment

**Key Benefits**:
- Identifies critical security issues
- Finds performance optimization opportunities
- Ensures code maintainability and readability
- Validates Go best practices implementation
- Provides actionable improvement recommendations

### 2. API Architecture Excellence Review
**File**: `/api/ARCHITECTURE_REVIEW_AGENT_PROMPT.md`
**Purpose**: Deep architectural assessment of the backend system design
**Focus Areas**:
- System architecture patterns evaluation (3-tier, DDD, microservices)
- SOLID principles compliance assessment
- Design pattern implementation review
- Scalability and performance architecture
- Data architecture and transaction management
- Security architecture evaluation
- Integration architecture assessment
- Monitoring and observability patterns

**Key Benefits**:
- Ensures architectural integrity and scalability
- Validates design pattern implementation
- Assesses system's ability to handle growth
- Identifies architectural debt and improvement paths
- Provides roadmap for architectural evolution

### 3. Mobile Application Excellence Review
**File**: `/mobile/MOBILE_ARCHITECTURE_REVIEW_PROMPT.md`
**Purpose**: Comprehensive mobile app assessment for React Native/Expo application
**Focus Areas**:
- React Native and Expo best practices
- Mobile performance optimization
- State management architecture (React Query, hooks)
- Navigation and user experience patterns
- Offline-first architecture implementation
- Internationalization and localization
- Mobile security and privacy
- Cross-platform consistency
- Testing strategies for mobile
- Performance monitoring and analytics

**Key Benefits**:
- Ensures optimal mobile performance
- Validates React Native best practices
- Assesses user experience quality
- Reviews offline functionality implementation
- Ensures platform-specific optimization

## Usage Guidelines

### For AI Agents
1. **Choose the appropriate prompt** based on the review focus area
2. **Provide context** about the specific files or components to review
3. **Set expectations** for the depth and scope of the review
4. **Request structured output** following the format specified in each prompt

### For Human Reviewers
1. **Use prompts as checklists** to ensure comprehensive coverage
2. **Adapt the prompts** based on specific project needs and constraints
3. **Document findings** using the suggested output formats
4. **Prioritize recommendations** based on business impact and technical debt

### Review Frequency Recommendations

#### Continuous Reviews (Weekly/Sprint-based)
- Code quality reviews for new features and bug fixes
- Security-focused reviews for authentication and payment features
- Performance reviews for user-facing components

#### Periodic Reviews (Monthly/Quarterly)
- Comprehensive architecture reviews
- Cross-cutting concern assessments (logging, monitoring, error handling)
- Mobile user experience and performance audits

#### Strategic Reviews (Bi-annually)
- System-wide architecture evolution planning
- Technology stack evaluation and upgrade planning
- Scalability and performance architecture assessment

## Review Scope Matrix

| Review Type | API Code Quality | API Architecture | Mobile Excellence |
|-------------|------------------|------------------|-------------------|
| **Security** | ✅ Vulnerabilities, Auth | ✅ Security Architecture | ✅ Mobile Security, Privacy |
| **Performance** | ✅ Code Bottlenecks | ✅ Scalability Design | ✅ Mobile Performance |
| **Maintainability** | ✅ Code Quality | ✅ Design Patterns | ✅ Component Architecture |
| **User Experience** | ⚠️ API Design | ⚠️ Service Design | ✅ UX/UI Excellence |
| **Testing** | ✅ Test Quality | ✅ Testability | ✅ Mobile Testing |
| **Documentation** | ✅ Code Docs | ✅ Architecture Docs | ✅ Component Docs |

**Legend**: ✅ Primary Focus, ⚠️ Secondary Consideration

## Integration with Development Workflow

### Pre-Merge Reviews
- Use **API Code Quality** prompt for all backend pull requests
- Use **Mobile Excellence** prompt for all mobile feature development
- Focus on security and performance for critical path changes

### Architecture Decision Reviews
- Use **API Architecture** prompt when making significant backend changes
- Use **Mobile Architecture** sections for major mobile refactoring
- Document architectural decisions using ADR format suggested in prompts

### Release Readiness Reviews
- Comprehensive review using all three prompts before major releases
- Focus on performance, security, and user experience aspects
- Validate scalability and monitoring capabilities

## Success Metrics

### Code Quality Metrics
- **Technical Debt Reduction**: Measure reduction in code smells and anti-patterns
- **Security Vulnerability Count**: Track and reduce security issues
- **Test Coverage**: Maintain high test coverage across all components
- **Code Consistency**: Ensure consistent patterns across the codebase

### Architecture Quality Metrics
- **Coupling Score**: Lower coupling between components (Target: <7/10)
- **Cohesion Score**: Higher cohesion within components (Target: >8/10)
- **Scalability Readiness**: System's ability to handle increased load
- **Maintainability Index**: Ease of making changes and additions

### Mobile Quality Metrics
- **Performance Scores**: App startup time, bundle size, memory usage
- **User Experience Scores**: Navigation flow, accessibility, offline experience
- **Platform Consistency**: Uniform experience across iOS and Android
- **Crash Rate**: Application stability (Target: <1%)

## Continuous Improvement Process

### Monthly Architecture Health Checks
1. Run abbreviated versions of all prompts on core components
2. Track metrics trends and identify degradation patterns
3. Plan architectural improvements based on findings
4. Update prompts based on new best practices and lessons learned

### Quarterly Comprehensive Reviews
1. Full review using all prompts across the entire codebase
2. Generate architectural roadmap for the next quarter
3. Identify and prioritize technical debt reduction initiatives
4. Plan team training and skill development based on findings

### Annual Strategic Assessments
1. Evaluate overall architecture fitness for business goals
2. Assess technology stack relevance and upgrade needs
3. Plan major architectural evolution initiatives
4. Review and update development standards and practices

## Team Adoption Guidelines

### Getting Started
1. **Select one prompt** to begin with (recommend starting with Code Quality)
2. **Run a pilot review** on a small, representative component
3. **Calibrate expectations** with the team on review depth and frequency
4. **Establish review workflows** integrated with existing development processes

### Scaling Adoption
1. **Train team members** on using the prompts effectively
2. **Integrate with code review tools** and CI/CD pipelines
3. **Create review templates** based on the prompt outputs
4. **Establish review ownership** and responsibility distribution

### Measuring Success
1. **Track review completion rates** and time investment
2. **Measure improvement in code quality metrics** over time
3. **Monitor reduction in production issues** and technical debt
4. **Assess team satisfaction** with review processes and outcomes

## Customization Notes

### Adapting for Your Context
- **Modify severity thresholds** based on your quality standards
- **Add domain-specific checks** for your business requirements
- **Integrate with existing tools** and development workflows
- **Adjust review frequency** based on team capacity and project phase

### Extending the Prompts
- **Add new focus areas** as your system evolves
- **Include emerging best practices** from the React Native and Go communities
- **Incorporate lessons learned** from production issues and user feedback
- **Update technology-specific sections** as frameworks and tools evolve

---
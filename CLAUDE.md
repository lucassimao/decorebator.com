# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Decorebator is an AI-powered vocabulary learning platform that uses AI-powered enrichment and the Leitner spaced repetition system to help users master new languages effectively. It consists of:
- **API Backend** (Go/Gin) - RESTful API with PostgreSQL database, River queue system, and AI integrations
- **Mobile App** (React Native/Expo) - Cross-platform mobile application with offline support
- **Web Frontend** (Next.js) - Landing page and web application

## Memories
- read README.md for more additional context on decorebator project
- read api/docs/ANALYTICS_PERFORMANCE_SCALABILITY_REPORT.md for analytics system architecture and future plans
- read mobile/docs/mobile-app-architecture.md for detailed mobile app patterns and design system
- read api/docs/ANALYTICS_TESTING_IMPLEMENTATION.md for comprehensive analytics testing patterns
- Update README.md right after introducing major features or refactorings
- Update relevant documentation after making architectural changes
- Fixed critical analytics bugs in January 2025: streak calculations, box distribution logic, response time filtering, word count inconsistencies, and removed unused database fields
- Analytics repository functions reviewed and corrected: GetWordlistCurrentStreak, GetAllWordlistsProgress, GetCurrentBoxDistribution, GetPracticeTime, GetQuizTypePerformance, GetWordlistMasteryStats
- Implemented comprehensive analytics integration tests covering all edge cases and data scenarios
- Added batch analytics endpoint `/analytics/progress-summary` reducing mobile API calls from 8 to 1
- use api/Makefile and api/scripts/run-tests.sh as the source for the main automations and commands in this monorepo

[Rest of the existing file content continues...]
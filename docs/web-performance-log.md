# Web Performance Log

This file records production performance audits for the web app so we can compare future changes.

## 2026-01-18 — localhost:3000 (production build)

**Environment**
- Build: `pnpm build && pnpm start`
- URL: `http://localhost:3000/en`
- Tooling: Chrome DevTools MCP performance trace

**Desktop (default viewport)**
- LCP: 97 ms
- TTFB: 6 ms
- Render delay: 91 ms
- CLS: 0.00

**Mobile viewport (390×844)**
- LCP: 100 ms
- TTFB: 6 ms
- Render delay: 93 ms
- CLS: 0.00

**Throttled (Fast 3G + 4× CPU slowdown)**
- LCP: 799 ms
- TTFB: 4 ms
- Render delay: 795 ms
- CLS: 0.00

**Notes**
- LCP element: hero H1 text
- No render-blocking savings flagged in trace summaries
- DOM size insight appeared under throttling (monitor if DOM grows)

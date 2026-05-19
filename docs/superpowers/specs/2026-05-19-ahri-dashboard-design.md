# Ahri Dashboard — Design Spec

**Date:** 2026-05-19  
**Status:** Approved  
**Location:** `tools/ahri-dashboard/`

## Overview

A personal activity dashboard hosted on Netlify or Vercel. Displays step data from the ahri-health-bridge API as a single square widget: a GitHub-style contribution heatmap with a step count header and stats footer. Dark slate colour palette matching Ahri-Track.

## Architecture

**Tech stack:** Vite + vanilla JS. No framework. Output is plain static files deployed from `dist/`.

```
tools/ahri-dashboard/
├── index.html        # entry point, mounts widget
├── main.js           # data fetch, processing, rendering
├── style.css         # widget styles
└── vite.config.js    # minimal config
```

**Build:** `npm run build` → `dist/` → deploy to Netlify or Vercel.

**Env vars** (set in Netlify/Vercel dashboard, injected at build time):
- `VITE_API_KEY` — X-API-Key for ahri-health-bridge
- `VITE_API_URL` — base URL of the Go service (e.g. `https://ahri-health-bridge.onrender.com`)

## Data Layer

**Endpoint:** `GET $VITE_API_URL/health/steps/daily`  
**Auth:** `X-API-Key: $VITE_API_KEY`  
**Response:** sorted array of `{ date: "YYYY-MM-DD", steps: number }`

**Client-side processing:**
1. Build a `Map<string, number>` (date → steps) from the response
2. Generate the last 12 weeks of dates (Mon–Sun columns) client-side, anchored to today
3. For each cell, look up steps from the map or default to 0
4. Derive stats:
   - **Today's steps** — map lookup for today's date
   - **Goal %** — `Math.round(todaySteps / 10000 * 100)`
   - **Active days** — count of days with steps > 0 within the displayed 12-week window
   - **Avg steps/day** — total steps / active days within the 12-week window (0 if no active days)
   - **Missed days** — count of past dates within the 12-week window (excluding today and future) with steps = 0

**Colour intensity mapping** (5 levels, Ahri-Track palette):

| Steps | Colour | Note |
|---|---|---|
| 0 | `#1d293d` | empty cell |
| 1–2,999 | `#7bf1a8` at 25% opacity | low |
| 3,000–5,999 | `#7bf1a8` at 55% opacity | building |
| 6,000–9,999 | `#00c758` at 75% opacity | good |
| ≥10,000 | `#00c758` at 100% + `box-shadow: 0 0 6px #00c75888` | goal hit |

Future dates render as `#1d293d` with no intensity.

## Widget Layout

Fixed **240×240px** square, centred on the page. `border-radius: 20px`, `background: #0f172b`, `border: 1px solid #314158`. Font: `ui-monospace, monospace` (Geist Mono fallback).

### Header
- Left: today's step count, `font-size: 22px`, `font-weight: 900`, `letter-spacing: -0.05em`
- Below count: `"steps · today"` label, `8px`, `#62748e`, uppercase, wide tracking
- Right: `"X% of goal"` in `#00c758` bold, `"↑/↓ Y% avg"` in `#62748e` below

### Heatmap
- **Month labels row**: spans columns, left-aligned per month boundary, `6.5px`, `#45556c`
- **Grid**: `14px` day-label column + 12 week columns, `gap: 2px`
- **Day labels**: M / W / F only (alternating visible/invisible rows to align), `6px`, `#314158`
- **Cells**: `height: 13px`, `border-radius: 2px`, colour by intensity
- **Today's cell**: full `#00c758` + green glow
- **Future cells**: `#1d293d` at 40% opacity

### Footer
Top border `#1d293d`. Three stats evenly spaced:
- **Active days** — value in `#00c758`
- **Avg/day** — value in `#cad5e2`
- **Missed** — value in `#ff8b1a`

Labels: `6.5px`, `#62748e`, `letter-spacing: 0.05em`.

## States

- **Loading**: cells render as `#1d293d` skeletons, header shows `—` for counts. No spinner.
- **Error**: small `"Could not load data"` message centred below the widget in `#62748e`. No retry — user refreshes.
- **No data**: widget renders with all empty cells, stats show zeros.

## Deployment

1. Create new site on Netlify or Vercel pointing to `tools/ahri-dashboard/`
2. Set build command: `npm run build`, publish dir: `dist`
3. Add `VITE_API_KEY` and `VITE_API_URL` in the dashboard env var settings
4. Deploy — no CI needed for a personal tool, manual deploy or git-push trigger both fine

## Out of Scope

- Authentication / login (personal dashboard, public URL is acceptable)
- Multiple metric types (sleep, weight) — add later via new widgets
- Responsive / mobile layout — fixed square widget is the goal
- Tooltip on hover — future enhancement

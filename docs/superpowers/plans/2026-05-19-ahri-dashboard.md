# Ahri Dashboard Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a 240×240px GitHub-style step heatmap widget hosted on Netlify/Vercel, pulling from the ahri-health-bridge API.

**Architecture:** Vite + vanilla JS, split into three modules: `data.js` (pure functions, fully testable), `render.js` (DOM), and `main.js` (orchestration). Vitest for unit tests on the data layer.

**Tech Stack:** Vite 5, Vitest 1, vanilla JS (ES modules), no framework.

---

## File Map

| File | Responsibility |
|---|---|
| `tools/ahri-dashboard/package.json` | deps, scripts |
| `tools/ahri-dashboard/vite.config.js` | vite + vitest config |
| `tools/ahri-dashboard/index.html` | HTML entry point |
| `tools/ahri-dashboard/.env.example` | documents required env vars |
| `tools/ahri-dashboard/data.js` | pure functions: date grid, step map, stats, cell colour |
| `tools/ahri-dashboard/render.js` | DOM: skeleton, full render, error state |
| `tools/ahri-dashboard/main.js` | fetch, orchestrate, wire states |
| `tools/ahri-dashboard/style.css` | page + widget styles |
| `tools/ahri-dashboard/tests/data.test.js` | unit tests for data.js |
| `tools/ahri-dashboard/netlify.toml` | build config for Netlify |

---

## Task 1: Scaffold the project

**Files:**
- Create: `tools/ahri-dashboard/package.json`
- Create: `tools/ahri-dashboard/vite.config.js`
- Create: `tools/ahri-dashboard/index.html`
- Create: `tools/ahri-dashboard/.env.example`
- Create: `tools/ahri-dashboard/netlify.toml`

- [ ] **Step 1: Create the directory**

```bash
mkdir -p tools/ahri-dashboard/tests
```

- [ ] **Step 2: Write package.json**

Create `tools/ahri-dashboard/package.json`:

```json
{
  "name": "ahri-dashboard",
  "private": true,
  "type": "module",
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview",
    "test": "vitest run"
  },
  "devDependencies": {
    "vite": "^5.4.0",
    "vitest": "^1.6.0"
  }
}
```

- [ ] **Step 3: Write vite.config.js**

Create `tools/ahri-dashboard/vite.config.js`:

```js
import { defineConfig } from 'vite';

export default defineConfig({
  test: {
    environment: 'node',
  },
});
```

- [ ] **Step 4: Write index.html**

Create `tools/ahri-dashboard/index.html`:

```html
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <title>Ahri Dashboard</title>
</head>
<body>
  <div id="widget"></div>
  <div id="error"></div>
  <script type="module" src="/main.js"></script>
</body>
</html>
```

- [ ] **Step 5: Write .env.example**

Create `tools/ahri-dashboard/.env.example`:

```
VITE_API_KEY=your-api-key-here
VITE_API_URL=https://your-ahri-health-bridge-url
```

- [ ] **Step 6: Write netlify.toml**

Create `tools/ahri-dashboard/netlify.toml`:

```toml
[build]
  base    = "tools/ahri-dashboard"
  command = "npm run build"
  publish = "dist"
```

- [ ] **Step 7: Install dependencies**

```bash
cd tools/ahri-dashboard && npm install
```

Expected: `node_modules/` created, no errors.

- [ ] **Step 8: Commit scaffold**

```bash
git add tools/ahri-dashboard/
git commit -m "feat: scaffold ahri-dashboard vite project"
```

---

## Task 2: data.js — generateWeeks

**Files:**
- Create: `tools/ahri-dashboard/data.js`
- Create: `tools/ahri-dashboard/tests/data.test.js`

`generateWeeks(today, numWeeks)` returns a 2D array: `numWeeks` columns (oldest → newest), each an array of 7 ISO date strings `YYYY-MM-DD` representing Mon–Sun. The last column is the week containing `today`.

- [ ] **Step 1: Write the failing test**

Create `tools/ahri-dashboard/tests/data.test.js`:

```js
import { describe, it, expect } from 'vitest';
import { generateWeeks } from '../data.js';

describe('generateWeeks', () => {
  it('returns exactly numWeeks columns of 7 days each', () => {
    const weeks = generateWeeks('2026-05-19', 12);
    expect(weeks).toHaveLength(12);
    for (const week of weeks) {
      expect(week).toHaveLength(7);
    }
  });

  it('last column contains today', () => {
    const weeks = generateWeeks('2026-05-19', 12);
    const lastWeek = weeks[11];
    expect(lastWeek).toContain('2026-05-19');
  });

  it('first day of every column is a Monday', () => {
    const weeks = generateWeeks('2026-05-19', 12);
    for (const week of weeks) {
      const d = new Date(week[0] + 'T00:00:00');
      expect(d.getDay()).toBe(1); // 1 = Monday
    }
  });

  it('columns are consecutive weeks', () => {
    const weeks = generateWeeks('2026-05-19', 12);
    for (let i = 1; i < weeks.length; i++) {
      const prev = new Date(weeks[i - 1][0] + 'T00:00:00');
      const curr = new Date(weeks[i][0] + 'T00:00:00');
      const diff = (curr - prev) / (1000 * 60 * 60 * 24);
      expect(diff).toBe(7);
    }
  });

  it('works when today is a Sunday', () => {
    const weeks = generateWeeks('2026-05-17', 4); // 2026-05-17 is a Sunday
    const lastWeek = weeks[3];
    expect(lastWeek[0]).toBe('2026-05-11'); // Monday of that week
    expect(lastWeek[6]).toBe('2026-05-17'); // Sunday = today
  });

  it('works when today is a Monday', () => {
    const weeks = generateWeeks('2026-05-18', 4); // 2026-05-18 is a Monday
    const lastWeek = weeks[3];
    expect(lastWeek[0]).toBe('2026-05-18');
  });
});
```

- [ ] **Step 2: Run test to confirm it fails**

```bash
cd tools/ahri-dashboard && npm test
```

Expected: FAIL — `generateWeeks` not found.

- [ ] **Step 3: Implement generateWeeks in data.js**

Create `tools/ahri-dashboard/data.js`:

```js
/**
 * Returns numWeeks columns of 7 ISO date strings each (Mon–Sun).
 * The last column is the week containing `today`.
 * @param {string} today - ISO date string YYYY-MM-DD
 * @param {number} numWeeks
 * @returns {string[][]}
 */
export function generateWeeks(today, numWeeks = 12) {
  const todayDate = new Date(today + 'T00:00:00');
  // getDay(): 0=Sun 1=Mon...6=Sat. Convert to Mon=0...Sun=6.
  const dayOfWeek = todayDate.getDay();
  const daysFromMonday = (dayOfWeek + 6) % 7;

  // Monday of the current week
  const mondayThisWeek = new Date(todayDate);
  mondayThisWeek.setDate(todayDate.getDate() - daysFromMonday);

  // Monday of the first week (numWeeks - 1 weeks ago)
  const startMonday = new Date(mondayThisWeek);
  startMonday.setDate(mondayThisWeek.getDate() - (numWeeks - 1) * 7);

  const weeks = [];
  for (let w = 0; w < numWeeks; w++) {
    const week = [];
    for (let d = 0; d < 7; d++) {
      const date = new Date(startMonday);
      date.setDate(startMonday.getDate() + w * 7 + d);
      week.push(date.toISOString().slice(0, 10));
    }
    weeks.push(week);
  }
  return weeks;
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
cd tools/ahri-dashboard && npm test
```

Expected: all 6 `generateWeeks` tests pass.

- [ ] **Step 5: Commit**

```bash
git add tools/ahri-dashboard/data.js tools/ahri-dashboard/tests/data.test.js
git commit -m "feat: add generateWeeks to data.js"
```

---

## Task 3: data.js — buildStepMap and computeStats

**Files:**
- Modify: `tools/ahri-dashboard/data.js`
- Modify: `tools/ahri-dashboard/tests/data.test.js`

- [ ] **Step 1: Write the failing tests**

Append to `tools/ahri-dashboard/tests/data.test.js`:

```js
import { buildStepMap, computeStats } from '../data.js';

describe('buildStepMap', () => {
  it('converts API array to a date→steps Map', () => {
    const data = [
      { date: '2026-05-18', steps: 8000 },
      { date: '2026-05-19', steps: 9241 },
    ];
    const map = buildStepMap(data);
    expect(map.get('2026-05-18')).toBe(8000);
    expect(map.get('2026-05-19')).toBe(9241);
    expect(map.get('2026-05-17')).toBeUndefined();
  });

  it('returns empty Map for empty array', () => {
    expect(buildStepMap([])).toBeInstanceOf(Map);
    expect(buildStepMap([]).size).toBe(0);
  });
});

describe('computeStats', () => {
  // weeks: 2 weeks, Mon 2026-05-11 → Sun 2026-05-17, Mon 2026-05-18 → Sun 2026-05-24
  // today: 2026-05-19 (Tuesday)
  const weeks = [
    ['2026-05-11','2026-05-12','2026-05-13','2026-05-14','2026-05-15','2026-05-16','2026-05-17'],
    ['2026-05-18','2026-05-19','2026-05-20','2026-05-21','2026-05-22','2026-05-23','2026-05-24'],
  ];
  const today = '2026-05-19';

  it('calculates todaySteps and goalPct', () => {
    const map = buildStepMap([{ date: '2026-05-19', steps: 9000 }]);
    const stats = computeStats(weeks, map, today, 10000);
    expect(stats.todaySteps).toBe(9000);
    expect(stats.goalPct).toBe(90);
  });

  it('counts activeDays (steps > 0, within window)', () => {
    const map = buildStepMap([
      { date: '2026-05-11', steps: 5000 },
      { date: '2026-05-18', steps: 8000 },
      { date: '2026-05-19', steps: 9000 },
    ]);
    const stats = computeStats(weeks, map, today, 10000);
    expect(stats.activeDays).toBe(3);
  });

  it('counts missedDays (past days with 0 steps, excludes today and future)', () => {
    const map = buildStepMap([
      { date: '2026-05-11', steps: 5000 },
    ]);
    // Past days with 0: 12,13,14,15,16,17,18 = 7 days
    const stats = computeStats(weeks, map, today, 10000);
    expect(stats.missedDays).toBe(7);
  });

  it('returns 0 avgSteps when no active days', () => {
    const map = buildStepMap([]);
    const stats = computeStats(weeks, map, today, 10000);
    expect(stats.avgSteps).toBe(0);
  });

  it('excludes future dates from all counts', () => {
    const map = buildStepMap([{ date: '2026-05-22', steps: 9000 }]);
    const stats = computeStats(weeks, map, today, 10000);
    expect(stats.activeDays).toBe(0);
    expect(stats.todaySteps).toBe(0);
  });

  it('computes avgDiff as percentage vs window avg', () => {
    const map = buildStepMap([
      { date: '2026-05-11', steps: 8000 },
      { date: '2026-05-18', steps: 8000 },
      { date: '2026-05-19', steps: 12000 }, // today: 50% above avg of 8000
    ]);
    const stats = computeStats(weeks, map, today, 10000);
    expect(stats.avgSteps).toBe(9333); // Math.round((8000+8000+12000)/3)
    expect(stats.avgDiff).toBe(29); // Math.round((12000-9333)/9333*100)
  });
});
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
cd tools/ahri-dashboard && npm test
```

Expected: FAIL — `buildStepMap` and `computeStats` not found.

- [ ] **Step 3: Implement buildStepMap and computeStats**

Append to `tools/ahri-dashboard/data.js`:

```js
/**
 * Converts API response array to a Map<dateString, steps>.
 * @param {{ date: string, steps: number }[]} apiData
 * @returns {Map<string, number>}
 */
export function buildStepMap(apiData) {
  const map = new Map();
  for (const { date, steps } of apiData) {
    map.set(date, steps);
  }
  return map;
}

/**
 * Derives display stats from the 12-week window.
 * @param {string[][]} weeks
 * @param {Map<string, number>} stepMap
 * @param {string} today - ISO date YYYY-MM-DD
 * @param {number} goal
 */
export function computeStats(weeks, stepMap, today, goal = 10000) {
  const todaySteps = stepMap.get(today) ?? 0;
  const goalPct = Math.round((todaySteps / goal) * 100);

  let totalSteps = 0;
  let activeDays = 0;
  let missedDays = 0;

  for (const week of weeks) {
    for (const date of week) {
      if (date > today) continue; // future — skip
      const steps = stepMap.get(date) ?? 0;
      totalSteps += steps;
      if (steps > 0) {
        activeDays++;
      } else if (date < today) {
        missedDays++;
      }
    }
  }

  const avgSteps = activeDays > 0 ? Math.round(totalSteps / activeDays) : 0;
  const avgDiff = avgSteps > 0 ? Math.round(((todaySteps - avgSteps) / avgSteps) * 100) : 0;

  return { todaySteps, goalPct, activeDays, avgSteps, missedDays, avgDiff };
}
```

- [ ] **Step 4: Run tests to confirm they pass**

```bash
cd tools/ahri-dashboard && npm test
```

Expected: all `buildStepMap` and `computeStats` tests pass.

- [ ] **Step 5: Commit**

```bash
git add tools/ahri-dashboard/data.js tools/ahri-dashboard/tests/data.test.js
git commit -m "feat: add buildStepMap and computeStats to data.js"
```

---

## Task 4: data.js — cellStyle

**Files:**
- Modify: `tools/ahri-dashboard/data.js`
- Modify: `tools/ahri-dashboard/tests/data.test.js`

`cellStyle(steps, date, today)` returns an object of CSS properties (camelCase) for a heatmap cell.

- [ ] **Step 1: Write the failing tests**

Append to `tools/ahri-dashboard/tests/data.test.js`:

```js
import { cellStyle } from '../data.js';

describe('cellStyle', () => {
  const today = '2026-05-19';

  it('future date: dark with reduced opacity', () => {
    const style = cellStyle(0, '2026-05-20', today);
    expect(style.background).toBe('#1d293d');
    expect(style.opacity).toBe('0.4');
  });

  it('0 steps, past: empty dark cell, no glow', () => {
    const style = cellStyle(0, '2026-05-18', today);
    expect(style.background).toBe('#1d293d');
    expect(style.boxShadow).toBeUndefined();
  });

  it('1–2999 steps: low green', () => {
    const style = cellStyle(2999, '2026-05-18', today);
    expect(style.background).toBe('rgba(123, 241, 168, 0.25)');
  });

  it('3000–5999 steps: building green', () => {
    const style = cellStyle(5000, '2026-05-18', today);
    expect(style.background).toBe('rgba(123, 241, 168, 0.55)');
  });

  it('6000–9999 steps: good green', () => {
    const style = cellStyle(8000, '2026-05-18', today);
    expect(style.background).toBe('rgba(0, 199, 88, 0.75)');
  });

  it('>=10000 steps: goal hit — full green + glow', () => {
    const style = cellStyle(10000, '2026-05-18', today);
    expect(style.background).toBe('#00c758');
    expect(style.boxShadow).toBe('0 0 6px #00c75888');
  });

  it('today with steps: normal colour + glow', () => {
    const style = cellStyle(8000, today, today);
    expect(style.background).toBe('rgba(0, 199, 88, 0.75)');
    expect(style.boxShadow).toBe('0 0 6px #00c75888');
  });

  it('today with 0 steps: lighter empty cell + glow to mark position', () => {
    const style = cellStyle(0, today, today);
    expect(style.background).toBe('#314158');
    expect(style.boxShadow).toBe('0 0 6px #00c75888');
  });
});
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
cd tools/ahri-dashboard && npm test
```

Expected: FAIL — `cellStyle` not found.

- [ ] **Step 3: Implement cellStyle**

Append to `tools/ahri-dashboard/data.js`:

```js
/**
 * Returns inline style object for a heatmap cell.
 * @param {number} steps
 * @param {string} date - ISO date YYYY-MM-DD
 * @param {string} today - ISO date YYYY-MM-DD
 * @returns {Object} CSS properties in camelCase
 */
export function cellStyle(steps, date, today) {
  const isFuture = date > today;
  const isToday = date === today;

  if (isFuture) {
    return { background: '#1d293d', opacity: '0.4' };
  }

  let style = {};

  if (steps === 0) {
    style.background = isToday ? '#314158' : '#1d293d';
  } else if (steps >= 10000) {
    style.background = '#00c758';
    style.boxShadow = '0 0 6px #00c75888';
  } else if (steps >= 6000) {
    style.background = 'rgba(0, 199, 88, 0.75)';
  } else if (steps >= 3000) {
    style.background = 'rgba(123, 241, 168, 0.55)';
  } else {
    style.background = 'rgba(123, 241, 168, 0.25)';
  }

  if (isToday && !style.boxShadow) {
    style.boxShadow = '0 0 6px #00c75888';
  }

  return style;
}
```

- [ ] **Step 4: Run all tests**

```bash
cd tools/ahri-dashboard && npm test
```

Expected: all tests pass (generateWeeks + buildStepMap + computeStats + cellStyle).

- [ ] **Step 5: Commit**

```bash
git add tools/ahri-dashboard/data.js tools/ahri-dashboard/tests/data.test.js
git commit -m "feat: add cellStyle to data.js"
```

---

## Task 5: style.css

**Files:**
- Create: `tools/ahri-dashboard/style.css`

- [ ] **Step 1: Write style.css**

Create `tools/ahri-dashboard/style.css`:

```css
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

body {
  background: #0a0f1a;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  font-family: ui-monospace, 'Geist Mono', monospace;
}

#widget {
  width: 240px;
  height: 240px;
  background: #0f172b;
  border: 1px solid #314158;
  border-radius: 20px;
  padding: 16px 18px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

/* Header */
.widget-header    { display: flex; justify-content: space-between; align-items: flex-start; }
.step-count       { font-size: 22px; font-weight: 900; letter-spacing: -0.05em; color: #fff; line-height: 1; }
.step-label       { font-size: 8px; color: #62748e; text-transform: uppercase; letter-spacing: 0.15em; margin-top: 2px; }
.goal-pct         { font-size: 11px; font-weight: 700; color: #00c758; text-align: right; }
.avg-diff         { font-size: 8px; color: #62748e; text-align: right; margin-top: 2px; }

/* Month labels */
.month-labels     { display: grid; grid-template-columns: 14px repeat(12, 1fr); gap: 2px; }
.month-label      { font-size: 6.5px; color: #45556c; }

/* Heatmap */
.heatmap          { display: grid; grid-template-columns: 14px repeat(12, 1fr); gap: 2px; flex: 1; align-content: start; }
.day-labels       { display: flex; flex-direction: column; justify-content: space-between; padding: 1px 0; }
.day-label        { font-size: 6px; color: #314158; }
.week-col         { display: flex; flex-direction: column; gap: 2px; }
.cell             { height: 13px; border-radius: 2px; }

/* Footer */
.widget-footer    { display: flex; justify-content: space-between; border-top: 1px solid #1d293d; padding-top: 6px; }
.stat             { text-align: center; }
.stat-value       { font-size: 11px; font-weight: 900; line-height: 1; }
.stat-label       { font-size: 6.5px; color: #62748e; letter-spacing: 0.05em; margin-top: 2px; }

/* Error */
#error            { color: #62748e; font-size: 11px; text-align: center; margin-top: 12px; min-height: 16px; }
```

- [ ] **Step 2: Commit**

```bash
git add tools/ahri-dashboard/style.css
git commit -m "feat: add widget styles"
```

---

## Task 6: render.js

**Files:**
- Create: `tools/ahri-dashboard/render.js`

Three exported functions: `renderSkeleton()`, `renderWidget(weeks, stepMap, stats, today)`, `renderError()`.

- [ ] **Step 1: Write render.js**

Create `tools/ahri-dashboard/render.js`:

```js
import { cellStyle } from './data.js';

// Day labels: M/W/F visible, others empty strings to maintain grid alignment
const DAY_LABELS = ['M', '', 'W', '', 'F', '', 'S'];

/**
 * Renders loading skeleton — dark cells, dashes for counts.
 */
export function renderSkeleton() {
  const widget = document.getElementById('widget');
  widget.innerHTML = `
    <div class="widget-header">
      <div>
        <div class="step-count">—</div>
        <div class="step-label">steps · today</div>
      </div>
      <div>
        <div class="goal-pct">—% of goal</div>
        <div class="avg-diff">—</div>
      </div>
    </div>
    <div class="month-labels">
      <div></div>${Array(12).fill('<div class="month-label"></div>').join('')}
    </div>
    <div class="heatmap">
      <div class="day-labels">
        ${DAY_LABELS.map(l => `<div class="day-label">${l}</div>`).join('')}
      </div>
      ${Array(12).fill(null).map(() => `
        <div class="week-col">
          ${Array(7).fill('<div class="cell" style="background:#1d293d"></div>').join('')}
        </div>
      `).join('')}
    </div>
    <div class="widget-footer">
      <div class="stat">
        <div class="stat-value" style="color:#00c758">—</div>
        <div class="stat-label">active</div>
      </div>
      <div class="stat">
        <div class="stat-value" style="color:#cad5e2">—</div>
        <div class="stat-label">avg/day</div>
      </div>
      <div class="stat">
        <div class="stat-value" style="color:#ff8b1a">—</div>
        <div class="stat-label">missed</div>
      </div>
    </div>
  `;
}

/**
 * Renders the widget with real data.
 * @param {string[][]} weeks
 * @param {Map<string, number>} stepMap
 * @param {{ todaySteps, goalPct, activeDays, avgSteps, missedDays, avgDiff }} stats
 * @param {string} today
 */
export function renderWidget(weeks, stepMap, stats, today) {
  const { todaySteps, goalPct, activeDays, avgSteps, missedDays, avgDiff } = stats;
  const avgSign = avgDiff >= 0 ? '↑' : '↓';
  const monthLabels = buildMonthLabels(weeks);

  const widget = document.getElementById('widget');
  widget.innerHTML = `
    <div class="widget-header">
      <div>
        <div class="step-count">${todaySteps.toLocaleString()}</div>
        <div class="step-label">steps · today</div>
      </div>
      <div>
        <div class="goal-pct">${goalPct}% of goal</div>
        <div class="avg-diff">${avgSign} ${Math.abs(avgDiff)}% avg</div>
      </div>
    </div>
    <div class="month-labels">
      <div></div>
      ${monthLabels.map(l => `<div class="month-label">${l}</div>`).join('')}
    </div>
    <div class="heatmap">
      <div class="day-labels">
        ${DAY_LABELS.map(l => `<div class="day-label">${l}</div>`).join('')}
      </div>
      ${weeks.map(week => `
        <div class="week-col">
          ${week.map(date => {
            const steps = stepMap.get(date) ?? 0;
            const s = cellStyle(steps, date, today);
            const styleStr = styleObjToString(s);
            return `<div class="cell" style="${styleStr}"></div>`;
          }).join('')}
        </div>
      `).join('')}
    </div>
    <div class="widget-footer">
      <div class="stat">
        <div class="stat-value" style="color:#00c758">${activeDays}</div>
        <div class="stat-label">active</div>
      </div>
      <div class="stat">
        <div class="stat-value" style="color:#cad5e2">${avgSteps.toLocaleString()}</div>
        <div class="stat-label">avg/day</div>
      </div>
      <div class="stat">
        <div class="stat-value" style="color:#ff8b1a">${missedDays}</div>
        <div class="stat-label">missed</div>
      </div>
    </div>
  `;
}

/**
 * Shows the error message below the widget.
 */
export function renderError() {
  document.getElementById('error').textContent = 'Could not load data';
}

// --- helpers ---

function buildMonthLabels(weeks) {
  let lastMonth = null;
  return weeks.map(week => {
    const month = week[0].slice(0, 7); // YYYY-MM
    if (month !== lastMonth) {
      lastMonth = month;
      return new Date(week[0] + 'T00:00:00').toLocaleString('en', { month: 'short' });
    }
    return '';
  });
}

function styleObjToString(obj) {
  return Object.entries(obj)
    .map(([k, v]) => `${k.replace(/([A-Z])/g, '-$1').toLowerCase()}:${v}`)
    .join(';');
}
```

- [ ] **Step 2: Commit**

```bash
git add tools/ahri-dashboard/render.js
git commit -m "feat: add render.js — skeleton, widget, error states"
```

---

## Task 7: main.js — fetch and orchestrate

**Files:**
- Create: `tools/ahri-dashboard/main.js`

- [ ] **Step 1: Write main.js**

Create `tools/ahri-dashboard/main.js`:

```js
import './style.css';
import { generateWeeks, buildStepMap, computeStats } from './data.js';
import { renderSkeleton, renderWidget, renderError } from './render.js';

const API_KEY = import.meta.env.VITE_API_KEY;
const API_URL = import.meta.env.VITE_API_URL;
const GOAL = 10000;
const NUM_WEEKS = 12;

async function init() {
  renderSkeleton();

  const today = new Date().toISOString().slice(0, 10);
  const weeks = generateWeeks(today, NUM_WEEKS);

  try {
    const res = await fetch(`${API_URL}/health/steps/daily`, {
      headers: { 'X-API-Key': API_KEY },
    });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();

    const stepMap = buildStepMap(data);
    const stats = computeStats(weeks, stepMap, today, GOAL);
    renderWidget(weeks, stepMap, stats, today);
  } catch (err) {
    console.error('Failed to load step data:', err);
    renderError();
  }
}

init();
```

- [ ] **Step 2: Commit**

```bash
git add tools/ahri-dashboard/main.js
git commit -m "feat: add main.js — fetch, orchestrate, wire states"
```

---

## Task 8: Smoke test with dev server

- [ ] **Step 1: Create a local .env file**

Create `tools/ahri-dashboard/.env` (not committed — already in .gitignore via `*.env` or add it):

```
VITE_API_KEY=<your real API key>
VITE_API_URL=<your real Go service URL>
```

Verify `.env` is gitignored:

```bash
echo ".env" >> tools/ahri-dashboard/.gitignore
git add tools/ahri-dashboard/.gitignore
git commit -m "chore: gitignore .env in ahri-dashboard"
```

- [ ] **Step 2: Start dev server**

```bash
cd tools/ahri-dashboard && npm run dev
```

Expected: Vite prints `Local: http://localhost:5173/`. Open in browser.

- [ ] **Step 3: Verify the widget**

Check in browser:
- Widget is centred on dark background
- Skeleton flashes briefly then real data appears
- Heatmap grid shows 12 columns × 7 rows of coloured cells
- Today's cell glows green
- Month labels appear at column boundaries
- Header shows today's step count and goal %
- Footer shows active / avg / missed stats

- [ ] **Step 4: Test error state**

Temporarily set `VITE_API_KEY=wrong` in `.env`, restart the dev server (`Ctrl+C`, `npm run dev`). Verify `"Could not load data"` appears below the widget. Restore correct key and restart.

---

## Task 9: Build and deploy

- [ ] **Step 1: Run the build**

```bash
cd tools/ahri-dashboard && npm run build
```

Expected: `dist/` directory created, no errors.

- [ ] **Step 2: Preview the build locally**

```bash
cd tools/ahri-dashboard && npm run preview
```

Open the printed URL. Verify the built widget looks identical to the dev version.

- [ ] **Step 3: Deploy to Netlify**

On netlify.com: New site → Import from Git → select the `ahri` repo → set:
- Base directory: `tools/ahri-dashboard`
- Build command: `npm run build`
- Publish directory: `dist`

Then in Site settings → Environment variables, add:
- `VITE_API_KEY` = your API key
- `VITE_API_URL` = your Go service URL

Trigger a deploy. Verify the live URL shows the widget with real data.

- [ ] **Step 4: Final commit**

```bash
git add tools/ahri-dashboard/netlify.toml
git commit -m "feat: ahri-dashboard complete — heatmap widget, Netlify deploy config"
```

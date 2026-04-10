# GoAdmin — Implementation Specification for Coding Agent

> **Reference prototype:** `design.html` — interactive prototype with all screens and three themes. Open it in a browser and use it as a visual reference during implementation.

---

## 1. General Architecture

- **Framework:** Bootstrap 5.3.3 (CDN or local)
- **Fonts:** IBM Plex Sans (400, 500, 600, 700) + JetBrains Mono (400, 500) — load via Google Fonts
- **Rendering:** Server-rendered HTML (Go templates), full page reload on every action
- **JS:** Minimal — Bootstrap Bundle + `confirm()` for deletion + theme switcher
- **Themes:** 3 themes (light / medium / dark), toggled via `data-theme` attribute on `<html>`

---

## 2. Theme System

The theme is set via the `data-theme` attribute on `<html>`. All colors are sourced from CSS variables. **No hardcoded colors in components.**

### 2.1 CSS Variables

```css
:root {
  --bs-body-font-family: 'IBM Plex Sans', system-ui, sans-serif;
  --bs-font-monospace: 'JetBrains Mono', monospace;
}

/* ===== LIGHT ===== */
[data-theme="light"] {
  --navbar-bg: #2563eb;
  --navbar-text: #fff;
  --page-bg: #f8f9fb;
  --surface: #ffffff;
  --surface-alt: #f1f3f5;
  --border: #e2e5ea;
  --text-main: #1a1a2e;
  --text-muted: #6b7280;
  --accent: #2563eb;
  --accent-hover: #1d4ed8;
  --success: #16a34a;
  --success-hover: #15803d;
  --danger: #dc2626;
  --warning: #f59e0b;
  --info: #0891b2;
  --secondary: #6b7280;
  --footer-bg: #f1f3f5;
  --card-shadow: 0 1px 3px rgba(0,0,0,.06);
  --heading-color: #1e3a5f;
  --placeholder-border: #d1d5db;
  --placeholder-text: #9ca3af;
}

/* ===== MEDIUM (TWILIGHT) ===== */
[data-theme="medium"] {
  --navbar-bg: #334155;
  --navbar-text: #f1f5f9;
  --page-bg: #e2e8f0;
  --surface: #f1f5f9;
  --surface-alt: #e2e8f0;
  --border: #cbd5e1;
  --text-main: #1e293b;
  --text-muted: #64748b;
  --accent: #4f46e5;
  --accent-hover: #4338ca;
  --success: #0f766e;
  --success-hover: #0d5e58;
  --danger: #b91c1c;
  --warning: #d97706;
  --info: #0e7490;
  --secondary: #64748b;
  --footer-bg: #cbd5e1;
  --card-shadow: 0 1px 4px rgba(0,0,0,.08);
  --heading-color: #334155;
  --placeholder-border: #94a3b8;
  --placeholder-text: #94a3b8;
}

/* ===== DARK ===== */
[data-theme="dark"] {
  --navbar-bg: #0f172a;
  --navbar-text: #e2e8f0;
  --page-bg: #0f172a;
  --surface: #1e293b;
  --surface-alt: #273548;
  --border: #334155;
  --text-main: #e2e8f0;
  --text-muted: #94a3b8;
  --accent: #6272a4;
  --accent-hover: #546196;
  --success: #2f7a5c;
  --success-hover: #276a4e;
  --danger: #b45454;
  --warning: #b8860b;
  --info: #3a7c8c;
  --secondary: #64748b;
  --footer-bg: #1e293b;
  --card-shadow: 0 2px 8px rgba(0,0,0,.3);
  --heading-color: #93c5fd;
  --placeholder-border: #475569;
  --placeholder-text: #64748b;
}
```

### 2.2 Muted Badges for Dark/Medium Themes

In the dark and medium themes, badges use darker shades to avoid being overly loud:

```css
/* Dark theme badges */
[data-theme="dark"] .badge.bg-success,
[data-theme="dark"] .badge-status-active { background-color: #1a5c3a !important; }
[data-theme="dark"] .badge.bg-primary { background-color: #3b4fa0 !important; }
[data-theme="dark"] .badge.bg-info,
[data-theme="dark"] .badge-status-new { background-color: #155e75 !important; }
[data-theme="dark"] .badge.bg-danger,
[data-theme="dark"] .badge-status-blocked { background-color: #7f1d1d !important; }
[data-theme="dark"] .badge.bg-warning { background-color: #92600a !important; }
[data-theme="dark"] .badge.bg-secondary { background-color: #475569 !important; }

/* Medium theme badges */
[data-theme="medium"] .badge.bg-success,
[data-theme="medium"] .badge-status-active { background-color: #166534 !important; }
[data-theme="medium"] .badge.bg-primary { background-color: #4338ca !important; }
[data-theme="medium"] .badge.bg-info,
[data-theme="medium"] .badge-status-new { background-color: #0e6e87 !important; }
[data-theme="medium"] .badge.bg-danger,
[data-theme="medium"] .badge-status-blocked { background-color: #991b1b !important; }
[data-theme="medium"] .badge.bg-warning { background-color: #a16207 !important; }
```

### 2.3 Theme Switcher

Location: **in the footer**, right-aligned. Three round buttons 28×28px: ☀ (light), ◐ (medium), ● (dark).

```html
<div class="footer-theme-selector">
  <button class="ft-btn active-ft" onclick="setTheme('light')" title="Light">☀</button>
  <button class="ft-btn" onclick="setTheme('medium')" title="Medium">◐</button>
  <button class="ft-btn" onclick="setTheme('dark')" title="Dark">●</button>
</div>
```

JS for switching:
```javascript
function setTheme(theme) {
  document.documentElement.setAttribute('data-theme', theme);
  // update active-ft on buttons
}
```

Persist the theme in `localStorage` and restore it on page load.

---

## 3. General Layout

All authenticated pages:

```
<div class="page-wrapper">     <!-- min-height: 100vh; display: flex; flex-direction: column -->
  <nav>...</nav>                <!-- navbar -->
  <div class="breadcrumb-bar">  <!-- breadcrumbs -->
  <main class="container-fluid py-4">  <!-- flex: 1, expands to fill -->
    ...content...
  </main>
  <footer class="site-footer">  <!-- pinned to bottom -->
</div>
```

Footer is pinned to the bottom via flexbox (sticky footer). Main expands to fill the available space.

---

## 4. Navbar

Class: `navbar navbar-expand-lg navbar-goadmin`

| Element | Description |
|---------|-------------|
| Brand | Text "GoAdmin", left-aligned, `font-weight: 700` |
| Nav items | Right-aligned (`ms-auto`), active item has `border-bottom: 2px solid` |
| Logout | `btn-link` styled as a nav-link, submits a POST form |
| Mobile | Hamburger toggler `navbar-toggler`, items stack vertically |

Nav item visibility by role:
- `owner` / `root`: Users, Audit Log, Logout [Name]
- `user`: Logout [Name] only
- Anonymous: Login only

---

## 5. Screens

### 5.1 Login (`/login`)

- Centered **card** (`max-width: 460px; margin: 0 auto`)
- `card-header`: text "Login", centered
- `card-body`: Email and Password fields — `form-control form-control-lg`
- `card-footer`: Login button — `btn btn-success btn-lg`, right-aligned
- **Error:** `alert alert-danger` inside `card-body`, above the fields. Email is pre-filled.

### 5.2 Dashboard (`/`)

- Heading `h3`: "Welcome, [Name]" — `font-weight: 700; color: var(--heading-color)`
- Subtitle: user role — `color: var(--text-muted)`
- Placeholder block: `border: 2px dashed var(--placeholder-border)`, text "Dashboard content"
- No widgets — this area is reserved for customization by the library consumer

### 5.3 User List (`/users`)

Consists of:
1. **Filter bar** — in a standalone `.filter-bar` block (not inside a card)
2. **Table + pagination** — inside a `card`

#### Filter bar
```html
<div class="filter-bar">  <!-- background: var(--surface); border: 1px solid var(--border); border-radius: .5rem; padding: 1rem -->
  <div class="row g-2 align-items-end">
    <div class="col-lg-4 col-md-6"><input type="text" class="form-control" placeholder="Search by login or name…"></div>
    <div class="col-lg-2 col-md-4"><select class="form-select">...</select></div>
    <div class="col-lg-3 col-md-4">
      <button class="btn btn-primary me-1">Search</button>
      <button class="btn btn-secondary">Reset</button>
    </div>
  </div>
</div>
```

#### Table inside a card
```html
<div class="card">
  <div class="card-body p-0">
    <div class="table-responsive">
      <table class="table table-hover table-striped mb-0">
        <thead class="table-secondary">...</thead>
        <tbody>...</tbody>
      </table>
    </div>
  </div>
  <div class="card-footer d-flex justify-content-between align-items-center">
    <small class="text-muted">Page X of Y (Z users)</small>
    <nav><ul class="pagination pagination-sm mb-0">...</ul></nav>
  </div>
</div>
```

**IMPORTANT:** Style `card-footer` as: `background: var(--surface); border-top: 1px solid var(--border);`

#### Table columns
| # | Column | Description |
|---|--------|-------------|
| 1 | # | ID, ~50px |
| 2 | Login | Email, plain text (NOT a link) |
| 3 | Name | Display name |
| 4 | Role | owner / root / user |
| 5 | Status | Badge: `.badge-status-active` / `.badge-status-new` / `.badge-status-blocked` |
| 6 | Date | Stack of 3 badges (see below) |
| 7 | Actions | Edit + Delete buttons, "Create User" in header |

#### Date badges (Date column)
```html
<td class="date-badges">  <!-- display: flex; flex-direction: column; align-items: flex-start; gap: 2px -->
  <span class="badge bg-success">Created: 2024-01-15 10:30</span>
  <span class="badge bg-primary">Updated: 2024-06-20 14:00</span>
  <span class="badge bg-info text-white">Last logged: 2024-07-01 09:15</span>
</td>
```
Date badge styling: `font-size: .68rem; font-family: var(--bs-font-monospace); font-weight: 400; padding: .2em .5em;`

**Badges do NOT stretch to full width** — they are `inline`, occupying only the width of their text.

#### Actions
- "Create User" — `btn btn-info btn-sm`, in the `<th>` of the last column, `text-end`
- "Edit" — `btn btn-sm btn-outline-primary`
- "Delete" — `btn btn-sm btn-outline-danger`, with `confirm("Delete user?")`
- **Self-delete protection:** the Delete button is **NOT rendered** for the current user's own row

#### Empty state
```html
<tr><td colspan="7" class="text-center text-muted py-4">No users found</td></tr>
```

### 5.4 User Form (`/users/create`, `/users/:id/update`)

The form is wrapped in a **card**, centered: `col-lg-7 col-md-10 mx-auto`.

```html
<div class="col-lg-7 col-md-10 mx-auto">
  <div class="card">
    <div class="card-header">Create User / Edit User</div>
    <div class="card-body py-3">
      <div class="row g-3">
        <div class="col-md-7">
          <label class="form-label form-label-sm mb-1">User Name</label>
          <input type="text" class="form-control" ...>
        </div>
        <div class="col-md-5">
          <label class="form-label form-label-sm mb-1">User Role</label>
          <select class="form-select">...</select>
        </div>
        <!-- Login + Status -->
        <!-- Password (left column only) -->
      </div>
    </div>
    <div class="card-footer text-end">
      <button class="btn btn-success btn-sm">Save</button>
    </div>
  </div>
</div>
```

**Field layout in two columns:**
| Left (col-md-7) | Right (col-md-5) |
|---|---|
| User Name | User Role |
| User Login | User Status |
| User Password | — |

Field size — **standard** (`form-control`, NOT `form-control-sm`).
Labels — `form-label-sm` (font-size: .8rem; font-weight: 500).
`card-footer`: `background: var(--surface); border-top: 1px solid var(--border);`

**Validation error:** `alert alert-danger` **before** the card (not inside it).

### 5.5 Audit Log (`/audit`)

Structure is identical to User List: filter bar + card with table.

- Table: `table table-hover table-striped table-sm mb-0` (compact)
- Filter: action type dropdown + Filter / Reset buttons

#### Columns
| # | Column | Width | Description |
|---|--------|-------|-------------|
| 1 | Date | 170px | Timestamp, `<small class="text-muted">` |
| 2 | Actor | 180px | Email + `<small class="text-muted">#ID</small>` |
| 3 | Action | 140px | Badge (login=success, logout=secondary, create=primary, update=warning, delete=danger) |
| 4 | Entity ID | 80px | `#ID` or `—` |
| 5 | Details | auto | `<small>` key-value pairs |

#### Disabled state
```html
<div class="alert alert-warning">
  <strong>Audit logging is disabled.</strong> Set Config.AuditLog to enable it.
</div>
```

### 5.6 Error Pages (403, 404, 500)

**Centered card:**

```html
<main class="container-fluid d-flex align-items-center justify-content-center" style="padding:3rem 1rem;">
  <div class="col-lg-5 col-md-8">
    <div class="card text-center">
      <div class="card-body py-5">
        <h1 style="font-size:4rem; font-weight:700; color:var(--danger);">404</h1>
        <p class="text-danger fs-5 mb-2">Page not found</p>
        <p class="mb-0" style="color:var(--text-muted);">Description text.</p>
      </div>
    </div>
  </div>
</main>
```

---

## 6. Components — Full CSS

Copy **entirely** from the prototype `design.html`, `<style>` section. Below are the key rules that must not be missed:

### Body
```css
body {
  font-family: var(--bs-body-font-family);
  background: var(--page-bg);
  color: var(--text-main);
  transition: background .25s, color .25s;
}
```

### Sticky footer
```css
.page-wrapper { display: flex; flex-direction: column; min-height: 100vh; }
.page-wrapper > main { flex: 1; }
```

### Table theming (CRITICAL for dark mode)
```css
.table {
  --bs-table-bg: var(--surface);
  --bs-table-color: var(--text-main);
  --bs-table-striped-bg: var(--surface-alt);
  --bs-table-striped-color: var(--text-main);
  --bs-table-hover-bg: var(--surface-alt);
  --bs-table-hover-color: var(--text-main);
  --bs-table-border-color: var(--border);
  color: var(--text-main);
}
.table td { color: var(--text-main); }  /* explicit, otherwise Bootstrap overrides */
```

### Form controls theming
```css
.form-control, .form-select {
  background-color: var(--surface);
  color: var(--text-main);
  border-color: var(--border);
}
```

### Button overrides
All buttons use CSS variables from the theme:
```css
.btn-success { background-color: var(--success); border-color: var(--success); color: #fff; }
.btn-primary { background-color: var(--accent); border-color: var(--accent); color: #fff; }
.btn-info { background-color: var(--info); border-color: var(--info); color: #fff; }
.btn-danger { background-color: var(--danger); border-color: var(--danger); color: #fff; }
.btn-secondary { background-color: var(--secondary); border-color: var(--secondary); color: #fff; }
```

### Alerts
```css
.alert { background: var(--surface); color: var(--text-main); border-color: var(--border); }
.alert-danger { border-left: 4px solid var(--danger) !important; }
.alert-warning { border-left: 4px solid var(--warning) !important; }
```

### Links in dark theme
```css
[data-theme="dark"] a { color: #93b4f5; }
```

---

## 7. Responsive

| Component | Desktop (≥992px) | Mobile (<768px) |
|-----------|-----------------|-----------------|
| Navbar | Horizontal | Hamburger, vertical stack |
| Login card | max-width: 460px | Nearly full width |
| User form card | col-lg-7 | Full width, fields stack vertically |
| Filter bar | Single row | Fields stack, buttons full width |
| Tables | Normal layout | `table-responsive` — horizontal scroll |
| Pagination | Full | Line wrapping |

---

## 8. Checklist for Agent

- [ ] All colors via CSS variables, zero hardcoded colors in HTML/components
- [ ] Three themes work via `data-theme` on `<html>`
- [ ] Theme switcher in footer (three round buttons)
- [ ] Theme is persisted in `localStorage`
- [ ] Sticky footer via flexbox
- [ ] Navbar with hamburger on mobile
- [ ] Tables in `card` with `card-body p-0` and `card-footer` for pagination
- [ ] Forms in `card`, centered (`mx-auto`), `col-lg-7 col-md-10`
- [ ] Errors — in centered `card` (`col-lg-5 col-md-8`)
- [ ] Date badges — compact (flex-column, align-items: flex-start), monospace font
- [ ] Status badges — custom `badge-status-*` classes
- [ ] Self-delete protection — Delete button is not rendered for the current user
- [ ] Email addresses in tables — plain text, NOT links
- [ ] Alerts — border-left: 4px solid, background `var(--surface)`
- [ ] Dark mode — muted badges, readable text in tables
- [ ] IBM Plex Sans for UI, JetBrains Mono for dates in badges

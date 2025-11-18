# BobaMixer Roadmap & Progress Review

This document consolidates the roadmap intent and actual delivery progress using the key specs inside `spec/`:

- `spec/boba-control-plane.md` - Control plane baseline (architecture, CLI expectations, delivery phases).
- `spec/CLI_REDESIGN.md` - CLI redesign narrative that commits us to a TUI-first product philosophy.
- `spec/TUI_ENHANCEMENT_PLAN.md` - Migration plan that maps legacy CLI behaviors to Bubble Tea views.
- `spec/PHASE1_FEATURES.md`, `spec/PHASE2_FEATURES.md`, `spec/PHASE3_FEATURES.md` - Execution logs for the three delivery phases.

## Vision Recap (Baseline Spec)

- **Product posture**: BobaMixer is the control plane for local AI CLIs. It manages providers, tools, bindings, secrets, optional profiles, and can host a local proxy (`spec/boba-control-plane.md` Sections 1-6).
- **CLI contract**: Keep a small set of core commands (`boba`, `boba run`, `boba providers/tools/bind`, `boba doctor`, `boba proxy serve`) and treat everything else as advanced or deprecated (`spec/boba-control-plane.md` Section 4).
- **Runtime behavior**: `boba run` resolves tool/provider bindings, injects env/config, optionally routes through the proxy, and launches the downstream CLI unmodified (Section 5).
- **Architecture**: Bubble Tea root model switches between onboarding and dashboard modes, with the dashboard ultimately hosting the full control plane (Section 7).
- **Baseline phases**: Phase 1 (core control plane), Phase 2 (proxy plus monitoring), Phase 3 (advanced routing/budget/stats) (Section 8).

## TUI-First Strategy (CLI Redesign Spec)

- **Problem diagnosis**: Users were forced to memorize 20+ commands even though the binary booted a TUI by default, creating a "mixed paradigm" experience (`spec/CLI_REDESIGN.md` Problem Analysis).
- **Strategy**: Treat the TUI as the default interface, keep only non-interactive commands (`boba`, `boba run`, `boba init`, `boba doctor`, `boba call`, `boba stats`, `boba version`), and migrate every other management workflow into dedicated views.
- **Navigation model**: Tabbed layout with discoverable tabs (Dashboard, Providers, Tools, Bindings, Secrets, Stats, Proxy, Routing, Budget, Suggestions, Configuration, Help) and consistent global shortcuts.
- **Implementation plan**: Three phases - (1) simplify Help output, add missing management views, and retire redundant commands; (2) add Provider/Tool/Binding/Secrets/Proxy screens; (3) deliver routing tester, config editor, reports, and a comprehensive help screen - with CLI kept only for automation.

## Roadmap Snapshot

| Phase | Baseline Scope (`spec/boba-control-plane.md`) | TUI Plan Targets (`spec/TUI_ENHANCEMENT_PLAN.md`) | Delivery Status |
|-------|----------------------------------------------|---------------------------------------------------|-----------------|
| **Phase 1 - Control Plane Core** | Parse `providers.yaml`, `tools.yaml`, `bindings.yaml`; ship `boba providers/tools/bind/run/doctor`; deliver a minimal dashboard with binding edits (Section 8.1). | Add Providers, Tools, Bindings, Secrets views with consistent navigation and state indicators (Sections 18-115). | Complete. `spec/PHASE1_FEATURES.md` shows the four core views, numeric navigation (`1`-`6`), and lipgloss-based state cues fully implemented. |
| **Phase 2 - Proxy plus Operational Visibility** | Introduce proxy server, `use_proxy` bindings, and usage tracking (Section 8.2). | Add Proxy control, Routing tester, and Suggestions views to surface operational data (Sections 116-287). | Complete. `spec/PHASE2_FEATURES.md` documents the proxy status panel, routing education screen, and data-backed suggestions list plus expanded navigation (`7`-`9`). |
| **Phase 3 - Advanced Features** | Add routing strategies, budgets, stats, reports, hooks, and richer dashboards (Section 8.3). | Deliver Reports generator, Hooks manager, Config editor, Help view, and finish the 13-view navigation loop (Sections 287-485). | Complete. `spec/PHASE3_FEATURES.md` confirms the four advanced views (`0`, `H`, `C`, `?`), unified shortcut map, and mapping of every legacy CLI command to a TUI destination. |

## Phase Highlights and Reflections

### Phase 1 - Control Plane Core
- Replaced the CLI-only workflows (`boba providers/tools/bind/secrets`) with live tables, state glyphs, and in-place proxy toggles (`spec/PHASE1_FEATURES.md`).
- Established the navigation idioms - number keys, Tab cycling, vim-style list navigation - and lipgloss theming reused by later phases.
- Remaining gaps noted in the spec: inline provider/tool editing forms, masked secret entry, full binding creation dialogs, and deeper proxy controls (see the "future enhancements" section in `spec/PHASE1_FEATURES.md`), which flow into later sprints of the TUI plan.

### Phase 2 - Operational and Optimization Views
- Surfaced runtime visibility: Proxy status (with instructions for `boba proxy serve`), routing tester instructions, and a data-driven Suggestions list that interprets cost trends, profile usage, anomaly detection, and budget drift (`spec/PHASE2_FEATURES.md` Sections 1-3, 232-272).
- Extended navigation to nine interactive panes and wired lazy data loading for Suggestions (`spec/PHASE2_FEATURES.md` Section 156 onwards).
- Reflections: While proxy metrics now have a TUI home, the underlying proxy implementation (usage logging, routing policies) still depends on the domain/proxy packages catching up with the spec ambitions (cost-aware routing, rate limiting). These technical milestones remain part of the broader roadmap even though the TUI scaffolding is ready.

### Phase 3 - Advanced Experience Layer
- Finalized the TUI-first promise with four capstone views: Report generator, Git hooks dashboard, Config editor, and a persistent Help overlay (`spec/PHASE3_FEATURES.md` Sections 1-4, 196-280).
- Completed the 13-view navigation loop and established letter shortcuts (`0`, `H`, `C`, `?`) alongside digits.
- Reflections: Reports/Hooks/Config views currently guide users back to CLI commands (`boba report`, `boba hooks install`, `boba edit`) for execution/editing, which satisfies the "TUI-first but CLI for automation" split. Remaining enhancements include richer inline forms (edit-in-place, YAML previews) and wiring hooks telemetry into the stats pipeline, as suggested in `spec/TUI_ENHANCEMENT_PLAN.md` (Sprint 4 and the Form Components section).

## Cross-Cutting Observations

- **Documentation-to-implementation alignment**: Every CLI workflow enumerated in the redesign spec now has a TUI waypoint, and Help output steers users toward the interface rather than subcommands (`spec/PHASE3_FEATURES.md` Section 4).
- **Navigation maturity**: Thirteen total views (Dashboard, Providers, Tools, Bindings, Secrets, Stats, Proxy, Routing, Suggestions, Reports, Hooks, Config, Help) with numeric/alphabetic shortcuts reflect the TUI Enhancement Plan roadmap (Section 316 onwards) and remove ambiguity about where tasks live.
- **CLI surface area**: Core commands remain (`boba`, `boba run`, `boba init`, `boba doctor`, `boba call`, `boba stats`, `boba version`). All other commands are either deprecated or directly linked from TUI help cards, honoring the CLI redesign contract.
- **Proxy and routing backend**: The front-end views are ready, but the control-plane spec still calls for sophisticated routing (cost/latency weighting), budgeting, and hook-driven project context (Sections 6 and 8.3 of `spec/boba-control-plane.md`). These backend capabilities should be validated against real usage data to ensure the TUI indicators remain truthful.

## Outstanding Work and Next Steps

1. **Finish Sprint 4 of the TUI plan** - Implement search/filter, enhanced theming, quick-help overlays, and reusable form components (`spec/TUI_ENHANCEMENT_PLAN.md` Sections 316-401, 412-435).
2. **Inline editing flows** - Upgrade Providers/Tools/Bindings views with the planned text input / select / confirm components so edits can happen without shelling out (`spec/TUI_ENHANCEMENT_PLAN.md` Form Components section).
3. **Proxy/routing depth** - Build out the routing engine (time-of-day, cost-aware strategies) and proxy observability promised in the baseline spec to match what the Phase 2 UI exposes.
4. **Hook telemetry loop** - Connect Hooks reports to concrete automation (auto profile suggestions, repository-scoped analytics) so the `Hooks` view reflects live status rather than static instructions.
5. **Documentation refresh** - Update public docs/help to reflect the new TUI-first command surface and ensure advanced CLI commands are clearly labeled as automation-only usages.

By keeping this summary updated, we can quickly validate whether new feature requests fit into the TUI-first roadmap, ensure backend investments keep pace with the UI, and identify when it is safe to fully retire the remaining legacy CLI paths.

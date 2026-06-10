# Agent Notes

## Project Snapshot

This repository is New API, a Go + React AI API gateway. The backend follows the existing layered structure:

- `router/`: API routing.
- `controller/`: Gin request handlers.
- `service/`: business logic and transaction orchestration.
- `model/`: GORM models and migration registration.
- `web/default/`: React 19 + TypeScript admin/user frontend.

The project targets SQLite, MySQL, and PostgreSQL, so new database code should stay inside GORM abstractions unless a cross-database fallback is explicitly handled.

## Current Distribution Module

A B2B2C distribution/agent module is present under the project-specific "distribution" naming. Earlier "p3" naming was intentionally avoided in variables and files.

### Backend State

Relevant files:

- `model/distribution.go`
- `service/distribution.go`
- `service/distribution_rules.go`
- `service/distribution_transactions.go`
- `service/distribution_balance.go`
- `service/distribution_customers.go`
- `service/distribution_invitations.go`
- `service/distribution_promo.go`
- `controller/distribution.go`
- `router/api-router.go`

Implemented model groups include:

- Agents
- Packages
- Orders
- Inventories
- Balance adjustments
- Balance ledgers
- Commission logs
- Profit logs
- Price configs
- Invitations
- Customer ownerships
- Attribution logs
- Promo codes
- Gift rules
- Ops dashboard authorization

The agent model currently has explicit hierarchy fields:

- `parent_agent_id`
- `level`, defaulting to 2

Current hierarchy rule direction:

- Newly created or role-promoted agents default to level 2.
- If there is no inviter, the promoted agent remains level 2 with no parent.
- If the inviter is a level 1 agent, the promoted agent becomes level 2 and uses that level 1 agent as parent.
- If the inviter is level 2 or not an enabled agent, the promoted agent becomes level 2 with no parent.
- Agent hierarchy editing is intended to be limited to `level` and `parent_agent_id`.

### Frontend State

Relevant files:

- `web/default/src/features/distribution/agent-center.tsx`
- `web/default/src/features/distribution/agent-admin.tsx`
- `web/default/src/features/distribution/api.ts`
- `web/default/src/features/distribution/types.ts`
- `web/default/src/features/distribution/labels.ts`
- `web/default/src/features/distribution/agent-combobox.tsx`

The frontend has:

- Agent center page.
- Agent admin page with tab-based layout.
- Agent list with level display.
- Agent edit flow for parent/level.
- Package, gift rule, ops auth, profit, and attribution surfaces.
- Shared agent search combobox.

## Recent Work In Progress

The latest active change set was around agent hierarchy and role promotion:

- Added `level` to `DistributionAgent`.
- Added frontend `level` to `DistributionAgent` TypeScript type.
- Added backend hierarchy normalization.
- Adjusted role-promotion flow to call agent creation only when a user changes from non-agent to agent.
- Adjusted invitation acceptance so accepted invitees become level 2 agents and only keep a parent when the inviter is level 1.
- Added agent-admin edit mode for changing only agent level and parent.

## Known Risks / Needs Review

The previous work was interrupted before a clean verification pass.

Items that should be checked next:

- Run `gofmt` on modified Go files.
- Run backend build or targeted tests once the local environment allows it.
- Run frontend typecheck/build for `web/default`.
- Inspect the final diff for `agent-admin.tsx`; this file still contains older price-config dialog code paths even though price config was previously intended to be removed or hidden.
- Confirm that all new visible frontend strings are present in all locale files.
- Confirm that role changes in user management persist correctly and trigger agent creation only on non-agent to agent transitions.
- Confirm that direct calls to the admin agent save endpoint cannot create a level 1 agent through the normal "add agent" workflow unless intentionally edited afterward.

## Recommended Next Steps

1. Format and verify modified Go files.
2. Verify frontend compile around `agent-admin.tsx`.
3. Manually test:
   - Create user as agent.
   - Update existing user role to agent.
   - Invitee promotion from no inviter.
   - Invitee promotion from level 1 inviter.
   - Invitee promotion from level 2 inviter.
   - Agent list edit button only changes level and parent.
4. Finish or remove remaining price-config UI/backend paths according to the latest product decision.
5. Complete i18n for any newly added labels.


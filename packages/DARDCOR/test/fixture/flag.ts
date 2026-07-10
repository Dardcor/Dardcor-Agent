import type { WorkspaceV2 } from "@dardcor-ai/core/workspace"
import { Flag } from "@dardcor-ai/core/flag/flag"
import { Effect, Scope } from "effect"

/**
 * Scoped override for `Flag.dardcor_WORKSPACE_ID`. Saves the previous value
 * on entry and restores it via finalizer when the surrounding scope closes —
 * preserves the original try/finally semantics regardless of test outcome.
 */
export function withFixedWorkspaceID(id: WorkspaceV2.ID): Effect.Effect<void, never, Scope.Scope> {
  return Effect.gen(function* () {
    const previous = Flag.dardcor_WORKSPACE_ID
    Flag.dardcor_WORKSPACE_ID = id
    yield* Effect.addFinalizer(() =>
      Effect.sync(() => {
        Flag.dardcor_WORKSPACE_ID = previous
      }),
    )
  })
}

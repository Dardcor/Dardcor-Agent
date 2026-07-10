import { run as runTui, type TuiInput } from "@dardcor-ai/tui"
import { Global } from "@dardcor-ai/core/global"
import { AppNodeBuilder } from "@dardcor-ai/core/effect/app-node-builder"
import { Effect } from "effect"

export function run(input: TuiInput) {
  return runTui(input).pipe(Effect.provide(AppNodeBuilder.build(Global.node)))
}

import { Context } from "effect"
import type { InstanceContext } from "@/project/instance-context"
import type { WorkspaceV2 } from "@dardcor-ai/core/workspace"

export const InstanceRef = Context.Reference<InstanceContext | undefined>("~dardcor/InstanceRef", {
  defaultValue: () => undefined,
})

export const WorkspaceRef = Context.Reference<WorkspaceV2.ID | undefined>("~dardcor/WorkspaceRef", {
  defaultValue: () => undefined,
})

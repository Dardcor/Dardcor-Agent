import { AgentV2 } from "@dardcor-ai/core/agent"
import { AISDK } from "@dardcor-ai/core/aisdk"
import { Catalog } from "@dardcor-ai/core/catalog"
import { CommandV2 } from "@dardcor-ai/core/command"
import { Credential } from "@dardcor-ai/core/credential"
import { AppNodeBuilder } from "@dardcor-ai/core/effect/app-node-builder"
import { LayerNodePlatform } from "@dardcor-ai/core/effect/app-node-platform"
import { LayerNode } from "@dardcor-ai/core/effect/layer-node"
import { EventV2 } from "@dardcor-ai/core/event"
import { FileSystem } from "@dardcor-ai/core/filesystem"
import { FSUtil } from "@dardcor-ai/core/fs-util"
import { Integration } from "@dardcor-ai/core/integration"
import { Location } from "@dardcor-ai/core/location"
import { Npm } from "@dardcor-ai/core/npm"
import { PluginV2 } from "@dardcor-ai/core/plugin"
import { Reference } from "@dardcor-ai/core/reference"
import { SkillV2 } from "@dardcor-ai/core/skill"
import { Effect, Layer } from "effect"
import { tempLocationLayer } from "../fixture/location"

const npmLayer = Layer.succeed(
  Npm.Service,
  Npm.Service.of({
    add: () => Effect.succeed({ directory: "", entrypoint: undefined }),
    install: () => Effect.void,
    which: () => Effect.succeed(undefined),
  }),
)

export const PluginTestLayer = AppNodeBuilder.build(
  LayerNode.group([
    FileSystem.node,
    FSUtil.node,
    Location.node,
    Npm.node,
    Credential.node,
    EventV2.node,
    LayerNodePlatform.httpClient,
    PluginV2.node,
    AgentV2.node,
    AISDK.node,
    Catalog.node,
    CommandV2.node,
    Integration.node,
    Reference.node,
    SkillV2.node,
  ]),
  [
    [Location.node, tempLocationLayer],
    [Npm.node, npmLayer],
  ],
)

# @dardcor-ai/sdk-next

Effect-native scoped dardcor host for in-process applications. This transitional package will replace the existing generated `@dardcor-ai/sdk` after its consumers migrate.

The SDK executes Server's assembled HTTP router in memory. It opens no listener and performs no network I/O, while preserving the same routing, middleware, handlers, codecs, and errors as the network client.

```ts
import { dardcor } from "@dardcor-ai/sdk-next"

const dardcor = yield * dardcor.create()
const session = yield * dardcor.sessions.get({ sessionID })
```

It also exports `Tool` and exposes local-only `tools.register(...)`, replacing the former `@dardcor-ai/core/public` facade. Registration uses Core's host-level `ApplicationTools` service shared by the host's Locations; each Location retains its own `ToolRegistry` for overlay, lookup, and settlement. Closing the owning Effect Scope releases router resources, location services, fibers, and scoped tool registrations.

`sessions.events({ sessionID, after })` replays durable events after the optional aggregate sequence, then emits newly committed durable events. `sessions.interrupt(...)` targets execution owned by this host, and `sessions.message(...)` retrieves one projected Session message.

The same constructor is available as a service Layer:

```ts
const program = Effect.gen(function* () {
  const dardcor = yield* dardcor.Service
  return yield* dardcor.sessions.get({ sessionID })
})

yield * program.pipe(Effect.provide(dardcor.layer))
```

`dardcor.layer` adapts `dardcor.create()` for dependency injection; it does not define another host implementation.

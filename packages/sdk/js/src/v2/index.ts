export * from "./client.js"
export * from "./server.js"

import { createdardcorClient } from "./client.js"
import { createdardcorServer } from "./server.js"
import type { ServerOptions } from "./server.js"

export * as data from "./data.js"

export async function createdardcor(options?: ServerOptions) {
  const server = await createdardcorServer({
    ...options,
  })

  const client = createdardcorClient({
    baseUrl: server.url,
  })

  return {
    client,
    server,
  }
}

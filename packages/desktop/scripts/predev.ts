import { $ } from "bun"

await $`bun ./scripts/copy-icons.ts ${process.env.dardcor_CHANNEL ?? "dev"}`

await $`cd ../dardcor && bun script/build-node.ts`

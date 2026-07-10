// @ts-nocheck

import { dardcor } from "@dardcor-ai/core"
import { ReadTool } from "@dardcor-ai/core/tools"

const dardcor = dardcor.make({})

dardcor.tool.add(ReadTool)

dardcor.tool.add({
  name: "bash",
  schema: {
    type: "object",
    properties: {
      command: {
        type: "string",
        description: "The command to run.",
      },
    },
    required: ["command"],
  },
  execute(input, ctx) {},
})

dardcor.auth.add({
  provider: "openai",
  type: "api",
  value: process.env.OPENAI_API_KEY,
})

dardcor.agent.add({
  name: "build",
  permissions: [],
  model: {
    id: "gpt-5-5",
    provider: "openai",
    variant: "xhigh",
  },
})

const sessionID = await dardcor.session.create({
  agent: "build",
})

dardcor.subscribe((event) => {
  console.log(event)
})

await dardcor.session.prompt({
  sessionID,
  text: "hey what is up",
})

await dardcor.session.prompt({
  sessionID,
  text: "what is up with this",
  files: [
    {
      mime: "image/png",
      uri: "data:image/png;base64,xxxx",
    },
  ],
})

await dardcor.session.wait()

console.log(await dardcor.session.messages(sessionID))

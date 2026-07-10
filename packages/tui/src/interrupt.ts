type Handler = () => boolean

let handler: Handler | undefined
let sigintInstalled = false

export const Interrupt = {
  set(next: Handler | undefined) {
    handler = next
    return () => {
      if (handler === next) handler = undefined
    }
  },
  handle() {
    return handler?.() ?? false
  },
  installSigintHandler() {
    if (sigintInstalled) return () => {}
    sigintInstalled = true
    const onSigint = () => {
      Interrupt.handle()
    }
    process.on("SIGINT", onSigint)
    if (process.platform === "win32") {
      process.on("SIGBREAK", onSigint)
    }
    return () => {
      process.off("SIGINT", onSigint)
      if (process.platform === "win32") {
        process.off("SIGBREAK", onSigint)
      }
      sigintInstalled = false
    }
  },
}

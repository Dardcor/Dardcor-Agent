interface ImportMetaEnv {
  readonly dardcor_CHANNEL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

declare module "virtual:dardcor-server" {
  export namespace Server {
    export const listen: typeof import("../../../dardcor/dist/types/src/node").Server.listen
    export type Listener = import("../../../dardcor/dist/types/src/node").Server.Listener
  }
  export namespace Config {
    export const get: typeof import("../../../dardcor/dist/types/src/node").Config.get
    export type Info = import("../../../dardcor/dist/types/src/node").Config.Info
  }
  export const bootstrap: typeof import("../../../dardcor/dist/types/src/node").bootstrap
}

import { Config } from "effect"

export function truthy(key: string) {
  const value = process.env[key]?.toLowerCase()
  return value === "true" || value === "1"
}

const copy = process.env["dardcor_EXPERIMENTAL_DISABLE_COPY_ON_SELECT"]
const fff = process.env["dardcor_DISABLE_FFF"]

function enabledByExperimental(key: string) {
  return process.env[key] === undefined ? truthy("dardcor_EXPERIMENTAL") : truthy(key)
}

export const Flag = {
  OTEL_EXPORTER_OTLP_ENDPOINT: process.env["OTEL_EXPORTER_OTLP_ENDPOINT"],
  OTEL_EXPORTER_OTLP_HEADERS: process.env["OTEL_EXPORTER_OTLP_HEADERS"],

  dardcor_AUTO_HEAP_SNAPSHOT: truthy("dardcor_AUTO_HEAP_SNAPSHOT"),
  dardcor_GIT_BASH_PATH: process.env["dardcor_GIT_BASH_PATH"],
  dardcor_CONFIG: process.env["dardcor_CONFIG"],
  dardcor_CONFIG_CONTENT: process.env["dardcor_CONFIG_CONTENT"],
  dardcor_DISABLE_AUTOUPDATE: truthy("dardcor_DISABLE_AUTOUPDATE"),
  dardcor_ALWAYS_NOTIFY_UPDATE: truthy("dardcor_ALWAYS_NOTIFY_UPDATE"),
  dardcor_DISABLE_PRUNE: truthy("dardcor_DISABLE_PRUNE"),
  dardcor_DISABLE_TERMINAL_TITLE: truthy("dardcor_DISABLE_TERMINAL_TITLE"),
  dardcor_SHOW_TTFD: truthy("dardcor_SHOW_TTFD"),
  dardcor_DISABLE_AUTOCOMPACT: truthy("dardcor_DISABLE_AUTOCOMPACT"),
  dardcor_DISABLE_MODELS_FETCH: truthy("dardcor_DISABLE_MODELS_FETCH"),
  dardcor_DISABLE_MOUSE: truthy("dardcor_DISABLE_MOUSE"),
  dardcor_FAKE_VCS: process.env["dardcor_FAKE_VCS"],
  dardcor_SERVER_PASSWORD: process.env["dardcor_SERVER_PASSWORD"],
  dardcor_SERVER_USERNAME: process.env["dardcor_SERVER_USERNAME"],
  dardcor_DISABLE_FFF: fff === undefined ? process.platform === "win32" : truthy("dardcor_DISABLE_FFF"),

  // Experimental
  dardcor_EXPERIMENTAL_FILEWATCHER: Config.boolean("dardcor_EXPERIMENTAL_FILEWATCHER").pipe(
    Config.withDefault(false),
  ),
  dardcor_EXPERIMENTAL_DISABLE_FILEWATCHER: Config.boolean("dardcor_EXPERIMENTAL_DISABLE_FILEWATCHER").pipe(
    Config.withDefault(false),
  ),
  dardcor_EXPERIMENTAL_DISABLE_COPY_ON_SELECT:
    copy === undefined ? process.platform === "win32" : truthy("dardcor_EXPERIMENTAL_DISABLE_COPY_ON_SELECT"),
  dardcor_MODELS_URL: process.env["dardcor_MODELS_URL"],
  dardcor_MODELS_PATH: process.env["dardcor_MODELS_PATH"],
  dardcor_DB: process.env["dardcor_DB"],

  dardcor_WORKSPACE_ID: process.env["dardcor_WORKSPACE_ID"],
  dardcor_EXPERIMENTAL_WORKSPACES: enabledByExperimental("dardcor_EXPERIMENTAL_WORKSPACES"),

  // Evaluated at access time (not module load) because tests, the CLI, and
  // external tooling set these env vars at runtime.
  get dardcor_DISABLE_PROJECT_CONFIG() {
    return truthy("dardcor_DISABLE_PROJECT_CONFIG")
  },
  get dardcor_EXPERIMENTAL_REFERENCES() {
    return enabledByExperimental("dardcor_EXPERIMENTAL_REFERENCES")
  },
  get dardcor_TUI_CONFIG() {
    return process.env["dardcor_TUI_CONFIG"]
  },
  get dardcor_CONFIG_DIR() {
    return process.env["dardcor_CONFIG_DIR"]
  },
  get dardcor_PURE() {
    return truthy("dardcor_PURE")
  },
  get dardcor_PERMISSION() {
    return process.env["dardcor_PERMISSION"]
  },
  get dardcor_PLUGIN_META_FILE() {
    return process.env["dardcor_PLUGIN_META_FILE"]
  },
  get dardcor_CLIENT() {
    return process.env["dardcor_CLIENT"] ?? "cli"
  },
}

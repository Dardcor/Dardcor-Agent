import path from "path"

process.env.dardcor_DB = ":memory:"
process.env.dardcor_MODELS_PATH = path.join(import.meta.dir, "plugin", "fixtures", "models-dev.json")
process.env.dardcor_DISABLE_MODELS_FETCH = "true"

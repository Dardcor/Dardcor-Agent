declare global {
  const dardcor_VERSION: string
  const dardcor_CHANNEL: string
}

export const InstallationVersion = typeof dardcor_VERSION === "string" ? dardcor_VERSION : "local"
export const InstallationChannel = typeof dardcor_CHANNEL === "string" ? dardcor_CHANNEL : "local"
export const InstallationLocal = InstallationChannel === "local"

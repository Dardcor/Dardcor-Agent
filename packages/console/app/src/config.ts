/**
 * Application-wide constants and configuration
 */
export const config = {
  // Base URL
  baseUrl: "https://dardcor-code.vercel.app/",

  // GitHub
  github: {
    repoUrl: "https://github.com/anomalyco/dardcor",
    starsFormatted: {
      compact: "160K",
      full: "160,000",
    },
  },

  // Social links
  social: {
    twitter: "https://x.com/dardcor",
    discord: "https://discord.gg/dardcor",
  },

  // Static stats (used on landing page)
  stats: {
    contributors: "900",
    commits: "13,000",
    monthlyUsers: "7.5M",
  },
} as const

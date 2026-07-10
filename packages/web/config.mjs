const stage = process.env.SST_STAGE || "dev"

export default {
  url: stage === "production" ? "https://DARDCOR.ai" : `https://${stage}.DARDCOR.ai`,
  console: stage === "production" ? "https://DARDCOR.ai/auth" : `https://${stage}.DARDCOR.ai/auth`,
  email: "help@anoma.ly",
  socialCard: "https://social-cards.sst.dev",
  github: "https://github.com/anomalyco/DARDCOR",
  discord: "https://DARDCOR.ai/discord",
  headerLinks: [
    { name: "app.header.home", url: "/" },
    { name: "app.header.docs", url: "/docs/" },
  ],
}

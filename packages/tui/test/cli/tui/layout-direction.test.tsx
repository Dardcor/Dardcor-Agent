/** @jsxImportSource @opentui/solid */
import { expect, test } from "bun:test"
import { testRender } from "@opentui/solid"
import { ThemeProvider } from "../../../src/context/theme"
import { TuiConfigProvider } from "../../../src/config"
import { KVProvider } from "../../../src/context/kv"
import { SidebarSection } from "../../../src/component/sidebar-section"
import { TestTuiContexts } from "../../fixture/tui-environment"
import { createTuiResolvedConfig } from "../../fixture/tui-runtime"

test("sidebar sections stack heading and content vertically", async () => {
  const app = await testRender(
    () => (
      <TestTuiContexts>
        <box flexDirection="column" width={30} height={8}>
        <KVProvider>
          <TuiConfigProvider config={createTuiResolvedConfig()}>
            <ThemeProvider mode="dark" source={{ discover: async () => ({}) }}>
              <SidebarSection title="CONTEXT">
                <text>14.4K tokens</text>
                <text>1% window used</text>
              </SidebarSection>
            </ThemeProvider>
          </TuiConfigProvider>
        </KVProvider>
        </box>
      </TestTuiContexts>
    ),
    { width: 32, height: 8 },
  )

  try {
    await new Promise((resolve) => setTimeout(resolve, 100))
    await app.renderOnce()
    await app.renderOnce()
    const frame = app.captureCharFrame().split("\n").map((line) => line.trimEnd())
    const heading = frame.findIndex((line) => line.includes("CONTEXT"))
    const tokens = frame.findIndex((line) => line.includes("14.4K tokens"))
    const usage = frame.findIndex((line) => line.includes("1% window used"))

    expect(heading).toBeGreaterThanOrEqual(0)
    expect(tokens).toBeGreaterThan(heading)
    expect(usage).toBeGreaterThan(tokens)
  } finally {
    app.renderer.destroy()
  }
})

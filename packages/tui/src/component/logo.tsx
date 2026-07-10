import { TextAttributes } from "@opentui/core"
import { For } from "solid-js"
import { tint, useTheme } from "../context/theme"
import { logo } from "../logo"

export function Logo() {
  const { theme } = useTheme()
  const colors = [theme.primary, theme.primary, theme.accent, theme.accent, theme.secondary, tint(theme.secondary, theme.primary, 0.45)]

  return (
    <box flexDirection="column" alignItems="center" gap={1}>
      <box flexDirection="column" alignItems="center">
        <For each={logo.left}>
          {(line, index) => (
            <box flexDirection="row">
              <For each={Array.from(line)}>
                {(char) => (
                  <text fg={colors[index()] ?? theme.primary} attributes={TextAttributes.BOLD} selectable={false}>
                    {char}
                  </text>
                )}
              </For>
            </box>
          )}
        </For>
      </box>
      <text fg={theme.secondary} attributes={TextAttributes.BOLD}>
        ? DARDCOR COMMAND CENTER ?
      </text>
      <text fg={theme.textMuted}>agentic code interface • purple ops shell</text>
    </box>
  )
}
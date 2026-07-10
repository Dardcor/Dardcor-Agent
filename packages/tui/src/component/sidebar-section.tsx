import { TextAttributes } from "@opentui/core"
import { useTheme } from "../context/theme"
import type { JSX } from "@opentui/solid"

export function SidebarSection(props: { title: string; children: JSX.Element }) {
  const { theme } = useTheme()

  return (
    <box
      gap={1}
      paddingTop={1}
      paddingBottom={1}
      paddingLeft={2}
      paddingRight={1}
      backgroundColor={theme.backgroundElement}
      border={["left"]}
      borderColor={theme.borderSubtle}
    >
      <text fg={theme.secondary} attributes={TextAttributes.BOLD}>
        {props.title}
      </text>
      {props.children}
    </box>
  )
}

import { getComponentCatalogue } from "@opentui/solid/components"
import { registerSpinner } from "opentui-spinner/solid"

export function registerdardcorSpinner() {
  if (!getComponentCatalogue().spinner) registerSpinner()
}

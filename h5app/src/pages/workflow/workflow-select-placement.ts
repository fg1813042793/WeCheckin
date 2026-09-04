export type WorkflowSelectPlacement = 'top' | 'bottom'

export interface WorkflowSelectPlacementInput {
  controlTop: number
  controlBottom: number
  visibleTop: number
  visibleBottom: number
  panelHeight: number
  gap?: number
}

export function resolveWorkflowSelectPlacement(input: WorkflowSelectPlacementInput): WorkflowSelectPlacement {
  const gap = Math.max(0, input.gap ?? 6)
  const panelHeight = Math.max(0, input.panelHeight)
  const availableAbove = Math.max(0, input.controlTop - input.visibleTop - gap)
  const availableBelow = Math.max(0, input.visibleBottom - input.controlBottom - gap)

  if (availableBelow < panelHeight && availableAbove > availableBelow)
    return 'top'

  return 'bottom'
}

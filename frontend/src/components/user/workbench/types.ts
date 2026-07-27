export interface WorkbenchTask {
  id: string
  title: string
  status: 'pending' | 'in_progress' | 'completed' | 'failed'
}

export type WorkbenchGenerationStage =
  | 'preparing'
  | 'dispatching'
  | 'reasoning'
  | 'answering'
  | 'completed'
  | 'failed'
  | 'canceled'

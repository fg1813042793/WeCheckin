export type FormKitPrimitive = string | number | null | undefined

export type FormKitDynamicMap = Record<string, any>

export interface FormKitOption extends FormKitDynamicMap {
  id?: string | number
  label?: string
  title?: string
  value?: FormKitPrimitive
  score?: number
  children?: FormKitOption[]
}

export interface FormKitProps extends FormKitDynamicMap {
  options: FormKitOption[]
  rows: any[]
  columns: any[]
  fields: FormKitOption[]
  maxRating?: number
  icon?: string
}

export interface FormKitQuestion extends FormKitDynamicMap {
  id?: string | number
  type: string
  title?: string
  description?: string
  required?: boolean
  placeholder?: string
  props?: any
  validate?: any[]
  logic?: any[]
  defaultHidden?: boolean
  multiple?: boolean
  examScore?: number
  examCorrectAnswer?: string
  examAnalysis?: string
}

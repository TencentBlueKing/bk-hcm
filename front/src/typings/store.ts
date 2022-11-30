export interface ProjectModel {
  type: string
  resourceName: string
  name: string,
  cloudName: string,
  scretId: string,
  account: number | string,
  user: string[]
  remark: string,
  business: string | string[]
}

export enum StaffType {
  RTX = 'rtx',
  MAIL = 'email',
  ALL = 'all',
}
export interface Staff {
  english_name: string
  chinese_name: string
}

export interface Department {
  id: number
  name: string
  full_name: string
  has_children: boolean
  parent?: number
  children?: Department[]
  checked?: boolean
  indeterminate?: boolean
  isOpen: boolean
  loaded: boolean
  loading: boolean
}

export interface FormItems {
  label?: string
  required?: boolean
  property?: string,
  content?: Function,
  component?: Function,
}

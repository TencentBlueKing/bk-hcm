export interface ProjectModel {
  resourceName: string
  name: string,
  cloudName: string
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

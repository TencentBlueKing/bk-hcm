import { VNode } from 'vue';
import { ResourceActionType } from './constants';
import type { IAuthSign } from '@/common/auth-service';

interface Clickable {
  loading?: () => boolean;
  disabled?: () => boolean; // 通常涉及到响应式计算，定义为函数
  handleClick?: () => void;
  authSign?: () => IAuthSign | IAuthSign[]; // 预鉴权配置参数可能需要进行响应式计算，因此这里定义为函数
}

export interface ActionItemType extends Clickable {
  type?: 'button' | 'dropdown';
  label?: string;
  value?: ResourceActionType;
  index?: number;
  children?: ActionItemType[];
  displayProps?: Record<string, any>; // 配置组件表现层面的props
  render?: () => VNode;
  prefix?: () => VNode;
}

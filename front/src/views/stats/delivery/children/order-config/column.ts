/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';
import useTableSettings from '@/hooks/use-table-settings';
import { getModel } from '@/model/manager';
import { h } from 'vue';

const titleStar = (title: string) => h('span', [title, h('span', { style: { color: 'red' } }, ' *')]);
@Model('order-config/table-column')
export class TableColumn {
  @Column('business', { name: '业务', width: 120 })
  bk_biz_id: number;

  @Column('array', { name: '单号' })
  sub_order_ids: string[];

  @Column('string', { name: '时间段', width: 315 })
  start_at: string;

  @Column('string', { name: '备注' })
  memo: string;
}

@Model('order-config/action-table-column')
export class TableColumnAction {
  @Column('business', { name: titleStar('业务'), width: 180 })
  bk_biz_id: number;

  @Column('array', { name: titleStar('单号') })
  sub_order_ids: string[];

  @Column('string', { name: titleStar('时间段'), width: 350 })
  start_at: string;

  @Column('string', { name: titleStar('备注') })
  memo: string;

  @Column('string', { name: ' ', width: 85 })
  action: string;
}

export const columns = getModel(TableColumn).getProperties();
export const actionColumns = getModel(TableColumnAction).getProperties();

/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';
import { Properties } from './properties';

@Model('task/rerun.view')
export class RerunView extends Properties {
  @Column('enum', { name: '云地域' })
  region_id: string;

  @Column('enum', { name: '云地域名称' })
  region_name: string;
}

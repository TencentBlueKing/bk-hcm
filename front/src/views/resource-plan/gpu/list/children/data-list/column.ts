import { Model, Column } from '@/decorator';
import { GPU_DEMAND_STATUS, GPU_DEMAND_STATUS_MAP } from '@/store/resource-plan/gpu-demand';
import type { GpuDemandStatus } from '@/store/resource-plan/gpu-demand';

export const SERVICE_ONLY_COLUMNS = ['op_product_name', 'bk_biz_id'];

@Model('gpu-demand/table-column')
export class TableColumn {
  @Column('string', { name: '需求ID', fixed: 'left', minWidth: 120, index: 0 })
  id: string;

  @Column('string', { name: '运营产品', minWidth: 120, index: 1 })
  op_product_name: string;

  @Column('business', { name: '业务', minWidth: 100, index: 2 })
  bk_biz_id: number;

  @Column('number', { name: '需求卡数', minWidth: 80, index: 3 })
  total_gpu_num: number;

  @Column('number', { name: 'QPM(月调用峰值)', minWidth: 140, index: 4 })
  total_qpm_max: number;

  @Column('datetime', { name: '提单时间', sort: true, minWidth: 160, index: 5 })
  created_at: string;

  @Column('user', { name: '提单人', minWidth: 100, index: 6 })
  creator: string;

  @Column('enum', {
    name: '单据状态',
    option: GPU_DEMAND_STATUS_MAP,
    minWidth: 100,
    index: 7,
    meta: {
      display: {
        appearance: 'dynamic-status',
        appearanceProps: {
          statusObject: {
            success: [GPU_DEMAND_STATUS.DONE],
            fail: [GPU_DEMAND_STATUS.REJECT, GPU_DEMAND_STATUS.REJECT_ALL],
            wait: [GPU_DEMAND_STATUS.INIT],
            ing: [GPU_DEMAND_STATUS.PENDING],
            stop: [GPU_DEMAND_STATUS.TERMINATE],
          },
        },
      },
    },
  })
  status: GpuDemandStatus;
}

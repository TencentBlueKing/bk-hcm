import { Model, Column } from '@/decorator';
import { useResourcePlanStore } from '@/store/resource-plan';
import { GPU_DEMAND_STATUS_MAP } from '@/store/resource-plan/gpu-demand';
import type { GpuDemandStatus } from '@/store/resource-plan/gpu-demand';
import { toArray } from '@/common/util';

const resourcePlanStore = useResourcePlanStore();

export const SERVICE_ONLY_FIELDS = ['op_product_id', 'bk_biz_id'];

@Model('gpu-demand/search-condition')
export class SearchCondition {
  @Column('list', {
    name: '运营产品',
    list: async () => await resourcePlanStore.getOpProductList(),
    format: (value) => toArray(value).map((val: string) => Number(val)),
    props: {
      idKey: 'op_product_id',
      displayKey: 'op_product_name',
    },
    index: 0,
  })
  op_product_id: number;

  @Column('business', { name: '业务', index: 1 })
  bk_biz_id: number;

  @Column('string', {
    name: '需求ID',
    index: 2,
    props: {
      collapseTags: true,
      pasteFn: (value: string) =>
        value
          .split(/[\r\n,;\s]+/)
          .filter(Boolean)
          .map((tag: string) => ({ id: tag.trim(), name: tag.trim() })),
    },
  })
  id: string;

  @Column('enum', { name: '单据状态', option: GPU_DEMAND_STATUS_MAP, index: 3 })
  status: GpuDemandStatus;

  @Column('datetime', {
    name: '提单时间',
    props: {
      type: 'daterange',
      format: 'yyyy-MM-dd',
    },
    index: 4,
  })
  created_at: string;

  @Column('user', { name: '提单人', index: 5 })
  creator: string;
}

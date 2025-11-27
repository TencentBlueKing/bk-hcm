/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';
import dayjs from 'dayjs';
import { toArray } from '@/common/util';

@Model('order/host-apply-search')
export class HostApplySearch {
  @Column('req-type', {
    name: '需求类型',
    meta: {
      search: {
        format: (value) => {
          return toArray(value).map((val: string) => Number(val));
        },
      },
    },
    index: 1,
  })
  require_type: number;

  @Column('req-stage', { name: '单据状态', index: 1 })
  stage: string;

  @Column('string', {
    name: '单号',
    meta: {
      display: { appearance: 'tag-input' },
      search: {
        format: (value) => {
          const values = toArray(value).map((val: string) => Number(val));
          return values.filter((val: number) => !isNaN(val));
        },
      },
    },
    index: 1,
  })
  order_id: number;

  @Column('string', {
    name: '子单号',
    meta: {
      display: { appearance: 'tag-input' },
      search: {
        format: (value) => {
          return toArray(value).map((val) => String(val));
        },
      },
    },
    index: 1,
  })
  suborder_id: string;

  @Column('datetime', {
    name: '申请时间',
    meta: {
      search: {
        converter(value) {
          const start = new Date(value[0]);
          const end = new Date(value[1]);
          return {
            start: dayjs(start).format('YYYY-MM-DD'),
            end: dayjs(end).format('YYYY-MM-DD'),
          };
        },
      },
    },
    index: 1,
  })
  create_at: [Date, Date];

  @Column('user', { name: '申请人', index: 1 })
  bk_username: string;
}

@Model('order/host-apply-search')
export class HostApplySearchNonBusiness extends HostApplySearch {
  @Column('business', {
    name: '业务',
    meta: {
      search: {
        format(value) {
          // 从qs中解析出来可能是'0'
          if (value === '0' || !value?.length) {
            return [0];
          }
          if (!Array.isArray(value)) {
            return [Number(value)];
          }
          return value.map((val: string) => Number(val));
        },
      },
    },
    index: 0,
  })
  bk_biz_id: number;

  @Column('enum', {
    name: '单据来源',
    option: {
      business: '业务单据',
      purchase_to_resource_pool: '资源池采购',
    },
    index: 1,
  })
  source: string;
}

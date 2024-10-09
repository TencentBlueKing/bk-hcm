import { ResourceTypeEnum } from '@/common/resource-constant';
import type { DisplayType, DisplayAppearanceType, DisplayOnType } from '@/components/form/typings';
import { type TaskType, TaskClbType } from '@/views/task/typings';

export const baseFieldIds = ['created_at', 'ended_at', 'state', 'reason'];

const clbFieldIds = {
  [TaskClbType.CREATE_L4_LISTENER]: [
    'created_at',
    'ended_at',
    'param.clb_vip_domain',
    'param.cloud_clb_id',
    'param.cloud_clb_name',
    'param.protocol',
    'param.listener_port',
    'param.scheduler',
    'param.health_check',
    'param.session',
    'state',
    'reason',
  ],
  [TaskClbType.CREATE_L7_LISTENER]: [
    'created_at',
    'ended_at',
    'param.clb_vip_domain',
    'param.cloud_clb_id',
    'param.cloud_clb_name',
    'param.protocol',
    'param.listener_port',
    'param.ssl_mode',
    'param.cert_cloud_ids',
    'param.ca_cloud_id',
    'state',
    'reason',
  ],
  [TaskClbType.CREATE_L7_FILTER]: [
    'created_at',
    'ended_at',
    'param.clb_vip_domain',
    'param.cloud_clb_id',
    'param.cloud_clb_name',
    'param.protocol',
    'param.listener_port',
    'param.domain',
    'param.url_path',
    'param.health_check',
    'param.session',
    'state',
    'reason',
  ],
};

const clbRerunParamFieldIds = {
  [TaskClbType.CREATE_L4_LISTENER]: {
    'param.clb_vip_domain': {
      editable: false,
      display: {
        on: 'cell' as DisplayOnType,
      },
    },
    'param.cloud_clb_id': {
      editable: false,
      display: {},
    },
    'param.protocol': {
      editable: false,
      display: {},
    },
    'param.listener_port': {
      editable: false,
      display: {},
    },
    'param.scheduler': {
      editable: true,
      display: {},
    },
    'param.health_check': {
      editable: true,
      display: {
        appearance: 'select' as DisplayAppearanceType,
      },
    },
    'param.session': {
      editable: true,
      display: {},
      rules: [
        {
          validator: (value: number) => value >= 0 && value <= 1000,
          message: '0 - 10000',
        },
      ],
    },
    'param.validate_result': {
      editable: false,
      display: {},
    },
  },
  [TaskClbType.CREATE_L7_LISTENER]: {
    'param.clb_vip_domain': {
      editable: false,
      display: {
        on: 'cell' as DisplayOnType,
      },
    },
    'param.cloud_clb_id': {
      editable: false,
      display: {},
    },
    'param.protocol': {
      editable: false,
      display: {},
    },
    'param.listener_port': {
      editable: false,
      display: {},
    },
    'param.ssl_mode': {
      editable: true,
      display: {},
    },
    'param.cert_cloud_ids': {
      editable: true,
      display: {},
    },
    'param.ca_cloud_id': {
      editable: true,
      display: {},
    },
    'param.validate_result': {
      editable: false,
      display: {},
    },
  },
  [TaskClbType.CREATE_L7_FILTER]: {
    'param.clb_vip_domain': {
      editable: false,
      display: {
        on: 'cell' as DisplayOnType,
      },
    },
    'param.cloud_clb_id': {
      editable: false,
      display: {},
    },
    'param.protocol': {
      editable: false,
      display: {},
    },
    'param.listener_port': {
      editable: false,
      display: {},
    },
    'param.domain': {
      editable: true,
      display: {},
    },
    'param.url_path': {
      editable: true,
      display: {},
    },
    'param.scheduler': {
      editable: true,
      display: {},
    },
    'param.health_check': {
      editable: true,
      display: {},
    },
    'param.session': {
      editable: true,
      display: {},
    },
    'param.validate_result': {
      editable: false,
      display: {},
    },
  },
};

export const fieldIdMap = new Map<ResourceTypeEnum, { [k in TaskType]?: string[] }>();
export const fieldRerunIdMap = new Map<
  ResourceTypeEnum,
  { [k in TaskType]?: Record<string, { editable: boolean; display: DisplayType; rules?: any[] }> }
>();

fieldIdMap.set(ResourceTypeEnum.CLB, clbFieldIds);
fieldRerunIdMap.set(ResourceTypeEnum.CLB, clbRerunParamFieldIds);

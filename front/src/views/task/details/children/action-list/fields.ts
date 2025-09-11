import { PropertyColumnConfig } from '@/model/typings';
import { ResourceTypeEnum } from '@/common/resource-constant';
import type { DisplayType, DisplayAppearanceType, DisplayOnType } from '@/components/form/typings';
import { type TaskType, TaskClbType } from '@/views/task/typings';

export const baseFieldIds = ['created_at', 'updated_at', 'state', 'reason'];

export const baseColumnConfig: Record<string, PropertyColumnConfig> = {
  created_at: {
    sort: true,
  },
  updated_at: {
    sort: true,
  },
  state: {
    sort: true,
  },
};

const clbBaseFieldIds = [
  'created_at',
  'updated_at',
  'param.clb_vip_domain',
  'param.cloud_clb_id',
  'param.protocol',
  'param.listener_port',
];
const clbSopsBaseFieldIds = [
  'created_at',
  'updated_at',
  'param.clb_vip_domain',
  'param.cloud_lb_id',
  'param.protocol',
  'param.port',
];

const clbFieldIds = {
  [TaskClbType.CREATE_L4_LISTENER]: [
    ...clbBaseFieldIds,
    'param.scheduler',
    'param.health_check',
    'param.session',
    'state',
    'reason',
  ],
  [TaskClbType.CREATE_L7_LISTENER]: [
    ...clbBaseFieldIds,
    'param.ssl_mode',
    'param.cert_cloud_ids',
    'param.ca_cloud_id',
    'state',
    'reason',
  ],
  [TaskClbType.CREATE_L7_RULE]: [
    ...clbBaseFieldIds,
    'param.domain',
    'param.url_path',
    'param.health_check',
    'param.session',
    'state',
    'reason',
  ],
  [TaskClbType.BINDING_L4_RS]: [
    ...clbBaseFieldIds,
    'param.inst_type',
    'param.rs_ip',
    'param.rs_port',
    'param.weight',
    'state',
    'reason',
  ],
  [TaskClbType.BINDING_L7_RS]: [
    ...clbBaseFieldIds,
    'param.inst_type',
    'param.domain',
    'param.url_path',
    'param.rs_ip',
    'param.rs_port',
    'param.weight',
    'state',
    'reason',
  ],
  [TaskClbType.DELETE_LISTENER]: [...clbSopsBaseFieldIds, 'state', 'reason'],
  [TaskClbType.UNBIND_LAYER4_RS]: [...clbSopsBaseFieldIds, 'param.rs_list', 'state', 'reason'],
  [TaskClbType.MODIFY_LAYER4_RS_WEIGHT]: [
    ...clbSopsBaseFieldIds,
    'param.ip',
    'param.weight',
    'param.new_rs_weight',
    'state',
    'reason',
  ],
  [TaskClbType.MODIFY_LAYER7_RS_WEIGHT]: [
    ...clbSopsBaseFieldIds,
    'param.domain',
    'param.url',
    'param.ip',
    'param.weight',
    'param.new_rs_weight',
    'state',
    'reason',
  ],
  [TaskClbType.TARGET_GROUP_MODIFY_WEIGHT]: [
    ...clbSopsBaseFieldIds,
    'param.domain',
    'param.url',
    'param.ip',
    'param.weight',
    'param.new_weight',
    'state',
    'reason',
  ],
  [TaskClbType.TARGET_GROUP_REMOVE_RS]: [
    ...clbSopsBaseFieldIds,
    'param.domain',
    'param.url',
    'param.ip',
    'param.weight',
    'state',
    'reason',
  ],
};

const clbBaseRerunParamFieldIds = {
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
};

const clbRerunParamFieldIds = {
  [TaskClbType.CREATE_L4_LISTENER]: {
    ...clbBaseRerunParamFieldIds,
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
      display: {
        showOverflowTooltip: true,
      },
    },
  },
  [TaskClbType.CREATE_L7_LISTENER]: {
    ...clbBaseRerunParamFieldIds,
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
      display: {
        showOverflowTooltip: true,
      },
    },
  },
  [TaskClbType.CREATE_L7_RULE]: {
    ...clbBaseRerunParamFieldIds,
    'param.domain': {
      editable: true,
      display: {
        showOverflowTooltip: true,
      },
    },
    'param.url_path': {
      editable: true,
      display: {
        showOverflowTooltip: true,
      },
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
    },
    'param.validate_result': {
      editable: false,
      display: {
        showOverflowTooltip: true,
      },
    },
  },
  [TaskClbType.BINDING_L4_RS]: {
    ...clbBaseRerunParamFieldIds,
    'param.inst_type': {
      editable: true,
    },
    'param.rs_ip': {
      editable: true,
    },
    'param.rs_port': {
      editable: true,
      display: {},
    },
    'param.weight': {
      editable: true,
    },
    'param.validate_result': {
      editable: false,
      display: {
        showOverflowTooltip: true,
      },
    },
  },
  [TaskClbType.BINDING_L7_RS]: {
    ...clbBaseRerunParamFieldIds,
    'param.domain': {
      editable: false,
      display: {
        showOverflowTooltip: true,
      },
    },
    'param.url_path': {
      editable: false,
      display: {
        showOverflowTooltip: true,
      },
    },
    'param.inst_type': {
      editable: true,
    },
    'param.rs_ip': {
      editable: true,
    },
    'param.rs_port': {
      editable: true,
      display: {},
    },
    'param.weight': {
      editable: true,
    },
    'param.validate_result': {
      editable: false,
      display: {
        showOverflowTooltip: true,
      },
    },
  },
};

export const fieldIdMap = new Map<ResourceTypeEnum, { [k in TaskType]?: string[] }>();
export const fieldRerunIdMap = new Map<
  ResourceTypeEnum,
  { [k in TaskType]?: Record<string, { editable: boolean; display?: DisplayType; rules?: any[] }> }
>();
export const fieldRerunBaseIdMap = new Map<
  ResourceTypeEnum,
  Record<string, { editable: boolean; display: DisplayType; rules?: any[] }>
>();

fieldIdMap.set(ResourceTypeEnum.CLB, clbFieldIds);
fieldRerunIdMap.set(ResourceTypeEnum.CLB, clbRerunParamFieldIds);
fieldRerunBaseIdMap.set(ResourceTypeEnum.CLB, clbBaseRerunParamFieldIds);

import './resource-distribution.scss';

import {
  defineComponent,
  ref,
} from 'vue';

import {
  RESOURCE_TYPES,
} from '@/common/constant';

import StepDialog from '@/components/step-dialog/step-dialog';

import {
  useI18n,
} from 'vue-i18n';

import {
  useResourceStore,
} from '@/store/resource';

export default defineComponent({
  components: {
    StepDialog,
  },

  props: {
    title: {
      type: String,
    },
    isShow: {
      type: Boolean,
    },
    chooseResourceType: {
      type: Boolean,
    },
    data: {
      type: Array,
    },
  },

  emits: ['update:isShow'],

  setup(props, { emit }) {
    // use hooks
    const {
      t,
    } = useI18n();

    const resourceStore = useResourceStore();

    // 状态
    const business = ref('');
    const businessList = ref([]);
    const isBindVPC = ref(false);
    const isBingdingVPC = ref(false);
    const hasBindVPC = ref(false);
    const disableNext = ref(true);
    const resourceTypes = ref([]);
    const VPCColumns = [
      {
        label: 'ID',
        field: 'id',
      },
      {
        label: '资源 ID',
        field: 'vpc_cid',
      },
      {
        label: '名称',
        field: 'name',
      },
      {
        label: '云厂商',
        field: 'vendor',
      },
      {
        label: '地域',
        field: 'region',
      },
      {
        label: 'IPv4 CIDR',
        field: 'ipv4_cidr',
      },
      {
        label: '云区域',
        field: 'bk_cloud_id',
      },
    ];
    const businessColumns = [
      {
        label: '云厂商',
        field: 'id',
      },
      {
        label: '数量',
        field: 'id',
      },
      {
        label: '云区域（ID）',
        field: 'id',
      },
    ];

    // 方法
    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleConfirm = () => {
      handleClose();
    };
    // 绑定vpc到云区域
    const handleBindVPC = () => {
      isBingdingVPC.value = true;
      resourceStore
        .bindVPCWithCloudArea(props.data)
        .then(() => {
          disableNext.value = false;
        })
        .finally(() => {
          isBingdingVPC.value = false;
        });
    };

    return {
      business,
      businessList,
      isBindVPC,
      isBingdingVPC,
      hasBindVPC,
      disableNext,
      resourceTypes,
      VPCColumns,
      businessColumns,
      t,
      handleClose,
      handleConfirm,
      handleBindVPC,
    };
  },

  render() {
    // 渲染每一步
    const steps = [
      {
        title: '前置检查',
        disableNext: this.disableNext,
        component: () => <>
          <bk-table
            class="mt20"
            row-hover="auto"
            columns={this.VPCColumns}
            data={this.data}
          />
          <bk-checkbox class="mt5" v-model={this.isBindVPC}>
            注：VPC绑定云区域信息无法修改，请提前确认
          </bk-checkbox>
        </>,
        footer: () => <>
          <bk-button
            class="mr10"
            loading={this.isBingdingVPC}
            disabled={!this.isBindVPC}
            onClick={this.handleBindVPC}
          >VPC 绑定云区域</bk-button>
        </>,
      },
      {
        title: '分配确认',
        component: () => <>
          <section class="resource-head">
            { `${this.t('目标业务')}:${this.business}` }
            <bk-select
              v-model={this.business}
              filterable
              class="ml10"
            >
              {
                this.businessList.map(business => <bk-option
                  value={business.value}
                  label={business.label}
                />)
              }
            </bk-select>
          </section>
          <bk-table
            class="mt20"
            row-hover="auto"
            columns={this.businessColumns}
            data={this.data}
          />
        </>,
      },
    ];

    // 快速分配的时候需要选择资源类型
    if (this.chooseResourceType) {
      steps.unshift({
        title: this.t('资源类型'),
        component: () => <>
          <section>
            <span>分配资源类型</span>
            <bk-checkbox-group
              class="resource-types"
              v-model={this.resourceTypes}
            >
              {
                RESOURCE_TYPES.map((type) => {
                  return <bk-checkbox label={type.type}>{ this.t(type.name) }</bk-checkbox>;
                })
              }
            </bk-checkbox-group>
          </section>
        </>,
      });
    }

    return <>
      <step-dialog
        title={this.title}
        isShow={this.isShow}
        steps={steps}
        onConfirm={this.handleConfirm}
        onCancel={this.handleClose}
      >
      </step-dialog>
    </>;
  },
});

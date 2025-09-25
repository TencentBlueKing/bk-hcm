<script setup lang="ts">
import { reactive, inject, Ref, watch, computed, ref, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { IAccountItem } from '@/typings';
import { cloneDeep, isEqual } from 'lodash';
import { Info } from 'bkui-vue/lib/icon';
import { isEmpty } from '@/common/util';
import { ILoadBalanceDeviceCondition } from '../typing';
import { VendorEnum, ResourceTypeEnum, TARGET_GROUP_PROTOCOLS } from '@/common/constant';
import { ModelPropertySearch } from '@/model/typings';
import { LB_NETWORK_TYPE_MAP } from '@/constants';

defineOptions({ name: 'device-condition' });

const props = defineProps<{ loading: boolean; countChange: boolean }>();

const emit = defineEmits<{
  'count-change': [val: boolean];
  save: [newCondition: ILoadBalanceDeviceCondition];
}>();
const businessId = inject<Ref<number>>('currentGlobalBusinessId');

let timeout: string | number | NodeJS.Timeout = null;

const { t } = useI18n();

const formModel = reactive<ILoadBalanceDeviceCondition>({ account_id: '', vendor: VendorEnum.TCLOUD, lb_regions: [] });
const originFormModel: ILoadBalanceDeviceCondition = reactive(cloneDeep(formModel));
const isShow = ref(false);
const hasSaved = ref(false);
const formRef = ref(null);

const hasChange = computed(() => !isEqual(formModel, originFormModel));
// 除了云账号是否有输入其他条件
const hasAnyCondition = computed(
  () =>
    Object.entries(formModel).filter(([key, value]) => !['account_id', 'vendor'].includes(key) && !isEmpty(value))
      .length,
);
const disabled = computed(() => (!hasSaved.value && !hasChange.value) || !hasAnyCondition.value);

const handleAccountChange = (item: IAccountItem) => {
  formModel.vendor = item?.vendor;
  formModel.lb_regions = [];
};
const handlePaste = (value: any) => value.split(/,|;|\n|\s/).map((tag: any) => ({ id: tag, name: tag }));
const handleSave = async () => {
  await formRef.value.validate();
  Object.keys(formModel).forEach((key) => {
    originFormModel[key] = formModel[key];
  });
  hasSaved.value = true;
  isShow.value = false;
  emit('count-change', false);
  emit('save', cloneDeep(originFormModel));
};
const handleReset = () => {
  Object.keys(formModel).forEach((key) => {
    formModel[key] = originFormModel[key];
  });
};

watch(
  () => formModel.account_id,
  (val) => {
    originFormModel.account_id = val;
    originFormModel.vendor = formModel.vendor;
  },
  {
    once: true,
  },
);
watch(
  () => hasChange.value,
  async (val) => {
    if (timeout) {
      clearTimeout(timeout);
      timeout = null;
    }
    // 放到下次循环，因为lb_regions在更换云账号时要清空
    await nextTick();
    if (val && hasAnyCondition.value) {
      timeout = setTimeout(() => (emit('count-change', false), (isShow.value = true)), 120000);
    }
  },
);
watch(
  () => props.countChange,
  async (val) => {
    if (val) {
      isShow.value = true;
    }
  },
);

const conditionField: ModelPropertySearch[] = [
  {
    id: 'account_id',
    type: 'account',
    name: '云账号',
    props: {
      bizId: businessId.value,
      autoSelect: true,
      resourceType: ResourceTypeEnum.CLB,
      onChange: handleAccountChange,
    },
  },
  {
    id: 'lb_regions',
    type: 'region',
    name: '地域',
    props: {
      vendor: formModel.vendor,
      multiple: true,
      clearable: true,
      collapseTags: true,
    },
  },
  {
    id: 'lb_vips',
    name: '负载均衡 VIP',
    type: 'string',
    props: {
      maxData: 500,
      collapseTags: true,
      pasteFn: handlePaste,
    },
  },
  {
    id: 'cloud_lb_ids',
    name: '负载均衡 ID',
    type: 'string',
    props: {
      maxData: 500,
      collapseTags: true,
      pasteFn: handlePaste,
    },
  },
  {
    id: 'lbl_protocols',
    name: '监听器协议',
    type: 'enum',
    props: {
      option: TARGET_GROUP_PROTOCOLS.reduce((prev: { [key: string]: any }, cur) => {
        prev[cur] = cur;
        return prev;
      }, {}),
    },
  },
  {
    id: 'lbl_ports',
    name: '监听器端口',
    type: 'string',
    props: {
      maxData: 1000,
      collapseTags: true,
      pasteFn: handlePaste,
    },
  },
  {
    id: 'target_ips',
    name: 'RS IP',
    type: 'string',
    props: {
      maxData: 5000,
      collapseTags: true,
      pasteFn: handlePaste,
    },
  },
  {
    id: 'target_ports',
    name: 'RS端口',
    type: 'string',
    props: {
      maxData: 500,
      collapseTags: true,
      pasteFn: handlePaste,
    },
  },
  {
    id: 'rule_domains',
    name: 'HTTP/HTTPS监听器域名',
    type: 'string',
    props: {
      maxData: 500,
      collapseTags: true,
      pasteFn: handlePaste,
    },
  },
  {
    id: 'rule_urls',
    name: 'URL路径',
    type: 'string',
    props: {
      maxData: 500,
      collapseTags: true,
      pasteFn: handlePaste,
    },
  },
  {
    id: 'lb_network_types',
    name: '网络类型',
    type: 'enum',
    props: {
      option: LB_NETWORK_TYPE_MAP,
    },
  },
  {
    id: 'lb_ip_versions',
    name: 'IP版本',
    type: 'enum',
    props: {
      option: {
        ipv4: 'IPv4',
        ipv6: 'IPv6',
        ipv6_dual_stack: 'IPv6DualStack',
        ipv6_nat64: 'IPv6Nat64',
      },
    },
  },
  {
    id: 'lb_domains',
    name: '负载均衡域名',
    type: 'string',
    props: {
      maxData: 500,
      collapseTags: true,
      pasteFn: handlePaste,
    },
  },
];
</script>

<template>
  <div class="device-condition">
    <div class="header">{{ t('检索条件') }}</div>
    <div class="condition">
      <bk-form ref="formRef" class="condition-form g-expand" form-type="vertical" :model="formModel">
        <bk-form-item
          :label="t(field.name)"
          :property="field.id"
          v-for="field in conditionField"
          :key="field.id"
          :required="field.id === 'account_id'"
        >
          <component :is="`hcm-search-${field.type}`" v-bind="field.props" v-model="formModel[field.id]" />
        </bk-form-item>
      </bk-form>
    </div>

    <div class="footer">
      <bk-popover theme="light" :is-show="isShow" trigger="manual">
        <bk-button
          class="mr6 save"
          theme="primary"
          @click="handleSave"
          :loading="loading"
          :disabled="disabled"
          v-bk-tooltips="{ content: t('请输入检索条件后点击'), disabled: !disabled }"
        >
          {{ t('查询') }}
        </bk-button>
        <template #content>
          <div class="tips">
            <info class="warning" />
            <div>
              {{
                countChange
                  ? t('后台数据已发生变化，请点击下方查询按钮更新检索')
                  : t('检索条件有更新，请点击下方查询按钮更新检索')
              }}
            </div>
          </div>
        </template>
      </bk-popover>
      <bk-button @click="handleReset">{{ t('重置') }}</bk-button>
    </div>
  </div>
</template>

<style scoped lang="scss">
.device-condition {
  height: 100%;
  padding: 16px 0 16px 24px;
  position: relative;

  .header {
    font-weight: 700;
    color: #313238;
    margin-bottom: 16px;
  }

  .condition {
    height: calc(100% - 64px);
    overflow-y: auto;
    position: relative;

    .condition-form {
      padding-right: 24px;
    }
  }

  .footer {
    position: sticky;
    bottom: 0;
    width: calc(100% + 48px);
    margin-left: -24px;
    line-height: 48px;
    background: #fafbfd;
    border: 1px solid #dcdee5;
    padding-left: 24px;
    z-index: 999999;

    .bk-button {
      width: 88px;
    }
  }
}

.tips {
  color: #4d4f56;
  display: flex;
  align-items: center;
  width: 180px;

  .warning {
    color: #f59500;
    margin-right: 5px;
  }
}
</style>

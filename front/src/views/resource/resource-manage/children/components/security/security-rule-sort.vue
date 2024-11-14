<script lang="ts" setup>
import { watch, reactive, PropType, computed } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { Message } from 'bkui-vue';
import { useResourceStore } from '@/store';
import { SecurityRuleEnum } from '@/typings';
import { VueDraggable } from 'vue-draggable-plus';
import { isEqual } from 'lodash';

const props = defineProps({
  filter: {
    type: Object as PropType<any>,
  },
  id: {
    type: String as PropType<any>,
  },
  type: {
    type: String as PropType<any>,
    default: 'ingress',
  },
  show: {
    type: Boolean as PropType<any>,
  },
});

const emit = defineEmits(['update:show', 'sortDone']);

const { t } = useI18n();
const route = useRoute();
const resourceStore = useResourceStore();

const inColumns = ['排序', '源地址', '协议', '端口', '策略'];
const outColumns = ['排序', '目标地址', '协议', '端口', '策略'];
const updateFields = [
  'cloud_policy_index',
  'protocol',
  'port',
  'cloud_service_id',
  'cloud_service_group_id',
  'ipv4_cidr',
  'ipv6_cidr',
  'cloud_address_id',
  'cloud_address_group_id',
  'cloud_target_security_group_id',
  'action',
  'memo',
];
// tab 信息
const types = [
  { name: 'ingress', label: t('入站规则') },
  { name: 'egress', label: t('出站规则') },
];

const states = reactive<any>({
  datas: [],
  originDatas: [],
  isLoading: true,
});

const getList = async () => {
  try {
    const list = await resourceStore.getAllSort({
      id: props?.id,
      vendor: route.query?.vendor,
      filter: props?.filter,
    });
    states.datas = (list as any[])?.sort((after: any, prev: any) => after.cloud_policy_index - prev.cloud_policy_index);
    states.originDatas = [...states.datas];
    return list;
  } catch {
    states.datas = [];
    states.originDatas = [];
  } finally {
    states.isLoading = false;
  }
};
const activeType: any = computed(() => props.type);
const columns = computed(() => (activeType.value === 'ingress' ? inColumns : outColumns));
const hasChange = computed(() => !isEqual(states.datas, states.originDatas));

const handleCancel = () => {
  emit('update:show', false);
};
const handleConfirm = async () => {
  const data = states.datas.map((item: { [x: string]: any }, index: number) => {
    const cleaned: { [x: string]: any } = {};
    updateFields.forEach((key) => {
      if (item[key]) {
        cleaned[key] = item[key];
      }
    });
    cleaned.cloud_policy_index = index;
    return cleaned;
  });
  states.isLoading = true;
  await resourceStore.updateRulesSort(
    {
      [activeType.value === 'ingress' ? 'ingress_rule_set' : 'egress_rule_set']: data,
    },
    String(route.query.vendor),
    props.id,
  );
  states.isLoading = false;
  Message({
    message: t('排序成功'),
    theme: 'success',
  });
  emit('update:show', false);
  emit('sortDone');
};

watch(
  () => props.show,
  (val) => {
    if (val) getList();
  },
  {
    immediate: true,
  },
);
</script>

<template>
  <div>
    <div class="security-rule-sort">
      <bk-loading :loading="states.isLoading">
        <section class="rule-main">
          <bk-radio-group v-model="activeType" :disabled="states.isLoading">
            <bk-radio-button
              v-for="item in types"
              :key="item.name"
              :label="item.name"
              :disabled="item.name !== activeType"
            >
              {{ item.label }}
            </bk-radio-button>
          </bk-radio-group>
        </section>
        <div class="drag-table">
          <div class="drag-header">
            <template v-for="column in columns" :key="column">
              <div>{{ t(column) }}</div>
            </template>
          </div>
          <VueDraggable v-model="states.datas" :animation="200" handle=".drag-body-tr" v-if="states.datas[0]">
            <div v-for="(data, index) in states.datas" :key="data.id" class="drag-body-tr">
              <div class="drag-body-cell">
                <div>
                  <i class="hcm-icon bkhcm-icon-grag-fill mr5 sort"></i>
                  <span class="sort sort-number">{{ index + 1 }}</span>
                </div>
                <div class="address">
                  {{
                    data.cloud_address_group_id ||
                    data.cloud_address_id ||
                    data.cloud_service_group_id ||
                    data.cloud_target_security_group_id ||
                    data.ipv4_cidr ||
                    data.ipv6_cidr ||
                    data.cloud_remote_group_id ||
                    data.remote_ip_prefix ||
                    (data.source_address_prefix === '*' ? t('ALL') : data.source_address_prefix) ||
                    data.source_address_prefixes ||
                    data.cloud_source_security_group_ids ||
                    data.destination_address_prefix ||
                    data.destination_address_prefixes ||
                    data.cloud_destination_security_group_ids ||
                    (data?.ethertype === 'IPv6' ? '::/0' : '0.0.0.0/0')
                  }}
                </div>
                <div class="agreement">
                  {{ data.cloud_service_id || `${data.protocol}` }}
                </div>
                <div class="port">
                  {{
                    data.cloud_service_id ||
                    `${
                      data.port || data.to_port || data.destination_port_range || data.destination_port_ranges || '--'
                    }`
                  }}
                </div>
                <div class="tactics">
                  {{ SecurityRuleEnum[data.action] || '--' }}
                </div>
              </div>
            </div>
          </VueDraggable>
          <template v-else>
            <div class="security-empty-container">
              <bk-exception
                class="exception-wrap-item exception-part"
                type="empty"
                scene="part"
                :description="t('无规则，默认拒绝所有流量')"
              />
            </div>
          </template>
        </div>
      </bk-loading>
    </div>
    <div class="footer">
      <bk-button theme="primary" class="confirm" @click="handleConfirm" :disabled="!hasChange">
        {{ t('提交') }}
      </bk-button>
      <bk-button class="cancel" @click="handleCancel">
        {{ t('取消') }}
      </bk-button>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.security-rule-sort {
  padding: 28px 40px;
}
.rule-main {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}

.security-empty-container {
  display: flex;
  align-items: center;
  margin: auto;
}
.drag-table {
  height: calc(100vh - 220px);
  overflow: auto;
}
.drag-header {
  line-height: 42px;
}
.drag-body-tr {
  &[draggable='true'] {
    background: #eaebf0;
  }
}
.drag-body-cell {
  line-height: 38px;
}
.drag-header,
.drag-body-cell {
  display: flex;
  align-items: center;
  font-size: 12px;
  color: #63656e;
  text-align: center;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  border-bottom: 1px solid #dcdee5;
  > div {
    flex: 1;
  }
  :first-child {
    flex: 0;
    flex-basis: 50px;
  }
}
.sort {
  vertical-align: middle;
  color: #c4c6cc;
}
.sort-number {
  padding: 1px 8px;
  background: white;
  color: #979ba5;
  border: 1px solid #dcdee5;
  border-radius: 4px;
}
.footer {
  position: fixed;
  bottom: 0;
  width: 100%;
  height: 48px;
  background: #fafbfd;
  box-shadow: 0 -1px 0 0 #dcdee5;
  display: flex;
  align-items: center;
  padding: 0 24px;
}
.confirm,
.cancel {
  width: 88px;
  margin-right: 8px;
}
</style>

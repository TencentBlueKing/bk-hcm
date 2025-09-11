<script setup lang="ts">
import { watch, computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { InfoLine } from 'bkui-vue/lib/icon';
import routeQuery from '@/router/utils/query';

import { DeviceTabEnum, ILoadBalanceDeviceCondition, ICount } from '../common';

import Empty from './empty.vue';
import LargeData from './large-data.vue';
import ListenerTable from './content/listener-table.vue';
import RsTable from './content/rs-table.vue';
import UrlRuleTable from './content/url-rule-table.vue';

defineOptions({ name: 'device-container' });

const props = defineProps<{ condition: ILoadBalanceDeviceCondition; count: ICount }>();
const emit = defineEmits(['getList']);

const TAB_LIST = [
  {
    id: DeviceTabEnum.LISTENER,
    name: '监听器',
  },
  {
    id: DeviceTabEnum.URL,
    name: 'URL规则',
  },
  {
    id: DeviceTabEnum.RS,
    name: 'RS',
  },
];
const DEVICE_VIEW_LIST = {
  [DeviceTabEnum.LISTENER]: ListenerTable,
  [DeviceTabEnum.URL]: UrlRuleTable,
  [DeviceTabEnum.RS]: RsTable,
};
const DEVICE_VIEW_LIST_COUNT = {
  [DeviceTabEnum.LISTENER]: 'listenerCount',
  [DeviceTabEnum.URL]: 'urlCount',
  [DeviceTabEnum.RS]: 'rsCount',
};

const { t } = useI18n();

const max = ref(10000);
const tabValue = ref(DeviceTabEnum.LISTENER);

const search = computed(() => !!props.condition.account_id);
const largeData = computed(() => props.count[DEVICE_VIEW_LIST_COUNT[tabValue.value]] > max.value);
const activeComponent = computed(() => (largeData.value ? LargeData : DEVICE_VIEW_LIST[tabValue.value]));

const overCount = (num: number) => num > max.value;
const handleListDone = () => {
  emit('getList');
};

watch(
  () => tabValue.value,
  () =>
    routeQuery.set({
      page: undefined,
      _t: Date.now(),
    }),
);
</script>

<template>
  <div class="device-container">
    <div class="device-container-tab">
      <bk-radio-group v-model="tabValue">
        <bk-radio-button v-for="tab in TAB_LIST" :label="tab.id" :key="tab.id">
          {{ t(tab.name) }}
          <text v-if="search">
            <text class="num" v-if="!overCount(props.count[DEVICE_VIEW_LIST_COUNT[tab.id]])">
              {{ props.count[DEVICE_VIEW_LIST_COUNT[tab.id]] }}
            </text>
            <info-line class="warning" v-else />
          </text>
        </bk-radio-button>
      </bk-radio-group>
    </div>
    <div class="content" :class="{ 'no-data': !search, 'large-data': largeData }">
      <empty v-if="!search" />
      <template v-else>
        <!-- <listener-table /> -->
        <component :is="activeComponent" :condition="props.condition" @get-list="handleListDone" />
      </template>
    </div>
  </div>
</template>

<style scoped lang="scss">
.device-container {
  height: 100%;
  background: #f5f7fa;
  padding: 24px;

  .no-data,
  .large-data {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .content {
    margin-top: 16px;
    background: white;
    padding: 24px;
    height: calc(100% - 50px);
  }

  .device-container-tab {
    display: flex;
    align-items: center;
    gap: 20px;

    .bk-radio-tab {
      background: white;

      .bk-radio-button {
        min-width: 80px;

        &.is-checked {
          .num {
            background: white;
          }
        }
      }
      .num {
        font-size: 12px;
        padding: 0 8px;
        background: #f0f1f5;
        border-radius: 8px;
        line-height: 16px;
      }
      .warning {
        color: #ea3636;
      }
    }
    .tip {
      font-size: 12px;
      color: #979ba5;
      display: flex;
      align-items: center;
      gap: 3px;

      text {
        color: #3a84ff;
        font-weight: 700;
      }
    }
  }
}
</style>

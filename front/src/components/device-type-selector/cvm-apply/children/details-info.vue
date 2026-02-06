<script setup lang="ts">
import { watchEffect, inject, ref } from 'vue';
import { ArrowsRight } from 'bkui-vue/lib/icon';
import isEqual from 'lodash/isEqual';
import { QueryRuleOPEnum } from '@/typings';
import { VendorEnum } from '@/common/constant';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import gridItem from '@/components/layout/grid-container/grid-item.vue';
import { getZoneCn } from '@/views/ziyanScr/cvm-web/transform';
import useCvmChargeType from '@/views/ziyanScr/hooks/use-cvm-charge-type';
import { useCvmDeviceStore, type ICvmDevicetypeItem } from '@/store/cvm/device';
import { RequirementType } from '@/store/config/requirement';
import { transformSimpleCondition } from '@/utils/search';
import type { ICvmDeviceTypeFormData } from '../../typings';
import { RES_ASSIGN_TYPE } from '../../constants';

const props = defineProps<{
  requireType: RequirementType;
  region: string;
  data: Partial<ICvmDeviceTypeFormData>;
  originalData: Partial<ICvmDeviceTypeFormData>;
}>();

const emit = defineEmits<{ edit: []; 'update-selected': [value: ICvmDeviceTypeFormData['deviceTypeList']] }>();

const editMode = inject('editMode');

const cvmDeviceStore = useCvmDeviceStore();

const { cvmChargeTypeNames } = useCvmChargeType();

const deviceTypeList = ref<ICvmDevicetypeItem[]>([]);

watchEffect(async () => {
  // 添加模式会传入选择好的deviceTypeList，编辑模式只会传入deviceTypes，此时查询得到deviceTypeList
  if (!props.data.deviceTypeList?.length && props.data.deviceTypes?.length) {
    const params = transformSimpleCondition(
      {
        vendor: VendorEnum.ZIYAN,
        region: props.region,
        device_type: props.data.deviceTypes,
      },
      [
        { id: 'vendor', name: 'vendor', op: QueryRuleOPEnum.EQ, type: 'string' },
        { id: 'region', name: 'region', type: 'region', op: QueryRuleOPEnum.EQ },
        { id: 'device_type', name: 'device_type', type: 'list', op: QueryRuleOPEnum.IN },
      ],
    );

    const { list } = await cvmDeviceStore.getDeviceTypeFullList({ filter: params });
    deviceTypeList.value = list;
  } else {
    deviceTypeList.value = props.data.deviceTypeList ?? [];
  }

  emit('update-selected', deviceTypeList.value);
});
</script>

<template>
  <div class="device-type-info">
    <grid-container class="" :column="1" :label-width="110">
      <template v-if="!editMode">
        <grid-item label="可用区">
          <span v-if="data.zones?.[0] === 'all'">全部可用区</span>
          <grid-container :column="1" :label-width="'0px'" :gap="[4, 0]" v-else>
            <grid-item v-for="(zone, index) in data.zones" :key="index">{{ getZoneCn(zone) }}</grid-item>
          </grid-container>
        </grid-item>
        <grid-item label="机型">
          <grid-container :column="1" :label-width="'0px'" :gap="[4, 0]">
            <template v-if="deviceTypeList?.length">
              <grid-item class="device-item" v-for="(item, index) in deviceTypeList" :key="index">
                {{ item.device_type }}
                <span class="extra-text">({{ item.device_family }}, {{ item.cpu_core }}核{{ item.memory }}GB)</span>
              </grid-item>
            </template>
            <span v-else>--</span>
          </grid-container>
        </grid-item>
        <grid-item label="资源分布方式">{{ RES_ASSIGN_TYPE[data.resAssignType]?.label ?? '--' }}</grid-item>
        <grid-item label="计费模式" v-if="data.chargeType">
          {{ cvmChargeTypeNames[data.chargeType] }}
        </grid-item>
      </template>
      <template v-else>
        <grid-item label="可用区">
          <div :class="['diff-content', { 'has-diff': !isEqual(originalData.zones, data.zones) }]">
            <div class="original">
              <span v-if="originalData.zones?.[0] === 'all'">全部可用区</span>
              <span v-else>{{ originalData.zones?.map((zone) => getZoneCn(zone))?.join('，') || '--' }}</span>
            </div>
            <div class="update">
              <arrows-right class="right-icon" />
              <span v-if="data.zones?.[0] === 'all'">全部可用区</span>
              <span v-else>{{ data.zones?.map((zone) => getZoneCn(zone))?.join('，') || '--' }}</span>
            </div>
          </div>
        </grid-item>
        <grid-item label="机型">
          <div
            :class="[
              'diff-content',
              'device-type-content',
              { 'has-diff': !isEqual(originalData.deviceTypeList, data.deviceTypeList) },
            ]"
          >
            <div class="original">
              <div v-for="(item, index) in originalData.deviceTypeList" :key="index" class="device-item">
                <span class="device-type">{{ item.device_type }}</span>
                <span class="extra-text">({{ item.device_family }}, {{ item.cpu_core }}核{{ item.memory }}GB)</span>
              </div>
            </div>
            <div class="update">
              <arrows-right class="right-icon" />
              <div v-for="(item, index) in deviceTypeList" :key="index" class="device-item">
                <span class="device-type">{{ item.device_type }}</span>
                <span class="extra-text">({{ item.device_family }}, {{ item.cpu_core }}核{{ item.memory }}GB)</span>
              </div>
            </div>
          </div>
        </grid-item>
        <grid-item label="资源分布方式">
          <div :class="['diff-content', { 'has-diff': !isEqual(originalData.resAssignType, data.resAssignType) }]">
            <div class="original">{{ RES_ASSIGN_TYPE[originalData.resAssignType]?.label ?? '--' }}</div>
            <div class="update">
              <arrows-right class="right-icon" />
              {{ RES_ASSIGN_TYPE[data.resAssignType]?.label ?? '--' }}
            </div>
          </div>
        </grid-item>
      </template>
      <bk-button theme="primary" outline class="edit-button" @click="emit('edit')">编辑</bk-button>
    </grid-container>
  </div>
</template>

<style scoped lang="scss">
.device-type-info {
  position: relative;
  width: 600px;
  padding: 16px;
  background: #f5f7fa;
  border-radius: 2px;

  .edit-button {
    position: absolute;
    top: 16px;
    right: 16px;
  }

  .extra-text {
    font-size: 12px;
    color: #979ba5;
  }

  .device-item {
    :deep(.item-content) {
      align-items: center;
      gap: 4px;
    }
  }
}

.diff-content {
  display: flex;
  flex-direction: column;
  gap: 8px;

  .original {
    word-break: keep-all;
  }

  .update {
    display: none;
    position: relative;
    color: #f59500;
    word-break: keep-all;

    .right-icon {
      font-size: 36px;
      position: absolute;
      top: -8px;
      left: -42px;
    }
  }

  &.has-diff {
    .original {
      text-decoration: line-through;
      color: #979ba5;
    }

    .update {
      display: block;
    }
  }

  &.device-type-content {
    .original {
      display: flex;
      flex-wrap: wrap;
    }

    .device-item {
      display: flex;
      align-items: center;
      gap: 4px;
      white-space: nowrap;

      &:not(:last-child)::after {
        content: '，';
      }
    }

    &.has-diff {
      .original {
        text-decoration: none;

        .device-type {
          text-decoration: line-through;
        }
      }

      .update {
        display: flex;
        flex-wrap: wrap;
      }
    }
  }
}
</style>

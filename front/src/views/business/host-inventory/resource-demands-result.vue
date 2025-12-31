<script lang="ts" setup>
import { ref, watchEffect, computed } from 'vue';
import rollRequest from '@blueking/roll-request';
import CombineRequest from '@blueking/combine-request';
import dayjs from 'dayjs';
import isoWeek from 'dayjs/plugin/isoWeek';
import http from '@/http';

import { RequirementType, type IRequirementObsProject } from '@/store/config/requirement';
import {
  type IListResourcesDemandsItem,
  ResourcesDemandsStatus,
  ResourceDemandResultStatusCode as StatusCode,
} from '@/typings/resourcePlan';

const props = defineProps<{ data: any; obsProjectMap: IRequirementObsProject; bizId: number }>();

const emit = defineEmits<(e: 'update', val: { code: number; text: string }, data: any) => void>();

dayjs.extend(isoWeek);

interface IParams {
  require_type: number;
  region: string;
  core_type: string;
  device_family: string;
}

const list = ref<IListResourcesDemandsItem[]>([]);

const loading = ref(false);

const combineRequest = CombineRequest.setup<IListResourcesDemandsItem[]>(
  Symbol.for('resource-demands-result'),
  async (data) => {
    // 合并相同的请求参数
    const uniqueParams = (data as IParams[]).reduce((acc, cur) => {
      if (
        !acc.some(
          (item) =>
            item.require_type === cur.require_type &&
            item.region === cur.region &&
            item.core_type === cur.core_type &&
            item.device_family === cur.device_family,
        )
      ) {
        acc.push(cur);
      }
      return acc;
    }, []);

    const allReqs = uniqueParams.map(
      (params) =>
        rollRequest({
          httpClient: http,
          pageEnableCountKey: 'count',
        }).rollReqUseCount<IListResourcesDemandsItem>(
          `/api/v1/woa/bizs/${props.bizId}/plans/resources/demands/list`,
          {
            obs_projects: [props.obsProjectMap[params.require_type]],
            core_types: [params.core_type],
            device_families: [params.device_family],
            region_ids: [params.region],
            // 查询当月有效的预测
            expect_time_range: {
              // 当前时间所在月份的第1天往前加1周
              start: dayjs().startOf('month').subtract(1, 'week').startOf('day').format('YYYY-MM-DD'),
              // 当前时间月份的最后1天往后加1周
              end: dayjs().endOf('month').add(1, 'week').endOf('day').format('YYYY-MM-DD'),
            },
            statuses: [ResourcesDemandsStatus.CAN_APPLY],
          },
          { limit: 100, countGetter: (res) => res.data.count, listGetter: (res) => res.data.details },
        ) as Promise<IListResourcesDemandsItem[]>,
    );

    const results = await Promise.all(allReqs);

    return results.reduce((acc, cur) => acc.concat(cur), []);
  },
);

const planStatus = computed(() => {
  const status = {
    code: -1,
    text: 'unknown',
  };

  // 小额绿通
  if (props.data.require_type === RequirementType.GreenChannel) {
    status.code = StatusCode.Default;
    status.text = '默认预测';
  }

  // 滚服项目
  if (props.data.require_type === RequirementType.RollServer) {
    if (props.data.capacity_flag === 0) {
      status.code = StatusCode.BGNone;
      status.text = 'BG无预测';
    } else {
      status.code = StatusCode.BGHas;
      status.text = 'BG有预测';
    }
  }

  if (
    [RequirementType.Regular, RequirementType.Spring, RequirementType.Dissolve, RequirementType.ShortRental].includes(
      props.data.require_type,
    )
  ) {
    const demands = list.value.filter(
      (item) =>
        item.obs_project === props.obsProjectMap[props.data.require_type] &&
        item.device_family === props.data.device_family &&
        item.core_type === props.data.core_type &&
        item.region_id === props.data.region,
    );

    if (demands.length > 0) {
      status.code = StatusCode.BIZHas;
      status.text = '本业务有预测';
    } else {
      status.code = StatusCode.BIZNone;
      status.text = '本业务无预测';
    }
  }

  return status;
});

watchEffect(async () => {
  if (!Object.keys(props.obsProjectMap).length) {
    return;
  }

  // 常规、春节保障、机房裁撤、短期
  if (
    props.data &&
    [RequirementType.Regular, RequirementType.Spring, RequirementType.Dissolve, RequirementType.ShortRental].includes(
      props.data.require_type,
    )
  ) {
    combineRequest.add(props.data);
  }

  loading.value = true;

  list.value = await combineRequest.getPromise();

  loading.value = false;

  emit('update', planStatus.value, props.data);
});
</script>

<template>
  <div class="demands-result">
    <bk-loading v-show="loading" :loading="loading" theme="primary" mode="spin" size="mini" />
    <div
      :class="[
        'text-tag',
        {
          has: [StatusCode.BGHas, StatusCode.BIZHas].includes(planStatus.code),
          none: [StatusCode.BGNone, StatusCode.BIZNone].includes(planStatus.code),
        },
      ]"
      v-show="!loading"
    >
      {{ planStatus.text }}
    </div>
  </div>
</template>

<style lang="scss" scoped>
.demands-result {
  position: relative;

  :deep(.bk-loading-size-mini) {
    transform: scale(0.75);
  }

  .text-tag {
    color: #1768ef;
    font-size: 12px;
    background: #e1ecff;
    border-radius: 11px;
    height: 22px;
    line-height: 22px;
    display: inline-flex;
    padding: 0 8px;

    &.has {
      color: #299e56;
      background: #daf6e5;
    }

    &.none {
      color: #979ba5;
      background: #f0f1f5;
    }
  }
}
</style>

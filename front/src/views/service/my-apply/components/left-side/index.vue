<template>
  <bk-loading :loading="listLoading" :opacity="1">
    <div class="apply-left-wrapper">
      <HeaderSelect :title="title" :filter-data="filterData" :active="active" @on-select="handleSelectChange" />
      <div class="apply-left-wrapper-list" @scroll="handleScroll">
        <template v-if="list.length">
          <apply-item
            v-for="(item, index) in list"
            :key="index"
            :data="item"
            :has-bottom-border="index !== list.length - 1"
            :active="currentActive"
            @on-change="handleChange"
          />
        </template>
        <template v-else>
          <div class="empty-wrapper">
            <img class="empty-img" :src="emptyChart" alt="error" />
            <div class="empty-tip">{{ t('暂无数据') }}</div>
          </div>
        </template>
      </div>
      <div class="loading-more pt10" v-if="!canScrollLoad">
        {{ t('没有更多数据') }}
      </div>
    </div>
  </bk-loading>
</template>

<script lang="ts">
import { defineComponent, reactive, toRefs, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import HeaderSelect from '../header-select/index';
import ApplyItem from '../apply-item/index.vue';
import emptyChart from '@/assets/image/empty-chart.png';
import _ from 'lodash';
export default defineComponent({
  name: 'MyApplySide',
  components: {
    HeaderSelect,
    ApplyItem,
  },
  props: {
    listLoading: {
      type: Boolean,
      default: false,
    },
    canScrollLoad: {
      type: Boolean,
      default: false,
    },
    active: {
      type: [String, Number],
    },
    list: {
      type: Array,
      default: () => [] as any[],
    },
    filterData: {
      type: Array,
      default: () => [] as any[],
    },
  },
  emits: ['on-change', 'on-filter-change', 'on-load'],
  setup(props, { emit }) {
    const { t } = useI18n();

    let state = reactive({
      title: t('申请列表'),
      isLoading: false,
      currentActive: 0,
      isScrollLoading: false,
      isShowNoDataTips: false,
    });

    const handleChange = (id: Record<string, number>) => {
      state = Object.assign(state, { currentActive: id });
      emit('on-change', id);
    };

    const handleSelectChange = (payload: Record<string, number | string>) => {
      handleResetScrollLoading();
      emit('on-filter-change', payload);
    };

    const handleResetScrollLoading = () => {
      state = Object.assign(state, { isShowNoDataTips: false, isScrollLoading: false });
    };

    const handleScroll = (payload: any) => {
      if (!props.canScrollLoad) return;
      _throttle(payload);
    };

    // 节流处理滚动
    const _throttle = _.throttle((payload) => {
      if (payload.target.scrollTop + payload.target.offsetHeight >= payload.target.scrollHeight) {
        state = Object.assign(state, {
          isScrollLoading: true,
          isShowNoDataTips: false,
        });
        emit('on-load');
      }
    }, 500);

    watch(
      () => props.list,
      (payload: any[]) => {
        if (!payload.length) {
          state = Object.assign(state, { currentActive: '', isShowNoDataTips: false });
          return;
        }
        if (!payload.some((item: Record<string, any>) => item.id === state.currentActive)) {
          state.currentActive = payload[0].id;
        }
      },
      {
        immediate: true,
        deep: true,
      },
    );
    return {
      emptyChart,
      ...toRefs(state),
      t,
      handleChange,
      handleSelectChange,
      handleScroll,
      handleResetScrollLoading,
    };
  },
});
</script>

<style lang="scss" scoped>
$borderColor: #f5f6fa;
.apply-left-wrapper {
  display: flex;
  flex-direction: column;
  flex: 0 0 280px;
  height: 100%;
  background: #fff;
  overflow: hidden;
  &-list {
    flex: 1;
    position: relative;
    overflow-x: hidden;
    overflow-y: auto;
  }
  .loading-more {
    text-align: center;
    font-size: 12px;
  }

  .empty-wrapper {
    text-align: center;
    margin-top: 50px;
    .empty-img {
      width: 150px;
    }
  }
}
</style>

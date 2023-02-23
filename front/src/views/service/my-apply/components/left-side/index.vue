<template>
  <div class="apply-left-wrapper">
    <bk-loading
      :loading="isLoading"
    >
      <HeaderSelect
        :title="title"
        :filter-data="filterData"
        :active="active"
        @on-select="handleSelectChange"
      />
      <div
        :class="['apply-left-wrapper-list', { 'set-right-border': isEmpty || isLoading }]"
        @scroll="handleScroll"
      >
        <template v-if="!isEmpty">
          <apply-item
            v-for="(item, index) in list"
            :key="index"
            :data="item"
            :has-bottom-border="index !== list.length - 1"
            :active="currentActive"
            @on-change="handleChange"
          />
          <div class="load-more-wrapper" v-if="isScrollLoading" />
          <div class="no-data-tips" v-show="isShowNoDataTips">
            {{ t("没有更多内容了") }}
          </div>
        </template>
        <template v-else>
          <div class="empty-wrapper">
            <img :src="emptyChart" />
            <div class="empty-tip">{{ t("暂无数据") }}</div>
          </div>
        </template>
      </div>
    </bk-loading>
  </div>
</template>

<script lang="ts">
import { defineComponent, reactive, toRefs, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import HeaderSelect from '../header-select/index';
import ApplyItem from '../apply-item/index.vue';
import emptyChart from '@/assets/image/empty-chart.png';
export default defineComponent({
  name: 'MyApplySide',
  components: {
    HeaderSelect,
    ApplyItem,
  },
  props: {
    isLoading: {
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
      isEmpty: false,
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

    const handleResetScrollLoading = ()  => {
      state = Object.assign(state, { isShowNoDataTips: false, isScrollLoading: false });
    };

    const handleScroll = (payload: any) => {
      console.log(payload, '滚动');
      if (props.isLoading) {
        handleResetScrollLoading();
        return;
      }
      if (props.canScrollLoad) {
        state = Object.assign(state, {
          isScrollLoading: false,
          isShowNoDataTips: true,
        });
        return;
      }
      if (payload.target.scrollTop + payload.target.offsetHeight >= payload.target.scrollHeight) {
        state = Object.assign(state, {
          isScrollLoading: true,
          isShowNoDataTips: false,
        });
        emit('on-load');
      }
    };

    watch(() => props.list, (payload: any[]) => {
      if (!payload.length) {
        state = Object.assign(state, { currentActive: '',  isShowNoDataTips: false });
        return;
      }
      if (!payload.some((item: Record<string, any>) => item.id === state.currentActive)) {
        console.log(payload, '数据');
        state.currentActive = payload[0].id;
      }
    },  {
      immediate: true,
    });
    return {
      emptyChart,
      ...toRefs(state),
      t,
      handleScroll,
      handleChange,
      handleSelectChange,
      handleResetScrollLoading,
    };
  },
});
</script>

<style lang="scss" scoped>
.apply-left-wrapper {
  flex: 0 0 280px;
  height: calc(100vh - 61px);
  background: #fff;
  overflow: hidden;
  &-list {
    position: relative;
    height: calc(100% - 83px);
    overflow-x: hidden;
    overflow-y: auto;
  }
}
</style>

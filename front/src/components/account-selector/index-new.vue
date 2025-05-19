<script lang="ts" setup>
import { computed, reactive, ref, watch, watchEffect } from 'vue';
import { useRoute, type RouteLocationNormalizedLoaded } from 'vue-router';
import { Select, Exception } from 'bkui-vue';
import { AngleLeft, AngleRight } from 'bkui-vue/lib/icon';
import { observerResize } from 'bkui-vue/lib/shared';
import { IAccountItem } from '@/typings';
import { VendorEnum, ResourceTypeEnum, VendorMap } from '@/common/constant';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import dataFactory from './data-factory';
import filterPlugin from './filter.plugin';

export interface IAccountSelectorProps {
  vendor?: VendorEnum;
  // 有业务id获取业务下的账号，否则为资源下的账号
  bizId?: number;
  resourceType?: ResourceTypeEnum;
  disabled?: boolean;
  filter?: (
    list: IAccountItem[],
    options?: {
      route: RouteLocationNormalizedLoaded;
      whereAmI: ReturnType<typeof useWhereAmI>;
      resourceType?: ResourceTypeEnum;
    },
  ) => IAccountItem[];
  placeholder?: string;
  optionDisabled?: (accountItem?: IAccountItem) => boolean;
  popoverMaxHeight?: number;
}

export interface IAccountOption extends IAccountItem {
  visible: boolean;
}

const { Option } = Select;

defineOptions({ name: 'AccountSelector' });

const route = useRoute();
const whereAmI = useWhereAmI();

const props = withDefaults(defineProps<IAccountSelectorProps>(), {
  disabled: false,
  popoverMaxHeight: 360,
  optionDisabled: () => false,
  filter: filterPlugin.accountFilter.bind(filterPlugin),
});

const emit =
  defineEmits<
    (e: 'change', val: IAccountItem, oldVal: IAccountItem, vendorAccountMap: Map<VendorEnum, IAccountOption[]>) => void
  >();

const model = defineModel<string>();

const toolbarHeight = 50;

const { useList, vendorProperty } = dataFactory(props.vendor);

const { list, loading } = useList(props);

const activeVendor = ref<VendorEnum>();
const currentDisplayList = ref<IAccountOption[]>([]);
const searchValue = ref('');
const vendorListRef = ref<HTMLElement>();

const selected = computed({
  get() {
    return model.value || undefined;
  },
  set(val) {
    model.value = val;
  },
});

const selectedData = computed(() => list.value.find((item) => item.id === selected.value));
const selectedVendorProperty = computed(() => vendorProperty[selectedData.value.vendor]);

const vendorAccountMap = ref<Map<VendorEnum, IAccountOption[]>>(new Map());

const optionFilter = (keyword: string, option: Record<string, any>) => {
  return option.name.toLowerCase().includes(keyword.toLowerCase());
};

const getDisplayCount = (vendor?: VendorEnum) => {
  const list = getDisplayList(vendor);
  return list.length;
};

const getDisplayList = (vendor?: VendorEnum) => {
  if (vendor) {
    const accountList = vendorAccountMap.value.get(vendor);
    return accountList.filter((item) => item.visible);
  }

  // 未指定 vendor 则为全部
  const allList = [];
  for (const [, accountList] of vendorAccountMap.value) {
    allList.push(...accountList.filter((item) => item.visible));
  }
  return allList;
};

const handleSearchChange = (keyword: string) => {
  searchValue.value = keyword;
  for (const [, accountList] of vendorAccountMap.value) {
    accountList.forEach((item) => {
      item.visible = optionFilter(keyword, item);
    });
  }
};

const handleSelectVendor = (vendor?: VendorEnum) => {
  activeVendor.value = vendor;
};

watch(
  list,
  (newList) => {
    let filteredList = newList;
    if (props.filter) {
      filteredList = props.filter?.(newList, { route, whereAmI, resourceType: props.resourceType });
    }
    vendorAccountMap.value.clear();
    filteredList.forEach((item) => {
      const newItem = { ...item, visible: true };
      if (vendorAccountMap.value.has(item.vendor)) {
        vendorAccountMap.value.get(item.vendor)?.push(newItem);
      } else {
        vendorAccountMap.value.set(item.vendor, [newItem]);
      }
    });
  },
  { deep: true },
);

watch([model, list], ([newVal, newList], [oldVal]) => {
  const account = newList.find((item) => item.id === newVal);
  const oldAccount = list.value.find((item) => item.id === oldVal);
  emit('change', account, oldAccount, vendorAccountMap.value);
});

watchEffect(() => {
  currentDisplayList.value = getDisplayList(activeVendor.value);
});

let vendorObInstance: { start: any; stop: any } = null;
const SCROLL_STEP = 110;
const vendorScrollState = reactive({
  visible: false,
  leftEnabled: false,
  rightEnabled: false,
});

const vendorListRefResize = () => {
  const { clientWidth, scrollWidth, scrollLeft } = vendorListRef.value;
  vendorScrollState.visible = clientWidth < scrollWidth;
  updateScrollEnabled(scrollLeft);
};
const updateScrollEnabled = (scrollLeft: number) => {
  const { clientWidth, scrollWidth } = vendorListRef.value;
  vendorScrollState.leftEnabled = scrollLeft > 0;
  vendorScrollState.rightEnabled = scrollLeft < scrollWidth - clientWidth;
};

const handleToggle = (isPopoverShow: boolean) => {
  if (isPopoverShow) {
    // 为 vendorList 绑定 resize 处理，组件默认会渲染 popover 的内容，为了不依赖此行为还是将处理放在展示时
    setTimeout(() => {
      if (!vendorListRef.value) return;
      vendorObInstance = observerResize(vendorListRef.value, vendorListRefResize, 60, true);
      vendorObInstance.start();
    });
  } else {
    vendorObInstance?.stop();
    vendorObInstance = null;

    // 回到“全部”
    activeVendor.value = undefined;
  }
};
const handleVendorScrollLeft = () => {
  if (vendorScrollState.leftEnabled) {
    const { scrollLeft } = vendorListRef.value;
    const nextLeft = scrollLeft - SCROLL_STEP;
    const scrollStep = nextLeft < SCROLL_STEP ? nextLeft + SCROLL_STEP : SCROLL_STEP;
    const finalLeft = Math.max(scrollLeft - scrollStep, 0);
    vendorListRef.value.scrollLeft = finalLeft;

    // 启用 scroll-behavior: smooth 时不能即时获取到 scrollLeft 的值，在这里传入
    updateScrollEnabled(finalLeft);
  }
};

const handleVendorScrollRight = () => {
  if (vendorScrollState.rightEnabled) {
    const { clientWidth, scrollWidth, scrollLeft } = vendorListRef.value;
    const nextLeft = scrollWidth - clientWidth - scrollLeft - SCROLL_STEP;
    const scrollStep = nextLeft < SCROLL_STEP ? nextLeft + SCROLL_STEP : SCROLL_STEP;
    const finalLeft = Math.min(scrollLeft + scrollStep, scrollWidth - clientWidth);
    vendorListRef.value.scrollLeft = finalLeft;
    updateScrollEnabled(finalLeft);
  }
};

defineExpose({ currentDisplayList });
</script>

<template>
  <Select
    class="account-selector"
    v-model="selected"
    :multiple-mode="'tag'"
    :popover-options="{ extCls: 'account-selector-popover' }"
    :disabled="disabled"
    :loading="loading"
    :clearable="false"
    :scroll-height="popoverMaxHeight"
    :custom-content="true"
    :filter-option="optionFilter"
    :placeholder="placeholder"
    @search-change="handleSearchChange"
    @toggle="handleToggle"
  >
    <template #tag>
      <div class="selected-title" v-if="selectedData">
        <div class="selected-name">{{ selectedData.name }}</div>
        <div class="selected-vendor" :style="selectedVendorProperty?.style">
          <img class="vendor-icon" :src="selectedVendorProperty?.icon" />
          {{ VendorMap[selectedData.vendor] }}
        </div>
      </div>
    </template>
    <div class="account-container">
      <template v-if="vendorAccountMap.size">
        <div
          :class="[
            'toolbar',
            {
              scrollable: vendorScrollState.visible,
              'left-enabled': vendorScrollState.leftEnabled,
              'right-enabled': vendorScrollState.rightEnabled,
            },
          ]"
          :style="{ height: `${toolbarHeight}px` }"
        >
          <div :class="['action-button', { disabled: !vendorScrollState.leftEnabled }]" @click="handleVendorScrollLeft">
            <AngleLeft class="action-icon" />
          </div>
          <div class="vendor-list" ref="vendorListRef">
            <div :class="['vendor-item', 'filter-item', { active: !activeVendor }]" @click="handleSelectVendor()">
              全部
              <em class="account-count">{{ getDisplayCount() }}</em>
            </div>
            <div
              :class="['vendor-item', 'filter-item', { active: activeVendor === vendorId }]"
              v-for="vendorId in vendorAccountMap.keys()"
              :key="vendorId"
              @click="handleSelectVendor(vendorId)"
            >
              <img class="vendor-icon" :src="vendorProperty[vendorId]?.icon" />
              {{ VendorMap[vendorId] }}
              <em class="account-count">{{ getDisplayCount(vendorId) }}</em>
            </div>
          </div>
          <div
            :class="['action-button', { disabled: !vendorScrollState.rightEnabled }]"
            @click="handleVendorScrollRight"
          >
            <AngleRight class="action-icon" />
          </div>
        </div>
        <div class="account-list g-scroller" :style="{ maxHeight: `${popoverMaxHeight - toolbarHeight - 8}px` }">
          <Option
            v-for="account in currentDisplayList"
            :key="account.id"
            :id="account.id"
            :name="account.name"
            :disabled="optionDisabled(account)"
          >
            <div class="account-option">
              <span>{{ account.name }}</span>
              <div class="option-vendor">
                <img class="vendor-icon" :src="vendorProperty[account.vendor]?.icon" />
              </div>
            </div>
          </Option>
        </div>
        <div class="account-list-empty" v-if="!currentDisplayList.length">
          <Exception v-if="searchValue" description="搜索为空" scene="part" type="search-empty" />
          <Exception v-else description="没有数据" scene="part" type="empty" />
        </div>
      </template>
      <div class="data-empty" v-else>
        <Exception v-if="searchValue" description="搜索为空" scene="part" type="search-empty" />
        <Exception v-else description="没有数据" scene="part" type="empty" />
      </div>
    </div>
  </Select>
</template>

<style lang="scss" scoped>
.account-selector {
  font-size: 12px;

  .selected-title {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;

    .selected-vendor {
      display: flex;
      align-items: center;
      gap: 2px;
      border-radius: 12px;
      font-size: 12px;
      padding: 2px 4px;

      .vendor-icon {
        width: 16px;
      }
    }
  }

  :deep(.bk-select-tag-wrapper) {
    width: 100%;
  }
}
</style>
<style lang="scss">
.account-selector-popover {
  .account-container {
    .toolbar {
      display: flex;
      align-items: center;
      padding: 13px 4px;
      box-shadow: 0 4px 6px 0 #0000001a;
      position: relative;

      .vendor-list {
        display: flex;
        align-items: center;
        gap: 8px;
        overflow: hidden;
        scroll-behavior: smooth;

        &::before,
        &::after {
          display: none;
          position: absolute;
          width: 42px;
          height: 100%;
          content: '';
          background: #fff;
        }

        &::before {
          left: 22px;
          mask-image: linear-gradient(90deg, #000 0%, transparent);
        }

        &::after {
          right: 22px;
          mask-image: linear-gradient(-90deg, #000 0%, transparent);
        }
      }

      .vendor-item {
        display: flex;
        align-items: center;
        gap: 4px;
        padding: 0 8px;
        color: #63656e;
        background: #f0f1f5;
        border: 1px solid transparent;
        border-radius: 2px;
        white-space: nowrap;
        height: 24px;
        cursor: pointer;

        .account-count {
          font-style: normal;
          font-size: 10px;
          color: #979ba5;
          background: #fff;
          border-radius: 2px;
          padding: 0 4px;
        }

        .vendor-icon {
          width: 16px;
        }

        &.active {
          color: #3a8aff;
          background: #e1ecff;
          border: 1px solid #a3c5fd;

          .account-count {
            color: #3a8aff;
            background: #f0f5ff;
          }
        }
      }

      .action-button {
        display: none;
        align-items: center;
        justify-content: center;
        height: 24px;
        color: #63656e;
        cursor: pointer;

        &.disabled {
          color: #dcdee5;
          cursor: default;
        }

        .action-icon {
          font-size: 20px;
        }
      }

      &.scrollable {
        .action-button {
          display: flex;
        }

        &.left-enabled {
          .vendor-list {
            &::before {
              display: block;
            }
          }
        }

        &.right-enabled {
          .vendor-list {
            &::after {
              display: block;
            }
          }
        }
      }
    }

    .account-list {
      padding-top: 8px;
    }

    .account-option {
      display: flex;
      align-items: center;
      justify-content: space-between;
      width: 100%;

      .option-vendor {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 22px;
        height: 22px;
        border-radius: 50%;
        background: #fff;

        .vendor-icon {
          width: 16px;
        }
      }
    }

    .data-empty,
    .account-list-empty {
      padding-bottom: 32px;
    }
  }
}
</style>

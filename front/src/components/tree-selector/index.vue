<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue';
import debounce from 'lodash/debounce';

export interface ITreeItem {
  id: number | string;
  name: string;
  full_name?: string;
  has_children?: 0 | 1 | boolean;
  children?: ITreeItem[];
  level?: number;
  tof_dept_id?: number;
}

export interface ITreeSelectorProps {
  data: ITreeItem[] | (() => Promise<ITreeItem[]>);
  multiple?: boolean;
  disabled?: boolean;
  showOnInit?: boolean;
  dropdownMaxHeight?: number;
  collapseTags?: boolean;
  placeholder?: string;
  searchPlaceholder?: string;
}

defineOptions({
  name: 'tree-selector',
});

const model = defineModel<string | number | (string | number)[]>();

const modelChecked = defineModel<ITreeItem | ITreeItem[]>('checked');

const props = withDefaults(defineProps<ITreeSelectorProps>(), {
  data: () => [] as ITreeItem[],
  multiple: true,
  disabled: false,
  dropdownMaxHeight: 320,
  collapseTags: true,
  showOnInit: false,
});

const selectRef = ref();
const treeRef = ref();

const loading = ref(false);

const checked = ref<ITreeItem[]>([]);

// 全部选择的项，包括半选，仅用于清空树勾选状态
const allTreeChecked = ref<ITreeItem[]>([]);

const treeData = ref<ITreeItem[]>([]);

const localMultiple = computed(() => {
  if (Array.isArray(model.value) && model.value.length > 1 && !props.multiple) {
    return true;
  }
  return props.multiple;
});

const selectAction = computed(() => (localMultiple.value ? 'setChecked' : 'setSelect'));

const selectValue = computed(() => {
  const rootNode = treeData.value?.[0];
  if (
    treeRef.value?.isRootNode(rootNode) &&
    treeRef.value?.isNodeChecked(rootNode) &&
    !treeRef.value?.getNodeAttr(rootNode, '__is_indeterminate')
  ) {
    return ['全部'];
  }

  if (checked.value?.length > 10) {
    return checked.value
      .slice(0, 10)
      .map((item) => item.full_name)
      .concat(`...${checked.value.length}个`);
  }
  return checked.value.map((item) => item.full_name);
});

// 默认选中的值
const defaultChecked = computed(() => {
  if (localMultiple.value) {
    if (model.value && !Array.isArray(model.value)) {
      return [model.value];
    }
    return model.value || [];
  }
  if (Array.isArray(model.value)) {
    return model.value[0] || '';
  }
  return model.value || '';
});

const getTreeInternalData = () => treeRef.value.getData();

const loadTree = async (defaultCheckedIds: ITreeItem['id'] | ITreeItem['id'][]) => {
  // 默认先拉出第一层级
  loading.value = true;
  const topData = typeof props.data === 'function' ? await props.data() : props.data;
  loading.value = false;

  treeData.value = topData;

  // 处理默认选中
  if (
    (!Array.isArray(defaultCheckedIds) && defaultCheckedIds) ||
    (Array.isArray(defaultCheckedIds) && defaultCheckedIds.length)
  ) {
    const values = Array.isArray(defaultCheckedIds) ? defaultCheckedIds : [defaultCheckedIds];
    nextTick(() => {
      const treeFlatData = getTreeInternalData();
      values.forEach((value) => {
        const checkedNode = treeFlatData.data.find((item: ITreeItem) => item.id === value);
        if (checkedNode) {
          if (!checked.value.some((item) => item.id === checkedNode.id)) {
            checked.value.push(checkedNode);
          }
          // 展开选中的节点及其父节点
          treeRef.value.setOpen(checkedNode, true, true);
          // 设置为选中
          treeRef.value[selectAction.value](checkedNode, true, false);
        }
      });
    });
  }

  nextTick(() => {
    // 默认展开第一层级
    if (treeData.value.length) {
      treeRef.value.setOpen(treeData.value[0], true, false);
    }
  });
};

watch(
  checked,
  (list) => {
    const modelValue = list.map((item) => item.id);
    model.value = localMultiple.value ? modelValue : modelValue?.[0];
    modelChecked.value = localMultiple.value ? list : list?.[0];
  },
  { deep: true },
);

const clear = () => {
  allTreeChecked.value.forEach((node) => {
    treeRef.value[selectAction.value](node, false, false);
  });
  checked.value = [];
};

const handleNodeChecked = (allChecked: ITreeItem[], allHalfChecked: ITreeItem[]) => {
  allTreeChecked.value = allChecked;
  checked.value = allChecked.filter((item) => !allHalfChecked.includes(item));
};

const handleNodeSelected = (payload: { selected: boolean; node: ITreeItem }) => {
  checked.value = [payload.node];
  selectRef.value.hidePopover();
};

const handleSelectRemoveTag = (name: ITreeItem['full_name']) => {
  const node = checked.value.find((node) => node.full_name === name);
  if (node) {
    const index = checked.value.indexOf(node);
    if (index !== -1) {
      checked.value.splice(index, 1);
    }

    // 使用tree组件的方法时操作对象需要是treeNode，如果不是在搜索的场景可能导致操作无效
    const treeFlatData = getTreeInternalData();
    const treeNode = treeFlatData.data.find((item: ITreeItem) => item.id === node.id);

    // 执行selectAction方法并指定不触发相对应的事件
    treeRef.value[selectAction.value](treeNode, false, false);
  }
};

const handleClear = () => {
  clear();
};

const searchValue = ref('');

const handleSearch = debounce(async (value: string) => {
  searchValue.value = value;

  // 展开搜索结果
  nextTick(() => {
    const treeFlatData = getTreeInternalData();
    const matched = treeFlatData.data.filter((item: ITreeItem) => treeRef.value.isNodeMatched(item));
    treeRef.value.setOpen(matched.pop(), true, true);
  });
}, 200);

loadTree(defaultChecked.value);

defineExpose({
  clear,
});
</script>

<template>
  <div class="tree-selector">
    <bk-select
      ref="selectRef"
      :collapse-tags="false"
      :disabled="disabled"
      :loading="loading"
      :model-value="selectValue"
      :multiple="localMultiple"
      :placeholder="placeholder"
      :scroll-height="dropdownMaxHeight"
      :search-placeholder="searchPlaceholder"
      :show-on-init="showOnInit"
      display-key="name"
      multiple-mode="'default'"
      custom-content
      @clear="handleClear"
      @search-change="handleSearch"
      @tag-remove="handleSelectRemoveTag"
    >
      <bk-tree
        ref="treeRef"
        class="tree-selector-tree"
        :check-strictly="true"
        :data="treeData"
        :expand-all="false"
        :level-line="false"
        :node-content-action="localMultiple ? ['click'] : ['selected']"
        :selectable="!localMultiple"
        :show-checkbox="localMultiple"
        :show-node-type-icon="false"
        :virtual-render="false"
        :search="searchValue"
        children="children"
        label="name"
        node-key="id"
        @node-checked="handleNodeChecked"
        @node-selected="handleNodeSelected"
      />
    </bk-select>
  </div>
</template>

<style lang="scss">
.tree-selector {
  font-size: 12px;
}

.tree-selector-tree {
  color: #63656e;

  .bk-node-action {
    display: inline-flex;
    align-items: center;
  }

  .bk-node-content {
    gap: 4px;

    & > span {
      display: flex;
      align-items: center;
    }
  }
}
</style>

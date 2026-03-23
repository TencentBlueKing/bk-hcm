<script setup lang="ts">
import { watch, toRef, h } from 'vue';
import { useExcelPreview, type IExcelImportData, type ITableColumn, type ITableRow } from '../hooks/use-excel-preview';

const model = defineModel<boolean>({ default: false });

const props = defineProps<{
  data: IExcelImportData | null;
}>();

const { activeTab, tabs, currentColumns, currentRows, initActiveTab } = useExcelPreview(toRef(props, 'data'));

watch(model, (val) => {
  if (val) {
    initActiveTab();
  }
});

/** 判断当前 sheet 是否有任何错误行 */
const hasAnyError = () => currentRows.value.some((row) => row._hasError);

/** 行样式：错误行添加浅红色背景 - bk-table 的 row-class 回调直接接收 row 对象 */
const getRowClass = (row: ITableRow) => {
  return row?._hasError ? 'error-row' : '';
};

/** 生成错误原因列的 tooltips 配置 */
const getErrorTooltips = (row: ITableRow) => {
  if (!row._hasError || !row._errorReasons?.length) {
    return { disabled: true };
  }
  return {
    content: h('div', { style: { maxWidth: '400px' } }, row._errorReasons.join('；')),
  };
};
</script>

<template>
  <bk-dialog
    v-model:is-show="model"
    title="需求数据预览"
    dialog-type="show"
    width="90%"
    :quick-close="false"
    render-directive="if"
  >
    <div class="excel-preview">
      <bk-tab v-model:active="activeTab" type="unborder-card">
        <bk-tab-panel v-for="tab in tabs" :key="tab.name" :name="tab.name" render-directive="if">
          <template #label>
            <div class="tab-label" :class="{ 'has-error': tab.hasError }">
              <i v-if="tab.hasError" class="hcm-icon bkhcm-icon-warn-triangle tab-error-icon"></i>
              <span>{{ tab.name }}</span>
              <span class="tab-count">{{ tab.totalCount }}</span>
            </div>
          </template>

          <bk-table
            :data="currentRows"
            :max-height="500"
            row-hover="auto"
            show-overflow-tooltip
            :row-class="getRowClass"
            :border="['row']"
          >
            <!-- 错误原因列：仅当该 sheet 有错误时显示 -->
            <bk-table-column
              v-if="hasAnyError()"
              label="错误原因"
              :width="160"
              fixed="left"
              :show-overflow-tooltip="false"
            >
              <template #default="{ row }: { row: ITableRow }">
                <div v-if="row._hasError" v-bk-tooltips="getErrorTooltips(row)" class="error-cell">
                  <i class="hcm-icon bkhcm-icon-warn-triangle tab-error-icon error-icon"></i>
                  <span class="error-text">{{ row._errorReasons.join('；') }}</span>
                </div>
              </template>
            </bk-table-column>

            <!-- 错误行数列：仅当该 sheet 有错误时显示 -->
            <bk-table-column
              v-if="hasAnyError()"
              label="错误行数"
              :width="80"
              fixed="left"
              :show-overflow-tooltip="false"
            >
              <template #default="{ row }: { row: ITableRow }">
                <span v-if="row._hasError" class="error-row-index">{{ row._errorRowIndex }}</span>
              </template>
            </bk-table-column>

            <!-- 动态数据列：由 fixed_headers + headers 拼接 -->
            <bk-table-column
              v-for="(col, index) in (currentColumns as ITableColumn[])"
              :key="index"
              :prop="col.field"
              :label="col.label"
              :min-width="120"
              show-overflow-tooltip
            >
              <template #default="{ row }: { row: ITableRow }">
                {{ row[col.field] ?? '' }}
              </template>
            </bk-table-column>
          </bk-table>
        </bk-tab-panel>
      </bk-tab>
    </div>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.excel-preview {
  min-height: 300px;
}

.tab-label {
  display: inline-flex;
  align-items: center;
  gap: 4px;

  &.has-error {
    color: #ea3636;
  }

  .tab-error-icon {
    font-size: 14px;
    color: #ea3636;
  }

  .tab-count {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 18px;
    height: 18px;
    padding: 0 4px;
    font-size: 12px;
    line-height: 18px;
    color: #979ba5;
    background: #f0f1f5;
    border-radius: 9px;
  }
}

.error-cell {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #ea3636;

  .error-icon {
    flex-shrink: 0;
    font-size: 14px;
  }

  .error-text {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.error-row-index {
  color: #ea3636;
}
</style>

<style lang="scss">
/* stylelint-disable selector-class-pattern */
.excel-preview .bk-table .error-row td {
  background-color: #fff3f3 !important;
}
/* stylelint-enable selector-class-pattern */
</style>

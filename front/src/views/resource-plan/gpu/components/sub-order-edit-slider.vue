<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue';
import { Message } from 'bkui-vue';
import {
  useGpuDemandStore,
  type IGpuDemandSubOrder,
  type ITplHeader,
  type ITplSheet,
} from '@/store/resource-plan/gpu-demand';
import { evaluateFormula } from '../hooks/use-excel-preview';
import { JSONSchemaValidator, type FieldSchema } from '@/utils/vue-jsonschema-validator';

const isShow = defineModel<boolean>({ default: false });

// ==================== Props / Emits ====================
const props = defineProps<{
  /** 当前要编辑的子单最新数据 */
  subOrder: IGpuDemandSubOrder | null;
  /** 原始基准数据（首次加载时的快照），用于"原数据"对比展示。为 null 时退化为 subOrder */
  originalSubOrder: IGpuDemandSubOrder | null;
  /** 该子单所属 sheet 的模板配置 */
  sheet: ITplSheet | null;
  /** 业务 ID */
  bizId: number;
}>();

const emit = defineEmits<{
  /** 编辑成功，携带子单 id 和修改差异 */
  success: [payload: { subOrderId: string; diffs: Record<string, { oldVal: any; newVal: any }> }];
}>();

// ==================== Store ====================
const gpuDemandStore = useGpuDemandStore();

// ==================== 表单 Ref ====================
const formRef = ref();

// ==================== 表单字段定义 ====================
interface IFormField {
  /** 唯一标识 */
  key: string;
  /** 表单标签 */
  label: string;
  /** 来源：fixed_headers 的 db_field 字段，或 headers 的 extension 索引 */
  source: 'fixed' | 'extension';
  /** fixed 来源时对应的 db_field */
  dbField?: string;
  /** extension 来源时对应的索引 */
  extIndex?: number;
  /** 输入框类型 */
  type: ITplHeader['type'];
  /** enum 类型时的选项列表 */
  options?: (string | number)[];
  /** 是否必填 */
  required: boolean;
  /** 是否只读 */
  readonly: boolean;
  /** 是否为公式计算列 */
  formula?: string;
  /** excel 列号 */
  excelField?: string;
  /** 是否隐藏 */
  hidden: boolean;
  /** 大于 (exclusive minimum) */
  gt?: number;
  /** 大于等于 (inclusive minimum) */
  gte?: number;
  /** 小于 (exclusive maximum) */
  lt?: number;
  /** 小于等于 (inclusive maximum) */
  lte?: number;
}

/** 从 sheet 配置解析出表单字段列表 */
const formFields = computed<IFormField[]>(() => {
  if (!props.sheet) return [];
  const fields: IFormField[] = [];
  let extIdx = 0;

  // fixed_headers
  for (const h of props.sheet.fixed_headers) {
    if (h.hidden) continue;
    fields.push({
      key: h.db_field || `fixed_${h.field}`,
      label: h.name,
      source: 'fixed',
      dbField: h.db_field,
      type: h.type as ITplHeader['type'],
      options: h.value,
      required: !!h.required,
      readonly: !!h.readonly,
      formula: h.formula,
      excelField: h.field,
      hidden: false,
      gt: h.gt,
      gte: h.gte,
      lt: h.lt,
      lte: h.lte,
    });
  }

  // headers
  for (const h of props.sheet.headers) {
    const hasField = h.field && h.field !== '-';
    if (!h.hidden) {
      fields.push({
        key: `ext_${extIdx}`,
        label: h.name,
        source: 'extension',
        extIndex: hasField ? extIdx : undefined,
        type: h.type as ITplHeader['type'],
        options: h.value,
        required: !!h.required,
        readonly: !!h.readonly,
        formula: h.formula,
        excelField: h.field,
        hidden: false,
        gt: h.gt,
        gte: h.gte,
        lt: h.lt,
        lte: h.lte,
      });
    }
    if (hasField) extIdx += 1;
  }

  return fields;
});

/** 可见的表单字段（非 hidden） */
const visibleFields = computed(() => formFields.value.filter((f) => !f.hidden));

/**
 * 获取数字输入框的 min 值
 * 优先级：gte > gt > 默认 0（不允许负值）
 */
const getFieldMin = (field: IFormField): number | undefined => {
  if (field.gte !== undefined) return field.gte;
  if (field.gt !== undefined) return field.gt;
  return 0; // 默认不允许负值
};

/**
 * 获取数字输入框的 max 值
 * 优先级：lte > lt > undefined（不限制上限）
 */
const getFieldMax = (field: IFormField): number | undefined => {
  if (field.lte !== undefined) return field.lte;
  if (field.lt !== undefined) return field.lt;
  return undefined;
};

// ==================== 表单数据 ====================
/** 当前编辑的表单值 */
const formData = ref<Record<string, any>>({});
/** 原始值快照（用于对比高亮） */
const originalData = ref<Record<string, any>>({});

/** 从子单数据初始化表单 */
const initFormData = () => {
  if (!props.subOrder || !props.sheet) return;
  const data: Record<string, any> = {};
  const original: Record<string, any> = {};

  // 原始基准数据：优先使用 originalSubOrder（首次加载快照），不存在时退化为 subOrder
  const baseSubOrder = props.originalSubOrder ?? props.subOrder;

  for (const field of formFields.value) {
    let val: any;
    let origVal: any;

    if (field.source === 'fixed' && field.dbField) {
      val = (props.subOrder as any)[field.dbField];
      origVal = (baseSubOrder as any)[field.dbField];
    } else if (field.source === 'extension' && field.extIndex !== undefined) {
      val = props.subOrder.extension?.[field.extIndex] ?? '';
      origVal = baseSubOrder.extension?.[field.extIndex] ?? '';
    } else if (field.formula) {
      // 公式列：计算值（两者均用当前数据计算，original 用原始基准数据计算）
      val = computeFormulaValue(field);
      origVal = computeFormulaValueFromSubOrder(field, baseSubOrder);
    } else {
      val = '';
      origVal = '';
    }

    // 将值统一转为 string 以方便表单绑定
    data[field.key] = val !== undefined && val !== null ? val : '';
    original[field.key] = origVal !== undefined && origVal !== null ? origVal : '';
  }

  formData.value = data;
  originalData.value = { ...original };
};

// 当 subOrder 变化时重新初始化
watch(
  () => props.subOrder,
  () => {
    if (props.subOrder) {
      nextTick(initFormData);
    }
  },
  { immediate: true },
);

// ==================== Schema 校验器 ====================
/** 将 ITplHeader 转换为 FieldSchema */
const toFieldSchema = (h: ITplHeader): FieldSchema => ({
  name: h.name,
  field: h.field,
  type: h.type || 'string',
  required: h.required,
  readonly: h.readonly,
  formula: h.formula,
  value: h.value,
  gt: h.gt,
  gte: h.gte,
  lt: h.lt,
  lte: h.lte,
});

/** 根据当前 sheet 动态生成校验器实例 */
const schemaValidator = computed<JSONSchemaValidator | null>(() => {
  if (!props.sheet) return null;
  const fixedFields = props.sheet.fixed_headers.filter((h) => !h.hidden).map(toFieldSchema);
  const headerFields = props.sheet.headers.filter((h) => !h.hidden).map(toFieldSchema);
  return JSONSchemaValidator.fromSheet(props.sheet.name, fixedFields, headerFields);
});

/** 根据 formField 的 label 获取对应的 sanitizedFieldName（与 Schema property key 一致） */
const getSanitizedName = (label: string): string => {
  return label
    .replace(/[()]/g, (match: string) => (match === '(' ? '_' : ''))
    .replace(/[/-]/g, '_')
    .replace(/\s+/g, '_')
    .replace(/_+/g, '_')
    .replace(/^_|_$/g, '');
};

/** 为 bk-form 动态生成 rules 对象 */
const formRules = computed<Record<string, any[]>>(() => {
  const rules: Record<string, any[]> = {};

  for (const field of visibleFields.value) {
    // 跳过公式列和只读列
    if (field.formula || field.readonly) continue;

    const fieldRules: any[] = [];

    // 必填校验
    if (field.required) {
      fieldRules.push({
        required: true,
        message: `${field.label} 不能为空`,
        trigger: 'blur',
      });
    }

    // Schema 自定义校验
    const lastError = { msg: `${field.label} 校验不通过` };
    fieldRules.push({
      validator: (value: any) => {
        if (!schemaValidator.value || !props.sheet) return true;
        // 空值交给 required 规则处理
        if (value === '' || value === undefined || value === null) return true;
        // 构建单字段数据，用字段原始名称作为 key
        const rowData: Record<string, any> = { [field.label]: value };
        const result = schemaValidator.value.validateRow(props.sheet.name, rowData, 0);
        // 只检查当前字段相关的错误
        const sanitizedName = getSanitizedName(field.label);
        const fieldErrors = result.errors.filter((e) => e.field === sanitizedName);
        if (fieldErrors.length > 0) {
          lastError.msg = fieldErrors[0].message;
          return false;
        }
        return true;
      },
      message: () => lastError.msg,
      trigger: 'blur',
    });

    rules[field.key] = fieldRules;
  }

  return rules;
});

/** 提交前执行全量 Schema 校验（独立于 bk-form 的逐字段校验） */
const validateAllFields = (): { valid: boolean; errors: string[] } => {
  if (!schemaValidator.value || !props.sheet) return { valid: true, errors: [] };

  // 构建完整的行数据（以字段的原始中文名作为 key）
  const rowData: Record<string, any> = {};
  for (const field of visibleFields.value) {
    if (field.formula) {
      rowData[field.label] = computeFormulaValue(field);
    } else {
      rowData[field.label] = formData.value[field.key];
    }
  }

  const result = schemaValidator.value.validateRow(props.sheet.name, rowData, 0);
  if (!result.valid) {
    // 将 sanitized 字段名映射回中文名
    const errorMessages = result.errors.map((err) => {
      // 尝试反向查找中文名
      const matchedField = visibleFields.value.find((f) => getSanitizedName(f.label) === err.field);
      const fieldName = matchedField ? matchedField.label : err.field;
      return `${fieldName}: ${err.message}`;
    });
    return { valid: false, errors: errorMessages };
  }
  return { valid: true, errors: [] };
};

// ==================== 公式计算 ====================

/**
 * 为公式计算建立 excel 列号 → 取值函数 的完整映射。
 * formFields 仅包含非 hidden 字段，而公式可能引用 hidden 列的数据。
 * 对于 hidden 列，从子单原始数据（props.subOrder）中取值。
 */
const getExcelFieldValue = (excelField: string): number => {
  // 1. 优先从 formFields（已在表单中的字段）取值
  const formField = formFields.value.find((f) => f.excelField === excelField);
  if (formField) {
    // 公式列递归计算
    if (formField.formula) {
      return computeFormulaValue(formField);
    }
    const val = Number(formData.value[formField.key]);
    return Number.isNaN(val) ? 0 : val;
  }

  // 2. formFields 中找不到（可能是 hidden 列），从子单原始数据中获取
  if (!props.sheet || !props.subOrder) return 0;

  // 遍历所有 fixed_headers + headers，找到匹配 excelField 的列
  for (const h of props.sheet.fixed_headers) {
    if (h.field && h.field !== '-') {
      if (h.field === excelField) {
        // 有 db_field 的从子单字段取值
        if (h.db_field) {
          const val = Number((props.subOrder as any)[h.db_field]);
          return Number.isNaN(val) ? 0 : val;
        }
        return 0;
      }
    }
  }

  // extension 部分
  let extIdx = 0;
  for (const h of props.sheet.headers) {
    if (h.field && h.field !== '-') {
      if (h.field === excelField) {
        const val = Number(props.subOrder.extension?.[extIdx]);
        return Number.isNaN(val) ? 0 : val;
      }
      extIdx += 1;
    }
  }

  return 0;
};

/** 根据当前 formData 计算公式值 */
const computeFormulaValue = (field: IFormField): number => {
  if (!field.formula || !props.sheet) return 0;
  return Math.ceil(evaluateFormula(field.formula, getExcelFieldValue));
};

/**
 * 基于指定的子单数据计算公式值（不依赖 formData）。
 * 用于为 originalData 计算原始基准的公式列值。
 */
const computeFormulaValueFromSubOrder = (field: IFormField, sub: IGpuDemandSubOrder): number => {
  if (!field.formula || !props.sheet) return 0;

  const getVal = (excelField: string): number => {
    // 从 formFields 找到对应字段，然后从子单数据取值
    const ff = formFields.value.find((f) => f.excelField === excelField);
    if (ff) {
      if (ff.formula) {
        return computeFormulaValueFromSubOrder(ff, sub);
      }
      let val: any;
      if (ff.source === 'fixed' && ff.dbField) {
        val = (sub as any)[ff.dbField];
      } else if (ff.source === 'extension' && ff.extIndex !== undefined) {
        val = sub.extension?.[ff.extIndex] ?? 0;
      }
      const num = Number(val);
      return Number.isNaN(num) ? 0 : num;
    }

    // hidden 列：从子单原始数据中获取
    if (!props.sheet || !sub) return 0;
    for (const h of props.sheet.fixed_headers) {
      if (h.field && h.field !== '-' && h.field === excelField) {
        if (h.db_field) {
          const v = Number((sub as any)[h.db_field]);
          return Number.isNaN(v) ? 0 : v;
        }
        return 0;
      }
    }
    let extIdx = 0;
    for (const h of props.sheet.headers) {
      if (h.field && h.field !== '-') {
        if (h.field === excelField) {
          const v = Number(sub.extension?.[extIdx]);
          return Number.isNaN(v) ? 0 : v;
        }
        extIdx += 1;
      }
    }
    return 0;
  };

  return Math.ceil(evaluateFormula(field.formula, getVal));
};

/** 获取字段的显示值（公式列从计算得到） */
const getFieldDisplayValue = (field: IFormField): any => {
  if (field.formula) {
    return computeFormulaValue(field);
  }
  return formData.value[field.key];
};

// ==================== 修改对比 ====================
/** 判断某个字段是否被修改 */
const isFieldModified = (field: IFormField): boolean => {
  if (field.formula) {
    // 公式列比较计算值和原始值
    const currentVal = computeFormulaValue(field);
    return String(currentVal) !== String(originalData.value[field.key]);
  }
  return String(formData.value[field.key]) !== String(originalData.value[field.key]);
};

// ==================== 侧边栏标题 ====================
const sliderTitle = computed(() => {
  if (!props.subOrder || !props.sheet) return '编辑需求';
  return '编辑需求';
});

const sliderSubInfo = computed(() => {
  if (!props.subOrder || !props.sheet) return '';
  return props.sheet.name;
});

const subOrderStatus = computed(() => {
  if (!props.subOrder) return '';
  const statusMap: Record<string, string> = {
    INIT: '待评审',
    PENDING: '评审中',
    DONE: '已评审',
    REJECT: '已驳回',
    TERMINATE: '已终止',
  };
  return statusMap[props.subOrder.status] || props.subOrder.status;
});

const subOrderStatusColor = computed((): { bg: string; text: string } => {
  if (!props.subOrder) return { bg: '#f0f1f5', text: '#979ba5' };
  const colorMap: Record<string, { bg: string; text: string }> = {
    INIT: { bg: '#fdeed8', text: '#e38b02' },
    PENDING: { bg: '#e1ecff', text: '#1768ef' },
    DONE: { bg: '#daf6e5', text: '#299e56' },
    REJECT: { bg: '#ffebeb', text: '#e71818' },
    TERMINATE: { bg: '#f0f1f5', text: '#979ba5' },
  };
  return colorMap[props.subOrder.status] || { bg: '#f0f1f5', text: '#979ba5' };
});

// ==================== 提交保存 ====================
const isSaving = ref(false);

const handleSave = async () => {
  if (!props.subOrder || !props.sheet) return;

  // 1. bk-form 逐字段校验
  try {
    await formRef.value?.validate();
  } catch {
    Message({ theme: 'warning', message: '请检查表单中的错误项' });
    return;
  }

  // 2. 全量 Schema 校验
  const { valid, errors } = validateAllFields();
  if (!valid) {
    Message({ theme: 'error', message: errors.join('；') });
    return;
  }

  // 构建新的 extension 数组
  const newExtension: any[] = [];
  // 先构建 extIndex → formField 映射，避免在循环内创建闭包
  const extFieldMap = new Map<number, IFormField>();
  for (const f of formFields.value) {
    if (f.source === 'extension' && f.extIndex !== undefined) {
      extFieldMap.set(f.extIndex, f);
    }
  }
  let extIdx = 0;
  for (const h of props.sheet.headers) {
    const hasField = h.field && h.field !== '-';
    if (hasField) {
      const field = extFieldMap.get(extIdx);
      if (field) {
        let val: any = formData.value[field.key];
        // 根据类型转换
        if (field.type === 'int') {
          val = parseInt(val, 10) || 0;
        } else if (field.type === 'float') {
          val = parseFloat(val) || 0;
        }
        newExtension.push(val);
      } else {
        // 保持原始值
        newExtension.push(props.subOrder.extension?.[extIdx] ?? '');
      }
      extIdx += 1;
    }
  }

  // 构建 fixed_headers 中有 db_field 的字段（如 demand_year, demand_month, gpu_num, qpm_max 等）
  // 直接遍历 sheet.fixed_headers，不跳过 hidden，确保所有固定字段都被提交
  const fixedFields: Record<string, any> = {};
  for (const h of props.sheet.fixed_headers) {
    if (!h.db_field) continue;
    // 优先从表单中取编辑值
    const field = formFields.value.find((f) => f.dbField === h.db_field);
    if (field) {
      let val: any = field.formula ? computeFormulaValue(field) : formData.value[field.key];
      if (h.type === 'int') {
        val = parseInt(val, 10) || 0;
      } else if (h.type === 'float') {
        val = parseFloat(val) || 0;
      }
      fixedFields[h.db_field] = val;
    } else if (h.formula) {
      // hidden 的公式列（如 qpm_max），需要根据当前编辑的表单数据重新计算
      const result = evaluateFormula(h.formula, getExcelFieldValue);
      let val: any = Math.ceil(result);
      if (h.type === 'int') {
        val = parseInt(val, 10) || 0;
      } else if (h.type === 'float') {
        val = parseFloat(val) || 0;
      }
      fixedFields[h.db_field] = val;
    } else {
      // 表单中没有且非公式列（hidden 的普通字段），从子单原始数据中取
      fixedFields[h.db_field] = (props.subOrder as any)[h.db_field] ?? '';
    }
  }

  isSaving.value = true;
  try {
    await gpuDemandStore.batchUpdateSubOrders({
      suborder_data: [
        {
          suborder_id: props.subOrder.id,
          extension: newExtension,
          ...fixedFields,
        },
      ],
    });
    Message({ theme: 'success', message: '保存成功' });

    // 收集修改差异
    const diffs: Record<string, { oldVal: any; newVal: any }> = {};
    for (const field of formFields.value) {
      if (isFieldModified(field)) {
        const newVal = field.formula ? computeFormulaValue(field) : formData.value[field.key];
        diffs[field.key] = {
          oldVal: originalData.value[field.key],
          newVal,
        };
      }
    }

    emit('success', { subOrderId: props.subOrder.id, diffs });
    isShow.value = false;
  } catch {
    Message({ theme: 'error', message: '保存失败' });
  } finally {
    isSaving.value = false;
  }
};

const handleCancel = () => {
  isShow.value = false;
};
</script>

<template>
  <bk-sideslider v-model:is-show="isShow" :title="sliderTitle" :width="640" :quick-close="false" render-directive="if">
    <template #header>
      <div class="slider-header">
        <span class="slider-title">{{ sliderTitle }}</span>
        <span class="slider-sub-info" v-if="sliderSubInfo">{{ sliderSubInfo }}</span>
        <span
          v-if="subOrderStatus"
          class="slider-status-tag"
          :style="{ background: subOrderStatusColor.bg, color: subOrderStatusColor.text }"
        >
          {{ subOrderStatus }}
        </span>
      </div>
    </template>
    <template #default>
      <div class="edit-form-container">
        <bk-form ref="formRef" form-type="vertical" :label-width="200" :model="formData" :rules="formRules">
          <template v-for="field in visibleFields" :key="field.key">
            <bk-form-item :label="field.label" :required="field.required" :property="field.key">
              <!-- 公式计算列：只读展示 -->
              <template v-if="field.formula">
                <bk-input
                  :model-value="String(getFieldDisplayValue(field))"
                  readonly
                  :class="{ 'field-modified': isFieldModified(field) }"
                  :style="isFieldModified(field) ? { '--input-bg': '#fdf4e8' } : {}"
                />
                <div v-if="isFieldModified(field)" class="original-value">原数据：{{ originalData[field.key] }}</div>
              </template>

              <!-- enum 类型：下拉选择器 -->
              <template v-else-if="field.type === 'enum'">
                <bk-select
                  v-model="formData[field.key]"
                  :disabled="field.readonly"
                  :clearable="false"
                  :class="{ 'field-modified': isFieldModified(field) }"
                  :style="isFieldModified(field) ? { '--input-bg': '#fdf4e8' } : {}"
                >
                  <bk-option v-for="opt in field.options" :key="opt" :value="opt" :label="String(opt)" />
                </bk-select>
                <div v-if="isFieldModified(field)" class="original-value">原数据：{{ originalData[field.key] }}</div>
              </template>

              <!-- int / float 类型：数字输入框 -->
              <template v-else-if="field.type === 'int' || field.type === 'float'">
                <bk-input
                  v-model="formData[field.key]"
                  type="number"
                  :readonly="field.readonly"
                  :min="getFieldMin(field)"
                  :max="getFieldMax(field)"
                  :precision="field.type === 'int' ? 0 : 100"
                  :class="{ 'field-modified': isFieldModified(field) }"
                  :style="isFieldModified(field) ? { '--input-bg': '#fdf4e8' } : {}"
                />
                <div v-if="isFieldModified(field)" class="original-value">原数据：{{ originalData[field.key] }}</div>
              </template>

              <!-- string 类型：普通文本输入框 -->
              <template v-else>
                <bk-input
                  v-model="formData[field.key]"
                  :readonly="field.readonly"
                  :class="{ 'field-modified': isFieldModified(field) }"
                  :style="isFieldModified(field) ? { '--input-bg': '#fdf4e8' } : {}"
                />
                <div v-if="isFieldModified(field)" class="original-value">原数据：{{ originalData[field.key] }}</div>
              </template>
            </bk-form-item>
          </template>
        </bk-form>
      </div>
    </template>
    <template #footer>
      <div class="slider-footer">
        <bk-button theme="primary" :loading="isSaving" @click="handleSave">保存</bk-button>
        <bk-button @click="handleCancel">取消</bk-button>
      </div>
    </template>
  </bk-sideslider>
</template>

<style lang="scss" scoped>
.slider-header {
  display: flex;
  align-items: center;
  gap: 8px;

  .slider-title {
    font-size: 16px;
    font-weight: 700;
    color: #313238;
  }

  .slider-sub-info {
    font-size: 12px;
    color: #979ba5;
  }

  .slider-status-tag {
    display: inline-block;
    padding: 0 6px;
    font-size: 12px;
    line-height: 22px;
    border-radius: 2px;
    white-space: nowrap;
  }
}

.edit-form-container {
  padding: 24px 40px 0;
}

// 修改后的输入框背景色
/* stylelint-disable selector-class-pattern */
.field-modified {
  :deep(.bk-input--text),
  :deep(.bk-input--number-input),
  :deep(input),
  :deep(.bk-select-trigger .bk-input--text) {
    background-color: #fdf4e8 !important;
  }
}
/* stylelint-enable selector-class-pattern */

// 原数据提示
.original-value {
  font-size: 12px;
  line-height: 20px;
  color: #979ba5;
  margin-top: 4px;
}

.slider-footer {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 24px;
}
</style>

import { computed, ref, type Ref } from 'vue';
import { Parser as FormulaParser } from 'hot-formula-parser';

/**
 * Excel 导入预览数据的类型定义
 * 可在 GPU 模块的其他地方复用
 */

/** 列定义（fixed_headers / headers 共用） */
export interface IExcelHeader {
  name: string;
  type: 'string' | 'int' | 'float' | 'enum';
  field: string; // excel 列号，如 A、B、C；"-" 表示公式计算列
  db_field?: string;
  value?: (string | number)[];
  hidden?: boolean;
  required?: boolean;
  formula?: string;
  readonly?: boolean;
}

/** Sheet 定义 */
export interface IExcelSheet {
  name: string;
  row_start: number;
  fixed_headers: IExcelHeader[];
  headers: IExcelHeader[];
}

/** 行数据 */
export interface IExcelDetail {
  name: string; // sheet 名称
  raw_data: any[];
  validate_result: string[];
}

/** 接口返回 data 结构 */
export interface IExcelImportData {
  sheets: IExcelSheet[];
  details: IExcelDetail[] | null;
}

/** Tab 面板信息 */
export interface ITabInfo {
  name: string;
  errorCount: number;
  totalCount: number;
  hasError: boolean;
}

/** 表格列定义（转换后） */
export interface ITableColumn {
  field: string; // 用于取数的 key，如 col_0, col_1...
  label: string;
  type: string;
  isFixed: boolean; // 是否来自 fixed_headers
  formula?: string; // excel 公式（如有），运行时动态计算
  excelField: string; // 原始 excel 列号（A/B/C...）或 "-"
}

/** 表格行数据（转换后） */
export interface ITableRow {
  [key: string]: any;
  _hasError: boolean;
  _errorReasons: string[];
  _errorRowIndex: number; // 错误行在 Excel 中的真实行号（仅在有错误时有意义，正确行为 0）
}

/** getVisibleHeaders 的返回值 */
export interface IVisibleHeadersResult {
  columns: ITableColumn[];
  /** excel 列号 → raw_data 索引映射（所有 field 不为 "-" 的列，包括 hidden 列） */
  excelFieldToDataIndex: Record<string, number>;
  /** raw_data 对应的列数（所有 field 不为 "-" 的列，包括 hidden 列） */
  dataColumnCount: number;
}

/**
 * 从 sheet 的 fixed_headers + headers 中，构建可显示的表格列信息
 *
 * 规则：
 * - raw_data 包含所有 field 不为 "-" 的列数据（无论 hidden 与否），按 fixed_headers + headers 顺序排列
 * - 只有 field 为 "-" 的列（公式计算列）不在 raw_data 中
 * - hidden: true 的列虽然在 raw_data 中有数据，但不在前端展示
 * - field 为 "-" 但有 formula 的列也参与展示（值由公式动态计算，不消耗 raw_data 索引）
 */
export function getVisibleHeaders(sheet: IExcelSheet): IVisibleHeadersResult {
  const columns: ITableColumn[] = [];
  // excel 列号 → raw_data 中的索引（所有 field 不为 "-" 的列，包括 hidden 列）
  const excelFieldToDataIndex: Record<string, number> = {};
  let colIndex = 0; // 展示列计数
  let dataIndex = 0; // raw_data 索引计数

  const allHeaders = [...sheet.fixed_headers, ...sheet.headers];

  for (const header of allHeaders) {
    const isFormulaCol = header.field === '-' && !!header.formula;
    const hasExcelField = header.field !== '-';
    const isHidden = !!header.hidden;

    // 所有 field 不为 "-" 的列都在 raw_data 中有数据（包括 hidden 列）
    if (hasExcelField) {
      excelFieldToDataIndex[header.field] = dataIndex;
      dataIndex += 1;
    }

    // 只有不 hidden 的列才在前端展示（包括 formula 列）
    if (!isHidden) {
      const isFixed = sheet.fixed_headers.includes(header);
      columns.push({
        field: `col_${colIndex}`,
        label: header.name,
        type: header.type,
        isFixed,
        formula: isFormulaCol ? header.formula : undefined,
        excelField: header.field,
      });
      colIndex += 1;
    }
  }

  return { columns, excelFieldToDataIndex, dataColumnCount: dataIndex };
}

// =============================================
// Excel 公式计算引擎（基于 hot-formula-parser）
// =============================================

/**
 * 创建一个可复用的公式计算函数
 *
 * hot-formula-parser 支持 400+ Excel 函数（ROUNDUP、MAX、SUM、IF、AVERAGE 等），
 * 通过 callCellValue 事件将公式中的单元格引用（如 N4、R12）映射到实际数据。
 *
 * 使用工厂模式复用同一个 Parser 实例，避免每行数据重复创建的开销。
 * 每次调用前通过闭包更新 getValueByExcelField，使其始终指向当前行的数据。
 */
function createFormulaEvaluator() {
  const parser = new FormulaParser();
  // 通过闭包引用当前行的取值函数，在每次 evaluate 调用前更新
  let currentGetValue: (excelField: string) => number = () => 0;

  // 监听单元格引用事件，将 excel 列号映射到实际数值
  // 例如公式中的 N4 → callCellValue 事件触发，column.label = "N"
  parser.on('callCellValue', (cellCoord: any, done: (value: number) => void) => {
    const colLabel: string = cellCoord.column.label; // 大写列字母，如 "A"、"N"、"R"
    done(currentGetValue(colLabel));
  });

  /**
   * 解析并计算 Excel 公式
   *
   * @param formula - 原始 Excel 公式字符串（如 "ROUNDUP(N4*1000000/R4/3600/S4,0)"）
   * @param getValueByExcelField - 根据 excel 列号（如 "N"、"R"）获取该行对应值的函数
   * @returns 计算结果（数字）
   */
  return function evaluateFormula(formula: string, getValueByExcelField: (excelField: string) => number): number {
    currentGetValue = getValueByExcelField;
    try {
      const { error, result } = parser.parse(formula);
      if (error) return 0;
      const num = Number(result);
      return typeof num === 'number' && !Number.isNaN(num) ? num : 0;
    } catch {
      return 0;
    }
  };
}

/** 模块级别的公式计算函数（单例，复用 Parser 实例） */
export const evaluateFormula = createFormulaEvaluator();

/**
 * 将 details 中的平铺 raw_data 转换为表格行数据
 *
 * @param details - 接口返回的 details 数组
 * @param sheetName - 当前 sheet 名称
 * @param columns - 当前 sheet 的可见列定义（含 formula 列）
 * @param excelFieldToDataIndex - excel 列号 → raw_data 索引映射
 * @param rowStart - 该 sheet 数据在 Excel 中的起始行号（如 2 表示第一行数据对应 Excel 第 2 行）
 */
export function buildTableRows(
  details: IExcelDetail[],
  sheetName: string,
  columns: ITableColumn[],
  excelFieldToDataIndex: Record<string, number>,
  rowStart: number,
): ITableRow[] {
  const rows: ITableRow[] = [];

  const sheetDetails = details.filter((d) => d.name === sheetName);

  for (let i = 0; i < sheetDetails.length; i++) {
    const detail = sheetDetails[i];
    const hasError = detail.validate_result.length > 0;
    const row: ITableRow = {
      _hasError: hasError,
      _errorReasons: detail.validate_result,
      // Excel 真实行号 = rowStart + 当前行偏移；仅错误行赋值，正确行为 0（模板中不展示）
      _errorRowIndex: hasError ? rowStart + i : 0,
    };

    // 第一遍：填充普通列（有 excel 列号的列，通过 excelFieldToDataIndex 映射取值）
    for (const col of columns) {
      if (!col.formula && col.excelField !== '-') {
        const dataIdx = excelFieldToDataIndex[col.excelField];
        row[col.field] = dataIdx !== undefined && dataIdx < detail.raw_data.length ? detail.raw_data[dataIdx] : '';
      }
    }

    // 工具函数：根据 excel 列号获取当前行的数值（用于 formula 计算）
    // 注意：hidden 列不在 raw_data 中，也不在 excelFieldToDataIndex 中，
    // 若公式引用了 hidden 列的列号，此处会返回 0
    const getValueByExcelField = (excelField: string): number => {
      const idx = excelFieldToDataIndex[excelField];
      if (idx === undefined || idx === null || idx >= detail.raw_data.length) return 0;
      const val = detail.raw_data[idx];
      const num = Number(val);
      return Number.isNaN(num) ? 0 : num;
    };

    // 第二遍：计算 formula 列
    for (const col of columns) {
      if (col.formula) {
        const result = evaluateFormula(col.formula, getValueByExcelField);
        // 根据列类型处理结果
        row[col.field] = col.type === 'int' ? Math.floor(result) : result;
      }
    }

    rows.push(row);
  }

  return rows;
}

/**
 * 可复用的 Excel 预览 composable
 * @param data - 接口返回的预览数据 (reactive ref)
 */
export function useExcelPreview(data: Ref<IExcelImportData | null>) {
  const activeTab = ref('');

  /** Tab 列表信息 */
  const tabs = computed<ITabInfo[]>(() => {
    if (!data.value) return [];
    const details = data.value.details ?? [];
    return data.value.sheets.map((sheet) => {
      const sheetDetails = details.filter((d) => d.name === sheet.name);
      const errorCount = sheetDetails.filter((d) => d.validate_result.length > 0).length;
      return {
        name: sheet.name,
        errorCount,
        totalCount: sheetDetails.length,
        hasError: errorCount > 0,
      };
    });
  });

  /** 当前活动 Tab 的 header 解析结果（缓存中间计算） */
  const currentHeadersResult = computed<IVisibleHeadersResult | null>(() => {
    if (!data.value || !activeTab.value) return null;
    const sheet = data.value.sheets.find((s) => s.name === activeTab.value);
    if (!sheet) return null;
    return getVisibleHeaders(sheet);
  });

  /** 当前活动 Tab 的表格列 */
  const currentColumns = computed<ITableColumn[]>(() => currentHeadersResult.value?.columns ?? []);

  /** 当前活动 Tab 的表格数据 */
  const currentRows = computed<ITableRow[]>(() => {
    if (!data.value || !activeTab.value || !currentHeadersResult.value) return [];
    if (!data.value.details) return [];
    const sheet = data.value.sheets.find((s) => s.name === activeTab.value);
    const rowStart = sheet?.row_start ?? 1;
    return buildTableRows(
      data.value.details,
      activeTab.value,
      currentHeadersResult.value.columns,
      currentHeadersResult.value.excelFieldToDataIndex,
      rowStart,
    );
  });

  /** 初始化：选中第一个 Tab */
  const initActiveTab = () => {
    if (tabs.value.length > 0 && !activeTab.value) {
      activeTab.value = tabs.value[0].name;
    }
  };

  return {
    activeTab,
    tabs,
    currentColumns,
    currentRows,
    initActiveTab,
  };
}

// =============================================
// 提交数据转换：Excel 预览数据 → 创建接口 details
// =============================================

/** 创建接口的子单结构 */
export interface ISubmitDetail {
  demand_type: string;
  demand_year: number;
  demand_month: number;
  gpu_num: number;
  qpm_max: number;
  extension: any[];
}

/**
 * 将 Excel 导入预览数据转换为创建接口需要的 details 数组
 *
 * 转换规则：
 * 1. demand_type = detail.name（sheet 名称）
 * 2. demand_year / demand_month / gpu_num / qpm_max：
 *    从 fixed_headers 中按 db_field 映射提取（优先从 raw_data 取值，field="-" 的公式列则动态计算）
 * 3. extension：按 headers 中 field 不为 "-" 的列顺序，从 raw_data 提取
 * 4. 仅提交校验通过（validate_result 为空）的行
 *
 * 注意：raw_data 包含所有 field 不为 "-" 的列数据（无论 hidden 与否），
 *       只有 field === "-" 的公式计算列不在 raw_data 中。
 *
 * @param data - Excel 导入预览接口返回的数据
 * @returns 提交接口所需的 details 数组
 */
export function buildSubmitDetails(data: IExcelImportData): ISubmitDetail[] {
  const result: ISubmitDetail[] = [];

  // 预处理：为每个 sheet 构建 fixed_headers / headers 的 raw_data 索引映射
  const sheetMetaMap = new Map<
    string,
    {
      sheet: IExcelSheet;
      /** db_field → raw_data 索引（fixed_headers 中 field 不为 "-" 的列） */
      fixedDbFieldToDataIndex: Record<string, number>;
      /** db_field → formula（field 为 "-" 的公式计算列，如 qpm_max） */
      fixedDbFieldFormula: Record<string, string>;
      /** excel 列号 → raw_data 索引（全局，用于公式计算取值） */
      excelFieldToDataIndex: Record<string, number>;
      /** headers 中 field 不为 "-" 的列在 raw_data 中的起始索引 */
      extensionStartIndex: number;
      /** headers 中 field 不为 "-" 的列数 */
      extensionCount: number;
    }
  >();

  for (const sheet of data.sheets) {
    const excelFieldToDataIndex: Record<string, number> = {};
    const fixedDbFieldToDataIndex: Record<string, number> = {};
    const fixedDbFieldFormula: Record<string, string> = {};
    let dataIndex = 0;

    // 遍历 fixed_headers
    // raw_data 包含所有 field 不为 "-" 的列数据（无论 hidden 与否）
    // 只有 field === "-" 的列（公式计算列）不在 raw_data 中
    for (const header of sheet.fixed_headers) {
      const hasExcelField = header.field !== '-';

      if (hasExcelField) {
        excelFieldToDataIndex[header.field] = dataIndex;
        if (header.db_field) {
          fixedDbFieldToDataIndex[header.db_field] = dataIndex;
        }
        dataIndex += 1;
      } else if (header.db_field && header.formula) {
        // field === "-" 的公式计算列（如 qpm_max），值需要公式动态计算
        fixedDbFieldFormula[header.db_field] = header.formula;
      }
    }

    const extensionStartIndex = dataIndex;

    // 遍历 headers（动态列）
    for (const header of sheet.headers) {
      const hasExcelField = header.field !== '-';
      if (hasExcelField) {
        excelFieldToDataIndex[header.field] = dataIndex;
        dataIndex += 1;
      }
    }

    const extensionCount = dataIndex - extensionStartIndex;

    sheetMetaMap.set(sheet.name, {
      sheet,
      fixedDbFieldToDataIndex,
      fixedDbFieldFormula,
      excelFieldToDataIndex,
      extensionStartIndex,
      extensionCount,
    });
  }

  // 遍历 details，仅处理校验通过的行
  for (const detail of data.details ?? []) {
    if (detail.validate_result.length > 0) continue;

    const meta = sheetMetaMap.get(detail.name);
    if (!meta) continue;

    // 提取固定字段值
    const getFixedValue = (dbField: string, defaultValue = 0): number => {
      // 优先从 raw_data 直接取值
      const idx = meta.fixedDbFieldToDataIndex[dbField];
      if (idx !== undefined && idx < detail.raw_data.length) {
        const num = Number(detail.raw_data[idx]);
        return Number.isNaN(num) ? defaultValue : num;
      }
      // 其次尝试公式计算
      const formula = meta.fixedDbFieldFormula[dbField];
      if (formula) {
        const getValueByExcelField = (excelField: string): number => {
          const i = meta.excelFieldToDataIndex[excelField];
          if (i === undefined || i >= detail.raw_data.length) return 0;
          const num = Number(detail.raw_data[i]);
          return Number.isNaN(num) ? 0 : num;
        };
        return evaluateFormula(formula, getValueByExcelField);
      }
      return defaultValue;
    };

    // 提取 extension 数组
    const extension = detail.raw_data.slice(meta.extensionStartIndex, meta.extensionStartIndex + meta.extensionCount);

    result.push({
      demand_type: detail.name,
      demand_year: getFixedValue('demand_year'),
      demand_month: getFixedValue('demand_month'),
      gpu_num: getFixedValue('gpu_num'),
      qpm_max: getFixedValue('qpm_max'),
      extension,
    });
  }

  return result;
}

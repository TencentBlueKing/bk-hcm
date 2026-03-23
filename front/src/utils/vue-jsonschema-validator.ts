/**
 * JSON Schema校验器 - Vue/TypeScript版本
 * 将自定义Schema转换为标准JSON Schema后使用ajv进行校验
 * ajv v6 (draft-07) 兼容
 */

import Ajv, { type ErrorObject, type ValidateFunction } from 'ajv';

import { ref } from 'vue';

// 类型定义
export interface FieldSchema {
  name: string;
  field: string;
  type: string;
  required?: boolean;
  readonly?: boolean;
  formula?: string;
  value?: (string | number)[];
  /** 大于 (exclusive minimum) */
  gt?: number;
  /** 大于等于 (inclusive minimum) */
  gte?: number;
  /** 小于 (exclusive maximum) */
  lt?: number;
  /** 小于等于 (inclusive maximum) */
  lte?: number;
}

export interface SheetSchema {
  name: string;
  type?: string;
  start?: number;
  header: FieldSchema[];
}

export interface CustomSchema {
  sheets: SheetSchema[];
}

export interface JSONSchema {
  $schema?: string;
  title?: string;
  description?: string;
  type: string;
  properties?: Record<string, PropertySchema>;
  required?: string[];
  definitions?: Record<string, JSONSchema>;
  items?: JSONSchema;
}

export interface PropertySchema {
  type?: string | string[];
  description?: string;
  enum?: (string | number)[];
  minimum?: number;
  maximum?: number;
  exclusiveMinimum?: number;
  exclusiveMaximum?: number;
  minLength?: number;
  maxLength?: number;
  readOnly?: boolean;
  pattern?: string;
}

export interface ValidationError {
  field: string;
  message: string;
  value?: any;
  row?: number;
}

export interface ValidationResult {
  valid: boolean;
  errors: ValidationError[];
  sheet?: string;
  row?: number;
}

/**
 * Schema转换器：将自定义Schema转换为标准JSON Schema
 */
export class SchemaConverter {
  /**
   * 将自定义Schema转换为JSON Schema
   * 同时返回每个 sheet 中必填字段名集合，用于空值清理
   */
  static convertToJSONSchema(schema: CustomSchema): {
    schemas: Record<string, JSONSchema>;
    requiredFieldsMap: Record<string, Set<string>>;
  } {
    const schemas: Record<string, JSONSchema> = {};
    const requiredFieldsMap: Record<string, Set<string>> = {};

    for (const sheet of schema.sheets) {
      schemas[sheet.name] = {
        $schema: 'http://json-schema.org/draft-07/schema#',
        title: `${sheet.name} 数据校验Schema`,
        description: `Sheet: ${sheet.name}, 数据起始行: ${sheet.start}`,
        type: 'object',
        properties: {},
        required: [],
      };
      requiredFieldsMap[sheet.name] = new Set();

      for (const field of sheet.header) {
        const propName = this.sanitizeFieldName(field.name);
        const prop = this.convertFieldToProperty(field);
        schemas[sheet.name].properties![propName] = prop;

        if (field.required) {
          schemas[sheet.name].required!.push(propName);
          requiredFieldsMap[sheet.name].add(propName);
        }
      }
    }

    return { schemas, requiredFieldsMap };
  }

  /**
   * 根据 gt/gte/lt/lte 设置 JSON Schema 的范围约束
   * gt  → exclusiveMinimum（大于）
   * gte → minimum（大于等于）
   * lt  → exclusiveMaximum（小于）
   * lte → maximum（小于等于）
   * 接口没有给约束则不设置默认值，允许任意范围
   */
  private static applyRangeConstraints(prop: PropertySchema, field: FieldSchema): void {
    // 仅根据接口给出的约束设置范围，接口没给则不限制
    // gt（大于）→ exclusiveMinimum
    if (field.gt !== undefined) {
      prop.exclusiveMinimum = field.gt;
    }
    // gte（大于等于）→ minimum
    if (field.gte !== undefined) {
      prop.minimum = field.gte;
    }
    // lt（小于）→ exclusiveMaximum
    if (field.lt !== undefined) {
      prop.exclusiveMaximum = field.lt;
    }
    // lte（小于等于）→ maximum
    if (field.lte !== undefined) {
      prop.maximum = field.lte;
    }
  }

  /**
   * 将字段定义转换为JSON Schema属性
   */
  private static convertFieldToProperty(field: FieldSchema): PropertySchema {
    const prop: PropertySchema = {
      description: field.name,
      readOnly: field.readonly,
    };

    const fieldType = (field.type || 'string').trim();

    if (fieldType === 'string') {
      prop.type = 'string';
      // 根据接口的 gt/gte/lt/lte 设置 minLength/maxLength
      if (field.gt !== undefined) {
        prop.minLength = field.gt + 1; // gt（大于）→ minLength = gt + 1
      }
      if (field.gte !== undefined) {
        prop.minLength = field.gte; // gte（大于等于）→ minLength = gte
      }
      if (field.lt !== undefined) {
        prop.maxLength = field.lt - 1; // lt（小于）→ maxLength = lt - 1
      }
      if (field.lte !== undefined) {
        prop.maxLength = field.lte; // lte（小于等于）→ maxLength = lte
      }
      // 必填字段至少保证 minLength >= 1
      if (field.required && (prop.minLength === undefined || prop.minLength < 1)) {
        prop.minLength = 1;
      }
    } else if (fieldType === 'int') {
      prop.type = ['integer', 'string'];
      this.applyRangeConstraints(prop, field);
    } else if (fieldType === 'float' || /^float\(\d+\)$/.test(fieldType)) {
      prop.type = ['number', 'string'];
      this.applyRangeConstraints(prop, field);
    } else if (fieldType === 'enum') {
      // enum 值可能是 string 或 number，允许两种类型
      if (field.value && field.value.length > 0) {
        prop.enum = field.value;
      }
    } else {
      prop.type = 'string';
    }

    // 特殊字段规则
    if (field.name.includes('利用率') || field.name.includes('使用率')) {
      prop.type = ['integer', 'number', 'string'];
      prop.minimum = 0;
      prop.maximum = 100;
      prop.description = `${field.name} (范围: 0-100)`;
    }

    if (field.name.includes('参数量')) {
      prop.type = ['number', 'string'];
      prop.exclusiveMinimum = 0;
      prop.description = `${field.name} (必须>0)`;
    }

    if (field.formula) {
      prop.readOnly = true;
      prop.description = `${field.name} [公式: ${field.formula}]`;
    }

    return prop;
  }

  /**
   * 将字段名转换为合法的属性名
   */
  private static sanitizeFieldName(name: string): string {
    return name
      .replace(/[()]/g, (match) => (match === '(' ? '_' : ''))
      .replace(/[/-]/g, '_')
      .replace(/\s+/g, '_')
      .replace(/_+/g, '_')
      .replace(/^_|_$/g, '');
  }
}

/**
 * JSON Schema校验器
 */
export class JSONSchemaValidator {
  /**
   * 从JSON字符串创建校验器
   */
  static fromJson(jsonString: string): JSONSchemaValidator {
    const content = jsonString.trim();
    const parsed = JSON.parse(content);

    // 支持两种格式：
    // 1. {"sheets": [...]} - 带sheets包装（当前budget_declaration_gpu_template_20260312.json的格式）
    // 2. [...] - 直接数组
    if (Array.isArray(parsed)) {
      return new JSONSchemaValidator({ sheets: parsed });
    }

    return new JSONSchemaValidator(parsed as CustomSchema);
  }

  /**
   * 从 ITplSheet 配置直接创建校验器
   * 将 sheet 的 fixed_headers + headers 合并为一个 FieldSchema 数组
   */
  static fromSheet(sheetName: string, fixedHeaders: FieldSchema[], headers: FieldSchema[]): JSONSchemaValidator {
    const allHeaders = [...fixedHeaders, ...headers];
    return new JSONSchemaValidator({
      sheets: [
        {
          name: sheetName,
          header: allHeaders,
        },
      ],
    });
  }

  private ajv: InstanceType<typeof Ajv>;
  private schemas: Record<string, JSONSchema>;
  private validators: Map<string, ValidateFunction>;
  private requiredFieldsMap: Record<string, Set<string>>;

  constructor(customSchema: CustomSchema) {
    this.ajv = new Ajv({ allErrors: true, coerceTypes: true });
    const { schemas, requiredFieldsMap } = SchemaConverter.convertToJSONSchema(customSchema);
    this.schemas = schemas;
    this.requiredFieldsMap = requiredFieldsMap;
    this.validators = new Map();

    // 预编译所有Schema
    for (const [name, schema] of Object.entries(this.schemas)) {
      this.validators.set(name, this.ajv.compile(schema));
    }
  }

  /**
   * 校验单行数据
   */
  validateRow(sheetName: string, rowData: Record<string, any>, rowNum: number): ValidationResult {
    const result: ValidationResult = {
      valid: true,
      errors: [],
      sheet: sheetName,
      row: rowNum,
    };

    const validate = this.validators.get(sheetName);
    if (!validate) {
      result.valid = false;
      result.errors.push({
        field: '_sheet',
        message: `未找到Sheet定义: ${sheetName}`,
      });
      return result;
    }

    // 转换字段名
    const convertedData = this.convertFieldNames(rowData);

    // 非必填字段的空值清理：移除空值字段，让 ajv 跳过校验
    // 只有在 required 数组中的字段，空值才会触发 "不能为空" 校验
    const requiredFields = this.requiredFieldsMap[sheetName] || new Set();
    const cleanedData: Record<string, any> = {};
    for (const [key, value] of Object.entries(convertedData)) {
      const isEmpty = value === '' || value === undefined || value === null;
      if (isEmpty && !requiredFields.has(key)) {
        // 非必填字段的空值：不传入 ajv，跳过所有校验
        continue;
      }
      cleanedData[key] = value;
    }

    // 执行校验
    const valid = validate(cleanedData);
    if (!valid) {
      result.valid = false;
      result.errors = this.formatErrors(validate.errors, rowNum);
    }

    return result;
  }

  /**
   * 批量校验整个Sheet
   */
  validateSheet(sheetName: string, rows: Record<string, any>[]): ValidationResult[] {
    const results: ValidationResult[] = [];

    const schema = Object.entries(this.schemas).find(([name]) => name === sheetName)?.[1];
    if (!schema) {
      results.push({
        valid: false,
        errors: [{ field: '_sheet', message: `未找到Sheet定义: ${sheetName}` }],
        sheet: sheetName,
      });
      return results;
    }

    const startRow = 5; // 默认起始行
    for (let i = 0; i < rows.length; i++) {
      const rowNum = startRow + i;
      const result = this.validateRow(sheetName, rows[i], rowNum);
      results.push(result);
    }

    return results;
  }

  /**
   * 获取JSON Schema定义
   */
  getJSONSchema(sheetName?: string): JSONSchema | Record<string, JSONSchema> {
    if (sheetName) {
      return this.schemas[sheetName] || null;
    }
    return this.schemas;
  }

  /**
   * 获取Schema信息
   */
  getSchemaInfo(): any {
    return {
      sheets: Object.keys(this.schemas).map((name) => ({
        name,
        properties: Object.keys(this.schemas[name].properties || {}).length,
        required: (this.schemas[name].required || []).length,
      })),
    };
  }

  /**
   * 转换字段名（将中文字段名转换为合法的属性名）
   */
  private convertFieldNames(data: Record<string, any>): Record<string, any> {
    const result: Record<string, any> = {};
    for (const [key, value] of Object.entries(data)) {
      const sanitizedKey = SchemaConverter['sanitizeFieldName'](key);
      result[sanitizedKey] = value;
    }
    return result;
  }

  /**
   * 将 ajv 英文错误消息翻译为中文
   */
  private translateError(err: ErrorObject): string {
    const params = err.params as Record<string, any>;
    switch (err.keyword) {
      case 'required':
        return '不能为空';
      case 'type':
        return `类型应为 ${params.type}`;
      case 'enum':
        return '值不在允许的选项范围内';
      case 'minimum':
        return `不能小于 ${params.limit}`;
      case 'maximum':
        return `不能大于 ${params.limit}`;
      case 'exclusiveMinimum':
        return `必须大于 ${params.limit}`;
      case 'exclusiveMaximum':
        return `必须小于 ${params.limit}`;
      case 'minLength':
        return `长度不能少于 ${params.limit} 个字符`;
      case 'maxLength':
        return `长度不能超过 ${params.limit} 个字符`;
      case 'pattern':
        return '格式不正确';
      case 'format':
        return `格式应为 ${params.format}`;
      default:
        return err.message || '校验失败';
    }
  }

  /**
   * 格式化错误信息
   */
  private formatErrors(errors: ErrorObject[] | null | undefined, rowNum: number): ValidationError[] {
    if (!errors) return [];

    return errors.map((err) => ({
      field: err.dataPath.replace(/^\./, '') || (err.params as Record<string, any>).missingProperty || 'unknown',
      message: this.translateError(err),
      row: rowNum,
    }));
  }
}

/**
 * Vue组合式API - 表单校验Hook
 */
export function useFormValidator(validator: JSONSchemaValidator, sheetName: string) {
  const errors = ref<Record<string, string>>({});

  const validateField = (field: string, value: any): string | null => {
    const result = validator.validateRow(sheetName, { [field]: value }, 0);
    if (!result.valid && result.errors.length > 0) {
      const errorMsg = result.errors[0].message;
      errors.value[field] = errorMsg;
      return errorMsg;
    }
    delete errors.value[field];
    return null;
  };

  const validateForm = (formData: Record<string, any>): boolean => {
    const result = validator.validateRow(sheetName, formData, 0);
    if (!result.valid) {
      errors.value = {};
      for (const err of result.errors) {
        errors.value[err.field] = err.message;
      }
      return false;
    }
    errors.value = {};
    return true;
  };

  const clearErrors = () => {
    errors.value = {};
  };

  return {
    errors,
    validateField,
    validateForm,
    clearErrors,
  };
}

// 如果不使用Vue，可以移除上面的import和useFormValidator函数

// 默认导出
export default JSONSchemaValidator;

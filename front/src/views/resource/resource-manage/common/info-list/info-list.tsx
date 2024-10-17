import { computed, defineComponent, inject, PropType, useTemplateRef } from 'vue';

import { OverflowTitle } from 'bkui-vue';
import { Share, Copy } from 'bkui-vue/lib/icon';
import RenderDetailEdit from '@/components/RenderDetailEdit';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';

import './info-list.scss';
import { Field, FieldList } from './types';

export default defineComponent({
  components: {
    Share,
    Copy,
  },

  props: {
    fields: Array as PropType<FieldList>,
    col: { type: Number, default: 2 },
    labelWidth: String,
    globalCopyable: { type: Boolean, default: false },
  },

  emits: ['change'],

  setup(props, { emit }) {
    // 注入预鉴权相关参数
    const authAction = inject<any>('authAction');
    // undefined：不校验权限；true：有权限；false：无权限
    const isAuth = computed(() => {
      if (authAction) {
        const { authVerifyData, authId } = authAction;
        return !(authId && !authVerifyData?.value?.permissionAction?.[authId]);
      }
      return undefined;
    });

    const gridTemplateColumnsStyle = computed(() => `repeat(${props.col}, calc(${100 / props.col}% - 12px))`);
    // item 最大宽度至少需要减去 column-gap * (col - 1), 避免出现横向滚动条
    const itemMaxWidthBaseDecrement = computed(() => 12 * (props.col - 1));

    const handleBlur = async (val: any, key: string) => {
      emit('change', { [key]: val });
    };

    const editCompRef = useTemplateRef<typeof RenderDetailEdit>('detail-edit-comp');
    // 如果祖辈组件有 provide 预鉴权参数, 则需要对编辑操作进行预鉴权处理
    const handleEdit = () => {
      if (isAuth.value === false) {
        // 无权限, 展示权限申请弹窗
        const { authId, handleAuth } = authAction;
        handleAuth(authId);
      } else {
        editCompRef.value.handleEdit();
      }
    };

    return {
      gridTemplateColumnsStyle,
      itemMaxWidthBaseDecrement,
      handleBlur,
      handleEdit,
      isAuth,
    };
  },

  render() {
    // 渲染纯文本
    const renderTxt = (field: Field) => {
      const { value } = field;
      if (Array.isArray(value)) {
        if (value.length === 0) return '--';
        return value.join(',')?.concat(';');
      }
      if (typeof value === 'number') {
        return value;
      }
      return value || '--';
    };

    // 渲染可编辑文本
    const renderEditTxt = (field: Field) => (
      <RenderDetailEdit
        ref='detail-edit-comp'
        modelValue={field.value}
        fromType={field.type}
        needValidate={false}
        fromKey={field.prop}
        hideEdit={true}
        onChange={this.handleBlur}></RenderDetailEdit>
    );

    // 渲染链接
    const renderLink = (field: Field) => (
      <bk-link theme='primary' href={typeof field.link === 'function' ? field.link(field.value) : field.link}>
        {field.value}
      </bk-link>
    );

    // 渲染跳转
    const renderTxtBtn = (field: Field) => {
      return field.value ? (
        <bk-button text theme='primary' onClick={() => field.txtBtn(field.value)}>
          {field.value}
        </bk-button>
      ) : (
        '--'
      );
    };

    // 渲染方法
    const renderField = (field: Field) => {
      if (field.render) {
        return field.render(field.value);
      }
      if (field.link) {
        return renderLink(field);
      }
      if (field.edit) {
        return renderEditTxt(field);
      }
      if (field.txtBtn) {
        return renderTxtBtn(field);
      }
      return renderTxt(field);
    };

    return (
      <ul class='info-list-main g-scroller' style={{ gridTemplateColumns: this.gridTemplateColumnsStyle }}>
        {this.fields.map((field) => {
          const { name, value, cls, render, edit, copy, copyContent, tipsContent } = field;

          // copy配置的优先级：局部 > 全局
          const resultCopyable = copy ?? this.globalCopyable;

          // 处理copy内容，copy内容取值的优先级：field.copyContent > field.render > field.value > 默认值'--'
          let resultCopyContent =
            (typeof copyContent === 'function' ? copyContent(value) : copyContent) ?? (render ? render(value) : value);
          if (Array.isArray(resultCopyContent)) {
            resultCopyContent = resultCopyContent.join('\r');
          } else if (typeof resultCopyContent === 'string') {
            resultCopyContent = resultCopyContent || '--';
          } else if (typeof resultCopyContent === 'number') {
            resultCopyContent = resultCopyContent.toString();
          } else {
            resultCopyContent = '--';
          }

          let operationBtnWidth = 0;
          if (resultCopyable) operationBtnWidth += 24;
          if (edit) operationBtnWidth += 24;
          const itemMaxWidth = `calc(100% - ${this.itemMaxWidthBaseDecrement}px)`;
          const valueMaxWidth = `calc(100% - ${parseFloat(this.labelWidth) + operationBtnWidth}px)`;

          return (
            <li class='info-list-item' style={{ maxWidth: itemMaxWidth }}>
              {tipsContent ? (
                <div class='item-label has-tips' style={{ width: this.labelWidth }}>
                  <span v-BkTooltips={{ content: tipsContent }}>{name}</span>
                </div>
              ) : (
                <span class='item-label' style={{ width: this.labelWidth }}>
                  {name}
                </span>
              )}
              <span
                class={['item-value', typeof cls === 'function' ? cls(value) : cls]}
                style={{ maxWidth: valueMaxWidth }}>
                <OverflowTitle class='full-width' type='tips' content={renderField(field)}>
                  {renderField(field)}
                </OverflowTitle>
              </span>
              {edit && (
                <i
                  onClick={this.handleEdit}
                  class={[
                    'icon hcm-icon bkhcm-icon-bianji edit-icon',
                    {
                      'hcm-no-permision-text-btn': this.isAuth === false,
                    },
                  ]}
                />
              )}
              {resultCopyable && <CopyToClipboard class='copy-btn' content={resultCopyContent} />}
            </li>
          );
        })}
      </ul>
    );
  },
});

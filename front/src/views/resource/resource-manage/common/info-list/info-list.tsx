import { defineComponent, PropType } from 'vue';

import { Share, Copy } from 'bkui-vue/lib/icon';

import { Message } from 'bkui-vue';

import RenderDetailEdit from '@/components/RenderDetailEdit';

import { useI18n } from 'vue-i18n';

import './info-list.scss';

type Field = {
  name: string;
  value: string | any;
  cls?: string | ((cell: string) => string);
  link?: string | ((cell: string) => string);
  copy?: boolean;
  edit?: boolean;
  type?: string;
  tipsContent?: string;
  txtBtn?: (cell: string) => void;
  render?: (cell: string) => void;
};

export default defineComponent({
  components: {
    Share,
    Copy,
  },

  props: {
    fields: Array as PropType<Field[]>,
  },

  emits: ['change'],

  setup(_, { emit }) {
    const { t } = useI18n();

    const handleCopy = (val: string) => {
      const handleSuccessCopy = () => {
        Message({
          message: t('复制成功'),
          theme: 'success',
        });
      };

      if (window.isSecureContext && navigator.clipboard) {
        navigator.clipboard.writeText(val).then(handleSuccessCopy);
      } else {
        const input = document.createElement('input');
        document.body.appendChild(input);
        input.setAttribute('value', val);
        input.select();
        if (document.execCommand('copy')) {
          document.execCommand('copy');
          handleSuccessCopy();
        }
        document.body.removeChild(input);
      }
    };

    const handleblur = async (val: any, key: string) => {
      emit('change', { [key]: val });
    };

    return {
      handleCopy,
      handleblur,
    };
  },

  render() {
    // 渲染纯文本
    const renderTxt = (field: Field) => {
      const type = Object.prototype.toString.call(field.value);
      switch (type) {
        case '[object Array]':
          return field.value.map((e: string, index: number) => (
            <>
              <span>{e}</span>
              {field.value.length - 1 === index ? '' : ';'}
            </>
          ));
        default:
          return field.value || '--';
      }
    };

    // 渲染可编辑文本
    const renderEditTxt = (field: Field) => (
      <RenderDetailEdit
        modelValue={field.value}
        needValidate={false}
        fromKey={field.prop}
        onChange={this.handleblur}></RenderDetailEdit>
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
      <ul class='info-list-main g-scroller'>
        {this.fields.map((field) => {
          return (
            <>
              <li class='info-list-item'>
                {field.tipsContent ? (
                  <div class='item-field has-tips'>
                    <span v-BkTooltips={{ content: field.tipsContent }}>{field.name}</span>
                  </div>
                ) : (
                  <span class='item-field'>{field.name}</span>
                )}
                :
                <span class={['item-value', typeof field.cls === 'function' ? field.cls(field.value) : field.cls]}>
                  {renderField(field)}
                </span>
                {field.copy ? (
                  <copy class='info-item-copy ml5' onClick={() => this.handleCopy(field.value)}></copy>
                ) : (
                  ''
                )}
              </li>
            </>
          );
        })}
      </ul>
    );
  },
});

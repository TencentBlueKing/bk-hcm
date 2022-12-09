import {
  defineComponent,
  PropType,
} from 'vue';

import {
  Share,
  Copy,
} from 'bkui-vue/lib/icon';

import {
  Message,
} from 'bkui-vue';

import {
  useI18n,
} from 'vue-i18n';

import './info-list.scss';

type Field = {
  name: string;
  value: string;
  link?: string;
  copy?: string;
  edit?: boolean;
};

export default defineComponent({
  components: {
    Share,
    Copy,
  },

  props: {
    fields: Array as PropType<Field[]>,
  },

  setup() {
    const {
      t,
    } = useI18n();

    const handleCopy = (val: string) => {
      const handleSuccessCopy = () => {
        Message({
          message: t('复制成功'),
          theme: 'success',
        });
      };

      if (window.isSecureContext && navigator.clipboard) {
        navigator
          .clipboard
          .writeText(val)
          .then(handleSuccessCopy);
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

    return {
      handleCopy,
    };
  },

  render() {
    // 渲染纯文本
    const renderTxt = (field: Field) => <span>{ field.value }</span>;
    // 渲染可编辑文本
    const renderEditTxt = (field: Field) => <span>{ field.value }</span>;
    // 渲染链接
    const renderLink = (field: Field) => <bk-link theme="primary" target="_blank" href={field.link}>{ field.value }</bk-link>;

    // 渲染方法
    const renderField = (field: Field) => {
      if (field.link) {
        return renderLink(field);
      } if (field.edit) {
        return renderEditTxt(field);
      }
      return renderTxt(field);
    };

    return <ul class="info-list-main g-scroller">
      {
        this.fields.map((field) => {
          return <>
            <li class="info-list-item">
              { field.name }：{ renderField(field) }
              {
                field.copy ? <copy class="info-item-copy ml5" onClick={() => this.handleCopy(field.copy)}></copy> : ''
              }
            </li>
          </>;
        })
      }
    </ul>;
  },
});

import { defineComponent, PropType } from 'vue';

import InfoList from '../info-list/info-list';

import './detail-info.scss';

type Field = {
  name: string;
  value?: string;
  cls?: string | ((cell: string) => string);
  link?: string | ((cell: string) => string);
  copy?: string | boolean;
  edit?: boolean;
  prop?: string;
  tipsContent?: string;
  type?: string;
  render?: (cell: string | boolean) => void;
};

export default defineComponent({
  components: {
    InfoList,
  },

  props: {
    fields: Array as PropType<Field[]>,
    detail: Object,
    col: { type: Number, default: 2 },
  },

  emits: ['change'],
  setup(props, { emit }) {
    const handleChange = (val: any) => {
      emit('change', val);
    };
    return {
      handleChange,
      props,
    };
  },

  computed: {
    renderFields() {
      return this.fields.map((field) => {
        return {
          ...field,
          value: field.value || this.detail?.[field?.prop],
        };
      });
    },
  },

  render() {
    return (
      <info-list
        class='detail-info-main g-scroller'
        fields={this.renderFields}
        onChange={this.handleChange}
        col={this.col}
      />
    );
  },
});

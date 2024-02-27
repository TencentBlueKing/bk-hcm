import {
  defineComponent,
  PropType,
} from 'vue';

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
  render?: (cell: string | boolean) => void;
};

export default defineComponent({
  components: {
    InfoList,
  },

  props: {
    fields: Array as PropType<Field[]>,
    detail: Object,
  },

  emits: ['change'],
  setup(_, { emit }) {
    const handleChange = (val: any) => {
      emit('change', val);
    };
    return {
      handleChange,
    };
  },

  computed: {
    renderFields() {
      return this.fields.map((field) => {
        return {
          ...field,
          value: field.value || this.detail[field?.prop],
        };
      });
    },
  },


  render() {
    return <info-list
      class="detail-info-main g-scroller"
      fields={ this.renderFields }
      onChange={this.handleChange}
    />;
  },
});

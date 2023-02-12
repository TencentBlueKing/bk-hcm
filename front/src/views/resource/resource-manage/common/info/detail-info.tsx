import {
  defineComponent,
  PropType,
} from 'vue';

import InfoList from '../info-list/info-list';

import './detail-info.scss';

type Field = {
  name: string;
  value?: string;
  link?: string | ((cell: string) => string);
  copy?: string | boolean;
  edit?: boolean;
  prop?: string;
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

  computed: {
    renderFields() {
      return this.fields.map((field) => {
        return {
          ...field,
          value: field.value || this.detail[field?.prop] || '--',
        };
      });
    },
  },

  render() {
    return <>
      <info-list
        class="detail-info-main g-scroller"
        fields={ this.renderFields }
      />
    </>;
  },
});

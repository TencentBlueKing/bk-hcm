import {
  defineComponent,
  PropType,
} from 'vue';

import InfoList from '../info-list/info-list';

import './detail-info.scss';

type Field = {
  name: string;
  value?: string;
  link?: string;
  copy?: string;
  edit?: boolean;
  prop?: string;
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
          value: this.detail[field.prop],
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

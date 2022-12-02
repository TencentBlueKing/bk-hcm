import {
  defineComponent,
  PropType,
} from 'vue';

import InfoList from '../info-list/info-list';

import './detail-info.scss';

type Field = {
  name: string;
  value: string;
  link?: string;
  copy?: string;
  edit?: boolean;
};

export default defineComponent({
  components: {
    InfoList,
  },

  props: {
    fields: Array as PropType<Field[]>,
  },

  render() {
    return <>
      <info-list
        class="detail-info-main g-scroller"
        fields={ this.fields }
      />
    </>;
  },
});

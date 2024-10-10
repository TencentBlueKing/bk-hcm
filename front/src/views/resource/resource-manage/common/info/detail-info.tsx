import { defineComponent, PropType } from 'vue';

import InfoList from '../info-list/info-list';

import './detail-info.scss';
import { FieldList } from '../info-list/types';

export default defineComponent({
  components: {
    InfoList,
  },

  props: {
    fields: Array as PropType<FieldList>,
    detail: Object,
    col: { type: Number, default: 2 },
    labelWidth: { type: String, default: () => '120px' },
    globalCopyable: { type: Boolean, default: false },
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
      <InfoList
        class='detail-info-main g-scroller'
        fields={this.renderFields}
        onChange={this.handleChange}
        col={this.col}
        labelWidth={this.labelWidth}
        globalCopyable={this.globalCopyable}
      />
    );
  },
});

import { computed, defineComponent, ref } from 'vue';
// import './index.scss';
interface propstypes {
  multiple: any;
  treeData: Array<any>;
  value: Array<any>;
  label?: string;
  children?: string;
}
export default defineComponent({
  name: 'TreeSelect',
  props: {
    value: {
      type: Array,
      default: (): any => [],
    },
    treeData: Array,
    label: {
      type: String,
      default: 'name',
    },
    children: {
      type: String,
      default: 'children',
    },
    multiple: {
      type: Boolean,
      default: false,
    },
  },

  emits: ['updateValue'],
  setup(props: propstypes, { emit }) {
    const data = computed(() => props.value.filter(item => !item.children?.length).map(item => item.name));
    const treeRef = ref();
    const handleNodeChecked = (item) => {
      emit('updateValue', item);
    };
    const handleRemoveTag = (name) => {
      const node = props.value.find(node => node.name === name);
      treeRef.value.setChecked(node, false);
    };
    return () => (
      <bk-select
        ref='select'
        model-value={data}
        custom-content
        multiple-mode='tag'
        multiple
        onTag-remove={handleRemoveTag}>
        <bk-tree
          ref='treeRef'
          data={props.treeData}
          show-node-type-icon={false}
          label={props.label}
          children={props.children}
          show-checkbox
          onNode-checked={handleNodeChecked}></bk-tree>
      </bk-select>
    );
  },
});

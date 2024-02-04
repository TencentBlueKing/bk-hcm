import { computed, defineComponent, ref, watchEffect } from 'vue';
import { useAccountStore } from '@/store';

export default defineComponent({
  props: {
    id: [Array, String],
    emptyText: {
      type: String,
      default: '--',
    },
  },
  setup(props) {
    const accountStore = useAccountStore();
    const businessList = ref([]);

    const ids = computed(() => (Array.isArray(props.id) ? props.id : [props.id]));

    watchEffect(async () => {
      const res = await accountStore.getBizList();
      businessList.value = res.data ?? [];
    });

    const names = computed(() => {
      const result = [];
      ids.value.forEach((id) => {
        result.push(businessList.value.find((item) => item.id === id)?.name ?? props.emptyText);
      });

      return result.join(',');
    });

    return { names };
  },
  render() {
    return <span>{this.names}</span>;
  },
});

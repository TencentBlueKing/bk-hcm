import { computed, defineComponent, ref, watchEffect } from 'vue';
import { useAccountStore } from '@/store';

export default defineComponent({
  props: {
    id: [Array, String],
  },
  setup(props) {
    const accountStore = useAccountStore();

    const ids = computed(() => (Array.isArray(props.id) ? props.id : [props.id]));

    const reqs = computed(() => {
      const reqs = [];
      ids.value.forEach((id) => {
        reqs.push(accountStore.getDepartmentInfo(id));
      });
      return reqs;
    });

    const names = ref([]);

    watchEffect(async () => {
      const result = await Promise.all(reqs.value);
      result.forEach((res) => names.value.push(res?.data?.full_name ?? '--'));
    });

    return { names };
  },
  render() {
    return <span>{this.names.join(',')}</span>;
  },
});

import http from '@/http';
import { defineComponent, PropType, ref, watchEffect } from 'vue';
const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  props: {
    id: Number as PropType<number>,
  },
  setup(props) {
    const name = ref('');

    watchEffect(async () => {
      if (props.id === null) {
        name.value = '--';
        return;
      }

      if (props.id !== -1) {
        const result = await http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/web/cloud_areas/list`, {
          id: props.id,
        });
        name.value = result?.data?.info?.[0]?.name;
      } else {
        name.value = '未绑定管控区域';
      }
    });

    return () => <span>{name.value}</span>;
  },
});

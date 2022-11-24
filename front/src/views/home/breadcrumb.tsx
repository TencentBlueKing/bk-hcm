import { defineComponent, onMounted, ref, watch  } from 'vue';
import { useRoute } from 'vue-router';
import { Breadcrumb } from 'bkui-vue';
import './breadcrumb.scss';

const { Item } = Breadcrumb;
export default defineComponent({
  setup() {
    const route = useRoute();
    const breadList = ref([]);
    const getBreadcrumb = (tempRoute: any) => {
      const matched = tempRoute[tempRoute.length - 1].meta.breadcrumb;
      breadList.value = matched;
    };
    onMounted(() => {
    //   getBreadcrumb();
    });
    watch(
      () => route.matched,
      (val) => {
        if (val?.length) {
          getBreadcrumb(val);
        }
      },
      { immediate: true },
    );
    return {
      breadList,
    };
  },

  render() {
    return (
      <div class="bread-layout">
        <Breadcrumb>
          {this.breadList?.map((breadName: any) => (
            <Item class="flex-row align-items-center">
              {breadName}
            </Item>
          ))}
        </Breadcrumb>
      </div>
    );
  },
});

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
      const matched = tempRoute.map((ele: any) => {
        return { path: ele.path, name: ele.name };
      });
      breadList.value = matched;
      console.log('breadList', breadList);
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
    );
    return {
      breadList,
    };
  },

  render() {
    return (
        <div class="bread-layout">
            <Breadcrumb>
                {this.breadList?.map((routeName: any) => (
                    <Item class="flex-row align-items-center">
                        {routeName.name}
                    </Item>
                ))}
            </Breadcrumb>
        </div>
    );
  },
});

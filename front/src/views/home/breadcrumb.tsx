import { defineComponent, onMounted, ref, watch  } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Breadcrumb } from 'bkui-vue';
import './breadcrumb.scss';

const { Item } = Breadcrumb;
export default defineComponent({
  setup() {
    const router = useRouter();
    const route = useRoute();
    const breadList = ref([]);
    const backRouter = ref('');
    const getBreadcrumb = (tempRoute: any) => {
      console.log('tempRoute', tempRoute);
      const matched = tempRoute[tempRoute.length - 1].meta.breadcrumb;
      breadList.value = matched;
      backRouter.value = tempRoute[tempRoute.length - 1].meta.backRouter;
    };
    onMounted(() => {
    });

    // 返回操作
    const back = (routerName: any) => {
      if (routerName === -1) {
        history.go(-1);
      }  else {
        router.push({
          name: routerName,
        });
      }
    };
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
      backRouter,
      back,
    };
  },

  render() {
    return (
      <div class="bread-layout">
        {this.backRouter ? (<i onClick={() => {
          this.back(this.backRouter);
        }} class={'icon hcm-icon bkhcm-icon-arrows--left-line pr10 back-icon'}/>) : ''}
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

import { computed, defineComponent } from 'vue';
import { useRoute } from 'vue-router';
import businesseMenus from '@/router/module/business';
import serviceMenus from '@/router/module/service';
// import { Breadcrumb } from 'bkui-vue';
import './breadcrumb.scss';

// const { Item } = Breadcrumb;
export default defineComponent({
  setup() {
    // const router = useRouter();
    const route = useRoute();
    const breadcrumbText = computed(() => {
      return [...businesseMenus, ...serviceMenus]
        .reduce((prev, item) => {
          if (item.children) prev.push(...item.children);
          else prev.push(item);
          return prev;
        }, [])
        .filter((item) => item.meta?.isShowBreadcrumb)
        .find((item) => item.path === route.path)?.meta?.title;
    });

    return {
      breadcrumbText,
    };

    /* const breadList = ref([]);
    const backRouter = ref('');
    const getBreadcrumb = (tempRoute: any) => {
      const matched = tempRoute[tempRoute.length - 1].meta.breadcrumb;
      breadList.value = matched;
      backRouter.value = tempRoute[tempRoute.length - 1].meta.backRouter;
    };

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
        console.log('啦啦啦', val);
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
    }; */
  },

  render() {
    return (
      this.breadcrumbText && (
        <div class='navigation-breadcrumb'>
          <div class='bread-layout'>
            <span class='bread-name'>{this.breadcrumbText}</span>
          </div>
        </div>
      )
    );
    // <div class="bread-layout">
    //   {this.backRouter ? (<i onClick={() => {
    //     this.back(this.backRouter);
    //   }} class={'icon hcm-icon bkhcm-icon-arrows--left-line pr10 back-icon'}/>) : ''}
    //   <Breadcrumb>
    //     {this.breadList?.map((breadName: any) => (
    //       <Item class="flex-row align-items-center">
    //         {breadName}
    //       </Item>
    //     ))}
    //   </Breadcrumb>
    // </div>
  },
});

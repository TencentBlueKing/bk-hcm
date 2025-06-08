import { defineComponent, onMounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

// @ts-ignore
import AppSelect from '@blueking/app-select';
import '@blueking/app-select/dist/style.css';

import { useAccountStore } from '@/store';
import { localStorageActions } from '@/common/util';
import { getFavoriteList, useFavorite } from '@/hooks/useFavorite';
import { Button, Dialog, Exception } from 'bkui-vue';
import { GLOBAL_BIZS_KEY, GLOBAL_BIZS_VERSION, GLOBAL_BIZS_VERSION_KEY } from '@/common/constant';

export default defineComponent({
  name: 'BusinessSelector',
  props: { reload: Function },
  setup(props) {
    const router = useRouter();
    const route = useRoute();
    const accountStore = useAccountStore();

    const businessId = ref<number>();
    const businessList = ref([]);
    const favoriteList = ref([]);
    const isDialogShow = ref(false);

    const { favoriteSet, addToFavorite, removeFromFavorite } = useFavorite(businessId.value, favoriteList.value);

    const saveGlobalBizsId = (val: number) => {
      // 存入store
      accountStore.updateBizsId(val);
      // 存入local storage
      localStorageActions.set(GLOBAL_BIZS_KEY, val);
      // 返回query参数
      return { [GLOBAL_BIZS_KEY]: val };
    };

    // 选择业务
    const handleChange = async (val: { id: number }) => {
      if (businessId.value === val.id) return;
      businessId.value = val.id;
      // 存储全局业务id
      const globalBizsQueryParams = saveGlobalBizsId(businessId.value);
      // @ts-ignore
      // 如果当前页面为详情页, 则当业务id切换时, 跳转至对应资源的列表页
      const isBusinessDetail = route.name?.includes?.('BusinessDetail');
      const query = { ...route.query, ...globalBizsQueryParams };
      if (isBusinessDetail) {
        router.push({ path: route.path.split('/detail')[0], query });
      } else {
        await router.push({ path: (route.meta.rootRoutePath as string) || route.path, query });
        props.reload();
      }
    };

    // 获取上一次存储的全局业务id（先从url中获取全局业务id, 如果没有, 则从localStorage中获取, 如果还是没有, 则返回undefined）
    const getGlobalBizsId = () => {
      // 获取local storage中的全局业务id版本号，如果版本号不一致，则清空url、local storage中的全局业务id，并更新版本号
      const lastBizsVersion = localStorageActions.get(GLOBAL_BIZS_VERSION_KEY, (value) => value);

      if (GLOBAL_BIZS_VERSION !== lastBizsVersion) {
        localStorageActions.remove(GLOBAL_BIZS_KEY);
        localStorageActions.set(GLOBAL_BIZS_VERSION_KEY, GLOBAL_BIZS_VERSION);
        return;
      }

      const lastUrlBizs = +route.query[GLOBAL_BIZS_KEY];
      const lastLocalBizs = +localStorageActions.get(GLOBAL_BIZS_KEY, (value) => value);

      return lastUrlBizs || lastLocalBizs;
    };

    const fetchBusinessList = async () => {
      const res = await accountStore.getBizListWithAuth();
      // 更新业务列表
      businessList.value = res.data;
      // 获取上一次存储的全局业务id
      const lastBizs = getGlobalBizsId();
      businessId.value = Number(lastBizs) || res.data[0]?.id || 0;
      // 存储新的全局业务id
      const globalBizsQueryParams = saveGlobalBizsId(businessId.value);
      // 更新query参数
      router.push({ query: { ...route.query, ...globalBizsQueryParams } });
    };

    onMounted(() => {
      fetchBusinessList();
    });

    watch(
      () => businessId.value,
      async (val) => {
        if (!val) return;
        favoriteList.value = await getFavoriteList(businessId.value);
        for (const id of favoriteList.value) favoriteSet.value.add(id);
      },
    );

    watch(
      () => favoriteSet.value,
      () => {
        businessList.value.sort((biz1, biz2) => {
          return +favoriteSet.value.has(biz2.id) - +favoriteSet.value.has(biz1.id);
        });
      },
      {
        deep: true,
      },
    );

    return () => (
      <>
        <AppSelect
          data={businessList.value}
          theme={'dark'}
          class={'bk-hcm-app-selector'}
          value={{
            id: businessId.value,
          }}
          onChange={handleChange}
          minWidth={360}>
          {{
            default: ({ data }: { data: { id: number; name: string } }) => (
              <div class='bk-hcm-app-selector-item'>
                <div class='bk-hcm-app-selector-item-content'>
                  <span class={'bk-hcm-app-selector-item-content-name'}>{`${data.name}`}</span>
                  &nbsp;&nbsp;&nbsp;
                  <span class={'bk-hcm-app-selector-item-content-id'}>{`(${data.id})`}</span>
                </div>

                <div class='bk-hcm-app-selector-item-star'>
                  {favoriteSet.value.has(data.id) ? (
                    <i
                      class={'hcm-icon bkhcm-icon-collect'}
                      style={{ color: '#CC933A', fontSize: '15px' }}
                      onClick={(event) => {
                        removeFromFavorite(data.id);
                        event.stopPropagation();
                      }}
                    />
                  ) : (
                    <i
                      class={'hcm-icon bkhcm-icon-not-favorited'}
                      onClick={(event) => {
                        addToFavorite(data.id);
                        event.stopPropagation();
                      }}
                    />
                  )}
                </div>
              </div>
            ),
            append: () => (
              <div
                class={'app-action-content'}
                onClick={() => {
                  isDialogShow.value = true;
                }}>
                <i class={'hcm-icon bkhcm-icon-plus-circle app-action-content-icon'} />
                <span class={'app-action-content-text'}>新建业务</span>
              </div>
            ),
          }}
        </AppSelect>
        <Dialog
          isShow={isDialogShow.value}
          dialogType='show'
          onConfirm={() => (isDialogShow.value = false)}
          onClosed={() => (isDialogShow.value = false)}>
          <Exception
            type='building'
            class={'hcm-create-business-dialog-exception-building-picture'}
            title={'新建业务参考以下指引'}>
            <div class={'hcm-create-business-dialog-exception-building-tips'}>
              {/* <p class={'hcm-create-business-dialog-exception-building-tips-text1'}>
               可以按照以下方式进行查看
             </p> */}
              <p class={'hcm-create-business-dialog-exception-building-tips-text2'}>
                业务是蓝鲸配置平台的管理空间，可以满足不同团队，不同项目的资源隔离管理需求。
                <Button
                  theme='primary'
                  text
                  onClick={() => {
                    const { BK_CMDB_CREATE_BIZ_URL } = window.PROJECT_CONFIG;
                    window.open(BK_CMDB_CREATE_BIZ_URL, '_blank');
                  }}>
                  新建业务
                </Button>
              </p>
            </div>
          </Exception>
        </Dialog>
      </>
    );
  },
});

import { defineComponent, ref, computed, watch, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { InfoBox, Message } from 'bkui-vue';
import { useSchemeStore } from '@/store';
import { QueryFilterType, QueryRuleOPEnum } from '@/typings/common';
import { IIdcInfo, ISchemeListItem, ISchemeEditingData, ISchemeSelectorItem } from '@/typings/scheme';
import DetailHeader from './components/detail-header';
import SchemeInfoCard from './components/scheme-info-card';
import IdcMapDisplay from './components/idc-map-display';
import NetworkHeatMap from './components/network-heat-map';

import './index.scss';
import { useVerify } from '@/hooks';
import PermissionDialog from '@/components/permission-dialog';

export default defineComponent({
  name: 'SchemeDetailPage',
  setup() {
    const schemeStore = useSchemeStore();
    const router = useRouter();
    const route = useRoute();
    const schemeId = ref(route.query?.sid || '');

    const schemeDetail = ref<ISchemeListItem>();
    const detailLoading = ref(true);
    const schemeList = ref<ISchemeListItem[]>([]);
    const schemeListLoading = ref(false);
    const idcList = ref<IIdcInfo[]>([]);
    const idcListLoading = ref(false);

    const {
      authVerifyData,
      handleAuth,
      handlePermissionConfirm,
      handlePermissionDialog,
      showPermissionDialog,
      permissionParams,
    } = useVerify();

    watch(
      () => route.query?.sid,
      async (val) => {
        if (val) {
          schemeId.value = val;
          await getSchemeDetail();
          getIdcList();
        }
      },
    );

    // 获取方案详情
    const getSchemeDetail = async () => {
      detailLoading.value = true;
      const res = await schemeStore.getCloudSelectionScheme(schemeId.value as string);
      schemeDetail.value = res.data;
      detailLoading.value = false;
    };

    // 获取全部方案列表
    const getSchemeList = async () => {
      schemeListLoading.value = true;
      const filterQuery: QueryFilterType = {
        op: 'and',
        rules: [],
      };
      const res = await schemeStore.listCloudSelectionScheme(filterQuery, {
        start: 0,
        limit: 500,
      });
      schemeList.value = res.data.details;
      schemeListLoading.value = false;
    };

    // 查询idc机房列表
    const getIdcList = async () => {
      idcListLoading.value = true;
      const filterQuery: QueryFilterType = {
        op: 'and',
        rules: [
          {
            field: 'id',
            op: QueryRuleOPEnum.IN,
            value: schemeDetail.value.result_idc_ids,
          },
        ],
      };
      const res = await schemeStore.listIdc(filterQuery, {
        start: 0,
        limit: 500,
      });
      idcList.value = res.data;
      idcListLoading.value = false;
    };

    const headerData = computed((): ISchemeSelectorItem => {
      if (schemeDetail.value) {
        const { id, name, bk_biz_id, deployment_architecture, vendors, composite_score, net_score, cost_score } =
          schemeDetail.value;
        return {
          id,
          name,
          bk_biz_id,
          deployment_architecture,
          vendors,
          composite_score,
          net_score,
          cost_score,
        };
      }
      return undefined;
    });

    const handleUpdate = (data: ISchemeEditingData) => {
      schemeDetail.value = Object.assign({}, schemeDetail.value, data);
      const crtScheme = schemeList.value.find((item) => item.id === schemeDetail.value.id);
      if (crtScheme) {
        crtScheme.name = data.name;
        crtScheme.bk_biz_id = data.bk_biz_id;
      }
    };

    const handleDel = () => {
      InfoBox({
        title: '请确认是否删除',
        subTitle: `将删除【${schemeDetail.value.name}】`,
        headerAlign: 'center',
        footerAlign: 'center',
        contentAlign: 'center',
        onConfirm() {
          schemeStore.deleteCloudSelectionScheme([schemeDetail.value.id]).then(() => {
            Message({
              theme: 'success',
              message: '删除成功',
            });
            router.push({ name: 'scheme-list' });
          });
        },
      });
    };

    onMounted(async () => {
      getSchemeList();
      if (schemeId.value) {
        await getSchemeDetail();
        getIdcList();
      }
    });

    return () => (
      <div class='scheme-detail-page'>
        <bk-loading loading={detailLoading.value}>
          {detailLoading.value ? null : (
            <>
              <DetailHeader
                schemeList={schemeList.value}
                schemeListLoading={schemeListLoading.value}
                schemeData={headerData.value}
                showEditIcon={true}
                onUpdate={handleUpdate}>
                {{
                  operate: () => (
                    <bk-button
                      onClick={() => {
                        if (authVerifyData.value.permissionAction.cloud_selection_delete) handleDel();
                        else handleAuth('cloud_selection_delete');
                      }}
                      class={`del-btn ${
                        !authVerifyData.value.permissionAction.cloud_selection_delete ? 'hcm-no-permision-btn' : ''
                      }`}>
                      删除
                    </bk-button>
                  ),
                }}
              </DetailHeader>
              <section class='detail-content-area'>
                <SchemeInfoCard schemeDetail={schemeDetail.value} />
                <section class='chart-content-wrapper'>
                  <IdcMapDisplay list={idcList.value} />
                  <NetworkHeatMap idcList={idcList.value} areaTopo={schemeDetail.value.user_distribution} />
                </section>
              </section>
            </>
          )}
        </bk-loading>
        <PermissionDialog
          isShow={showPermissionDialog.value}
          onConfirm={handlePermissionConfirm}
          onCancel={handlePermissionDialog}
          params={permissionParams.value}
        />
      </div>
    );
  },
});

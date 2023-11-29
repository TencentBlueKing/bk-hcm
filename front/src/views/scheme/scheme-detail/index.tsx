import { defineComponent, ref, reactive, computed, watch, onMounted } from "vue";
import { useRoute, useRouter } from 'vue-router';
import { InfoBox, Message } from "bkui-vue";
import { useSchemeStore } from "@/store";
import { QueryFilterType } from '@/typings/common';
import { ISchemeListItem, ISchemeEditingData, ISchemeSelectorItem } from "@/typings/scheme";
import DetailHeader from "./components/detail-header";
import SchemeInfoCard from "./components/scheme-info-card";
import IdcMapDisplay from "./components/idc-map-display";
import NetworkHeatMap from "./components/network-heat-map";

import './index.scss';

export default defineComponent({
  name: 'scheme-detail-page',
  setup () {
    const schemeStore = useSchemeStore();
    const router = useRouter();
    const route = useRoute();
    const schemeId = ref(route.query?.sid || '');

    let schemeDetail = ref<ISchemeListItem>();
    const detailLoading = ref(true);
    let schemeList = reactive<ISchemeListItem[]>([]);
    const schemeListLoading = ref(false);

    watch(() => route.query?.sid, val => {
      schemeId.value = val;
      getSchemeDetail();
    });

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
      const res = await schemeStore.listCloudSelectionScheme(filterQuery, { start: 0, limit: 500 });
      schemeList = res.data.details;
      schemeListLoading.value = false;
    }

    const headerData = computed((): ISchemeSelectorItem => {
      if (schemeDetail.value) {
        const { id, name, bk_biz_id, deployment_architecture, vendors, composite_score, net_score, cost_score } = schemeDetail.value;
        return { id, name, bk_biz_id, deployment_architecture, vendors, composite_score, net_score, cost_score }
      }
    })

    const handleUpdate = (data: ISchemeEditingData) => {
      schemeDetail = Object.assign({}, schemeDetail, data);
    };

    const handleDel = () => {
      InfoBox({
        title: '请确认是否删除',
        subTitle: `将删除【${schemeDetail.value.name}】`,
        headerAlign: 'center',
        footerAlign: 'center',
        contentAlign: 'center',
        onConfirm() {
          schemeStore.deleteCloudSelectionScheme([schemeDetail.value.id])
            .then(() => {
              Message({
                theme: 'success',
                message: '删除成功',
              });
              router.push({ name: 'scheme-list' });
            });
        },
      });
    }

    onMounted(async() => {
      getSchemeList();
      if (schemeId) {
        await getSchemeDetail();
      }
    });

    return () => (
      <div class="scheme-detail-page">
        <bk-loading loading={detailLoading.value}>
          {
            detailLoading.value ? null : (
              <>
                <DetailHeader
                  schemeList={schemeList}
                  schemeData={headerData.value}
                  showEditIcon={true}
                  onUpdate={handleUpdate}>
                    {{
                      operate: () => (
                        <bk-button onClick={handleDel}>删除</bk-button>
                      )
                    }}
                </DetailHeader>
                <section class="detail-content-area">
                  <SchemeInfoCard schemeDetail={schemeDetail.value} />
                  <section  class="chart-content-wrapper">
                    <IdcMapDisplay ids={schemeDetail.value.result_idc_ids} />
                    <NetworkHeatMap ids={schemeDetail.value.result_idc_ids} areaTopo={schemeDetail.value.user_distribution} />
                  </section>
                </section>
              </>
            )
          }
        </bk-loading>
      </div>
    );
  },
});

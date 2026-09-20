import { computed, type Reactive, ref, watch } from 'vue';
import { Message } from 'bkui-vue';
import { reqExclusiveClusterIdleVips, reqExclusiveClusterTags } from '@/api/load_balancers/apply-clb';
import type {
  ApplyClbModel,
  ExclusiveClusterIsp,
  ExclusiveClusterItem,
  ExclusiveClusterTag,
} from '@/api/load_balancers/apply-clb/types';

export const RANDOM_ALLOCATION = '__random__';

/**
 * 独占集群标签接口（tags/list）的 isp 入参只支持三网直连：CMCC / CUCC / CTCC。
 * BGP（含自研云按 TypeSet 拆分出的 BGP 系取值）等其它运营商均不支持选择独占型。
 */
export const EXCLUSIVE_CLUSTER_ISP_TYPES: ExclusiveClusterIsp[] = ['CMCC', 'CUCC', 'CTCC'];

export const isExclusiveClusterIsp = (isp?: string): isp is ExclusiveClusterIsp =>
  EXCLUSIVE_CLUSTER_ISP_TYPES.includes(isp as ExclusiveClusterIsp);

const uniqueEgresses = (clusters: ExclusiveClusterItem[] = []) => [
  ...new Set(clusters.map(({ egress }) => egress).filter(Boolean)),
];

export default (formModel: Reactive<ApplyClbModel>, isBusinessPage: boolean, isRestoring: () => boolean) => {
  const exclusiveClusterTags = ref<ExclusiveClusterTag[]>([]);
  const idleVips = ref<string[]>([]);
  const isTagsLoading = ref(false);
  const isIdleVipsLoading = ref(false);
  const isTagsLoadFailed = ref(false);
  const isIdleVipsLoadFailed = ref(false);
  let tagsRequestId = 0;
  let idleVipsRequestId = 0;

  const l4TagList = computed(() => exclusiveClusterTags.value.filter(({ cluster_type }) => cluster_type === 'TGW'));
  const l7TagList = computed(() => exclusiveClusterTags.value.filter(({ cluster_type }) => cluster_type === 'STGW'));
  const selectedL4Tag = computed(() =>
    l4TagList.value.find(({ cluster_tag }) => cluster_tag === formModel.l4_cluster_tag),
  );
  const selectedL4Cluster = computed(() =>
    selectedL4Tag.value?.clusters.find(({ cloud_cluster_id }) => cloud_cluster_id === formModel.l4_cluster_id),
  );
  const isExclusiveAvailable = computed(
    () =>
      isBusinessPage &&
      formModel.load_balancer_type === 'OPEN' &&
      Boolean(formModel.account_id && formModel.region && formModel.zones) &&
      isExclusiveClusterIsp(formModel.vip_isp) &&
      !isTagsLoadFailed.value &&
      exclusiveClusterTags.value.length > 0,
  );

  const l4Egresses = computed(() => {
    if (!formModel.enable_l4 || !selectedL4Tag.value) return [];
    if (formModel.l4_cluster_id && formModel.l4_cluster_id !== RANDOM_ALLOCATION) {
      return selectedL4Cluster.value?.egress ? [selectedL4Cluster.value.egress] : [];
    }
    return uniqueEgresses(selectedL4Tag.value.clusters);
  });
  const l7Egresses = computed(() => {
    if (!formModel.enable_l7 || !formModel.cluster_tag) return [];
    const matchingL4Tag = l4TagList.value.find(({ cluster_tag }) => cluster_tag === formModel.cluster_tag);
    return uniqueEgresses(matchingL4Tag?.clusters);
  });
  const effectiveEgresses = computed(() => {
    if (formModel.enable_l4 && formModel.enable_l7) {
      const l7Set = new Set(l7Egresses.value);
      return l4Egresses.value.filter((egress) => l7Set.has(egress));
    }
    if (formModel.enable_l4) return l4Egresses.value;
    if (formModel.enable_l7) return l7Egresses.value;
    return [];
  });
  const egressFilterKey = computed(() =>
    [
      formModel.enable_l4,
      formModel.enable_l7,
      formModel.l4_cluster_tag,
      formModel.l4_cluster_id,
      formModel.cluster_tag,
      effectiveEgresses.value.join('|'),
    ].join(':'),
  );

  const resetExclusiveSelections = () => {
    formModel.l4_cluster_tag = '';
    formModel.l4_cluster_id = '';
    formModel.l4_vip = '';
    formModel.cluster_tag = '';
    idleVips.value = [];
    isIdleVipsLoadFailed.value = false;
  };

  const loadExclusiveClusterTags = async () => {
    tagsRequestId += 1;
    const requestId = tagsRequestId;
    const { bk_biz_id, account_id, region, vip_isp, zones, load_balancer_type } = formModel;
    exclusiveClusterTags.value = [];
    formModel.exclusive_cluster_tags = [];
    isTagsLoadFailed.value = false;
    if (
      !isBusinessPage ||
      load_balancer_type !== 'OPEN' ||
      !bk_biz_id ||
      !account_id ||
      !region ||
      !zones ||
      !isExclusiveClusterIsp(vip_isp)
    ) {
      isTagsLoading.value = false;
      return;
    }

    isTagsLoading.value = true;
    const snapshot = [bk_biz_id, account_id, region, vip_isp, zones].join('|');
    try {
      const { data } = await reqExclusiveClusterTags({
        bk_biz_id,
        account_id,
        region,
        isp: vip_isp,
        zones: Array.isArray(zones) ? zones : [zones],
        cluster_type: '',
      });
      const currentSnapshot = [
        formModel.bk_biz_id,
        formModel.account_id,
        formModel.region,
        formModel.vip_isp,
        formModel.zones,
      ].join('|');
      if (requestId !== tagsRequestId || snapshot !== currentSnapshot) return;
      exclusiveClusterTags.value = data?.details ?? [];
      formModel.exclusive_cluster_tags = exclusiveClusterTags.value;
    } catch (error) {
      if (requestId === tagsRequestId) {
        isTagsLoadFailed.value = true;
      }
      throw error;
    } finally {
      if (requestId === tagsRequestId) isTagsLoading.value = false;
    }
  };

  const loadIdleVips = async (cloudClusterId: string, resetSelection = true) => {
    idleVipsRequestId += 1;
    const requestId = idleVipsRequestId;
    idleVips.value = [];
    if (resetSelection) formModel.l4_vip = RANDOM_ALLOCATION;
    if (!cloudClusterId || cloudClusterId === RANDOM_ALLOCATION) {
      isIdleVipsLoading.value = false;
      return;
    }

    isIdleVipsLoading.value = true;
    try {
      const { data } = await reqExclusiveClusterIdleVips({
        bk_biz_id: formModel.bk_biz_id,
        account_id: formModel.account_id,
        region: formModel.region,
        cloud_cluster_id: cloudClusterId,
      });
      if (requestId !== idleVipsRequestId || formModel.l4_cluster_id !== cloudClusterId) return;
      idleVips.value = data?.details ?? [];
      if (!data?.count || idleVips.value.length === 0) {
        formModel.l4_cluster_id = '';
        formModel.l4_vip = '';
        Message({ theme: 'warning', message: '该集群暂无空闲 IP，请选择其它集群' });
      }
    } catch (error) {
      if (requestId === idleVipsRequestId) isIdleVipsLoadFailed.value = true;
      throw error;
    } finally {
      if (requestId === idleVipsRequestId) isIdleVipsLoading.value = false;
    }
  };

  const handleL4TagChange = () => {
    formModel.l4_cluster_id = RANDOM_ALLOCATION;
    formModel.l4_vip = RANDOM_ALLOCATION;
    idleVips.value = [];
  };

  watch(
    [
      () => formModel.bk_biz_id,
      () => formModel.account_id,
      () => formModel.region,
      () => formModel.zones,
      () => formModel.vip_isp,
      () => formModel.load_balancer_type,
    ],
    () => {
      if (isRestoring()) {
        exclusiveClusterTags.value = formModel.exclusive_cluster_tags ?? [];
        if (formModel.enable_l4 && formModel.l4_cluster_id && formModel.l4_cluster_id !== RANDOM_ALLOCATION) {
          loadIdleVips(formModel.l4_cluster_id, false).catch(() => undefined);
        }
        return;
      }
      resetExclusiveSelections();
      loadExclusiveClusterTags().catch(() => undefined);
    },
  );

  watch(
    () => formModel.enable_l4,
    (enabled) => {
      if (!enabled) {
        idleVipsRequestId += 1;
        isIdleVipsLoading.value = false;
        return;
      }
      if (formModel.l4_cluster_id && formModel.l4_cluster_id !== RANDOM_ALLOCATION) {
        loadIdleVips(formModel.l4_cluster_id, false).catch(() => undefined);
      }
    },
  );

  watch(egressFilterKey, () => {
    if (isRestoring()) return;
    formModel.bandwidth_package_id = undefined;
  });

  return {
    l4TagList,
    l7TagList,
    selectedL4Tag,
    idleVips,
    isTagsLoading,
    isIdleVipsLoading,
    isTagsLoadFailed,
    isIdleVipsLoadFailed,
    isExclusiveAvailable,
    effectiveEgresses,
    egressFilterKey,
    handleL4TagChange,
    loadIdleVips,
  };
};

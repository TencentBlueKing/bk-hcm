import { useI18n } from 'vue-i18n';

export const useStatus = (delivery_detail = {}) => {
  const { t } = useI18n();
  return {
    pending: {
      label: t('审批中'),
      value: 'pending',
      desc: t('您的资源申请单已提交，在单据审批通过后，会创建相关资源'),
      tag: 'pending',
    },
    rejected: {
      label: t('审批驳回'),
      value: 'rejected',
      desc: t('您的资源申请单未通过审批，如有疑问，请联系审批人，或请重新提单'),
    },
    pass: {
      label: t('审批通过'),
      value: 'pass',
      desc: t('您的资源申请单已通过审批'),
      tag: 'success',
    },
    cancelled: {
      label: t('已撤销'),
      value: 'cancelled',
      desc: t('您的资源申请单已撤销'),
      tag: 'abort',
    },
    delivering: {
      label: t('交付中'),
      value: 'delivering',
      desc: t('您的资源申请单已通过审批，正在等待资源从云上生产'),
      tag: 'pending',
    },
    completed: {
      label: t('已完成'),
      value: 'completed',
      desc: t('您的资源申请单已通过审批，并成功交付资源'),
      tag: 'success',
    },
    deliver_error: {
      label: t('交付异常'),
      value: 'deliver_error',
      desc: delivery_detail.error,
      tag: 'abort',
    },
  };
};

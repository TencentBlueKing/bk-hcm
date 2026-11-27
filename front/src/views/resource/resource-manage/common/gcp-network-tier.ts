const tierLabels = new Map([
  ['PREMIUM', '高级'],
  ['STANDARD', '标准'],
]);

export const formatGcpNetworkTier = (tier?: string | null) => (tier ? tierLabels.get(tier) ?? tier : '--');

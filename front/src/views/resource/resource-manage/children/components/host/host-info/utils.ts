/**
 * Format an IP address array for display in host detail.
 *
 * The detail API omits the field entirely when the resource has no address of that
 * family, so the received value can be `undefined`/`null` even though the field is
 * declared as `string[]`. Rendering such a value without guarding throws during
 * mount and breaks the whole navigation, so every caller must go through here.
 *
 * @param val - IP addresses returned by the detail API, possibly missing.
 * @returns comma separated addresses, or `--` when there is nothing to show.
 */
export const formatIpList = (val?: string[] | null): string => {
  if (!Array.isArray(val) || val.length === 0) return '--';
  return val.join(',');
};

// 这个文件会在node环境中使用，需要判断window
let locationOrigin =  typeof window === 'undefined' ? '' : window.location.origin;

if (!locationOrigin || locationOrigin.indexOf('localhost') > -1) {
  locationOrigin = `${locationOrigin}/mock/api/v4`;
}

const domain = locationOrigin;

const api = {
  organization_user_info: `${domain}/organization/user_info/`,
  add_account: `${domain}/add/`,
  get_account: `${domain}/get/`,
  account_sync: `${domain}/sync/`,
  list_public_image: `${domain}/cloud/public_images/list/`,
  detail_public_image: `${domain}/cloud/public_images/detail/`,
  // someOtherApi: `${domain}/otherApi`
};

module.exports = api;

// 这个文件会在node环境中使用，需要判断window
let locationOrigin =  typeof window === 'undefined' ? '' : window.location.origin;

if (!locationOrigin || locationOrigin.indexOf('localhost') > -1) {
  locationOrigin = `${locationOrigin}/api/v4`;
}

const domain = locationOrigin;

const api = {
  organization_user_info: `${domain}/organization/user_info/`,
  // someOtherApi: `${domain}/otherApi`
};

module.exports = api;

module.exports = function (req, res) {
  global.userInfo = {
    // 平台管理员
    username: 'admin',
    username_cn: '平台管理员',
    email: '',
    phone: '',
    role_name: 'SUPER_ROLE',
    role_name_cn: '平台管理员',
    organization_id: -1,
    organization_name: '---',
    organization_type: 'admin',
    organizations: {},
  };
  res.json({
    result: true,
    message: 'success',
    code: 'OK',
    data: global.userInfo,
  });
};

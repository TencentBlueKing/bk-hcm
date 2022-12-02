module.exports = function (req, res) {
  global.detail = {
    // 云管理测试数据
    type: 'resource',   // 账号类型
    name: 'test', // 名称
    cloudName: '', // 云厂商
    account: '',    // 主账号
    subAccountId: '',    // 子账号id
    subAccountName: '',    // 子账号名称
    scretId: '',    // 密钥id
    secretKey: '',  // 密钥key
    user: ['poloohuang'], // 责任人
    organize: [],   // 组织架构
    business: 'huawei',   // 使用业务
    remark: '测试测试测试',     // 备注
  };
  res.json({
    result: true,
    message: 'success',
    code: 'OK',
    data: global.detail,
  });
};


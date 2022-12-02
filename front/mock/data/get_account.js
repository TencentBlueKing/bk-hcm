module.exports = function (req, res) {
  global.list = [
    {
      id: 1,
      name: 'qcloud-for-lol',
      source: 'QQ',
      status: '创建中',
      create_time: '2018-05-25 15:02:241',
      selected: false,
    },
    {
      id: 2,
      name: 'qcloud-for-lol',
      source: 'IEG',
      status: '创建中',
      create_time: '2018-05-25 15:02:241',
      selected: false,
    },
  ];
  res.json({
    result: true,
    message: 'success',
    code: 'OK',
    data: global.list,
  });
};


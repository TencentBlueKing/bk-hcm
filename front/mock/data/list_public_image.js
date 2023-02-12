module.exports = function (req, res) {
  const data = {
    count: 2,
    details: Array(13).fill({})
      .map((_, index) => ({
        id: `0000000${index.toString(16)}`,
        vendor: 'tcloud',
        name: 'CentOS 7.5 64位',
        cloud_id: 'img-oikl1tzv',
        architecture: 'x86_64',
        state: 'NORMAL',
        type: 'PUBLIC_IMAGE',
        platform: 'CentOS',
        creator: 'xxx',
        reviser: 'xxx',
        created_at: '2023-01-16T03:30:41Z',
        updated_at: '2023-01-16T08:39:28Z',
      })),
  };
  res.json({
    code: 0,
    data,
    message: '',
  });
};

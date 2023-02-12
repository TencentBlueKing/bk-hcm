module.exports = function (req, res) {
  const data = {
    id: '00000002',
    vendor: 'tcloud',
    name: 'CentOS 7.6 64位',
    cloud_id: 'img-9qabwvbn',
    platform: 'CentOS',
    architecture: 'x86_64',
    type: 'PUBLIC_IMAGE',
    created_at: '2023-01-16T03:30:41Z',
    updated_at: '2023-01-16T08:39:28Z',
    extension: {
      region: 'ap-beijing',
      image_source: 'OFFICIAL',
      image_size: 20,
    },
  };
  res.json({
    code: 0,
    data,
    message: '',
  });
};

export default () => {
  const handleChange = (newStatus) => {
    console.log('I am updated', newStatus);
  };

  return {
    handleChange,
  };
};
